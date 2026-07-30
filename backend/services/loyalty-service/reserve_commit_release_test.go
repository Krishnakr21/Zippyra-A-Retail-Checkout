package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbName := fmt.Sprintf("file:loyaltydb_%s?mode=memory&cache=shared", t.Name())
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		t.Fatalf("Failed to open sqlite memory DB: %v", err)
	}

	schema := `
		CREATE TABLE loyalty_accounts (
			user_id TEXT PRIMARY KEY,
			points_balance INTEGER NOT NULL DEFAULT 0,
			points_reserved INTEGER NOT NULL DEFAULT 0,
			lifetime_points_earned INTEGER NOT NULL DEFAULT 0,
			tier TEXT NOT NULL DEFAULT 'BRONZE',
			referral_code TEXT UNIQUE,
			tier_updated_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (points_balance >= 0 AND points_reserved >= 0)
		);

		CREATE TABLE referral_events (
			id TEXT PRIMARY KEY,
			referrer_user_id TEXT NOT NULL,
			referred_user_id TEXT UNIQUE NOT NULL,
			referral_code TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			first_order_id TEXT,
			rewarded_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL
		);

		CREATE TABLE subscription_plans (
			id TEXT PRIMARY KEY,
			chain_id TEXT NOT NULL,
			name TEXT NOT NULL,
			price_paise INTEGER NOT NULL,
			billing_interval TEXT NOT NULL,
			benefits TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE member_subscriptions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			plan_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			razorpay_subscription_id TEXT UNIQUE,
			current_period_end TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE loyalty_ledger (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			entry_type TEXT NOT NULL,
			points_delta INTEGER NOT NULL,
			reference_type TEXT,
			reference_id TEXT,
			idempotency_key TEXT UNIQUE NOT NULL,
			balance_after INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE loyalty_tier_config (
			tier TEXT PRIMARY KEY,
			min_lifetime_points INTEGER NOT NULL,
			earn_multiplier REAL NOT NULL DEFAULT 1.0,
			display_name TEXT NOT NULL,
			display_order INTEGER NOT NULL
		);

		INSERT INTO loyalty_tier_config (tier, min_lifetime_points, earn_multiplier, display_name, display_order)
		VALUES 
			('BRONZE', 0, 1.00, 'Bronze Tier', 1),
			('SILVER', 5000, 1.20, 'Silver Tier', 2),
			('GOLD', 20000, 1.50, 'Gold Tier', 3),
			('PLATINUM', 50000, 2.00, 'Platinum Tier', 4);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestReserveCommit_HappyPath(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	ctx := context.Background()
	userID := "usr-happy-100"

	// Setup initial account with 500 points
	_, _ = db.Exec(`INSERT INTO loyalty_accounts (user_id, points_balance, points_reserved, lifetime_points_earned, tier) VALUES ($1, 500, 0, 500, 'BRONZE')`, userID)

	// 1. Reserve 100 points
	reserved, balanceAfter, err := repo.ReservePointsTx(ctx, userID, 100, "reserve:pay-100")
	if err != nil || !reserved {
		t.Fatalf("ReservePointsTx failed: %v", err)
	}
	if balanceAfter != 400 {
		t.Errorf("Expected balanceAfter 400, got %d", balanceAfter)
	}

	// Verify DB state
	acc, _ := repo.GetAccountByUserID(ctx, userID)
	if acc.PointsBalance != 400 || acc.PointsReserved != 100 {
		t.Errorf("After reserve: balance=%d, reserved=%d; want 400, 100", acc.PointsBalance, acc.PointsReserved)
	}

	// 2. Commit 100 points
	committed, err := repo.CommitPointsTx(ctx, userID, 100, "commit:pay-100")
	if err != nil || !committed {
		t.Fatalf("CommitPointsTx failed: %v", err)
	}

	// Verify DB state
	acc, _ = repo.GetAccountByUserID(ctx, userID)
	if acc.PointsBalance != 400 || acc.PointsReserved != 0 {
		t.Errorf("After commit: balance=%d, reserved=%d; want 400, 0", acc.PointsBalance, acc.PointsReserved)
	}
}

func TestReserveRelease_HappyPath(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	ctx := context.Background()
	userID := "usr-release-100"

	_, _ = db.Exec(`INSERT INTO loyalty_accounts (user_id, points_balance, points_reserved, lifetime_points_earned, tier) VALUES ($1, 500, 0, 500, 'BRONZE')`, userID)

	// 1. Reserve 200 points
	_, _, err := repo.ReservePointsTx(ctx, userID, 200, "reserve:pay-200")
	if err != nil {
		t.Fatalf("Reserve failed: %v", err)
	}

	// 2. Release 200 points
	released, balanceAfter, err := repo.ReleasePointsTx(ctx, userID, 200, "release:pay-200")
	if err != nil || !released {
		t.Fatalf("Release failed: %v", err)
	}
	if balanceAfter != 500 {
		t.Errorf("Expected restored balanceAfter 500, got %d", balanceAfter)
	}

	acc, _ := repo.GetAccountByUserID(ctx, userID)
	if acc.PointsBalance != 500 || acc.PointsReserved != 0 {
		t.Errorf("After release: balance=%d, reserved=%d; want 500, 0", acc.PointsBalance, acc.PointsReserved)
	}
}

func TestDuplicateReserve_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	ctx := context.Background()
	userID := "usr-dupe-100"

	_, _ = db.Exec(`INSERT INTO loyalty_accounts (user_id, points_balance, points_reserved, lifetime_points_earned, tier) VALUES ($1, 500, 0, 500, 'BRONZE')`, userID)

	// Call 1
	res1, bal1, err1 := repo.ReservePointsTx(ctx, userID, 100, "reserve:pay-dupe")
	if err1 != nil || !res1 {
		t.Fatalf("First reserve failed: %v", err1)
	}

	// Call 2 (duplicate idempotency key)
	res2, bal2, err2 := repo.ReservePointsTx(ctx, userID, 100, "reserve:pay-dupe")
	if err2 != nil || !res2 {
		t.Fatalf("Second reserve failed: %v", err2)
	}

	if bal1 != bal2 || bal2 != 400 {
		t.Errorf("Idempotent reserve output mismatch: bal1=%d, bal2=%d", bal1, bal2)
	}

	// Verify points only deducted ONCE
	acc, _ := repo.GetAccountByUserID(ctx, userID)
	if acc.PointsBalance != 400 || acc.PointsReserved != 100 {
		t.Errorf("Points double deducted! balance=%d, reserved=%d", acc.PointsBalance, acc.PointsReserved)
	}
}

func TestInsufficientBalance_Rejection(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	ctx := context.Background()
	userID := "usr-low-100"

	_, _ = db.Exec(`INSERT INTO loyalty_accounts (user_id, points_balance, points_reserved, lifetime_points_earned, tier) VALUES ($1, 50, 0, 50, 'BRONZE')`, userID)

	// Attempt to reserve 100 points
	reserved, _, err := repo.ReservePointsTx(ctx, userID, 100, "reserve:pay-insufficient")
	if err == nil || reserved {
		t.Fatalf("Expected error for insufficient balance, got success")
	}

	acc, _ := repo.GetAccountByUserID(ctx, userID)
	if acc.PointsBalance != 50 || acc.PointsReserved != 0 {
		t.Errorf("Balance mutated after failed reserve attempt: balance=%d", acc.PointsBalance)
	}
}

func TestConcurrentReserve_OverflowProtection(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	ctx := context.Background()
	userID := "usr-race-100"

	// User has 100 points
	_, _ = db.Exec(`INSERT INTO loyalty_accounts (user_id, points_balance, points_reserved, lifetime_points_earned, tier) VALUES ($1, 100, 0, 100, 'BRONZE')`, userID)

	// 10 concurrent goroutines trying to reserve 100 points each with distinct idempotency keys
	var wg sync.WaitGroup
	successCount := 0
	failureCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			idemKey := "reserve:race-" + string(rune('A'+idx))
			res, _, err := repo.ReservePointsTx(ctx, userID, 100, idemKey)
			mu.Lock()
			if err == nil && res {
				successCount++
			} else {
				failureCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if successCount != 1 || failureCount != 9 {
		t.Fatalf("Race test failure: successCount=%d, failureCount=%d; want 1, 9", successCount, failureCount)
	}

	acc, _ := repo.GetAccountByUserID(ctx, userID)
	if acc.PointsBalance != 0 || acc.PointsReserved != 100 {
		t.Errorf("Final account balance invalid: balance=%d, reserved=%d", acc.PointsBalance, acc.PointsReserved)
	}
}
