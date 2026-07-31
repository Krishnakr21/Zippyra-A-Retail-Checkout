package main

import (
	"context"
	"encoding/json"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestLoyaltyDPDPConsumer_PurgesLoyaltyData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewPostgresRepository(db)

	ctx := context.Background()

	// Ensure loyalty account exists
	acc, err := repo.EnsureAccountExists(ctx, "user-loyalty-dpdp-1")
	if err != nil || acc == nil {
		t.Fatalf("Failed to ensure account exists: %v", err)
	}

	consumer := NewLoyaltyDPDPConsumer(repo, nil)

	reqBytes, _ := json.Marshal(DPDPLoyaltyDeletionRequestPayload{
		UserID:        "user-loyalty-dpdp-1",
		DPDPRequestID: "req-dpdp-loyalty-001",
	})

	if err := consumer.HandleUserDataDeletionRequested(ctx, reqBytes); err != nil {
		t.Fatalf("HandleUserDataDeletionRequested failed: %v", err)
	}

	// Verify account is purged
	fetched, _ := repo.GetAccountByUserID(ctx, "user-loyalty-dpdp-1")
	if fetched != nil {
		t.Fatalf("Expected loyalty account to be purged, but it still exists")
	}
}
