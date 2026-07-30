package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

type CouponReconciliationJob struct {
	repo     OfferRepository
	compiler *CouponCompiler
}

func NewCouponReconciliationJob(repo OfferRepository, compiler *CouponCompiler) *CouponReconciliationJob {
	return &CouponReconciliationJob{
		repo:     repo,
		compiler: compiler,
	}
}

func (j *CouponReconciliationJob) RunReconciliationSweep(ctx context.Context) error {
	coupons, err := j.repo.ListActiveCoupons(ctx)
	if err != nil {
		return fmt.Errorf("failed to list active coupons: %w", err)
	}

	reconciledCount := 0
	for _, c := range coupons {
		var storeIDs []string
		if c.StoreID == nil || *c.StoreID == "" {
			storeIDs, _ = j.repo.GetStoreIDsForChain(ctx, c.ChainID)
		}
		if err := j.compiler.SyncCouponToRedis(ctx, c, storeIDs); err == nil {
			reconciledCount++
		}
	}

	log.Printf("[CouponReconciliationJob] Hourly sweep completed: %d active coupons reconciled to Redis", reconciledCount)
	return nil
}

func (j *CouponReconciliationJob) StartHourlyLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = j.RunReconciliationSweep(ctx)
		}
	}
}
