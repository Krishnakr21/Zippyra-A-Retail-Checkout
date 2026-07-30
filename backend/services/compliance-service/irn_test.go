package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestIRNConsumer_HandleOrderCompleted_HappyPath(t *testing.T) {
	repo := NewMemoryRepository()
	irpClient := NewHTTPIRPClient("", "", "", "") // Dev mock mode

	consumer := NewIRNConsumer(repo, irpClient)

	orderCompletedPayload, _ := json.Marshal(OrderCompletedEventPayload{
		OrderID:    "ord-irn-001",
		StoreID:    "store-001",
		ChainID:    "chain-001",
		TotalPaise: 11800,
		CGSTPaise:  900,
		SGSTPaise:  900,
		IGSTPaise:  0,
		Items: []map[string]interface{}{
			{
				"hsn_code":    "8471",
				"name":        "Wireless Mouse",
				"qty":         1,
				"price_paise": 10000,
			},
		},
		Timestamp: time.Now().UTC(),
	})

	err := consumer.HandleOrderCompleted(context.Background(), orderCompletedPayload)
	if err != nil {
		t.Fatalf("HandleOrderCompleted failed: %v", err)
	}

	rec, err := repo.GetIRNRecordByOrderID(context.Background(), "ord-irn-001")
	if err != nil || rec == nil {
		t.Fatalf("IRN record not found: %v", err)
	}

	if rec.Status != IRNStatusIssued {
		t.Fatalf("Expected IRN status ISSUED, got %s", rec.Status)
	}

	if rec.IRN == nil || *rec.IRN == "" {
		t.Fatalf("Expected valid IRN string, got nil or empty")
	}

	outbox, _ := repo.GetPendingOutbox(context.Background(), 10)
	if len(outbox) == 0 || outbox[0].Topic != "compliance.irn_issued" {
		t.Fatalf("Expected compliance.irn_issued outbox event")
	}
}

func TestIRNRetryJob_RetriesFailedRecords(t *testing.T) {
	repo := NewMemoryRepository()
	irpClient := NewHTTPIRPClient("", "", "", "")

	// Create a failed IRN record
	payloadBytes, _ := json.Marshal(map[string]interface{}{
		"DocDtls": map[string]interface{}{"No": "ord-retry-1"},
	})
	rec := &IRNRecord{
		OrderID:    "ord-retry-1",
		StoreID:    "store-1",
		ChainID:    "chain-1",
		Status:     IRNStatusFailed,
		IRPPayload: string(payloadBytes),
		RetryCount: 1,
	}
	_, _ = repo.CreateIRNRecord(context.Background(), rec)

	job := NewIRNRetryJob(repo, irpClient)
	job.RunOnce(context.Background())

	updatedRec, _ := repo.GetIRNRecordByOrderID(context.Background(), "ord-retry-1")
	if updatedRec.Status != IRNStatusIssued {
		t.Fatalf("Expected retry job to issue IRN, got status %s", updatedRec.Status)
	}
}
