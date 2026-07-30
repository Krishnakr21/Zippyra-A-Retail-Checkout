package main

import (
	"context"
	"log"
	"time"
)

type OfferScheduleJob struct {
	compiler *OfferCompiler
	repo     OfferRepository
}

func NewOfferScheduleJob(compiler *OfferCompiler, repo OfferRepository) *OfferScheduleJob {
	return &OfferScheduleJob{
		compiler: compiler,
		repo:     repo,
	}
}

func (j *OfferScheduleJob) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[INFO] [OfferScheduleJob] Stopping background schedule worker...")
			return
		case <-ticker.C:
			j.RunOnce(ctx)
		}
	}
}

func (j *OfferScheduleJob) RunOnce(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Println("[INFO] [OfferScheduleJob] Running 5-minute schedule boundary check...")

	// In real DB environment or mock, find stores with recently started/expired offers
	// For all active stores in system, recompile if active schedule boundaries crossed
	stores, err := j.repo.ListStoresForChain(ctx, "chain-default-001")
	if err != nil {
		log.Printf("[WARN] [OfferScheduleJob] Failed to list stores for schedule check: %v", err)
		return
	}

	for _, storeID := range stores {
		if err := j.compiler.CompileAndPublish(ctx, storeID); err != nil {
			log.Printf("[WARN] [OfferScheduleJob] Failed to recompile store %s: %v", storeID, err)
		}
	}
}
