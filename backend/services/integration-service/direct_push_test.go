package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zippyra/backend/shared/crypto"
)

func TestDirectPushWorker_RetryCeiling(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	masterKey := "12345678901234567890123456789012"
	pushWorker := NewDirectPushWorker(repo, masterKey)

	// Mock server that always fails with 500
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "SAP Internal Server Error", http.StatusInternalServerError)
	}))
	defer failServer.Close()

	outboundCfg := OutboundConfig{BaseURL: failServer.URL}
	rawCfg, _ := crypto.Encrypt([]byte(`{"base_url":"`+failServer.URL+`"}`), masterKey)
	_ = outboundCfg

	conn := &ERPConnection{
		ID:                      "conn-failing-sap",
		ChainID:                 "chain-1",
		ERPType:                 ERPTypeSAP,
		IntegrationMode:         IntegrationModeDirect,
		OutboundConfigEncrypted: rawCfg,
		Status:                  ConnectionStatusActive,
	}
	_ = repo.CreateConnection(context.Background(), conn)

	job := &ERPSyncJob{
		ID:              "job-fail-1",
		ConnectionID:    "conn-failing-sap",
		Direction:       "OUTBOUND",
		SourceEventType: "order.completed",
		SourceEventID:   "ord-fail-1",
		Payload:         []byte(`{"order_id":"ord-fail-1"}`),
		Status:          SyncJobStatusPending,
		AttemptCount:    9, // Set attempt count to 9 (1 before ceiling)
	}
	_, _ = repo.CreateSyncJob(context.Background(), job)

	// Run retry push -> attempt_count becomes 10, status becomes FAILED
	err := pushWorker.PushSyncJob(context.Background(), job, conn)
	if err == nil {
		t.Fatalf("Expected error when target server returns 500, got nil")
	}

	jAfter, _ := repo.GetSyncJobByID(context.Background(), "job-fail-1")
	if jAfter.AttemptCount != 10 {
		t.Fatalf("Expected AttemptCount 10, got %d", jAfter.AttemptCount)
	}
	if jAfter.Status != SyncJobStatusFailed {
		t.Fatalf("Expected Status FAILED, got %s", jAfter.Status)
	}

	// Verify runRetrySweep ignores jobs with attempt_count >= 10
	pushWorker.runRetrySweep(context.Background())

	jAfterSweep, _ := repo.GetSyncJobByID(context.Background(), "job-fail-1")
	if jAfterSweep.AttemptCount != 10 {
		t.Fatalf("Expected AttemptCount to remain 10 after retry sweep (ceiling reached), got %d", jAfterSweep.AttemptCount)
	}
}
