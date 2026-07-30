package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupReferralRaceDB(t *testing.T) *sql.DB {
	dbName := fmt.Sprintf("file:refracedb_%s?mode=memory&cache=shared", t.Name())
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		t.Fatalf("failed to open sqlite memory DB: %v", err)
	}

	schema := `
		CREATE TABLE referral_events (
			id TEXT PRIMARY KEY,
			referrer_user_id TEXT NOT NULL,
			referred_user_id TEXT UNIQUE NOT NULL,
			referral_code TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	return db
}

func TestReferralReward_ConcurrentOrderCompletedEvents_PreventsDoubleReward(t *testing.T) {
	db := setupReferralRaceDB(t)
	defer db.Close()

	referrerID := "usr-referrer-001"
	referredID := "usr-referred-001"
	code := "REF999"

	// Simulate 10 concurrent order.completed delivery goroutines for the SAME first order of referredID
	var wg sync.WaitGroup
	numGoroutines := 10
	successCount := 0
	conflictCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			eventID := fmt.Sprintf("evt-ref-%d", idx)
			query := `INSERT INTO referral_events (id, referrer_user_id, referred_user_id, referral_code, status, created_at)
					  VALUES (?, ?, ?, ?, 'REWARDED', ?)`

			_, err := db.ExecContext(context.Background(), query, eventID, referrerID, referredID, code, time.Now())
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				conflictCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful referral reward, got %d", successCount)
	}
	if conflictCount != numGoroutines-1 {
		t.Fatalf("expected %d UNIQUE constraint conflicts, got %d", numGoroutines-1, conflictCount)
	}

	// Verify only 1 row in referral_events table
	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM referral_events WHERE referred_user_id = ?", referredID).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 referral event record, got %d", count)
	}
}
