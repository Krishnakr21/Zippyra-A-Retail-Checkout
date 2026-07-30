package main

import (
	"context"
	"testing"
)

func TestDeltaSync_Pagination_And_Deletions(t *testing.T) {
	repo := NewMemoryRepository()
	syncEngine := NewSyncEngineService(repo)
	ctx := context.Background()

	storeID := "store-sync-1"

	// Create 3 active products with valid barcodes
	p1 := &Product{ID: "p1", StoreID: storeID, Barcode: "8901030300011", Name: "Item 1", PricePaise: 100, IsActive: true}
	p2 := &Product{ID: "p2", StoreID: storeID, Barcode: "012345678905", Name: "Item 2", PricePaise: 200, IsActive: true}
	p3 := &Product{ID: "p3", StoreID: storeID, Barcode: "4006381333931", Name: "Item 3", PricePaise: 300, IsActive: true}

	_ = repo.CreateProduct(ctx, p1) // sync_seq = 2
	_ = repo.CreateProduct(ctx, p2) // sync_seq = 3
	_ = repo.CreateProduct(ctx, p3) // sync_seq = 4

	// Soft-delete p2 -> bumps sync_seq to 5
	_, _ = repo.SoftDeleteProduct(ctx, p2.ID)

	// 1. Delta sync with since_seq = 0 (Full Sync)
	req1 := &CatalogSyncRequest{StoreID: storeID, SinceSeq: 0}
	resp1, err := syncEngine.PerformDeltaSync(ctx, req1)
	if err != nil {
		t.Fatalf("Unexpected error on delta sync: %v", err)
	}

	if len(resp1.Products) != 2 { // p1 and p3 (p2 is soft deleted)
		t.Errorf("Expected 2 active products in delta sync, got %d", len(resp1.Products))
	}
	if len(resp1.DeletedIDs) != 1 || resp1.DeletedIDs[0] != "p2" {
		t.Errorf("Expected p2 in deleted_ids, got %v", resp1.DeletedIDs)
	}
	if resp1.HasMore {
		t.Errorf("Expected has_more to be false")
	}

	// 2. Delta sync with since_seq = latest seq -> should return 0 changed products
	req2 := &CatalogSyncRequest{StoreID: storeID, SinceSeq: resp1.NewMaxSeq}
	resp2, err := syncEngine.PerformDeltaSync(ctx, req2)
	if err != nil {
		t.Fatalf("Unexpected error on delta sync: %v", err)
	}

	if len(resp2.Products) != 0 || len(resp2.DeletedIDs) != 0 {
		t.Errorf("Expected 0 changes for up-to-date sync_seq, got %d products and %d deleted", len(resp2.Products), len(resp2.DeletedIDs))
	}
	if resp2.HasMore {
		t.Errorf("Expected has_more to be false")
	}
}
