package main

import (
	"context"
	"testing"
	"time"
)

func TestOfferScheduleJob(t *testing.T) {
	rdb := NewMemoryRedis()
	repo := NewMemoryOfferRepository()
	compiler := NewOfferCompiler(repo, rdb)
	job := NewOfferScheduleJob(compiler, repo)

	storeID := "store-sched-001"
	chainID := "chain-default-001"
	repo.SetStoreChain(storeID, chainID)

	now := time.Now().UTC()
	pastFrom := now.Add(-10 * time.Minute)
	expiredUntil := now.Add(-2 * time.Minute) // Expired 2 minutes ago

	expiredOffer := &Offer{
		ID:                "off-expired",
		ChainID:           chainID,
		StoreID:           &storeID,
		Type:              "PERCENT_OFF",
		AppliesTo:         "ALL",
		RuleConfig:        map[string]interface{}{"percent": 15.0},
		MinCartValuePaise: 0,
		Priority:          10,
		ActiveFrom:        pastFrom,
		ActiveUntil:       &expiredUntil,
		IsActive:          true,
		CreatedBy:         "manager",
	}
	_ = repo.CreateOffer(context.Background(), expiredOffer)

	// Run schedule job tick
	job.RunOnce(context.Background())

	// Verify Redis rules exclude expired offer
	val, err := rdb.Get(context.Background(), "offer_rules:"+storeID).Result()
	if err != nil {
		t.Fatalf("failed to get redis key: %v", err)
	}

	if indexOf(val, "off-expired") != -1 {
		t.Fatalf("expected expired offer off-expired to be excluded after schedule tick, got json: %s", val)
	}
	if val != "[]" {
		t.Errorf("expected '[]' after expired offer excluded, got %s", val)
	}
}
