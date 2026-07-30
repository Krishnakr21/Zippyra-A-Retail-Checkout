package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestOutbound_RedeliveredEvent_CreatesOnlyOneSyncJob(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	masterKey := "12345678901234567890123456789012"
	pushWorker := NewDirectPushWorker(repo, masterKey)
	consumer := NewEventConsumer(repo, pushWorker)

	conn := &ERPConnection{
		ID:                    "conn-idem-1",
		ChainID:               "chain-1",
		ERPType:               ERPTypeSAP,
		IntegrationMode:       IntegrationModeDirect,
		EnabledOutboundEvents: []string{"order.completed"},
		Status:                ConnectionStatusActive,
	}
	_ = repo.CreateConnection(context.Background(), conn)

	eventBytes, _ := json.Marshal(map[string]interface{}{
		"event_id": "evt-redelivered-100",
		"order_id": "ord-100",
		"chain_id": "chain-1",
	})

	// Process event 3 times (Kafka redelivery simulation)
	_ = consumer.ProcessEvent(context.Background(), "order.completed", eventBytes)
	_ = consumer.ProcessEvent(context.Background(), "order.completed", eventBytes)
	_ = consumer.ProcessEvent(context.Background(), "order.completed", eventBytes)

	jobs, _ := repo.ListSyncJobs(context.Background(), "conn-idem-1", nil, "OUTBOUND")
	if len(jobs) != 1 {
		t.Fatalf("Expected exactly 1 sync job for redelivered event, got %d", len(jobs))
	}
}

func TestConnectionCreation_OneTimeSecrets_NotReturnedInGet(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	masterKey := "12345678901234567890123456789012"
	handler := NewConnectionHandler(repo, nil, masterKey)

	r := mux.NewRouter()
	r.HandleFunc("/v1/integration/connections", handler.CreateConnection).Methods("POST")
	r.HandleFunc("/v1/integration/connections/{id}", handler.GetConnection).Methods("GET")

	// 1. Create Connection
	createPayload := map[string]interface{}{
		"chain_id":                "chain-hq-1",
		"erp_type":                "TALLY",
		"integration_mode":        "AGENT_POLLED",
		"display_name":            "Store 1 Tally Connection",
		"enabled_outbound_events": []string{"order.completed"},
	}
	bodyBytes, _ := json.Marshal(createPayload)

	reqCreate := httptest.NewRequest("POST", "/v1/integration/connections", bytes.NewReader(bodyBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("X-User-Role", "OWNER")
	reqCreate.Header.Set("X-Chain-ID", "chain-hq-1")
	reqCreate.Header.Set("X-User-ID", "usr-owner-1")

	rrCreate := httptest.NewRecorder()
	r.ServeHTTP(rrCreate, reqCreate)

	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created on connection creation, got %d (body: %s)", rrCreate.Code, rrCreate.Body.String())
	}

	var createResp CreateConnectionResponse
	_ = json.NewDecoder(rrCreate.Body).Decode(&createResp)

	if createResp.PlaintextSecret == "" {
		t.Fatalf("Expected webhook_secret returned in creation response")
	}
	if createResp.PlaintextAgentAPIKey == nil || *createResp.PlaintextAgentAPIKey == "" {
		t.Fatalf("Expected agent_api_key returned in creation response for AGENT_POLLED mode")
	}

	connID := createResp.Connection.ID

	// 2. GET Connection by ID
	reqGet := httptest.NewRequest("GET", "/v1/integration/connections/"+connID, nil)
	reqGet.Header.Set("X-User-Role", "OWNER")
	reqGet.Header.Set("X-Chain-ID", "chain-hq-1")

	rrGet := httptest.NewRecorder()
	r.ServeHTTP(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on GET connection, got %d", rrGet.Code)
	}

	var getConn ERPConnection
	_ = json.NewDecoder(rrGet.Body).Decode(&getConn)

	// Verify plaintext secret & agent API key are NOT exposed in ERPConnection struct json
	rawGetJSON := rrGet.Body.String()
	if bytes.Contains([]byte(rawGetJSON), []byte(createResp.PlaintextSecret)) {
		t.Fatalf("GET response MUST NOT contain plaintext webhook secret!")
	}
	if bytes.Contains([]byte(rawGetJSON), []byte(*createResp.PlaintextAgentAPIKey)) {
		t.Fatalf("GET response MUST NOT contain plaintext agent_api_key!")
	}
}
