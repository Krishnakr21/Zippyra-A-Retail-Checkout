package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type OrderCompletedEventPayload struct {
	OrderID    string                   `json:"order_id"`
	StoreID    string                   `json:"store_id"`
	ChainID    string                   `json:"chain_id"`
	TotalPaise int64                    `json:"total_paise"`
	CGSTPaise  int64                    `json:"cgst_paise"`
	SGSTPaise  int64                    `json:"sgst_paise"`
	IGSTPaise  int64                    `json:"igst_paise"`
	Items      []map[string]interface{} `json:"items"`
	Timestamp  time.Time                `json:"ts"`
}

type IRNConsumer struct {
	repo      Repository
	irpClient IRPClient
}

func NewIRNConsumer(repo Repository, irpClient IRPClient) *IRNConsumer {
	return &IRNConsumer{repo: repo, irpClient: irpClient}
}

func (c *IRNConsumer) HandleOrderCompleted(ctx context.Context, value []byte) error {
	var payload OrderCompletedEventPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal order.completed payload: %w", err)
	}

	if payload.OrderID == "" || payload.StoreID == "" {
		logger.Warn("[Compliance IRN Consumer] Empty order_id or store_id in order.completed payload")
		return nil
	}

	// 1. Build canonical payload
	irpPayloadMap := BuildCanonicalIRPPayload(
		payload.OrderID,
		payload.StoreID,
		"29ABCDE1234F1ZW", // store GSTIN default fallback
		payload.Items,
		payload.TotalPaise,
		payload.CGSTPaise,
		payload.SGSTPaise,
		payload.IGSTPaise,
	)

	payloadBytes, _ := json.Marshal(irpPayloadMap)

	rec := &IRNRecord{
		OrderID:    payload.OrderID,
		StoreID:    payload.StoreID,
		ChainID:    payload.ChainID,
		Status:     IRNStatusPending,
		IRPPayload: string(payloadBytes),
		RetryCount: 0,
	}

	// 2. Insert IRN record (Idempotent ON CONFLICT DO NOTHING)
	inserted, err := c.repo.CreateIRNRecord(ctx, rec)
	if err != nil {
		return fmt.Errorf("failed to create irn_records entry: %w", err)
	}
	if !inserted {
		logger.Info("[Compliance IRN Consumer] Order %s already has an IRN record lineage. Idempotent skip.", payload.OrderID)
		return nil
	}

	// 3. Submit to IRP (15s timeout)
	irpCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	irpResp, err := c.irpClient.SubmitEInvoice(irpCtx, irpPayloadMap)
	if err != nil || irpResp == nil || irpResp.Status != "SUCCESS" {
		failReason := "IRP submission failed"
		if err != nil {
			failReason = err.Error()
		} else if irpResp != nil && irpResp.ErrorDetails != "" {
			failReason = irpResp.ErrorDetails
		}

		logger.Warn("[Compliance IRN Consumer] IRN generation failed for order %s: %s", payload.OrderID, failReason)
		_ = c.repo.UpdateIRNStatus(ctx, rec.ID, IRNStatusFailed, nil, nil, nil, nil, nil, &failReason)

		// Write failure outbox event
		failedOutbox, _ := json.Marshal(map[string]interface{}{
			"order_id": payload.OrderID,
			"reason":   failReason,
		})
		_ = c.repo.InsertIRNOutbox(ctx, "compliance.irn_failed", failedOutbox)
		return nil
	}

	// 4. Success: update record and write outbox event
	ackDateParsed, _ := time.Parse("2006-01-02 15:04:05", irpResp.AckDate)
	if ackDateParsed.IsZero() {
		ackDateParsed = time.Now().UTC()
	}

	irpRespBytes, _ := json.Marshal(irpResp)
	irpRespStr := string(irpRespBytes)

	_ = c.repo.UpdateIRNStatus(
		ctx,
		rec.ID,
		IRNStatusIssued,
		&irpResp.IRN,
		&irpResp.AckNo,
		&ackDateParsed,
		&irpResp.SignedQRCode,
		&irpRespStr,
		nil,
	)

	issuedOutbox, _ := json.Marshal(map[string]interface{}{
		"order_id":       payload.OrderID,
		"irn":            irpResp.IRN,
		"ack_no":         irpResp.AckNo,
		"ack_date":       ackDateParsed,
		"signed_qr_code": irpResp.SignedQRCode,
	})
	_ = c.repo.InsertIRNOutbox(ctx, "compliance.irn_issued", issuedOutbox)

	logger.Info("[Compliance IRN Consumer] Successfully issued IRN for order %s: %s", payload.OrderID, irpResp.IRN)
	return nil
}
