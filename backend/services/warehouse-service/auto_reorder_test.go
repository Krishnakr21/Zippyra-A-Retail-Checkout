package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/kafka"
)

func TestAutoReorder_IdempotentPerStoreBarcodeDay(t *testing.T) {
	db, repo := setupWarehouseTestDB(t)
	defer db.Close()

	producer := kafka.NewProducer("localhost:9092")
	consumer := NewEventConsumer(repo, producer)

	ctx := context.Background()
	storeID := "store-auto-1"
	barcode := "8907777777777"

	payload := LowStockKafkaPayload{
		StoreID:      storeID,
		Barcode:      barcode,
		CurrentQty:   5,
		ReorderPoint: 10,
		ReorderQty:   50,
		Timestamp:    time.Now(),
	}
	bytesVal, _ := json.Marshal(payload)

	// 1. First low stock event -> Creates Auto PO
	if err := consumer.ProcessLowStockEvent(ctx, bytesVal); err != nil {
		t.Fatalf("First ProcessLowStockEvent failed: %v", err)
	}

	pos1, _ := repo.ListPOs(ctx, storeID, "", 10, 0)
	if len(pos1) != 1 {
		t.Fatalf("Expected 1 auto PO created on first event, got %d", len(pos1))
	}
	if pos1[0].Source != POSourceAutoReorder {
		t.Errorf("Expected source AUTO_REORDER, got %s", pos1[0].Source)
	}

	// 2. Second low stock event (flapping event on same day) -> SKIPPED via UNIQUE constraint / idempotency
	if err := consumer.ProcessLowStockEvent(ctx, bytesVal); err != nil {
		t.Fatalf("Second ProcessLowStockEvent failed: %v", err)
	}

	pos2, _ := repo.ListPOs(ctx, storeID, "", 10, 0)
	if len(pos2) != 1 {
		t.Errorf("Expected ONLY 1 auto PO to exist for store+barcode+day, got %d", len(pos2))
	}
}
