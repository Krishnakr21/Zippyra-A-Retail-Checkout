package main

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type OutboxRelay struct {
	db                  *sql.DB
	producer            kafka.EventPublisher
	pollInterval        time.Duration
	maxRetries          int
	mu                  sync.RWMutex
	lastSuccessfulPoll time.Time
	stopChan            chan struct{}
	wg                  sync.WaitGroup
}

func NewOutboxRelay(db *sql.DB, producer kafka.EventPublisher, pollInterval time.Duration) *OutboxRelay {
	if pollInterval <= 0 {
		pollInterval = 200 * time.Millisecond
	}
	return &OutboxRelay{
		db:                  db,
		producer:            producer,
		pollInterval:        pollInterval,
		maxRetries:          20,
		lastSuccessfulPoll: time.Now(),
		stopChan:            make(chan struct{}),
	}
}

func (r *OutboxRelay) Start() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(r.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-r.stopChan:
				return
			case <-ticker.C:
				r.processBatch()
			}
		}
	}()
}

func (r *OutboxRelay) Stop() {
	close(r.stopChan)
	r.wg.Wait()
}

func (r *OutboxRelay) LastPollTime() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastSuccessfulPoll
}

func (r *OutboxRelay) processBatch() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to start outbox transaction: %v", err)
		return
	}
	defer tx.Rollback()

	query := `
		SELECT id, topic, payload, retry_count
		FROM payment_outbox
		WHERE published_at IS NULL AND retry_count < $1
		ORDER BY created_at ASC
		LIMIT 100
		FOR UPDATE SKIP LOCKED
	`
	rows, err := tx.QueryContext(ctx, query, r.maxRetries)
	if err != nil {
		// Fallback without FOR UPDATE SKIP LOCKED for SQLite in unit tests
		fallbackQuery := `
			SELECT id, topic, payload, retry_count
			FROM payment_outbox
			WHERE published_at IS NULL AND retry_count < $1
			ORDER BY created_at ASC
			LIMIT 100
		`
		rows, err = tx.QueryContext(ctx, fallbackQuery, r.maxRetries)
		if err != nil {
			logger.Error("Outbox poll query failed: %v", err)
			return
		}
	}

	type pendingEvent struct {
		id         string
		topic      string
		payload    []byte
		retryCount int
	}

	var events []pendingEvent
	for rows.Next() {
		var ev pendingEvent
		if err := rows.Scan(&ev.id, &ev.topic, &ev.payload, &ev.retryCount); err != nil {
			logger.Error("Failed to scan outbox row: %v", err)
			rows.Close()
			return
		}
		events = append(events, ev)
	}
	rows.Close()

	for _, ev := range events {
		// Publish event to Kafka
		pubErr := r.producer.PublishEvent(ctx, ev.topic, ev.id, ev.payload)
		if pubErr == nil {
			// Success: update published_at
			_, updateErr := tx.ExecContext(ctx, `UPDATE payment_outbox SET published_at = CURRENT_TIMESTAMP WHERE id = $1`, ev.id)
			if updateErr != nil {
				logger.Error("Failed to mark outbox event %s published: %v", ev.id, updateErr)
			}
		} else {
			// Failure: increment retry_count
			newRetry := ev.retryCount + 1
			logger.Warn("Kafka publish failed for outbox event %s (retry %d/%d): %v", ev.id, newRetry, r.maxRetries, pubErr)
			_, updateErr := tx.ExecContext(ctx, `UPDATE payment_outbox SET retry_count = $1 WHERE id = $2`, newRetry, ev.id)
			if updateErr != nil {
				logger.Error("Failed to increment retry count for %s: %v", ev.id, updateErr)
			}

			if newRetry >= r.maxRetries {
				logger.Error("CRITICAL: Outbox event %s reached max retries (%d). Alerting ops!", ev.id, r.maxRetries)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit outbox relay transaction: %v", err)
		return
	}

	r.mu.Lock()
	r.lastSuccessfulPoll = time.Now()
	r.mu.Unlock()
}
