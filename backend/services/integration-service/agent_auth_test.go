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

func TestAgentAuth_WrongKey_Returns401(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	handler := NewAgentHandler(repo)

	validKey := "agent_key_secret123"
	keyHash := hashString(validKey)

	conn := &ERPConnection{
		ID:              "conn-tally-1",
		ChainID:         "chain-1",
		ERPType:         ERPTypeTally,
		IntegrationMode: IntegrationModeAgentPolled,
		AgentAPIKeyHash: &keyHash,
		Status:          ConnectionStatusActive,
	}
	_ = repo.CreateConnection(context.Background(), conn)

	req := httptest.NewRequest("GET", "/v1/integration/connections/conn-tally-1/pull-queue", nil)
	req.Header.Set("Authorization", "Bearer agent_key_WRONG_TOKEN")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/v1/integration/connections/{id}/pull-queue", handler.PullQueue)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized for wrong agent key, got %d", rr.Code)
	}
}

func TestAgentAuth_DirectConnection_Returns400(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	handler := NewAgentHandler(repo)

	keyHash := hashString("agent_key_123")

	conn := &ERPConnection{
		ID:              "conn-sap-direct",
		ChainID:         "chain-1",
		ERPType:         ERPTypeSAP,
		IntegrationMode: IntegrationModeDirect,
		AgentAPIKeyHash: &keyHash,
		Status:          ConnectionStatusActive,
	}
	_ = repo.CreateConnection(context.Background(), conn)

	req := httptest.NewRequest("GET", "/v1/integration/connections/conn-sap-direct/pull-queue", nil)
	req.Header.Set("Authorization", "Bearer agent_key_123")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/v1/integration/connections/{id}/pull-queue", handler.PullQueue)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for pull-queue on DIRECT connection, got %d", rr.Code)
	}
}

func TestAgentPullAndAck_Flow(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	handler := NewAgentHandler(repo)

	validKey := "agent_key_valid_999"
	keyHash := hashString(validKey)

	conn := &ERPConnection{
		ID:              "conn-busy-1",
		ChainID:         "chain-1",
		ERPType:         ERPTypeBusy,
		IntegrationMode: IntegrationModeAgentPolled,
		AgentAPIKeyHash: &keyHash,
		Status:          ConnectionStatusPendingSetup,
	}
	_ = repo.CreateConnection(context.Background(), conn)

	// Seed 2 PENDING sync jobs
	job1 := &ERPSyncJob{ID: "job-1", ConnectionID: "conn-busy-1", Direction: "OUTBOUND", SourceEventType: "order.completed", SourceEventID: "ord-1", Status: SyncJobStatusPending}
	job2 := &ERPSyncJob{ID: "job-2", ConnectionID: "conn-busy-1", Direction: "OUTBOUND", SourceEventType: "order.completed", SourceEventID: "ord-2", Status: SyncJobStatusPending}
	_, _ = repo.CreateSyncJob(context.Background(), job1)
	_, _ = repo.CreateSyncJob(context.Background(), job2)

	r := mux.NewRouter()
	r.HandleFunc("/v1/integration/connections/{id}/pull-queue", handler.PullQueue).Methods("GET")
	r.HandleFunc("/v1/integration/connections/{id}/pull-queue/ack", handler.AckQueue).Methods("POST")

	// 1. GET /pull-queue -> returns 2 jobs, marks DELIVERED
	reqPull := httptest.NewRequest("GET", "/v1/integration/connections/conn-busy-1/pull-queue", nil)
	reqPull.Header.Set("Authorization", "Bearer "+validKey)

	rrPull := httptest.NewRecorder()
	r.ServeHTTP(rrPull, reqPull)

	if rrPull.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on pull-queue, got %d", rrPull.Code)
	}

	var pullResp struct {
		Jobs  []*ERPSyncJob `json:"jobs"`
		Count int           `json:"count"`
	}
	_ = json.NewDecoder(rrPull.Body).Decode(&pullResp)

	if pullResp.Count != 2 {
		t.Fatalf("Expected 2 jobs pulled, got %d", pullResp.Count)
	}

	// Verify job1 status is now DELIVERED
	j1, _ := repo.GetSyncJobByID(context.Background(), "job-1")
	if j1.Status != SyncJobStatusDelivered {
		t.Fatalf("Expected job-1 status DELIVERED, got %s", j1.Status)
	}

	// Verify connection flipped PENDING_SETUP -> ACTIVE on first poll
	connAfter, _ := repo.GetConnectionByID(context.Background(), "conn-busy-1")
	if connAfter.Status != ConnectionStatusActive {
		t.Fatalf("Expected connection status ACTIVE after first agent poll, got %s", connAfter.Status)
	}

	// 2. POST /pull-queue/ack for job-1 only
	ackBody, _ := json.Marshal(map[string]interface{}{
		"job_ids": []string{"job-1"},
	})
	reqAck := httptest.NewRequest("POST", "/v1/integration/connections/conn-busy-1/pull-queue/ack", bytes.NewReader(ackBody))
	reqAck.Header.Set("Content-Type", "application/json")
	reqAck.Header.Set("Authorization", "Bearer "+validKey)

	rrAck := httptest.NewRecorder()
	r.ServeHTTP(rrAck, reqAck)

	if rrAck.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on ack, got %d", rrAck.Code)
	}

	// Verify job-1 is ACKNOWLEDGED, job-2 remains DELIVERED
	j1After, _ := repo.GetSyncJobByID(context.Background(), "job-1")
	j2After, _ := repo.GetSyncJobByID(context.Background(), "job-2")

	if j1After.Status != SyncJobStatusAcknowledged {
		t.Fatalf("Expected job-1 ACKNOWLEDGED, got %s", j1After.Status)
	}
	if j2After.Status != SyncJobStatusDelivered {
		t.Fatalf("Expected job-2 DELIVERED, got %s", j2After.Status)
	}
}
