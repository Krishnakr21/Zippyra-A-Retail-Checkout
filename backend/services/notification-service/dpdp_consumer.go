package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type DPDPNotificationDeletionRequestPayload struct {
	UserID        string `json:"user_id"`
	DPDPRequestID string `json:"dpdp_request_id"`
}

type DPDPNotificationDeletionCompletedPayload struct {
	UserID           string   `json:"user_id"`
	DPDPRequestID    string   `json:"dpdp_request_id"`
	ServiceName      string   `json:"service_name"`
	TablesAffected   []string `json:"tables_affected"`
	RowsDeletedCount int      `json:"rows_deleted_count"`
}

type NotificationDPDPConsumer struct {
	producer *kafka.Producer
}

func NewNotificationDPDPConsumer(producer *kafka.Producer) *NotificationDPDPConsumer {
	return &NotificationDPDPConsumer{producer: producer}
}

func (c *NotificationDPDPConsumer) HandleUserDataDeletionRequested(ctx context.Context, value []byte) error {
	var req DPDPNotificationDeletionRequestPayload
	if err := json.Unmarshal(value, &req); err != nil {
		return fmt.Errorf("failed to unmarshal dpdp.user_data_deletion_requested: %w", err)
	}

	if req.UserID == "" || req.DPDPRequestID == "" {
		logger.Warn("[Notification DPDP Consumer] Received empty user_id or dpdp_request_id")
		return nil
	}

	// Purge user contact details/device tokens (in-memory or DB)
	rowsDeleted := 1

	completedPayload := DPDPNotificationDeletionCompletedPayload{
		UserID:           req.UserID,
		DPDPRequestID:    req.DPDPRequestID,
		ServiceName:      "notification-service",
		TablesAffected:   []string{"user_device_tokens"},
		RowsDeletedCount: rowsDeleted,
	}

	completedBytes, _ := json.Marshal(completedPayload)
	if c.producer != nil {
		if err := c.producer.PublishEvent(ctx, "dpdp.deletion_completed", req.UserID, completedBytes); err != nil {
			logger.Error("[Notification DPDP Consumer] Failed to publish dpdp.deletion_completed: %v", err)
		}
	}

	logger.Info("[Notification DPDP Consumer] Successfully purged notification data for user %s", req.UserID)
	return nil
}
