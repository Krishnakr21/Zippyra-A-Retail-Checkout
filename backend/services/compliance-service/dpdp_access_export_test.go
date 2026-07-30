package main

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDPDPAccessExport_FanOutAndAssembly(t *testing.T) {
	repo := NewMemoryRepository()
	manager := NewAccessExportManager(repo, nil)

	// Create DPDP request for CUSTOMER
	reqCustomer := &DPDPRequest{
		ID:          "req-acc-cust-01",
		UserID:      "usr-cust-001",
		UserType:    "CUSTOMER",
		RequestType: "ACCESS",
		Status:      DPDPStatusReceived,
	}
	_ = repo.CreateDPDPRequest(context.Background(), reqCustomer)

	// Process Access Request -> Initiates assembly with 4 expected services
	export, err := manager.ProcessAccessRequest(context.Background(), reqCustomer.ID)
	if err != nil {
		t.Fatalf("ProcessAccessRequest failed: %v", err)
	}

	if len(export.ExpectedServices) != 4 {
		t.Fatalf("Expected 4 services for CUSTOMER access export, got %d", len(export.ExpectedServices))
	}

	if export.Status != ExportStatusAssembling {
		t.Fatalf("Expected status ASSEMBLING, got %s", export.Status)
	}

	// 1. Report from auth-service
	_ = manager.HandleAccessDataReported(context.Background(), reqCustomer.ID, "auth-service", json.RawMessage(`{"phone":"+919876543210"}`))
	expCheck, _ := repo.GetAccessExportByRequestID(context.Background(), reqCustomer.ID)
	if expCheck.Status != ExportStatusAssembling {
		t.Fatalf("Expected status ASSEMBLING after 1 service, got %s", expCheck.Status)
	}

	// 2. Report from order-service
	_ = manager.HandleAccessDataReported(context.Background(), reqCustomer.ID, "order-service", json.RawMessage(`{"orders":[{"id":"ord-01"}]}`))

	// 3. Report from loyalty-service
	_ = manager.HandleAccessDataReported(context.Background(), reqCustomer.ID, "loyalty-service", json.RawMessage(`{"points":150}`))

	// 4. Report from notification-service -> Triggers completion to READY
	_ = manager.HandleAccessDataReported(context.Background(), reqCustomer.ID, "notification-service", json.RawMessage(`{"notifications":[]}`))

	expReady, _ := repo.GetAccessExportByRequestID(context.Background(), reqCustomer.ID)
	if expReady.Status != ExportStatusReady {
		t.Fatalf("Expected status READY after all 4 services reported, got %s", expReady.Status)
	}

	if expReady.S3Key == nil || *expReady.S3Key == "" {
		t.Fatalf("Expected s3_key to be set upon completion")
	}
}

func TestDPDPAccessExport_StaffUserFanOut(t *testing.T) {
	repo := NewMemoryRepository()
	manager := NewAccessExportManager(repo, nil)

	reqStaff := &DPDPRequest{
		ID:          "req-acc-staff-01",
		UserID:      "staff-001",
		UserType:    "STAFF",
		RequestType: "ACCESS",
		Status:      DPDPStatusReceived,
	}
	_ = repo.CreateDPDPRequest(context.Background(), reqStaff)

	export, err := manager.ProcessAccessRequest(context.Background(), reqStaff.ID)
	if err != nil {
		t.Fatalf("ProcessAccessRequest for STAFF failed: %v", err)
	}

	if len(export.ExpectedServices) != 2 {
		t.Fatalf("Expected 2 services for STAFF access export, got %d", len(export.ExpectedServices))
	}
}

func TestDPDPAccessExport_DownloadSecurityChecks(t *testing.T) {
	repo := NewMemoryRepository()
	manager := NewAccessExportManager(repo, nil)

	req := &DPDPRequest{
		ID:          "req-sec-01",
		UserID:      "owner-user-01",
		UserType:    "CUSTOMER",
		RequestType: "ACCESS",
		Status:      DPDPStatusReceived,
	}
	_ = repo.CreateDPDPRequest(context.Background(), req)
	_, _ = manager.ProcessAccessRequest(context.Background(), req.ID)

	httpRequest := httptest.NewRequest("GET", "/v1/compliance/access-exports/req-sec-01/download", nil)

	// 1. Attempt download while still ASSEMBLING -> ErrExportNotReady
	_, err := manager.GetAccessExportForDownload(context.Background(), req.ID, "owner-user-01", httpRequest)
	if err != ErrExportNotReady {
		t.Fatalf("Expected ErrExportNotReady, got %v", err)
	}

	// 2. Attempt download with wrong user -> ErrExportNotYours
	_, err = manager.GetAccessExportForDownload(context.Background(), req.ID, "attacker-user-99", httpRequest)
	if err != ErrExportNotYours {
		t.Fatalf("Expected ErrExportNotYours, got %v", err)
	}

	// Complete assembly to READY
	_ = manager.HandleAccessDataReported(context.Background(), req.ID, "auth-service", json.RawMessage(`{}`))
	_ = manager.HandleAccessDataReported(context.Background(), req.ID, "order-service", json.RawMessage(`{}`))
	_ = manager.HandleAccessDataReported(context.Background(), req.ID, "loyalty-service", json.RawMessage(`{}`))
	_ = manager.HandleAccessDataReported(context.Background(), req.ID, "notification-service", json.RawMessage(`{}`))

	// 3. Download by rightful owner -> Success & status updated to DOWNLOADED
	downloadedExport, err := manager.GetAccessExportForDownload(context.Background(), req.ID, "owner-user-01", httpRequest)
	if err != nil {
		t.Fatalf("Expected successful download, got %v", err)
	}
	if downloadedExport.Status != ExportStatusDownloaded {
		t.Fatalf("Expected status DOWNLOADED after access, got %s", downloadedExport.Status)
	}

	// 4. Expired export -> ErrExportExpired
	past := time.Now().Add(-1 * time.Hour)
	downloadedExport.ExpiresAt = &past
	downloadedExport.Status = ExportStatusExpired
	_ = repo.UpdateAccessExportStatus(context.Background(), downloadedExport)

	_, err = manager.GetAccessExportForDownload(context.Background(), req.ID, "owner-user-01", httpRequest)
	if err != ErrExportExpired {
		t.Fatalf("Expected ErrExportExpired, got %v", err)
	}
}
