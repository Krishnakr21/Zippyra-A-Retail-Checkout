package main

import (
	"context"
	"testing"
)

func TestSplitPayment_ReserveAndCommit(t *testing.T) {
	loyalty := NewMockLoyaltyServiceClient()
	ctx := context.Background()
	userID := "user-split-1"

	loyalty.Balances[userID] = 500 // 500 points = ₹5 (500 paise)

	// 1. Reserve 200 points
	err := loyalty.ReservePoints(ctx, userID, 200)
	if err != nil {
		t.Fatalf("ReservePoints failed: %v", err)
	}

	bal, _ := loyalty.GetPointsBalance(ctx, userID)
	if bal != 300 {
		t.Fatalf("Expected balance 300 after reservation, got %d", bal)
	}

	// 2. Commit points on webhook captured
	err = loyalty.CommitReservedPoints(ctx, userID, 200)
	if err != nil {
		t.Fatalf("CommitReservedPoints failed: %v", err)
	}

	balAfter, _ := loyalty.GetPointsBalance(ctx, userID)
	if balAfter != 300 {
		t.Fatalf("Expected balance 300 after commit, got %d", balAfter)
	}
}

func TestSplitPayment_ReserveAndReleaseOnFailure(t *testing.T) {
	loyalty := NewMockLoyaltyServiceClient()
	ctx := context.Background()
	userID := "user-split-2"

	loyalty.Balances[userID] = 500

	// 1. Reserve 200 points
	_ = loyalty.ReservePoints(ctx, userID, 200)

	// 2. Payment fails -> Release reserved points
	_ = loyalty.ReleaseReservedPoints(ctx, userID, 200)

	balAfter, _ := loyalty.GetPointsBalance(ctx, userID)
	if balAfter != 500 {
		t.Fatalf("Expected points restored to 500 after release, got %d", balAfter)
	}
}
