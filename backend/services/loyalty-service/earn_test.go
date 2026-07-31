package main

import (
	"context"
	"testing"
)

func TestEarnPoints_TierUpgradeAndIdempotency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	ctx := context.Background()
	userID := "usr-earn-100"

	// 1. Initial order of ₹45,000 (4,500,000 paise) -> Bronze 1.0x -> 4,500 points
	earned1, oldTier1, newTier1, upgraded1, balance1, err := repo.EarnPointsTx(ctx, "ord-earn-1", userID, 4500000)
	if err != nil {
		t.Fatalf("EarnPointsTx 1 failed: %v", err)
	}

	if earned1 != 4500 || oldTier1 != "BRONZE" || newTier1 != "BRONZE" || upgraded1 {
		t.Errorf("Earn 1 output invalid: earned=%d, old=%s, new=%s, upgraded=%v", earned1, oldTier1, newTier1, upgraded1)
	}
	if balance1 != 4500 {
		t.Errorf("Expected balance 4500, got %d", balance1)
	}

	// 2. Second order of ₹10,000 (1,000,000 paise) -> Bronze 1.0x -> 1,000 points -> total lifetime 5,500 points -> upgrades to SILVER!
	earned2, oldTier2, newTier2, upgraded2, balance2, err := repo.EarnPointsTx(ctx, "ord-earn-2", userID, 1000000)
	if err != nil {
		t.Fatalf("EarnPointsTx 2 failed: %v", err)
	}

	if earned2 != 1000 || oldTier2 != "BRONZE" || newTier2 != "SILVER" || !upgraded2 {
		t.Errorf("Earn 2 tier upgrade failed: earned=%d, old=%s, new=%s, upgraded=%v", earned2, oldTier2, newTier2, upgraded2)
	}
	if balance2 != 5500 {
		t.Errorf("Expected balance 5500, got %d", balance2)
	}

	// 3. Duplicate delivery of order "ord-earn-2" (idempotency check)
	earned3, _, _, upgraded3, balance3, err := repo.EarnPointsTx(ctx, "ord-earn-2", userID, 1000000)
	if err != nil {
		t.Fatalf("Duplicate EarnPointsTx failed: %v", err)
	}

	if earned3 != 0 || upgraded3 || balance3 != 5500 {
		t.Errorf("Duplicate earn mutated balance! earned=%d, upgraded=%v, balance=%d", earned3, upgraded3, balance3)
	}
}

func TestEarnPoints_SubscriptionBonusMultiplier(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	ctx := context.Background()
	userID := "usr-member-777"

	// 1. Without subscription: ₹100 (10,000 paise) -> Bronze 1.0x -> 10 points
	earned1, _, _, _, _, _ := repo.EarnPointsTx(ctx, "ord-sub-base", userID, 10000)
	if earned1 != 10 {
		t.Errorf("expected 10 points without subscription, got %d", earned1)
	}

	// 2. Insert active subscription for user
	db.Exec("INSERT INTO subscription_plans (id, chain_id, name, price_paise, billing_interval, benefits) VALUES ('plan-saver', 'chain-1', 'Saver', 19900, 'MONTHLY', '{\"loyalty_multiplier_bonus\": 0.5}')")
	db.Exec("INSERT INTO member_subscriptions (id, user_id, plan_id, status) VALUES ('sub-777', 'usr-member-777', 'plan-saver', 'ACTIVE')")

	// 3. With active subscription: ₹100 (10,000 paise) -> Bronze 1.0x + 0.5x = 1.5x -> 15 points!
	earned2, _, _, _, _, _ := repo.EarnPointsTx(ctx, "ord-sub-bonus", userID, 10000)
	if earned2 != 15 {
		t.Errorf("expected 15 points with active subscription bonus, got %d", earned2)
	}

	// 4. Cancel subscription -> returns to base 1.0x multiplier (future billing stops, past points preserved)
	db.Exec("UPDATE member_subscriptions SET status = 'CANCELLED' WHERE user_id = 'usr-member-777'")
	earned3, _, _, _, _, _ := repo.EarnPointsTx(ctx, "ord-sub-after-cancel", userID, 10000)
	if earned3 != 10 {
		t.Errorf("expected 10 points after subscription cancellation, got %d", earned3)
	}
}
