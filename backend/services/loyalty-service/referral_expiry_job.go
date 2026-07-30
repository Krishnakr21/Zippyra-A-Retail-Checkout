package main

import (
	"context"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type ReferralExpiryJob struct {
	repo Repository
}

func NewReferralExpiryJob(repo Repository) *ReferralExpiryJob {
	return &ReferralExpiryJob{repo: repo}
}

func (j *ReferralExpiryJob) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial execution on startup
	j.runOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping Referral Expiry Job...")
			return
		case <-ticker.C:
			j.runOnce(ctx)
		}
	}
}

func (j *ReferralExpiryJob) runOnce(ctx context.Context) {
	expiredCount, err := j.repo.ExpirePendingReferrals(ctx)
	if err != nil {
		logger.Error("Referral Expiry Job failed: %v", err)
	} else if expiredCount > 0 {
		logger.Info("Referral Expiry Job: expired %d pending referral events", expiredCount)
	}
}
