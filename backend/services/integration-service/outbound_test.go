package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestOutbound_MatchingConnection_CreatesSyncJob(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	masterKey := "12345678901234567890123456789012"
	pushWorker := NewDirectPushWorker(repo, masterKey)
	consumer := NewEventConsumer(repo, pushWorker)

	conn := &ERPConnection{
		ID:                    "conn-active-sap",
		ChainID:               "chain-1",
		ERPType:               ERPTypeSAP,
		IntegrationMode:       IntegrationModeDirect,
		EnabledOutboundEvents: []string{"order.completed"},
		Status:                ConnectionStatusActive,
	}
	_ = repo.CreateConnection(context.Background(), conn)

	eventBytes, _ := json.Marshal(map[string]interface{}{
		"event_id": "evt-order-999",
		"order_id": "ord-999",
		"chain_id": "chain-1",
		"store_id": "store-1",
	})

	err := consumer.ProcessEvent(context.Background(), "order.completed", eventBytes)
	if err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	jobs, err := repo.ListSyncJobs(context.Background(), "conn-active-sap", nil, "OUTBOUND")
	if err != nil {
		t.Fatalf("ListSyncJobs failed: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("Expected 1 sync job created, got %d", len(jobs))
	}
	if jobs[0].SourceEventID != "evt-order-999" {
		t.Fatalf("Expected SourceEventID 'evt-order-999', got %s", jobs[0].SourceEventID)
	}
}

func TestOutbound_NoMatchingConnection_CreatesZeroSyncJobs(t *testing.T) {
	repo := NewMemoryIntegrationRepository()
	masterKey := "12345678901234567890123456789012"
	pushWorker := NewDirectPushWorker(repo, masterKey)
	consumer := NewEventConsumer(repo, pushWorker)

	// Connection only enabled for inventory.stock_updated
	conn := &ERPConnection{
		ID:                    "conn-inventory-only",
		ChainID:               "chain-2",
		ERPType:               ERPTypeSAP,
		IntegrationMode:       IntegrationModeDirect,
		EnabledOutboundEvents: []string{"inventory.stock_updated"},
		Status:                ConnectionStatusActive,
	}
	_ = repo.CreateConnection(context.Background(), conn)

	eventBytes, _ := json.Marshal(map[string]interface{}{
		"event_id": "evt-order-888",
		"order_id": "ord-888",
		"chain_id": "chain-2",
	})

	// Order completed event emitted for chain-2
	err := consumer.ProcessEvent(context.Background(), "order.completed", eventBytes)
	if err != nil {
		t.Fatalf("ProcessEvent failed: %v", err)
	}

	jobs, _ := repo.ListSyncJobs(context.Background(), "conn-inventory-only", nil, "OUTBOUND")
	if len(jobs) != 0 {
		t.Fatalf("Expected 0 sync jobs for non-matching event type, got %d", len(jobs))
	}
}
