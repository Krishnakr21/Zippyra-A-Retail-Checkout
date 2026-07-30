package main

import (
	"context"
	"testing"
)

func TestBackfillOfferRulesMode(t *testing.T) {
	rdb := NewMemoryRedis()
	repo := NewMemoryOfferRepository()
	compiler := NewOfferCompiler(repo, rdb)

	store1 := "store-backfill-001"
	store2 := "store-backfill-002"
	chainID := "chain-default-001"
	repo.SetStoreChain(store1, chainID)
	repo.SetStoreChain(store2, chainID)

	// store1 pre-exists with audit row
	_ = repo.LogOfferRulesAudit(context.Background(), store1, []byte("[]"))

	// Verify store2 has no audit row
	has2, _ := repo.HasAuditRow(context.Background(), store2)
	if has2 {
		t.Fatalf("expected store2 to have no audit row before backfill")
	}

	// List stores with no audit
	uncompiledStores, err := repo.ListStoresWithNoAudit(context.Background())
	if err != nil {
		t.Fatalf("failed to list stores with no audit: %v", err)
	}

	if len(uncompiledStores) != 1 || uncompiledStores[0] != store2 {
		t.Fatalf("expected [store-backfill-002] for uncompiled stores, got %v", uncompiledStores)
	}

	// Run backfill for store2
	for _, s := range uncompiledStores {
		_ = compiler.CompileAndPublish(context.Background(), s)
	}

	// Now store2 must have an audit row and Redis key!
	has2After, _ := repo.HasAuditRow(context.Background(), store2)
	if !has2After {
		t.Errorf("expected store2 to have audit row after backfill")
	}

	val, err := rdb.Get(context.Background(), "offer_rules:"+store2).Result()
	if err != nil || val != "[]" {
		t.Errorf("expected Redis key offer_rules:store-backfill-002 to be '[]', got %s (%v)", val, err)
	}
}
