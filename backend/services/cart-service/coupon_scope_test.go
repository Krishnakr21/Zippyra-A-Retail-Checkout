package main

import (
	"context"
	"testing"
	"time"
)

func TestCouponScope_ValidationBoundsAndFanOut(t *testing.T) {
	repo := NewMemoryOfferRepository()
	compiler := NewCouponCompiler(nil) // nil = no Redis in unit tests
	adminH := NewCouponAdminHandler(repo, compiler)

	// 1. Percent OFF bounds check (Must be 1–90)
	err1 := adminH.validateCouponRules("PERCENT_OFF", 95.0, 1, nil, nil)
	if err1 == nil {
		t.Fatalf("expected error for PERCENT_OFF > 90, got nil")
	}

	err2 := adminH.validateCouponRules("PERCENT_OFF", 0.0, 1, nil, nil)
	if err2 == nil {
		t.Fatalf("expected error for PERCENT_OFF < 1, got nil")
	}

	err3 := adminH.validateCouponRules("PERCENT_OFF", 20.0, 1, nil, nil)
	if err3 != nil {
		t.Fatalf("expected valid PERCENT_OFF (20%%), got error: %v", err3)
	}

	// 2. FLAT_OFF bounds
	err4 := adminH.validateCouponRules("FLAT_OFF", 0.0, 1, nil, nil)
	if err4 == nil {
		t.Fatalf("expected error for FLAT_OFF <= 0, got nil")
	}

	// 3. active_until before active_from
	tFrom := time.Now()
	tUntil := tFrom.Add(-1 * time.Hour)
	err5 := adminH.validateCouponRules("PERCENT_OFF", 20.0, 1, &tFrom, &tUntil)
	if err5 == nil {
		t.Fatalf("expected error for active_until before active_from, got nil")
	}

	// 4. Chain-wide Fan-Out Test
	repo.SetStoreChain("store-001", "chain-001")
	repo.SetStoreChain("store-002", "chain-001")

	coupon := &Coupon{
		ID:                 "coup-chain-wide",
		ChainID:            "chain-001",
		StoreID:            nil, // Chain-wide
		Code:               "SUPER50",
		DiscountType:       "PERCENT_OFF",
		DiscountValue:      50.0,
		MaxUsesPerCustomer: 1,
		IsActive:           true,
		ActiveFrom:         time.Now(),
	}
	_ = repo.CreateCoupon(context.Background(), coupon)

	// Use MemoryRedis via its redis.Cmdable embedding
	memRedis := NewMemoryRedis()
	fanOutCompiler := NewCouponCompiler(memRedis)

	storeIDs, _ := repo.GetStoreIDsForChain(context.Background(), "chain-001")
	if err := fanOutCompiler.SyncCouponToRedis(context.Background(), coupon, storeIDs); err != nil {
		t.Fatalf("SyncCouponToRedis failed: %v", err)
	}

	// Verify both store-001 and store-002 have Redis keys
	key1 := "coupon:store-001:SUPER50"
	key2 := "coupon:store-002:SUPER50"

	cmd1 := memRedis.Get(context.Background(), key1)
	if cmd1.Val() == "" {
		t.Errorf("expected Redis key %s to be populated for store-001", key1)
	}
	cmd2 := memRedis.Get(context.Background(), key2)
	if cmd2.Val() == "" {
		t.Errorf("expected Redis key %s to be populated for store-002", key2)
	}
}
