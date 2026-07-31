package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/jwt"
)

type MockOrderVerifier struct {
	validOrders map[string]bool
}

func (m *MockOrderVerifier) VerifyOrderCompleted(ctx context.Context, orderID, storeID string) (bool, error) {
	return m.validOrders[orderID], nil
}

func TestStaffOverride_ValidCompletedOrder_OpensGate(t *testing.T) {
	db, redisClient, mqttClient, handler := setupTestEnvironment(t)
	defer db.Close()
	defer redisClient.Close()

	secret := "secret-32bytes-secret-32bytes-12"

	// Generate Staff Token (SECURITY role)
	staffToken, err := jwt.GenerateSessionToken("staff-usr-1", "store-1", "sess-1", "STAFF", secret, 4*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate staff token: %v", err)
	}

	verifier := &MockOrderVerifier{validOrders: map[string]bool{"ord-completed-100": true}}

	body, _ := json.Marshal(StaffOverrideRequest{
		OrderID: "ord-completed-100",
		GateID:  "gate-A",
		Reason:  "SCANNER_MALFUNCTION",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/exit/staff-override", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+staffToken)

	rec := httptest.NewRecorder()
	handler.StaffOverrideHandler(rec, req, verifier)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	// Verify MQTT published OPEN command
	published := mqttClient.GetPublished()
	if len(published) != 1 || published[0].Cmd.Cmd != "OPEN" {
		t.Fatalf("Expected 1 MQTT OPEN command on staff override, got %d", len(published))
	}

	// Verify exit_attempts row created
	attempt, err := handler.repo.GetLatestExitAttemptByOrderID(context.Background(), "ord-completed-100")
	if err != nil || attempt == nil || attempt.Result != ResultStaffOverride {
		t.Fatalf("Expected STAFF_OVERRIDE exit attempt row, got %+v, err=%v", attempt, err)
	}
}

func TestStaffOverride_UnpaidOrder_Rejection(t *testing.T) {
	db, redisClient, mqttClient, handler := setupTestEnvironment(t)
	defer db.Close()
	defer redisClient.Close()

	secret := "secret-32bytes-secret-32bytes-12"
	staffToken, _ := jwt.GenerateSessionToken("staff-usr-1", "store-1", "sess-1", "STAFF", secret, 4*time.Hour)

	verifier := &MockOrderVerifier{validOrders: map[string]bool{"ord-unpaid-999": false}}

	body, _ := json.Marshal(StaffOverrideRequest{
		OrderID: "ord-unpaid-999",
		GateID:  "gate-A",
		Reason:  "CUSTOMER_ASSISTANCE",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/exit/staff-override", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+staffToken)

	rec := httptest.NewRecorder()
	handler.StaffOverrideHandler(rec, req, verifier)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404 for unpaid order override attempt, got %d", rec.Code)
	}

	// Verify NO MQTT command published
	if len(mqttClient.GetPublished()) != 0 {
		t.Fatalf("Expected 0 MQTT OPEN commands on invalid order override")
	}
}
