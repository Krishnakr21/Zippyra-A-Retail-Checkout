package main

import (
	"context"
	"log"
	"time"
)

type ExportExpiryJob struct {
	repo     Repository
	interval time.Duration
	stopChan chan struct{}
}

func NewExportExpiryJob(repo Repository, interval time.Duration) *ExportExpiryJob {
	if interval == 0 {
		interval = 1 * time.Hour
	}
	return &ExportExpiryJob{
		repo:     repo,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

func (j *ExportExpiryJob) Start() {
	ticker := time.NewTicker(j.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				j.RunSweep()
			case <-j.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
	log.Printf("[Export Expiry Job] Started background sweeper with interval %v", j.interval)
}

func (j *ExportExpiryJob) Stop() {
	close(j.stopChan)
}

func (j *ExportExpiryJob) RunSweep() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expiredCount, err := j.repo.SweepExpiredAccessExports(ctx)
	if err != nil {
		log.Printf("[Export Expiry Job] Error sweeping expired access exports: %v", err)
		return
	}

	if expiredCount > 0 {
		log.Printf("[Export Expiry Job] Swept and marked %d access exports as EXPIRED", expiredCount)
	}
}
