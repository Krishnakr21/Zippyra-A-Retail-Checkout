package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/zippyra/backend/shared/kafka"
)

type DPDPAccessRequestedEvent struct {
	DPDPRequestID string `json:"dpdp_request_id"`
	UserID        string `json:"user_id"`
	UserType      string `json:"user_type"`
}

type DPDPAccessDataReportedEvent struct {
	DPDPRequestID string          `json:"dpdp_request_id"`
	UserID        string          `json:"user_id"`
	ServiceName   string          `json:"service_name"`
	Data          json.RawMessage `json:"data"`
}

type DPDPAccessConsumer struct {
	repo     Repository
	producer *kafka.Producer
}

func NewDPDPAccessConsumer(repo Repository, producer *kafka.Producer) *DPDPAccessConsumer {
	return &DPDPAccessConsumer{
		repo:     repo,
		producer: producer,
	}
}

func (c *DPDPAccessConsumer) HandleAccessRequested(ctx context.Context, msg []byte) error {
	var evt DPDPAccessRequestedEvent
	if err := json.Unmarshal(msg, &evt); err != nil {
		log.Printf("[DPDP Access Consumer - Chain HQ] Failed to unmarshal event: %v", err)
		return err
	}

	if evt.UserType != "CHAIN_HQ" {
		return nil
	}

	user, err := c.repo.GetUserByID(ctx, evt.UserID)
	var hqData map[string]interface{}
	if err != nil || user == nil {
		hqData = map[string]interface{}{
			"status": "not_found",
		}
	} else {
		hqData = map[string]interface{}{
			"user_id":    user.ID,
			"name":       user.Name,
			"phone":      user.Phone,
			"chain_id":   user.ChainID,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		}
	}

	dataBytes, _ := json.Marshal(hqData)

	reportedEvt := DPDPAccessDataReportedEvent{
		DPDPRequestID: evt.DPDPRequestID,
		UserID:        evt.UserID,
		ServiceName:   "chain-hq-service",
		Data:          dataBytes,
	}

	payload, _ := json.Marshal(reportedEvt)
	if c.producer != nil {
		_ = c.producer.PublishEvent(ctx, "dpdp.access_data_reported", evt.UserID, payload)
	}

	log.Printf("[DPDP Access Consumer - Chain HQ] Successfully reported HQ user data for user %s (request: %s)", evt.UserID, evt.DPDPRequestID)
	return nil
}
