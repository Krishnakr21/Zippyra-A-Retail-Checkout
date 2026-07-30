package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/zippyra/backend/shared/kafka"
)

type DPDPDeletionConsumer struct {
	repo     Repository
	producer *kafka.Producer
}

func NewDPDPDeletionConsumer(repo Repository, producer *kafka.Producer) *DPDPDeletionConsumer {
	return &DPDPDeletionConsumer{repo: repo, producer: producer}
}

type DPDPDeletionRequestedPayload struct {
	UserID        string `json:"user_id"`
	UserType      string `json:"user_type"`
	DPDPRequestID string `json:"dpdp_request_id"`
}

func (c *DPDPDeletionConsumer) ProcessDeletionRequest(ctx context.Context, value []byte) error {
	var payload DPDPDeletionRequestedPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal dpdp.user_data_deletion_requested: %w", err)
	}

	if payload.UserType != "" && payload.UserType != "CHAIN_HQ" {
		// Ignore requests for other user types (CUSTOMER, STAFF, ADMIN)
		return nil
	}

	if err := c.repo.AnonymizeUser(ctx, payload.UserID); err != nil && err != ErrUserNotFound {
		log.Printf("[DPDP Consumer - Chain HQ] Failed to anonymize user %s: %v", payload.UserID, err)
		return err
	}

	log.Printf("[DPDP Consumer - Chain HQ] Successfully anonymized user %s", payload.UserID)

	// Publish dpdp.deletion_completed
	completedPayload, _ := json.Marshal(map[string]interface{}{
		"user_id":           payload.UserID,
		"dpdp_request_id":    payload.DPDPRequestID,
		"service_name":      "chain-hq-service",
		"tables_affected":   []string{"chain_hq_users"},
		"rows_deleted_count": 1,
	})

	if c.producer != nil {
		if err := c.producer.PublishEvent(ctx, "dpdp.deletion_completed", payload.UserID, completedPayload); err != nil {
			log.Printf("[DPDP Consumer - Chain HQ] Failed to publish deletion completed event: %v", err)
			return err
		}
	}

	return nil
}
