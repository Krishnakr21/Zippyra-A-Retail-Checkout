package main

import (
	"context"
	"fmt"
	"time"
)

type DPDPRequestService struct {
	repo Repository
}

func NewDPDPRequestService(repo Repository) *DPDPRequestService {
	return &DPDPRequestService{repo: repo}
}

func (s *DPDPRequestService) CreateRequest(ctx context.Context, userID, userType, reqType, detail string) (*DPDPRequest, error) {
	if reqType != "ACCESS" && reqType != "CORRECTION" && reqType != "DELETION" {
		return nil, fmt.Errorf("invalid request_type: %s", reqType)
	}

	if userType == "" {
		userType = "CUSTOMER"
	}

	// ADMIN self-service deletion exclusion rule under DPDP Act governance
	if userType == "ADMIN" && reqType == "DELETION" {
		return nil, fmt.Errorf("ADMIN_DELETION_EXCLUDED: Admin accounts are excluded from self-service DPDP deletion; manual Super-Admin review and support escalation is required")
	}

	now := time.Now().UTC()
	slaDueAt := now.AddDate(0, 0, 30) // 30-day statutory SLA under DPDP Act 2023

	var dPtr *string
	if detail != "" {
		dPtr = &detail
	}

	req := &DPDPRequest{
		UserID:      userID,
		UserType:    userType,
		RequestType: reqType,
		Status:      DPDPStatusReceived,
		Detail:      dPtr,
		SLADueAt:    slaDueAt,
		CreatedAt:   now,
	}

	if err := s.repo.CreateDPDPRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to create DPDP request: %w", err)
	}

	return req, nil
}

func (s *DPDPRequestService) ReviewRequest(ctx context.Context, reqID, action, handledBy, reason string) (*DPDPRequest, error) {
	req, err := s.repo.GetDPDPRequestByID(ctx, reqID)
	if err != nil {
		return nil, err
	}

	var status string
	switch action {
	case "APPROVE":
		status = DPDPStatusCompleted
	case "REJECT":
		status = DPDPStatusRejected
	case "IN_PROGRESS":
		status = DPDPStatusInProgress
	default:
		return nil, fmt.Errorf("invalid review action: %s", action)
	}

	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}
	hPtr := &handledBy

	if err := s.repo.UpdateDPDPRequestStatus(ctx, reqID, status, hPtr, reasonPtr); err != nil {
		return nil, fmt.Errorf("failed to update DPDP request status: %w", err)
	}

	req.Status = status
	req.HandledBy = hPtr
	req.RejectionReason = reasonPtr
	if status == DPDPStatusCompleted {
		now := time.Now().UTC()
		req.CompletedAt = &now
	}

	return req, nil
}
