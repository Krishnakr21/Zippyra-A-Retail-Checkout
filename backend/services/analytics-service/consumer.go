package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type AnalyticsConsumer struct {
	repo       Repository
	dedupGuard HourlyDedupGuard
}

func NewAnalyticsConsumer(repo Repository, dedupGuard HourlyDedupGuard) *AnalyticsConsumer {
	return &AnalyticsConsumer{
		repo:       repo,
		dedupGuard: dedupGuard,
	}
}

func (c *AnalyticsConsumer) ConsumeOrderCompleted(ctx context.Context, payload []byte) error {
	var msg map[string]interface{}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return err
	}

	orderID, _ := msg["order_id"].(string)
	sessionID, _ := msg["session_id"].(string)
	storeID, _ := msg["store_id"].(string)
	chainID, _ := msg["chain_id"].(string)
	paymentMethod, _ := msg["payment_method"].(string)

	if orderID == "" || storeID == "" {
		return nil
	}
	if chainID == "" {
		chainID = "chain-default-1"
	}

	totalPaise := int64(getFloat(msg, "total_paise"))
	discountPaise := int64(getFloat(msg, "discount_paise"))
	cgstPaise := int64(getFloat(msg, "cgst_paise"))
	sgstPaise := int64(getFloat(msg, "sgst_paise"))
	igstPaise := int64(getFloat(msg, "igst_paise"))

	ts := time.Now().UTC()
	if tsStr, ok := msg["ts"].(string); ok && tsStr != "" {
		if parsed, err := time.Parse(time.RFC3339, tsStr); err == nil {
			ts = parsed
		}
	}

	// 1. Insert into sales_events
	salesEvt := &SalesEvent{
		EventDate:     ts,
		EventTime:     ts,
		StoreID:       storeID,
		ChainID:       chainID,
		OrderID:       orderID,
		TotalPaise:    totalPaise,
		DiscountPaise: discountPaise,
		CGSTPaise:     cgstPaise,
		SGSTPaise:     sgstPaise,
		IGSTPaise:     igstPaise,
		PaymentMethod: paymentMethod,
	}

	var itemEvents []*OrderItemEvent
	if rawItems, ok := msg["items"].([]interface{}); ok {
		salesEvt.ItemCount = uint16(len(rawItems))
		for _, rawItem := range rawItems {
			if itemMap, ok := rawItem.(map[string]interface{}); ok {
				barcode, _ := itemMap["barcode"].(string)
				name, _ := itemMap["name"].(string)
				qty := uint16(getFloat(itemMap, "qty"))
				lineTotal := int64(getFloat(itemMap, "line_total_paise"))

				itemEvents = append(itemEvents, &OrderItemEvent{
					EventDate:      ts,
					OrderID:        orderID,
					StoreID:        storeID,
					ChainID:        chainID,
					Barcode:        barcode,
					ProductName:    name,
					Qty:            qty,
					LineTotalPaise: lineTotal,
				})
			}
		}
	}

	_ = c.repo.InsertSalesEvent(ctx, salesEvt)
	if len(itemEvents) > 0 {
		_ = c.repo.InsertOrderItemEvents(ctx, itemEvents)
	}

	// 2. Increment transaction_hourly protected by Redis dedup guard
	shouldIncrement, _ := c.dedupGuard.ShouldIncrementHourly(ctx, orderID)
	if shouldIncrement {
		_ = c.repo.IncrementTransactionHourly(ctx, storeID, ts)
	} else {
		logger.Warn("[Analytics Redelivery Guard] Duplicate order.completed for order %s skipped in hourly count", orderID)
	}

	// 3. Insert funnel events
	if sessionID != "" {
		_ = c.repo.InsertFunnelEvent(ctx, &FunnelEvent{
			EventDate: ts,
			EventTime: ts,
			StoreID:   storeID,
			SessionID: sessionID,
			Stage:     StagePaymentConfirmed,
		})
		_ = c.repo.InsertFunnelEvent(ctx, &FunnelEvent{
			EventDate: ts,
			EventTime: ts,
			StoreID:   storeID,
			SessionID: sessionID,
			Stage:     StageOrderCompleted,
		})
	} else {
		logger.Warn("[Analytics Funnel] order.completed missing session_id for order %s — sales recorded, funnel stage skipped", orderID)
	}

	return nil
}

func (c *AnalyticsConsumer) ConsumeSessionStarted(ctx context.Context, payload []byte) error {
	var msg map[string]interface{}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return err
	}
	sessionID, _ := msg["session_id"].(string)
	storeID, _ := msg["store_id"].(string)
	if sessionID == "" || storeID == "" {
		return nil
	}

	ts := time.Now().UTC()
	_ = c.repo.InsertFunnelEvent(ctx, &FunnelEvent{
		EventDate: ts,
		EventTime: ts,
		StoreID:   storeID,
		SessionID: sessionID,
		Stage:     StageSessionStarted,
	})
	return nil
}

func (c *AnalyticsConsumer) ConsumeCheckoutInitiated(ctx context.Context, payload []byte) error {
	var msg map[string]interface{}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return err
	}
	sessionID, _ := msg["session_id"].(string)
	storeID, _ := msg["store_id"].(string)
	if sessionID == "" || storeID == "" {
		return nil
	}

	ts := time.Now().UTC()
	_ = c.repo.InsertFunnelEvent(ctx, &FunnelEvent{
		EventDate: ts,
		EventTime: ts,
		StoreID:   storeID,
		SessionID: sessionID,
		Stage:     StageCheckoutInitiated,
	})
	return nil
}

func (c *AnalyticsConsumer) ConsumeExitValidated(ctx context.Context, payload []byte) error {
	var msg map[string]interface{}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return err
	}
	sessionID, _ := msg["session_id"].(string)
	storeID, _ := msg["store_id"].(string)
	if sessionID == "" || storeID == "" {
		return nil
	}

	ts := time.Now().UTC()
	_ = c.repo.InsertFunnelEvent(ctx, &FunnelEvent{
		EventDate: ts,
		EventTime: ts,
		StoreID:   storeID,
		SessionID: sessionID,
		Stage:     StageExitValidated,
	})
	return nil
}

func getFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}
