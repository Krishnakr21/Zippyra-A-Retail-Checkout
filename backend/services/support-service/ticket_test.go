package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestCreateTicket_OrderNotOwned_Returns403(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewTicketHandler(repo, nil, nil)
	r := mux.NewRouter()
	SetupRoutes(r, handler, nil)

	// Mock order verifier returns false (order belongs to different user)
	handler.SetOrderVerifier(func(ctx context.Context, orderID, customerID string) (bool, error) {
		return false, nil
	})

	orderID := "ord-other-user-999"
	payload := CreateTicketRequest{
		Category:       CategoryOrderIssue,
		RelatedOrderID: &orderID,
		Subject:        "Wrong item delivered",
		Description:    "I received the wrong item in my order.",
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/v1/support/tickets", bytes.NewBuffer(bodyBytes))
	req.Header.Set("X-User-ID", "cust-user-1")
	req.Header.Set("X-User-Role", "CUSTOMER")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 when order belongs to different user, got %d", rr.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "ORDER_NOT_OWNED_BY_REQUESTER" {
		t.Errorf("Expected code ORDER_NOT_OWNED_BY_REQUESTER, got %v", errObj["code"])
	}
}

func TestCreateTicket_ExitGateIssue_AutoUrgentAnd4hSLA(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewTicketHandler(repo, nil, nil)
	r := mux.NewRouter()
	SetupRoutes(r, handler, nil)

	storeID := "store-mumbai-01"
	payload := CreateTicketRequest{
		Category:    CategoryExitGateIssue,
		Subject:     "Stuck at barrier",
		Description: "Gate alarm triggered, turnstile won't unlock.",
		StoreID:     &storeID,
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/v1/support/tickets", bytes.NewBuffer(bodyBytes))
	req.Header.Set("X-User-ID", "cust-stuck-1")
	req.Header.Set("X-User-Role", "CUSTOMER")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d", rr.Code)
	}

	var created SupportTicket
	_ = json.Unmarshal(rr.Body.Bytes(), &created)

	if created.Priority != PriorityUrgent {
		t.Errorf("Expected auto-priority URGENT for EXIT_GATE_ISSUE, got %s", created.Priority)
	}

	expectedSLABound := time.Now().Add(4 * time.Hour)
	if created.SLADueAt.Before(time.Now()) || created.SLADueAt.After(expectedSLABound.Add(1*time.Minute)) {
		t.Errorf("Expected sla_due_at to be ~4h from now, got %v", created.SLADueAt)
	}
}
