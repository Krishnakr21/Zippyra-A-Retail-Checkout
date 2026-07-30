package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type EventProcessor struct {
	cartStore   CartStore
	holdManager HoldManager
	producer    *kafka.Producer
}

func NewEventProcessor(cartStore CartStore, holdManager HoldManager, producer *kafka.Producer) *EventProcessor {
	return &EventProcessor{
		cartStore:   cartStore,
		holdManager: holdManager,
		producer:    producer,
	}
}

func (ep *EventProcessor) ProcessInventoryStockUpdated(ctx context.Context, value []byte) error {
	var evt struct {
		StoreID      string `json:"store_id"`
		Barcode      string `json:"barcode"`
		AvailableQty int    `json:"available_qty"`
	}

	if err := json.Unmarshal(value, &evt); err != nil {
		return err
	}

	if evt.StoreID != "" && evt.Barcode != "" {
		if err := ep.holdManager.SetAvailableQty(ctx, evt.StoreID, evt.Barcode, evt.AvailableQty); err != nil {
			logger.Error("Failed to update available_qty in Redis for %s/%s: %v", evt.StoreID, evt.Barcode, err)
			return err
		}
		logger.Info("Updated available_qty for %s/%s to %d from Kafka", evt.StoreID, evt.Barcode, evt.AvailableQty)
	}

	return nil
}

func (ep *EventProcessor) ProcessStoreSessionEnded(ctx context.Context, value []byte) error {
	var evt struct {
		UserID  string `json:"user_id"`
		StoreID string `json:"store_id"`
	}

	if err := json.Unmarshal(value, &evt); err != nil {
		return err
	}

	if evt.UserID != "" && evt.StoreID != "" {
		items, _, _ := ep.cartStore.GetCart(ctx, evt.StoreID, evt.UserID)
		_ = ep.holdManager.ReleaseAllUserHolds(ctx, evt.StoreID, evt.UserID, items)
		_ = ep.cartStore.ClearCart(ctx, evt.StoreID, evt.UserID)
		logger.Info("Cleared cart & released holds for user %s at store %s following session end event", evt.UserID, evt.StoreID)
	}

	return nil
}

func (ep *EventProcessor) PublishCheckoutInitiated(ctx context.Context, session *CheckoutSession) {
	if ep.producer == nil {
		return
	}

	payload := map[string]interface{}{
		"checkout_session_id": session.ID,
		"session_id":          session.SessionID,
		"user_id":             session.UserID,
		"store_id":            session.StoreID,
		"total_paise":         session.TotalPaise,
		"ts":                  time.Now().Unix(),
	}

	_ = ep.producer.PublishEvent(ctx, "cart.checkout_initiated", session.StoreID, payload)
}
