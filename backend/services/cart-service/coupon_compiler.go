package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type CouponCompiler struct {
	rdb redis.Cmdable
}

func NewCouponCompiler(rdb redis.Cmdable) *CouponCompiler {
	return &CouponCompiler{rdb: rdb}
}

func (c *CouponCompiler) SyncCouponToRedis(ctx context.Context, coupon *Coupon, storeIDs []string) error {
	if c.rdb == nil {
		return nil
	}

	cfg := CouponConfigJSON{
		ID:                 coupon.ID,
		ChainID:            coupon.ChainID,
		StoreID:            coupon.StoreID,
		Code:               coupon.Code,
		DiscountType:       coupon.DiscountType,
		DiscountValue:      coupon.DiscountValue,
		MinCartValuePaise:  coupon.MinCartValuePaise,
		MaxUses:            coupon.MaxUses,
		MaxUsesPerCustomer: coupon.MaxUsesPerCustomer,
		CurrentUseCount:    coupon.CurrentUseCount,
		ActiveFrom:         coupon.ActiveFrom,
		ActiveUntil:        coupon.ActiveUntil,
		IsActive:           coupon.IsActive,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	// Store-specific vs Chain-wide
	if coupon.StoreID != nil && *coupon.StoreID != "" {
		key := fmt.Sprintf("coupon:%s:%s", *coupon.StoreID, coupon.Code)
		if !coupon.IsActive {
			_ = c.rdb.Del(ctx, key)
		} else {
			_ = c.rdb.Set(ctx, key, string(data), 0)
		}
	} else {
		// Chain-wide: Fan out to every store in the chain
		for _, storeID := range storeIDs {
			key := fmt.Sprintf("coupon:%s:%s", storeID, coupon.Code)
			if !coupon.IsActive {
				_ = c.rdb.Del(ctx, key)
			} else {
				_ = c.rdb.Set(ctx, key, string(data), 0)
			}
		}
	}

	return nil
}

func (c *CouponCompiler) DeleteCouponFromRedis(ctx context.Context, storeID, code string, storeIDs []string) error {
	if c.rdb == nil {
		return nil
	}

	if storeID != "" {
		key := fmt.Sprintf("coupon:%s:%s", storeID, code)
		_ = c.rdb.Del(ctx, key)
	} else {
		for _, sID := range storeIDs {
			key := fmt.Sprintf("coupon:%s:%s", sID, code)
			_ = c.rdb.Del(ctx, key)
		}
	}
	return nil
}
