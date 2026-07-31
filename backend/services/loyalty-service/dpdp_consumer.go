package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type DPDPLoyaltyDeletionRequestPayload struct {
	UserID        string `json:"user_id"`
	DPDPRequestID string `json:"dpdp_request_id"`
}

type DPDPLoyaltyDeletionCompletedPayload struct {
	UserID           string   `json:"user_id"`
	DPDPRequestID    string   `json:"dpdp_request_id"`
	ServiceName      string   `json:"service_name"`
	TablesAffected   []string `json:"tables_affected"`
	RowsDeletedCount int      `json:"rows_deleted_count"`
}

type LoyaltyDPDPConsumer struct {
	repo     Repository
	producer *kafka.Producer
}

func NewLoyaltyDPDPConsumer(repo Repository, producer *kafka.Producer) *LoyaltyDPDPConsumer {
	return &LoyaltyDPDPConsumer{repo: repo, producer: producer}
}

func (c *LoyaltyDPDPConsumer) HandleUserDataDeletionRequested(ctx context.Context, value []byte) error {
	var req DPDPLoyaltyDeletionRequestPayload
	if err := json.Unmarshal(value, &req); err != nil {
		return fmt.Errorf("failed to unmarshal dpdp.user_data_deletion_requested: %w", err)
	}

	if req.UserID == "" || req.DPDPRequestID == "" {
		logger.Warn("[Loyalty DPDP Consumer] Received empty user_id or dpdp_request_id")
		return nil
	}

	rowsPurged, err := c.repo.DeleteUserLoyaltyData(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete user loyalty data for %s: %w", req.UserID, err)
	}

	completedPayload := DPDPLoyaltyDeletionCompletedPayload{
		UserID:           req.UserID,
		DPDPRequestID:    req.DPDPRequestID,
		ServiceName:      "loyalty-service",
		TablesAffected:   []string{"loyalty_accounts", "loyalty_ledger"},
		RowsDeletedCount: rowsPurged,
	}

	completedBytes, _ := json.Marshal(completedPayload)
	if c.producer != nil {
		if err := c.producer.PublishEvent(ctx, "dpdp.deletion_completed", req.UserID, completedBytes); err != nil {
			logger.Error("[Loyalty DPDP Consumer] Failed to publish dpdp.deletion_completed: %v", err)
		}
	}

	logger.Info("[Loyalty DPDP Consumer] Successfully purged loyalty data for user %s (%d rows)", req.UserID, rowsPurged)
	return nil
}
