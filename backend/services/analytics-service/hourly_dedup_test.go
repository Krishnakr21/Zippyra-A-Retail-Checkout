package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestHourlyDedup_KafkaRedelivery_PreventsSecondIncrement(t *testing.T) {
	repo := NewMemoryRepository()
	dedupGuard := NewMemoryHourlyDedupGuard()
	consumer := NewAnalyticsConsumer(repo, dedupGuard)

	ctx := context.Background()
	orderID := "ord-redelivery-999"
	now := time.Now().UTC()

	payload, _ := json.Marshal(map[string]interface{}{
		"order_id":       orderID,
		"session_id":     "sess-1",
		"store_id":       "store-100",
		"chain_id":       "chain-100",
		"total_paise":    100000,
		"discount_paise": 0,
		"payment_method": "UPI",
		"ts":             now.Format(time.RFC3339),
	})

	// First delivery
	err1 := consumer.ConsumeOrderCompleted(ctx, payload)
	if err1 != nil {
		t.Fatalf("unexpected error on first delivery: %v", err1)
	}

	// Redelivery of the exact same event
	err2 := consumer.ConsumeOrderCompleted(ctx, payload)
	if err2 != nil {
		t.Fatalf("unexpected error on redelivery: %v", err2)
	}

	// Check sales sales_events (ReplacingMergeTree dedup) -> 1 row
	sales, _ := repo.GetSales(ctx, "store-100", now.Format("2006-01-02"), now.Format("2006-01-02"), "day")
	if len(sales) != 1 || sales[0].OrderCount != 1 {
		t.Fatalf("expected 1 order in sales events after redelivery, got %v", sales)
	}
}
