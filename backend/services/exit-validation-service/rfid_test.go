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

func TestRFID_Flow_Success(t *testing.T) {
	db, redisClient, mqttClient, handler := setupTestEnvironment(t)
	defer db.Close()
	defer redisClient.Close()

	ctx := context.Background()
	secret := "secret-32bytes-secret-32bytes-12"

	// Enable RFID for store-rfid
	_ = redisClient.Set(ctx, "store_rfid_enabled:store-rfid", "true", 5*time.Minute)

	exitToken, _ := jwt.GenerateExitToken("ord-rfid-1", "user-1", "store-rfid", "", secret, 10*time.Minute)
	deviceToken, _ := jwt.GenerateDeviceToken("dev-rfid", "gate-R1", "store-rfid", secret, 24*time.Hour)

	// Step 1: Validate Exit
	body1, _ := json.Marshal(ValidateExitRequest{ExitToken: exitToken})
	req1 := httptest.NewRequest(http.MethodPost, "/v1/exit/validate", bytes.NewBuffer(body1))
	req1.Header.Set("Authorization", "Bearer "+deviceToken)

	rec1 := httptest.NewRecorder()
	handler.ValidateExitHandler(rec1, req1)

	var resp1 ValidateExitResponse
	_ = json.Unmarshal(rec1.Body.Bytes(), &resp1)

	if resp1.Result != ResultAwaitingRFID || resp1.OrderID != "ord-rfid-1" {
		t.Fatalf("Expected AWAITING_RFID response, got %+v", resp1)
	}

	// Verify gate hasn't opened yet
	if len(mqttClient.GetPublished()) != 0 {
		t.Fatalf("Expected 0 MQTT commands before RFID confirm")
	}

	// Step 2: RFID Confirm with deactivation_success = true
	body2, _ := json.Marshal(RFIDConfirmRequest{
		OrderID:             "ord-rfid-1",
		TagIDs:              []string{"TAG-001", "TAG-002"},
		DeactivationSuccess: true,
	})
	req2 := httptest.NewRequest(http.MethodPost, "/v1/exit/rfid-confirm", bytes.NewBuffer(body2))
	req2.Header.Set("Authorization", "Bearer "+deviceToken)

	rec2 := httptest.NewRecorder()
	handler.RFIDConfirmHandler(rec2, req2)

	var resp2 ValidateExitResponse
	_ = json.Unmarshal(rec2.Body.Bytes(), &resp2)

	if resp2.Result != "OPEN" {
		t.Fatalf("Expected OPEN response on RFID confirm, got %+v", resp2)
	}

	// Verify MQTT published OPEN command
	published := mqttClient.GetPublished()
	if len(published) != 1 || published[0].Cmd.Cmd != "OPEN" {
		t.Fatalf("Expected 1 MQTT OPEN command after RFID confirm, got %d", len(published))
	}
}

func TestRFID_Flow_FailureDeactivation(t *testing.T) {
	db, redisClient, mqttClient, handler := setupTestEnvironment(t)
	defer db.Close()
	defer redisClient.Close()

	ctx := context.Background()
	secret := "secret-32bytes-secret-32bytes-12"

	_ = redisClient.Set(ctx, "store_rfid_enabled:store-rfid", "true", 5*time.Minute)

	exitToken, _ := jwt.GenerateExitToken("ord-rfid-fail", "user-1", "store-rfid", "", secret, 10*time.Minute)
	deviceToken, _ := jwt.GenerateDeviceToken("dev-rfid", "gate-R1", "store-rfid", secret, 24*time.Hour)

	// Step 1: Validate Exit
	body1, _ := json.Marshal(ValidateExitRequest{ExitToken: exitToken})
	req1 := httptest.NewRequest(http.MethodPost, "/v1/exit/validate", bytes.NewBuffer(body1))
	req1.Header.Set("Authorization", "Bearer "+deviceToken)

	rec1 := httptest.NewRecorder()
	handler.ValidateExitHandler(rec1, req1)

	// Step 2: RFID Confirm with deactivation_success = false
	body2, _ := json.Marshal(RFIDConfirmRequest{
		OrderID:             "ord-rfid-fail",
		TagIDs:              []string{"TAG-BAD"},
		DeactivationSuccess: false,
	})
	req2 := httptest.NewRequest(http.MethodPost, "/v1/exit/rfid-confirm", bytes.NewBuffer(body2))
	req2.Header.Set("Authorization", "Bearer "+deviceToken)

	rec2 := httptest.NewRecorder()
	handler.RFIDConfirmHandler(rec2, req2)

	var resp2 ValidateExitResponse
	_ = json.Unmarshal(rec2.Body.Bytes(), &resp2)

	if resp2.Result != "DENY" || resp2.Reason != ResultRFIDTimeout {
		t.Fatalf("Expected DENY with RFID_TIMEOUT reason, got %+v", resp2)
	}

	// Verify NO gate OPEN command was sent
	if len(mqttClient.GetPublished()) != 0 {
		t.Fatalf("Expected 0 MQTT OPEN commands on RFID failure")
	}
}
