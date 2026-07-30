package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type DPDPOrderDeletionRequestPayload struct {
	UserID        string `json:"user_id"`
	DPDPRequestID string `json:"dpdp_request_id"`
}

type DPDPOrderDeletionCompletedPayload struct {
	UserID           string   `json:"user_id"`
	DPDPRequestID    string   `json:"dpdp_request_id"`
	ServiceName      string   `json:"service_name"`
	TablesAffected   []string `json:"tables_affected"`
	RowsDeletedCount int      `json:"rows_deleted_count"`
}

type OrderDPDPConsumer struct {
	repo     Repository
	producer *kafka.Producer
}

func NewOrderDPDPConsumer(repo Repository, producer *kafka.Producer) *OrderDPDPConsumer {
	return &OrderDPDPConsumer{repo: repo, producer: producer}
}

func (c *OrderDPDPConsumer) HandleUserDataDeletionRequested(ctx context.Context, value []byte) error {
	var req DPDPOrderDeletionRequestPayload
	if err := json.Unmarshal(value, &req); err != nil {
		return fmt.Errorf("failed to unmarshal dpdp.user_data_deletion_requested: %w", err)
	}

	if req.UserID == "" || req.DPDPRequestID == "" {
		logger.Warn("[Order DPDP Consumer] Received empty user_id or dpdp_request_id")
		return nil
	}

	rowsAnonymized, err := c.repo.AnonymizeUserOrders(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to anonymize user orders for %s: %w", req.UserID, err)
	}

	completedPayload := DPDPOrderDeletionCompletedPayload{
		UserID:           req.UserID,
		DPDPRequestID:    req.DPDPRequestID,
		ServiceName:      "order-service",
		TablesAffected:   []string{"orders"},
		RowsDeletedCount: rowsAnonymized,
	}

	completedBytes, _ := json.Marshal(completedPayload)
	if c.producer != nil {
		if err := c.producer.PublishEvent(ctx, "dpdp.deletion_completed", req.UserID, completedBytes); err != nil {
			logger.Error("[Order DPDP Consumer] Failed to publish dpdp.deletion_completed: %v", err)
		}
	}

	logger.Info("[Order DPDP Consumer] Successfully anonymized orders for user %s (%d rows)", req.UserID, rowsAnonymized)
	return nil
}
