package main

import (
	"context"
	"testing"
)

func TestReversal_ProportionalDeductionAndFloorZero(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	ctx := context.Background()
	userID := "usr-reverse-100"

	// 1. User earns 1,000 points on a ₹10,000 order (1,000,000 paise)
	_, _, _, _, _, _ = repo.EarnPointsTx(ctx, "ord-rev-100", userID, 1000000)

	// 2. Partial return of 50% of the order (₹5,000 / 500,000 paise) -> reverses 500 points
	reversed1, balance1, err := repo.ReversePointsTx(ctx, "ord-rev-100", userID, "ret-1", 500000, 1000000)
	if err != nil {
		t.Fatalf("ReversePointsTx 1 failed: %v", err)
	}

	if reversed1 != 500 || balance1 != 500 {
		t.Errorf("Reversal 1 output invalid: reversed=%d, balance=%d; want 500, 500", reversed1, balance1)
	}

	acc1, _ := repo.GetAccountByUserID(ctx, userID)
	if acc1.LifetimePointsEarned != 1000 || acc1.Tier != "BRONZE" {
		t.Errorf("Reversal mutated lifetime points or tier! lifetime=%d, tier=%s", acc1.LifetimePointsEarned, acc1.Tier)
	}

	// 3. User spends remaining 500 points (balance drops to 0)
	_, _, _ = repo.ReservePointsTx(ctx, userID, 500, "reserve:rev-spend")
	_, _ = repo.CommitPointsTx(ctx, userID, 500, "commit:rev-spend")

	acc2, _ := repo.GetAccountByUserID(ctx, userID)
	if acc2.PointsBalance != 0 {
		t.Fatalf("Setup balance after spend failed: balance=%d", acc2.PointsBalance)
	}

	// 4. Second partial return of remaining 50% (₹5,000) -> attempts to reverse 500 points, but balance is 0!
	reversed2, balance2, err := repo.ReversePointsTx(ctx, "ord-rev-100", userID, "ret-2", 500000, 1000000)
	if err != nil {
		t.Fatalf("ReversePointsTx 2 failed: %v", err)
	}

	// Must floor at 0 (never negative)
	if reversed2 != 0 || balance2 != 0 {
		t.Errorf("Zero floor protection failed: reversed=%d, balance=%d; want 0, 0", reversed2, balance2)
	}
}
