package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/crypto"
)

func TestInboundWebhook_InvalidSignature_Returns400(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	masterKey := "12345678901234567890123456789012"
	downstream := &MockDownstreamClient{}
	handler := NewWebhookHandler(repo, masterKey, downstream)

	secret := "whsec_test123"
	encSecret, _ := crypto.Encrypt([]byte(secret), masterKey)

	conn := &ERPConnection{
		ID:                            "conn-wh-1",
		ChainID:                       "chain-1",
		ERPType:                       ERPTypeSAP,
		IntegrationMode:               IntegrationModeDirect,
		InboundWebhookSecretEncrypted: encSecret,
		Status:                        ConnectionStatusPendingSetup,
	}
	_ = repo.CreateConnection(context.Background(), conn)

	payload := map[string]interface{}{
		"event_type": "PRICE_UPDATE",
		"payload": map[string]interface{}{
			"barcode":     "8901030300011",
			"store_id":    "store-1",
			"price_paise": 1500,
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/v1/integration/connections/conn-wh-1/webhook", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "invalid_signature_hex")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/v1/integration/connections/{id}/webhook", handler.HandleInboundWebhook)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request on invalid signature, got %d", rr.Code)
	}

	// Verify zero rows in erp_webhook_events
	events, _ := repo.ListWebhookEvents(context.Background(), "conn-wh-1", nil)
	if len(events) != 0 {
		t.Fatalf("Expected 0 webhook events recorded on invalid signature, got %d", len(events))
	}
}

func TestInboundWebhook_DuplicateEvent_Returns200_DoesNotCallDownstreamTwice(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	masterKey := "12345678901234567890123456789012"
	downstream := &MockDownstreamClient{}
	handler := NewWebhookHandler(repo, masterKey, downstream)

	secret := "whsec_test123"
	encSecret, _ := crypto.Encrypt([]byte(secret), masterKey)

	conn := &ERPConnection{
		ID:                            "conn-wh-dup",
		ChainID:                       "chain-1",
		ERPType:                       ERPTypeSAP,
		IntegrationMode:               IntegrationModeDirect,
		InboundWebhookSecretEncrypted: encSecret,
		Status:                        ConnectionStatusActive,
	}
	_ = repo.CreateConnection(context.Background(), conn)

	payload := map[string]interface{}{
		"event_type": "PRICE_UPDATE",
		"payload": map[string]interface{}{
			"barcode":     "8901030300011",
			"store_id":    "store-1",
			"price_paise": 1500,
		},
	}
	bodyBytes, _ := json.Marshal(payload)
	sig := ComputeHMACSignature(bodyBytes, secret)

	r := mux.NewRouter()
	r.HandleFunc("/v1/integration/connections/{id}/webhook", handler.HandleInboundWebhook)

	// First Request
	req1 := httptest.NewRequest("POST", "/v1/integration/connections/conn-wh-dup/webhook", bytes.NewReader(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Signature", sig)
	req1.Header.Set("X-Event-Id", "evt-unique-100")

	rr1 := httptest.NewRecorder()
	r.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Fatalf("First request expected 200 OK, got %d", rr1.Code)
	}
	if downstream.PriceUpdateCalls != 1 {
		t.Fatalf("Expected 1 price update call on first request, got %d", downstream.PriceUpdateCalls)
	}

	// Second Request with SAME X-Event-Id
	req2 := httptest.NewRequest("POST", "/v1/integration/connections/conn-wh-dup/webhook", bytes.NewReader(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Signature", sig)
	req2.Header.Set("X-Event-Id", "evt-unique-100")

	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("Second request expected 200 OK, got %d", rr2.Code)
	}

	// Verify downstream was NOT called a second time
	if downstream.PriceUpdateCalls != 1 {
		t.Fatalf("Expected downstream price update calls to remain 1 after duplicate request, got %d", downstream.PriceUpdateCalls)
	}
}
