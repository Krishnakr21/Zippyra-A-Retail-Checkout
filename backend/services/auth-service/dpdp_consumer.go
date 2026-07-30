package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type DPDPDeletionRequestPayload struct {
	UserID        string `json:"user_id"`
	DPDPRequestID string `json:"dpdp_request_id"`
}

type DPDPDeletionCompletedPayload struct {
	UserID           string   `json:"user_id"`
	DPDPRequestID    string   `json:"dpdp_request_id"`
	ServiceName      string   `json:"service_name"`
	TablesAffected   []string `json:"tables_affected"`
	RowsDeletedCount int      `json:"rows_deleted_count"`
}

type AuthDPDPConsumer struct {
	repo     Repository
	producer *kafka.Producer
}

func NewAuthDPDPConsumer(repo Repository, producer *kafka.Producer) *AuthDPDPConsumer {
	return &AuthDPDPConsumer{repo: repo, producer: producer}
}

func (c *AuthDPDPConsumer) HandleUserDataDeletionRequested(ctx context.Context, value []byte) error {
	var req DPDPDeletionRequestPayload
	if err := json.Unmarshal(value, &req); err != nil {
		return fmt.Errorf("failed to unmarshal dpdp.user_data_deletion_requested: %w", err)
	}

	if req.UserID == "" || req.DPDPRequestID == "" {
		logger.Warn("[Auth DPDP Consumer] Received empty user_id or dpdp_request_id")
		return nil
	}

	rowsDeleted, err := c.repo.DeleteUserPII(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete user PII for %s: %w", req.UserID, err)
	}

	completedPayload := DPDPDeletionCompletedPayload{
		UserID:           req.UserID,
		DPDPRequestID:    req.DPDPRequestID,
		ServiceName:      "auth-service",
		TablesAffected:   []string{"users"},
		RowsDeletedCount: rowsDeleted,
	}

	completedBytes, _ := json.Marshal(completedPayload)
	if c.producer != nil {
		if err := c.producer.PublishEvent(ctx, "dpdp.deletion_completed", req.UserID, completedBytes); err != nil {
			logger.Error("[Auth DPDP Consumer] Failed to publish dpdp.deletion_completed: %v", err)
		}
	}

	logger.Info("[Auth DPDP Consumer] Successfully purged PII for user %s (%d rows)", req.UserID, rowsDeleted)
	return nil
}
