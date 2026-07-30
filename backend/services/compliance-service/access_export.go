package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/audit"
)

var (
	ErrExportNotReady = errors.New("EXPORT_NOT_READY")
	ErrExportExpired  = errors.New("EXPORT_EXPIRED")
	ErrExportNotYours = errors.New("EXPORT_NOT_YOURS")
)

type AccessExportManager struct {
	repo     Repository
	auditPub *audit.Publisher
}

func NewAccessExportManager(repo Repository, auditPub *audit.Publisher) *AccessExportManager {
	return &AccessExportManager{
		repo:     repo,
		auditPub: auditPub,
	}
}

// ProcessAccessRequest initiates the multi-service fan-out for a DPDP ACCESS request
func (m *AccessExportManager) ProcessAccessRequest(ctx context.Context, requestID string) (*DPDPAccessExport, error) {
	req, err := m.repo.GetDPDPRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if req.RequestType != "ACCESS" {
		return nil, fmt.Errorf("invalid request type %s for access export fulfillment", req.RequestType)
	}

	expectedServices := DetermineExpectedServices(req.UserType)

	export := &DPDPAccessExport{
		DPDPRequestID:    req.ID,
		Status:           ExportStatusAssembling,
		ExpectedServices: expectedServices,
		ServicesReported: []string{},
		Sections:         "{}",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := m.repo.CreateAccessExport(ctx, export); err != nil {
		return nil, err
	}

	// Update DPDP request status to IN_PROGRESS
	reason := "Admin initiated access export assembly"
	_ = m.repo.UpdateDPDPRequestStatus(ctx, req.ID, DPDPStatusInProgress, nil, &reason)

	log.Printf("[DPDP Access Fulfillment] Initiated access export assembly for request %s (user_type: %s, expected: %v)", req.ID, req.UserType, expectedServices)
	return export, nil
}

// HandleAccessDataReported appends reported service data and completes export assembly when all services respond
func (m *AccessExportManager) HandleAccessDataReported(ctx context.Context, requestID string, serviceName string, rawData json.RawMessage) error {
	export, err := m.repo.GetAccessExportByRequestID(ctx, requestID)
	if err != nil {
		return err
	}

	if export.Status != ExportStatusAssembling {
		log.Printf("[DPDP Access Fulfillment] Export %s for request %s is in status %s (ignoring late payload from %s)", export.ID, requestID, export.Status, serviceName)
		return nil
	}

	// Check if service already reported
	for _, s := range export.ServicesReported {
		if s == serviceName {
			log.Printf("[DPDP Access Fulfillment] Service %s already reported for export %s", serviceName, export.ID)
			return nil
		}
	}

	// Parse sections map
	sectionsMap := make(map[string]json.RawMessage)
	if export.Sections != "" && export.Sections != "{}" {
		_ = json.Unmarshal([]byte(export.Sections), &sectionsMap)
	}
	sectionsMap[serviceName] = rawData

	updatedSectionsJSON, _ := json.Marshal(sectionsMap)
	export.ServicesReported = append(export.ServicesReported, serviceName)
	export.Sections = string(updatedSectionsJSON)
	export.UpdatedAt = time.Now()

	// Check if all expected services have reported
	if isComplete(export.ExpectedServices, export.ServicesReported) {
		s3Key := fmt.Sprintf("dpdp-exports/%s.json", requestID)
		now := time.Now()
		expiresAt := now.Add(7 * 24 * time.Hour) // Presigned URL valid for 7 days

		export.Status = ExportStatusReady
		export.S3Key = &s3Key
		export.ReadyAt = &now
		export.ExpiresAt = &expiresAt

		if err := m.repo.UpdateAccessExportStatus(ctx, export); err != nil {
			return err
		}

		// Update DPDP Request status to COMPLETED
		completeReason := "Access export generated successfully"
		_ = m.repo.UpdateDPDPRequestStatus(ctx, requestID, DPDPStatusCompleted, nil, &completeReason)

		log.Printf("[DPDP Access Fulfillment] DPDP Access Export for request %s FULLY ASSEMBLED & READY (s3: %s)", requestID, s3Key)
	} else {
		if err := m.repo.UpdateAccessExportSections(ctx, export); err != nil {
			return err
		}
		log.Printf("[DPDP Access Fulfillment] Export %s updated with section %s (%d/%d reported)", export.ID, serviceName, len(export.ServicesReported), len(export.ExpectedServices))
	}

	return nil
}

// GetAccessExportForDownload validates ownership, status, and expiry before releasing presigned download URL
func (m *AccessExportManager) GetAccessExportForDownload(ctx context.Context, requestID string, requestingUserID string, req *http.Request) (*DPDPAccessExport, error) {
	dpdpReq, err := m.repo.GetDPDPRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Reject different user_id
	if dpdpReq.UserID != requestingUserID {
		return nil, ErrExportNotYours
	}

	export, err := m.repo.GetAccessExportByRequestID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if export.Status == ExportStatusAssembling {
		return nil, ErrExportNotReady
	}

	if export.Status == ExportStatusExpired || (export.ExpiresAt != nil && time.Now().After(*export.ExpiresAt)) {
		return nil, ErrExportExpired
	}

	// Generate presigned download URL (mock S3 URL with token expiry)
	mockDownloadURL := fmt.Sprintf("https://s3.ap-south-1.amazonaws.com/zippyra-dpdp-exports/dpdp-exports/%s.json?token=exp_%d", requestID, time.Now().Add(1*time.Hour).Unix())
	export.DownloadURL = &mockDownloadURL

	// Mark status = DOWNLOADED on first access
	if export.Status == ExportStatusReady {
		export.Status = ExportStatusDownloaded
		_ = m.repo.UpdateAccessExportStatus(ctx, export)
	}

	log.Printf("[DPDP Access Audit] User %s downloaded data export for request %s (IP: %s)", requestingUserID, requestID, req.RemoteAddr)

	return export, nil
}

func isComplete(expected, reported []string) bool {
	if len(reported) < len(expected) {
		return false
	}
	repMap := make(map[string]bool)
	for _, r := range reported {
		repMap[r] = true
	}
	for _, e := range expected {
		if !repMap[e] {
			return false
		}
	}
	return true
}
