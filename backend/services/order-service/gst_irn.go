package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type IRNIssuedPayload struct {
	OrderID      string     `json:"order_id"`
	IRN          string     `json:"irn"`
	AckNo        string     `json:"ack_no"`
	AckDate      *time.Time `json:"ack_date,omitempty"`
	SignedQRCode string     `json:"signed_qr_code"`
}

type IRNFailedPayload struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

type GSTIRNConsumer struct {
	repo       Repository
	invoiceSvc InvoiceService
}

func NewGSTIRNConsumer(repo Repository, invoiceSvc InvoiceService) *GSTIRNConsumer {
	return &GSTIRNConsumer{repo: repo, invoiceSvc: invoiceSvc}
}

func (c *GSTIRNConsumer) HandleIRNIssued(ctx context.Context, value []byte) error {
	var payload IRNIssuedPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal compliance.irn_issued payload: %w", err)
	}

	if payload.OrderID == "" || payload.IRN == "" {
		logger.Warn("[GST IRN Consumer] Invalid irn_issued payload missing order_id or irn")
		return nil
	}

	ackDate := payload.AckDate
	if ackDate == nil {
		now := time.Now().UTC()
		ackDate = &now
	}

	err := c.repo.UpdateOrderInvoiceAndIRN(
		ctx,
		payload.OrderID,
		nil,
		&payload.IRN,
		&payload.AckNo,
		ackDate,
		&payload.SignedQRCode,
	)
	if err != nil {
		return fmt.Errorf("failed to update order IRN fields for order %s: %w", payload.OrderID, err)
	}

	if c.invoiceSvc != nil {
		if err := c.invoiceSvc.RegenerateInvoiceWithIRN(ctx, payload.OrderID, payload.IRN, payload.AckNo, *ackDate, payload.SignedQRCode); err != nil {
			logger.Error("[GST IRN Consumer] Failed Phase-2 PDF regeneration for order %s: %v", payload.OrderID, err)
		}
	}

	logger.Info("[GST IRN Consumer] Successfully updated IRN and regenerated Phase-2 PDF for order %s", payload.OrderID)
	return nil
}

func (c *GSTIRNConsumer) HandleIRNFailed(ctx context.Context, value []byte) error {
	var payload IRNFailedPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal compliance.irn_failed payload: %w", err)
	}

	logger.Warn("[GST IRN Consumer] IRN issuance failed for order %s: %s", payload.OrderID, payload.Reason)
	// Idempotent no-op on orders table: IRN fields remain NULL
	return nil
}
