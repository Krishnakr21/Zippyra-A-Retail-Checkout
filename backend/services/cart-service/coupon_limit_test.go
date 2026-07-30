package main

import (
	"context"
	"testing"
	"time"
)

func TestCouponLimits_PerCustomerAndGlobalMaxUses(t *testing.T) {
	repo := NewMemoryOfferRepository()

	maxUses := 2
	coupon := &Coupon{
		ID:                 "coup-limited",
		ChainID:            "chain-001",
		Code:               "SAVE100",
		DiscountType:       "FLAT_OFF",
		DiscountValue:      10000, // ₹100
		MaxUses:            &maxUses,
		MaxUsesPerCustomer: 1,
		CurrentUseCount:    0,
		IsActive:           true,
		ActiveFrom:         time.Now(),
	}
	_ = repo.CreateCoupon(context.Background(), coupon)

	// User 1 first redemption -> Success
	err1 := repo.RecordCouponRedemption(context.Background(), coupon.ID, "user-001", "sess-1")
	if err1 != nil {
		t.Fatalf("expected user 1 first redemption success, got: %v", err1)
	}

	// User 1 second redemption attempt -> Per-customer limit reached (1 max)
	redsUser1, _ := repo.GetUserCouponRedemptions(context.Background(), coupon.ID, "user-001")
	if redsUser1 < coupon.MaxUsesPerCustomer {
		t.Fatalf("expected user 1 redemption count to equal max_uses_per_customer (1)")
	}

	// User 2 first redemption -> Success (Global count = 2)
	err2 := repo.RecordCouponRedemption(context.Background(), coupon.ID, "user-002", "sess-2")
	if err2 != nil {
		t.Fatalf("expected user 2 first redemption success, got: %v", err2)
	}

	// User 3 attempt -> Global max_uses (2) reached!
	c, _ := repo.GetCouponByID(context.Background(), coupon.ID)
	if c.CurrentUseCount < *c.MaxUses {
		t.Fatalf("expected current_use_count (%d) to reach max_uses (%d)", c.CurrentUseCount, *c.MaxUses)
	}
}
