package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type DPDPDeletionProcessor struct {
	repo     Repository
	producer *kafka.Producer
}

func NewDPDPDeletionProcessor(repo Repository, producer *kafka.Producer) *DPDPDeletionProcessor {
	return &DPDPDeletionProcessor{repo: repo, producer: producer}
}

func (p *DPDPDeletionProcessor) ProcessDeletionRequest(ctx context.Context, reqID, adminID string, stepUpVerified bool) (*DPDPRequest, error) {
	if !stepUpVerified {
		return nil, fmt.Errorf("STEP_UP_REQUIRED: Step-up authentication re-verification is required for irreversible DPDP deletion")
	}

	req, err := p.repo.GetDPDPRequestByID(ctx, reqID)
	if err != nil {
		return nil, err
	}
	if req.RequestType != "DELETION" {
		return nil, fmt.Errorf("request %s is not a DELETION request", reqID)
	}

	// Set in-progress
	hPtr := &adminID
	if err := p.repo.UpdateDPDPRequestStatus(ctx, reqID, DPDPStatusInProgress, hPtr, nil); err != nil {
		return nil, fmt.Errorf("failed to mark request in-progress: %w", err)
	}

	// Publish fan-out event dpdp.user_data_deletion_requested
	userType := req.UserType
	if userType == "" {
		userType = "CUSTOMER"
	}

	deletionPayload, _ := json.Marshal(map[string]interface{}{
		"user_id":         req.UserID,
		"user_type":       userType,
		"dpdp_request_id": req.ID,
	})

	if p.producer != nil {
		if err := p.producer.PublishEvent(ctx, "dpdp.user_data_deletion_requested", req.UserID, deletionPayload); err != nil {
			logger.Error("[DPDP Deletion Processor] Failed to publish fan-out deletion event: %v", err)
			return nil, fmt.Errorf("failed to publish deletion event to Kafka: %w", err)
		}
	}

	req.Status = DPDPStatusInProgress
	req.HandledBy = hPtr
	return req, nil
}

type DPDPDeletionCompletedPayload struct {
	UserID           string   `json:"user_id"`
	DPDPRequestID    string   `json:"dpdp_request_id"`
	ServiceName      string   `json:"service_name"`
	TablesAffected   []string `json:"tables_affected"`
	RowsDeletedCount int      `json:"rows_deleted_count"`
}

func (p *DPDPDeletionProcessor) HandleDeletionCompletedConsumer(ctx context.Context, value []byte) error {
	var payload DPDPDeletionCompletedPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal dpdp.deletion_completed: %w", err)
	}

	if payload.DPDPRequestID == "" || payload.ServiceName == "" {
		logger.Warn("[DPDP Deletion Processor] Empty request_id or service_name in deletion completed event")
		return nil
	}

	// 1. Log to dpdp_deletion_audit
	audit := &DPDPDeletionAudit{
		DPDPRequestID:    payload.DPDPRequestID,
		ServiceName:      payload.ServiceName,
		TablesAffected:   payload.TablesAffected,
		RowsDeletedCount: payload.RowsDeletedCount,
		ExecutedAt:       time.Now().UTC(),
	}
	if err := p.repo.InsertDeletionAudit(ctx, audit); err != nil {
		logger.Error("[DPDP Deletion Processor] Failed to insert deletion audit: %v", err)
	}

	// 2. Fetch DPDP request to check user_type
	req, err := p.repo.GetDPDPRequestByID(ctx, payload.DPDPRequestID)
	if err != nil {
		logger.Error("[DPDP Deletion Processor] Failed to find request %s: %v", payload.DPDPRequestID, err)
		return nil
	}

	audits, err := p.repo.GetDeletionAuditsByRequestID(ctx, payload.DPDPRequestID)
	if err == nil {
		servicesReported := make(map[string]bool)
		for _, a := range audits {
			servicesReported[a.ServiceName] = true
		}

		userType := req.UserType
		if userType == "" {
			userType = "CUSTOMER"
		}

		allReported := false
		switch userType {
		case "STAFF":
			allReported = servicesReported["retailer-auth-service"] && servicesReported["notification-service"]
		case "CHAIN_HQ":
			allReported = servicesReported["chain-hq-service"] && servicesReported["notification-service"]
		default: // CUSTOMER
			allReported = servicesReported["auth-service"] &&
				servicesReported["order-service"] &&
				servicesReported["loyalty-service"] &&
				servicesReported["notification-service"]
		}

		if allReported {
			_ = p.repo.UpdateDPDPRequestStatus(ctx, payload.DPDPRequestID, DPDPStatusCompleted, nil, nil)
			logger.Info("[DPDP Deletion Processor] DPDP Deletion Request %s (%s) fully COMPLETED!", payload.DPDPRequestID, userType)
		}
	}

	return nil
}

func DetermineExpectedServices(userType string) []string {
	switch userType {
	case "STAFF":
		return []string{"retailer-auth-service", "notification-service"}
	case "CHAIN_HQ":
		return []string{"chain-hq-service", "notification-service"}
	default: // CUSTOMER
		return []string{"auth-service", "order-service", "loyalty-service", "notification-service"}
	}
}
