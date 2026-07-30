package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestVelocityMonitor_TriggersHighValueAlert(t *testing.T) {
	repo := NewMemoryRepository()
	vm := NewVelocityMonitor(repo)

	ctx := context.Background()

	// High value transaction > ₹2,000,000 (200,000,000 paise)
	payload, _ := json.Marshal(PaymentConfirmedVelocityPayload{
		PaymentID:   "pay-high-val-1",
		StoreID:     "store-luxury-1",
		AmountPaise: 250000000, // ₹2.5 Million
		Timestamp:   time.Now().UTC(),
	})

	err := vm.HandlePaymentConfirmed(ctx, payload)
	if err != nil {
		t.Fatalf("HandlePaymentConfirmed failed: %v", err)
	}

	alerts, err := repo.ListVelocityAlerts(ctx, "store-luxury-1", true)
	if err != nil || len(alerts) != 1 {
		t.Fatalf("Expected 1 velocity alert for store, got %d", len(alerts))
	}

	if alerts[0].AlertType != "UNUSUAL_TRANSACTION_VALUE" {
		t.Fatalf("Expected alert type UNUSUAL_TRANSACTION_VALUE, got %s", alerts[0].AlertType)
	}
}
