package main

import (
	"context"
	"testing"
	"time"
)

func TestOfferCompiler(t *testing.T) {
	rdb := NewMemoryRedis()
	repo := NewMemoryOfferRepository()
	compiler := NewOfferCompiler(repo, rdb)

	storeID := "store-comp-001"
	chainID := "chain-comp-001"
	repo.SetStoreChain(storeID, chainID)

	t.Run("Priority sorting and Redis + audit log writing", func(t *testing.T) {
		now := time.Now().UTC().Add(-1 * time.Minute)

		// 1. Chain-wide offer (Priority 200)
		cwOffer := &Offer{
			ID:                "off-cw",
			ChainID:           chainID,
			StoreID:           nil,
			Type:              "PERCENT_OFF",
			AppliesTo:         "ALL",
			RuleConfig:        map[string]interface{}{"percent": 10.0},
			MinCartValuePaise: 0,
			Priority:          200,
			ActiveFrom:        now,
			IsActive:          true,
			CreatedBy:         "owner",
		}
		_ = repo.CreateOffer(context.Background(), cwOffer)

		// 2. Store-specific offer (Priority 50 - evaluated first)
		ssOffer := &Offer{
			ID:                "off-ss",
			ChainID:           chainID,
			StoreID:           &storeID,
			Type:              "FLAT_OFF",
			AppliesTo:         "ALL",
			RuleConfig:        map[string]interface{}{"flat_paise": 1000.0},
			MinCartValuePaise: 0,
			Priority:          50,
			ActiveFrom:        now,
			IsActive:          true,
			CreatedBy:         "manager",
		}
		_ = repo.CreateOffer(context.Background(), ssOffer)

		err := compiler.CompileAndPublish(context.Background(), storeID)
		if err != nil {
			t.Fatalf("compiler failed: %v", err)
		}

		// Verify Redis key
		val, err := rdb.Get(context.Background(), "offer_rules:"+storeID).Result()
		if err != nil {
			t.Fatalf("expected redis key offer_rules:%s, got error: %v", storeID, err)
		}

		// Check order: off-ss (priority 50) must come before off-cw (priority 200)
		idxSS := indexOf(val, "off-ss")
		idxCW := indexOf(val, "off-cw")
		if idxSS == -1 || idxCW == -1 || idxSS >= idxCW {
			t.Fatalf("expected off-ss (priority 50) before off-cw (priority 200) in json: %s", val)
		}

		// Verify Audit row written
		hasAudit, _ := repo.HasAuditRow(context.Background(), storeID)
		if !hasAudit {
			t.Errorf("expected offer_rules_audit row for store %s", storeID)
		}
	})

	t.Run("Store with zero applicable offers writes [] to Redis and logs audit row", func(t *testing.T) {
		emptyStore := "store-empty-001"
		repo.SetStoreChain(emptyStore, "chain-empty-001")

		err := compiler.CompileAndPublish(context.Background(), emptyStore)
		if err != nil {
			t.Fatalf("compiler failed for empty store: %v", err)
		}

		val, err := rdb.Get(context.Background(), "offer_rules:"+emptyStore).Result()
		if err != nil {
			t.Fatalf("expected redis key for empty store, got error: %v", err)
		}
		if val != "[]" {
			t.Errorf("expected '[]' in redis for empty store, got %s", val)
		}

		hasAudit, _ := repo.HasAuditRow(context.Background(), emptyStore)
		if !hasAudit {
			t.Errorf("expected audit row for empty store")
		}
	})
}

func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
