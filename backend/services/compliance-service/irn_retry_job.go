package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type IRNRetryJob struct {
	repo      Repository
	irpClient IRPClient
}

func NewIRNRetryJob(repo Repository, irpClient IRPClient) *IRNRetryJob {
	return &IRNRetryJob{repo: repo, irpClient: irpClient}
}

func (j *IRNRetryJob) RunOnce(ctx context.Context) {
	failedRecords, err := j.repo.GetFailedIRNRecordsForRetry(ctx, 3)
	if err != nil {
		logger.Error("[IRN Retry Job] Failed to query failed IRN records: %v", err)
		return
	}

	if len(failedRecords) == 0 {
		return
	}

	logger.Info("[IRN Retry Job] Retrying IRN issuance for %d failed records...", len(failedRecords))

	for _, rec := range failedRecords {
		var payloadMap map[string]interface{}
		if err := json.Unmarshal([]byte(rec.IRPPayload), &payloadMap); err != nil {
			logger.Error("[IRN Retry Job] Invalid payload for record %s: %v", rec.ID, err)
			continue
		}

		irpCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		irpResp, err := j.irpClient.SubmitEInvoice(irpCtx, payloadMap)
		cancel()

		if err != nil || irpResp == nil || irpResp.Status != "SUCCESS" {
			failReason := "Retry failed"
			if err != nil {
				failReason = err.Error()
			} else if irpResp != nil && irpResp.ErrorDetails != "" {
				failReason = irpResp.ErrorDetails
			}
			_ = j.repo.IncrementIRNRetry(ctx, rec.ID, failReason)
			logger.Warn("[IRN Retry Job] Retry failed for order %s (attempt %d/3): %s", rec.OrderID, rec.RetryCount+1, failReason)
			continue
		}

		// Success
		ackDateParsed, _ := time.Parse("2006-01-02 15:04:05", irpResp.AckDate)
		if ackDateParsed.IsZero() {
			ackDateParsed = time.Now().UTC()
		}
		irpRespBytes, _ := json.Marshal(irpResp)
		irpRespStr := string(irpRespBytes)

		_ = j.repo.UpdateIRNStatus(
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
			"order_id":       rec.OrderID,
			"irn":            irpResp.IRN,
			"ack_no":         irpResp.AckNo,
			"ack_date":       ackDateParsed,
			"signed_qr_code": irpResp.SignedQRCode,
		})
		_ = j.repo.InsertIRNOutbox(ctx, "compliance.irn_issued", issuedOutbox)
		logger.Info("[IRN Retry Job] Successfully issued IRN on retry for order %s: %s", rec.OrderID, irpResp.IRN)
	}
}

func (j *IRNRetryJob) StartWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.RunOnce(ctx)
		}
	}
}
