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
		log.Printf("[DPDP Access Consumer - Retailer Auth] Failed to unmarshal event: %v", err)
		return err
	}

	if evt.UserType != "STAFF" {
		return nil
	}

	staff, err := c.repo.GetStaffByID(ctx, evt.UserID)
	var staffData map[string]interface{}
	if err != nil || staff == nil {
		staffData = map[string]interface{}{
			"status": "not_found",
		}
	} else {
		staffData = map[string]interface{}{
			"staff_id":   staff.ID,
			"name":       staff.Name,
			"phone":      staff.Phone,
			"store_id":   staff.StoreID,
			"role":       staff.Role,
			"created_at": staff.CreatedAt,
		}
	}

	dataBytes, _ := json.Marshal(staffData)

	reportedEvt := DPDPAccessDataReportedEvent{
		DPDPRequestID: evt.DPDPRequestID,
		UserID:        evt.UserID,
		ServiceName:   "retailer-auth-service",
		Data:          dataBytes,
	}

	payload, _ := json.Marshal(reportedEvt)
	if c.producer != nil {
		_ = c.producer.PublishEvent(ctx, "dpdp.access_data_reported", evt.UserID, payload)
	}

	log.Printf("[DPDP Access Consumer - Retailer Auth] Successfully reported staff data for user %s (request: %s)", evt.UserID, evt.DPDPRequestID)
	return nil
}
