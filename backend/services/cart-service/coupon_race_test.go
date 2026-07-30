package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCouponRaceCondition_ConcurrentRedemptionsAtMaxUsesBoundary(t *testing.T) {
	repo := NewMemoryOfferRepository()

	maxUses := 5
	coupon := &Coupon{
		ID:                 "coup-race-test",
		ChainID:            "chain-001",
		Code:               "FLASH50",
		DiscountType:       "PERCENT_OFF",
		DiscountValue:      50.0,
		MaxUses:            &maxUses,
		MaxUsesPerCustomer: 1,
		CurrentUseCount:    0,
		IsActive:           true,
		ActiveFrom:         time.Now(),
	}
	_ = repo.CreateCoupon(context.Background(), coupon)

	// Simulate 10 concurrent customer redemption attempts for a coupon capped at max_uses=5
	var wg sync.WaitGroup
	var mu sync.Mutex
	numGoroutines := 10
	successCount := 0
	rejectedCount := 0

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := fmt.Sprintf("user-race-%d", idx)
			sessionID := fmt.Sprintf("sess-race-%d", idx)

			// Pre-check limits under lock simulation
			mu.Lock()
			c, _ := repo.GetCouponByID(context.Background(), coupon.ID)
			if c.CurrentUseCount < *c.MaxUses {
				_ = repo.RecordCouponRedemption(context.Background(), coupon.ID, userID, sessionID)
				successCount++
			} else {
				rejectedCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if successCount != 5 {
		t.Fatalf("expected exactly 5 successful redemptions, got %d", successCount)
	}
	if rejectedCount != 5 {
		t.Fatalf("expected 5 rejected redemptions, got %d", rejectedCount)
	}

	finalCoupon, _ := repo.GetCouponByID(context.Background(), coupon.ID)
	if finalCoupon.CurrentUseCount != 5 {
		t.Errorf("expected final current_use_count to be 5, got %d", finalCoupon.CurrentUseCount)
	}
}
