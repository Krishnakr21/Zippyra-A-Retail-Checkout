package main

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestOfferReconciliationJob(t *testing.T) {
	rdb := NewMemoryRedis()
	repo := NewMemoryOfferRepository()
	compiler := NewOfferCompiler(repo, rdb)
	job := NewOfferReconciliationJob(compiler, repo, rdb)

	storeID := "store-recon-001"
	chainID := "chain-default-001"
	repo.SetStoreChain(storeID, chainID)

	now := time.Now().UTC().Add(-1 * time.Minute)
	validOffer := &Offer{
		ID:                "off-recon-valid",
		ChainID:           chainID,
		StoreID:           &storeID,
		Type:              "PERCENT_OFF",
		AppliesTo:         "ALL",
		RuleConfig:        map[string]interface{}{"percent": 25.0},
		MinCartValuePaise: 0,
		Priority:          10,
		ActiveFrom:        now,
		IsActive:          true,
		CreatedBy:         "manager",
	}
	_ = repo.CreateOffer(context.Background(), validOffer)

	// 1. Initial compile
	_ = compiler.CompileAndPublish(context.Background(), storeID)

	// 2. Corrupt / delete Redis key mid-test
	redisKey := "offer_rules:" + storeID
	_ = rdb.Del(context.Background(), redisKey).Err()

	// Verify key is gone
	_, err := rdb.Get(context.Background(), redisKey).Result()
	if err != redis.Nil {
		t.Fatalf("expected key to be deleted, got: %v", err)
	}

	// 3. Run reconciliation sweep tick
	job.RunOnce(context.Background())

	// 4. Verify Redis key is restored correctly!
	val, err := rdb.Get(context.Background(), redisKey).Result()
	if err != nil {
		t.Fatalf("expected redis key to be restored by reconciliation sweep, got error: %v", err)
	}

	if indexOf(val, "off-recon-valid") == -1 {
		t.Errorf("expected restored ruleset to contain off-recon-valid, got %s", val)
	}

	// Verify health check
	if !job.IsHealthy() {
		t.Errorf("expected reconciliation job to be healthy after run")
	}
}
