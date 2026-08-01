package sync_loop

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/local_status_server"
	"github.com/zippyra/zippyra-connector/internal/logging"
	"github.com/zippyra/zippyra-connector/internal/zippyra_client"
)

type SyncLoop struct {
	client       *zippyra_client.Client
	adapter      erp_adapter.ErpAdapter
	metrics      *local_status_server.StatusMetrics
	pollInterval time.Duration
	logger       *logging.Logger
	lastPollTime time.Time
}

func NewSyncLoop(
	client *zippyra_client.Client,
	adapter erp_adapter.ErpAdapter,
	metrics *local_status_server.StatusMetrics,
	pollIntervalSeconds int,
	logger *logging.Logger,
) *SyncLoop {
	if pollIntervalSeconds <= 0 {
		pollIntervalSeconds = 60
	}
	return &SyncLoop{
		client:       client,
		adapter:      adapter,
		metrics:      metrics,
		pollInterval: time.Duration(pollIntervalSeconds) * time.Second,
		logger:       logger,
		lastPollTime: time.Now().Add(-1 * time.Hour),
	}
}

func (s *SyncLoop) Start(ctx context.Context) {
	s.logger.Info("[SyncLoop] Starting ERP connector sync loop (poll interval: %v)...", s.pollInterval)

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	// Initial immediate tick on startup
	s.RunTick(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("[SyncLoop] Stopping sync loop context canceled.")
			return
		case <-ticker.C:
			s.RunTick(ctx)
		}
	}
}

func (s *SyncLoop) RunTick(ctx context.Context) {
	// PANIC GUARD: Process MUST NOT CRASH under any circumstances
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			err := fmt.Errorf("PANIC RECOVERED in SyncLoop tick: %v\nStack:\n%s", r, stack)
			s.logger.Error("[SyncLoop] CRITICAL: %v", err)
			if s.metrics != nil {
				s.metrics.UpdatePoll(0, err)
			}
		}
	}()

	s.logger.Info("[SyncLoop] Starting sync cycle tick at %s", time.Now().Format(time.RFC3339))

	// 1. INBOUND SYNC: Pull Queue from Zippyra Integration Service
	jobs, err := s.client.PullQueue(ctx)
	if err != nil {
		s.logger.Error("[SyncLoop] PullQueue failed: %v", err)
		if s.metrics != nil {
			s.metrics.UpdatePoll(0, err)
		}
		return
	}

	var succeededJobIDs []string
	var failedCount int

	for _, job := range jobs {
		if err := s.processJob(ctx, job); err != nil {
			failedCount++
			s.logger.Error("[SyncLoop] Failed to process job %s (type=%s): %v. LEAVING UN-ACKNOWLEDGED.", job.JobID, job.SourceEventType, err)
		} else {
			succeededJobIDs = append(succeededJobIDs, job.JobID)
		}
	}

	// Batched Ack call for all successful jobs in this tick cycle
	if len(succeededJobIDs) > 0 {
		if err := s.client.AckQueue(ctx, succeededJobIDs); err != nil {
			s.logger.Error("[SyncLoop] Batched AckQueue failed for %d jobs: %v", len(succeededJobIDs), err)
			if s.metrics != nil {
				s.metrics.UpdatePoll(failedCount, err)
			}
		} else {
			s.logger.Info("[SyncLoop] Batched acked %d jobs successfully", len(succeededJobIDs))
			if s.metrics != nil {
				s.metrics.UpdatePoll(failedCount, nil)
			}
		}
	} else {
		if s.metrics != nil {
			s.metrics.UpdatePoll(failedCount, nil)
		}
	}

	// 2. OUTBOUND SYNC: Poll Local ERP Changes and Push via Webhook
	now := time.Now()
	localChanges, err := s.adapter.PollLocalChanges(ctx, s.lastPollTime)
	if err != nil {
		s.logger.Error("[SyncLoop] PollLocalChanges failed: %v", err)
		if s.metrics != nil {
			s.metrics.UpdatePush(err)
		}
	} else {
		s.lastPollTime = now
		var pushErr error
		for _, change := range localChanges {
			if err := s.client.SendWebhook(ctx, change); err != nil {
				s.logger.Error("[SyncLoop] SendWebhook failed for barcode %s: %v", change.Barcode, err)
				pushErr = err
			}
		}
		if s.metrics != nil {
			s.metrics.UpdatePush(pushErr)
		}
	}
}

func (s *SyncLoop) processJob(ctx context.Context, job zippyra_client.SyncJob) error {
	switch job.SourceEventType {
	case "CATALOG_PRICE_CHANGED":
		barcode, _ := job.Payload["barcode"].(string)
		pricePaise, _ := parseAsInt64(job.Payload["price_paise"])
		if barcode == "" || pricePaise <= 0 {
			return fmt.Errorf("invalid price update payload: barcode=%s pricePaise=%d", barcode, pricePaise)
		}
		return s.adapter.ApplyPriceUpdate(ctx, barcode, pricePaise)

	case "STOCK_ADJUSTED":
		barcode, _ := job.Payload["barcode"].(string)
		qtyDelta, _ := parseAsInt64(job.Payload["qty_delta"])
		reason, _ := job.Payload["reason"].(string)
		if barcode == "" {
			return fmt.Errorf("invalid stock adjustment payload: barcode=%s", barcode)
		}
		return s.adapter.ApplyStockAdjustment(ctx, barcode, qtyDelta, reason)

	case "GRN_CREATED":
		itemsRaw, ok := job.Payload["items"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid GRN payload: missing items array")
		}

		var items []erp_adapter.GrnItem
		for _, itemRaw := range itemsRaw {
			if itemMap, ok := itemRaw.(map[string]interface{}); ok {
				bc, _ := itemMap["barcode"].(string)
				qty, _ := parseAsInt64(itemMap["qty"])
				cost, _ := parseAsInt64(itemMap["cost_paise"])
				items = append(items, erp_adapter.GrnItem{
					Barcode:   bc,
					Qty:       qty,
					CostPaise: cost,
				})
			}
		}
		return s.adapter.ApplyGrn(ctx, items)

	default:
		return fmt.Errorf("unsupported source event type: %s", job.SourceEventType)
	}
}

func parseAsInt64(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case json.Number:
		i, err := v.Int64()
		return i, err == nil
	default:
		return 0, false
	}
}
