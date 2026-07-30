package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestAuthDPDPConsumer_PurgesUserPII(t *testing.T) {
	repo := NewMemoryRepository()
	u, _ := repo.CreateUserWithPhone(context.Background(), "+919876543210")

	consumer := NewAuthDPDPConsumer(repo, nil)

	reqPayload, _ := json.Marshal(DPDPDeletionRequestPayload{
		UserID:        u.ID,
		DPDPRequestID: "req-dpdp-001",
	})

	err := consumer.HandleUserDataDeletionRequested(context.Background(), reqPayload)
	if err != nil {
		t.Fatalf("HandleUserDataDeletionRequested failed: %v", err)
	}

	// Verify user is purged from repo
	fetched, _ := repo.GetUserByID(context.Background(), u.ID)
	if fetched != nil {
		t.Fatalf("Expected user PII to be purged, but user still exists")
	}
}
