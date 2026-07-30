package main

import (
	"context"
	"testing"
	"time"
)

func TestFunnelAnalytics_IncompleteSession_ReturnsAll5StagesWithZeros(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	now := time.Now().UTC()

	sessionID := "sess-incomplete-100"
	storeID := "store-100"

	// Insert stage 1 & stage 2 only
	_ = repo.InsertFunnelEvent(ctx, &FunnelEvent{
		EventDate: now,
		EventTime: now,
		StoreID:   storeID,
		SessionID: sessionID,
		Stage:     StageSessionStarted,
	})
	_ = repo.InsertFunnelEvent(ctx, &FunnelEvent{
		EventDate: now,
		EventTime: now.Add(10 * time.Second),
		StoreID:   storeID,
		SessionID: sessionID,
		Stage:     StageCheckoutInitiated,
	})

	stages, err := repo.GetFunnel(ctx, storeID, now.Format("2006-01-02"), now.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("unexpected error fetching funnel: %v", err)
	}

	if len(stages) != 5 {
		t.Fatalf("expected all 5 stages in fixed funnel response array, got %d", len(stages))
	}

	// Verify stage 1 & 2 have count 1, stage 3, 4, 5 have count 0
	if stages[0].Stage != StageSessionStarted || stages[0].SessionCount != 1 {
		t.Fatalf("expected SESSION_STARTED count 1, got %v", stages[0])
	}
	if stages[1].Stage != StageCheckoutInitiated || stages[1].SessionCount != 1 {
		t.Fatalf("expected CHECKOUT_INITIATED count 1, got %v", stages[1])
	}
	if stages[2].Stage != StagePaymentConfirmed || stages[2].SessionCount != 0 {
		t.Fatalf("expected PAYMENT_CONFIRMED count 0, got %v", stages[2])
	}
	if stages[3].Stage != StageOrderCompleted || stages[3].SessionCount != 0 {
		t.Fatalf("expected ORDER_COMPLETED count 0, got %v", stages[3])
	}
	if stages[4].Stage != StageExitValidated || stages[4].SessionCount != 0 {
		t.Fatalf("expected EXIT_VALIDATED count 0, got %v", stages[4])
	}
}
