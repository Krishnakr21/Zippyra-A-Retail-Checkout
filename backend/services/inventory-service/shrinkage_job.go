package main

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type ShrinkageRollupJob struct {
	db       *sql.DB
	repo     Repository
	producer *kafka.Producer
}

func NewShrinkageRollupJob(db *sql.DB, repo Repository, producer *kafka.Producer) *ShrinkageRollupJob {
	return &ShrinkageRollupJob{
		db:       db,
		repo:     repo,
		producer: producer,
	}
}

func (j *ShrinkageRollupJob) Run(ctx context.Context, targetDate string) error {
	if targetDate == "" {
		// Default to yesterday
		targetDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	logger.Info("[SHRINKAGE ROLLUP JOB] Starting rollup for date %s...", targetDate)

	query := `
		SELECT store_id, SUM(variance_qty), SUM(expected_qty), COUNT(DISTINCT barcode)
		FROM stock_counts
		WHERE date(counted_at) = $1
		GROUP BY store_id
	`
	rows, err := j.db.QueryContext(ctx, query, targetDate)
	if err != nil {
		return fmt.Errorf("failed to query stock counts for date %s: %w", targetDate, err)
	}
	defer rows.Close()

	type storeRollup struct {
		storeID       string
		totalVariance int64
		totalExpected int64
		itemCount     int
	}
	var rollups []storeRollup

	for rows.Next() {
		var r storeRollup
		if err := rows.Scan(&r.storeID, &r.totalVariance, &r.totalExpected, &r.itemCount); err != nil {
			return fmt.Errorf("failed to scan stock counts rollup: %w", err)
		}
		rollups = append(rollups, r)
	}

	for _, r := range rollups {
		shrinkagePercent := 0.0
		if r.totalExpected > 0 {
			absVariance := math.Abs(float64(r.totalVariance))
			shrinkagePercent = (absVariance / float64(r.totalExpected)) * 100.0
		}

		err := j.repo.UpsertShrinkageDaily(ctx, r.storeID, targetDate, r.totalVariance, r.totalExpected, shrinkagePercent, r.itemCount)
		if err != nil {
			logger.Error("[SHRINKAGE ROLLUP JOB] Failed to upsert shrinkage for store %s: %v", r.storeID, err)
			continue
		}

		logger.Info("[SHRINKAGE ROLLUP JOB] Store %s on %s: shrinkage = %.2f%% (variance: %d, expected: %d)", r.storeID, targetDate, shrinkagePercent, r.totalVariance, r.totalExpected)

		// Alert threshold > 0.5%
		if shrinkagePercent > 0.5 {
			alertMsg := ShrinkageAlertPayload{
				StoreID:          r.storeID,
				Date:             targetDate,
				ShrinkagePercent: shrinkagePercent,
				Timestamp:        time.Now(),
			}
			_ = j.producer.PublishEvent(ctx, TopicShrinkageAlert, r.storeID, alertMsg)
			logger.Warn("[SHRINKAGE ALERT] Store %s exceeded 0.5%% shrinkage threshold on %s (%.2f%%)", r.storeID, targetDate, shrinkagePercent)
		}
	}

	logger.Info("[SHRINKAGE ROLLUP JOB] Rollup completed successfully for date %s.", targetDate)
	return nil
}
