package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDPDPDeletionConsumer_AnonymizesChainHQUser(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// 1. Create a chain HQ user
	user := &ChainHQUser{
		ID:       "hq-usr-dpdp-001",
		ChainID:  "chain-001",
		Name:     "Vikram Patel",
		Phone:    "+919876543210",
		Role:     "FINANCE",
		IsActive: true,
	}
	_ = repo.CreateUser(ctx, user)

	consumer := NewDPDPDeletionConsumer(repo, nil)

	// 2. Process DPDP deletion request for CHAIN_HQ
	payload, _ := json.Marshal(DPDPDeletionRequestedPayload{
		UserID:        "hq-usr-dpdp-001",
		UserType:      "CHAIN_HQ",
		DPDPRequestID: "req-dpdp-hq-1",
	})

	if err := consumer.ProcessDeletionRequest(ctx, payload); err != nil {
		t.Fatalf("ProcessDeletionRequest failed: %v", err)
	}

	// 3. Verify user anonymized
	updatedUser, err := repo.GetUserByID(ctx, "hq-usr-dpdp-001")
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	if updatedUser.Name != "Anonymized HQ User" {
		t.Errorf("Expected name 'Anonymized HQ User', got '%s'", updatedUser.Name)
	}

	if updatedUser.IsActive != false {
		t.Errorf("Expected IsActive = false, got true")
	}

	if updatedUser.Phone != "deleted_hq-usr-dpdp-001" {
		t.Errorf("Expected tombstone phone 'deleted_hq-usr-dpdp-001', got '%s'", updatedUser.Phone)
	}
}
