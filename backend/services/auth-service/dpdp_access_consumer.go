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
		log.Printf("[DPDP Access Consumer - Auth] Failed to unmarshal event: %v", err)
		return err
	}

	if evt.UserType != "CUSTOMER" {
		return nil
	}

	user, err := c.repo.GetUserByID(ctx, evt.UserID)
	var accountData map[string]interface{}
	if err != nil || user == nil {
		accountData = map[string]interface{}{
			"status": "not_found",
		}
	} else {
		phoneVal := ""
		if user.Phone != nil {
			phoneVal = *user.Phone
		}
		emailVal := ""
		if user.Email != nil {
			emailVal = *user.Email
		}
		accountData = map[string]interface{}{
			"user_id":            user.ID,
			"phone":              phoneVal,
			"email":              emailVal,
			"google_linked":      user.GoogleSub != nil,
			"phone_verified_at": user.PhoneVerifiedAt,
		}
	}

	dataBytes, _ := json.Marshal(accountData)

	reportedEvt := DPDPAccessDataReportedEvent{
		DPDPRequestID: evt.DPDPRequestID,
		UserID:        evt.UserID,
		ServiceName:   "auth-service",
		Data:          dataBytes,
	}

	payload, _ := json.Marshal(reportedEvt)
	if c.producer != nil {
		_ = c.producer.PublishEvent(ctx, "dpdp.access_data_reported", evt.UserID, payload)
	}

	log.Printf("[DPDP Access Consumer - Auth] Successfully reported data for user %s (request: %s)", evt.UserID, evt.DPDPRequestID)
	return nil
}
