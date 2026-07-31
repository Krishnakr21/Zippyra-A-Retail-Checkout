package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type PaymentConfirmedVelocityPayload struct {
	PaymentID   string    `json:"payment_id"`
	StoreID     string    `json:"store_id"`
	AmountPaise int64     `json:"amount_paise"`
	Timestamp   time.Time `json:"ts"`
}

type PaymentRefundVelocityPayload struct {
	PaymentID string    `json:"payment_id"`
	StoreID   string    `json:"store_id"`
	RefundID  string    `json:"refund_id"`
	Timestamp time.Time `json:"ts"`
}

type VelocityMonitor struct {
	repo Repository
}

func NewVelocityMonitor(repo Repository) *VelocityMonitor {
	return &VelocityMonitor{repo: repo}
}

func (v *VelocityMonitor) HandlePaymentConfirmed(ctx context.Context, value []byte) error {
	var payload PaymentConfirmedVelocityPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payment.confirmed for velocity check: %w", err)
	}

	if payload.StoreID == "" {
		return nil
	}

	// Rule 1: High single transaction value (> ₹2,000,000 = 200,000,000 paise)
	if payload.AmountPaise > 200000000 {
		detailJSON, _ := json.Marshal(map[string]interface{}{
			"payment_id":   payload.PaymentID,
			"amount_paise": payload.AmountPaise,
			"threshold":    200000000,
		})
		alert := &VelocityAlert{
			StoreID:   payload.StoreID,
			AlertType: "UNUSUAL_TRANSACTION_VALUE",
			Detail:    string(detailJSON),
		}
		_ = v.repo.CreateVelocityAlert(ctx, alert)
		logger.Warn("[Velocity Monitor] Alert UNUSUAL_TRANSACTION_VALUE triggered for store %s, payment %s", payload.StoreID, payload.PaymentID)
	}

	return nil
}

func (v *VelocityMonitor) HandleRefundInitiated(ctx context.Context, value []byte) error {
	var payload PaymentRefundVelocityPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payment.refund_initiated for velocity check: %w", err)
	}

	if payload.StoreID == "" {
		return nil
	}

	// Rule 2: Rapid refunds alert
	detailJSON, _ := json.Marshal(map[string]interface{}{
		"refund_id":  payload.RefundID,
		"payment_id": payload.PaymentID,
	})
	alert := &VelocityAlert{
		StoreID:   payload.StoreID,
		AlertType: "RAPID_REFUNDS",
		Detail:    string(detailJSON),
	}
	_ = v.repo.CreateVelocityAlert(ctx, alert)
	logger.Warn("[Velocity Monitor] Alert RAPID_REFUNDS logged for store %s, refund %s", payload.StoreID, payload.RefundID)
	return nil
}
