package main

import (
	"context"
	"encoding/json"

	"github.com/zippyra/backend/shared/logger"
)

type EventConsumer struct {
	engine *MovementEngine
}

func NewEventConsumer(engine *MovementEngine) *EventConsumer {
	return &EventConsumer{engine: engine}
}

func (c *EventConsumer) ProcessOrderCompleted(ctx context.Context, value []byte) error {
	var payload OrderCompletedKafkaPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		logger.Error("Failed to unmarshal order.completed event: %v", err)
		return err
	}

	if payload.OrderID == "" || payload.StoreID == "" || len(payload.Items) == 0 {
		logger.Warn("Received order.completed event with missing order_id, store_id, or items")
		return nil
	}

	noteStr := "Order Sale"
	for _, item := range payload.Items {
		_, newOnHand, err := c.engine.ApplyMovement(
			ctx,
			nil,
			payload.StoreID,
			item.Barcode,
			MovementSale,
			-item.Qty,
			RefOrder,
			payload.OrderID,
			nil,
			&noteStr,
			true, // allowNegative = true for sales! Log warning if negative
		)
		if err != nil {
			logger.Error("Failed to apply SALE movement for order %s barcode %s: %v", payload.OrderID, item.Barcode, err)
		} else if newOnHand < 0 {
			logger.Warn("[NEGATIVE STOCK ALERT] Order %s caused store %s barcode %s on-hand quantity to drop to %d", payload.OrderID, payload.StoreID, item.Barcode, newOnHand)
		}
	}

	logger.Info("Processed inventory movements for order.completed order %s at store %s", payload.OrderID, payload.StoreID)
	return nil
}

func (c *EventConsumer) ProcessOrderReturned(ctx context.Context, value []byte) error {
	var payload OrderReturnedKafkaPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		logger.Error("Failed to unmarshal order.returned event: %v", err)
		return err
	}

	if payload.OrderID == "" || payload.StoreID == "" || len(payload.Items) == 0 {
		logger.Warn("Received order.returned event with missing order_id, store_id, or items")
		return nil
	}

	noteStr := "Order Return Restock"
	for _, item := range payload.Items {
		_, _, err := c.engine.ApplyMovement(
			ctx,
			nil,
			payload.StoreID,
			item.Barcode,
			MovementReturn,
			item.Qty,
			RefOrder,
			payload.OrderID,
			nil,
			&noteStr,
			true,
		)
		if err != nil {
			logger.Error("Failed to apply RETURN movement for order %s barcode %s: %v", payload.OrderID, item.Barcode, err)
		}
	}

	logger.Info("Processed inventory movements for order.returned order %s at store %s", payload.OrderID, payload.StoreID)
	return nil
}
