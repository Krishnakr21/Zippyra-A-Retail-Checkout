package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zippyra/backend/shared/featureflags"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type EventConsumer struct {
	repo     Repository
	producer *kafka.Producer
}

func NewEventConsumer(repo Repository, producer *kafka.Producer) *EventConsumer {
	return &EventConsumer{
		repo:     repo,
		producer: producer,
	}
}

func (c *EventConsumer) ProcessLowStockEvent(ctx context.Context, value []byte) error {
	var payload LowStockKafkaPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		logger.Error("Failed to unmarshal inventory.low_stock event: %v", err)
		return err
	}

	if payload.StoreID == "" || payload.Barcode == "" {
		logger.Warn("Received inventory.low_stock event with missing store_id or barcode")
		return nil
	}

	// Feature Flag Check: auto_reorder per store
	if !featureflags.IsEnabled(ctx, nil, nil, "auto_reorder", payload.StoreID) {
		logger.Info("[AUTO-REORDER SKIPPED] auto_reorder feature flag disabled for store %s", payload.StoreID)
		return nil
	}

	chainID := "chain-default-1"

	reorderQty := payload.ReorderQty
	if reorderQty <= 0 {
		reorderQty = 50
	}

	// Idempotency: CreateAutoPO enforces UNIQUE constraint on (store_id, auto_reorder_item_barcode, auto_reorder_date)
	po, err := c.repo.CreateAutoPO(ctx, payload.StoreID, chainID, payload.Barcode, reorderQty)
	if err != nil {
		logger.Error("Failed to create auto PO for store %s barcode %s: %v", payload.StoreID, payload.Barcode, err)
		return err
	}

	if po == nil {
		logger.Info("[AUTO-REORDER DUP SKIPPED] Auto PO already created today for store %s barcode %s", payload.StoreID, payload.Barcode)
		return nil
	}

	// Publish warehouse.po_auto_created event
	autoEvent := POAutoCreatedPayload{
		POID:      po.ID,
		StoreID:   payload.StoreID,
		Barcode:   payload.Barcode,
		Qty:       reorderQty,
		Timestamp: time.Now(),
	}
	_ = c.producer.PublishEvent(ctx, TopicPOAutoCreated, po.ID, autoEvent)

	logger.Info("[AUTO-REORDER CREATED] Auto PO %s created for store %s barcode %s qty %d", po.ID, payload.StoreID, payload.Barcode, reorderQty)
	return nil
}
