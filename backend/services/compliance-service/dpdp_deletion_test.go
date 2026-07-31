package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestDPDPDeletionProcessor_StepUpRequirementAndAggregation(t *testing.T) {
	repo := NewMemoryRepository()
	reqSvc := NewDPDPRequestService(repo)
	proc := NewDPDPDeletionProcessor(repo, nil)

	ctx := context.Background()

	// 1. Create a customer deletion request
	req, err := reqSvc.CreateRequest(ctx, "usr-deletion-1", "CUSTOMER", "DELETION", "User requested account deletion")
	if err != nil {
		t.Fatalf("CreateRequest failed: %v", err)
	}

	// 2. Attempt process deletion WITHOUT step-up -> error STEP_UP_REQUIRED
	_, err = proc.ProcessDeletionRequest(ctx, req.ID, "admin-1", false)
	if err == nil || err.Error() == "" {
		t.Fatalf("Expected error when step_up_verified is false")
	}

	// 3. Process deletion WITH step-up -> SUCCESS (status IN_PROGRESS)
	updatedReq, err := proc.ProcessDeletionRequest(ctx, req.ID, "admin-1", true)
	if err != nil || updatedReq.Status != DPDPStatusInProgress {
		t.Fatalf("Expected status IN_PROGRESS, got %v (err: %v)", updatedReq.Status, err)
	}

	// 4. Simulate responses from 4 target services via HandleDeletionCompletedConsumer
	services := []string{"auth-service", "order-service", "loyalty-service", "notification-service"}
	for _, svcName := range services {
		payload, _ := json.Marshal(DPDPDeletionCompletedPayload{
			UserID:           "usr-deletion-1",
			DPDPRequestID:    req.ID,
			ServiceName:      svcName,
			TablesAffected:   []string{"table_1"},
			RowsDeletedCount: 1,
		})
		if err := proc.HandleDeletionCompletedConsumer(ctx, payload); err != nil {
			t.Fatalf("HandleDeletionCompletedConsumer for %s failed: %v", svcName, err)
		}
	}

	// 5. Verify request status is now COMPLETED
	finalReq, err := repo.GetDPDPRequestByID(ctx, req.ID)
	if err != nil || finalReq.Status != DPDPStatusCompleted {
		t.Fatalf("Expected status COMPLETED after all 4 services reported, got %v", finalReq.Status)
	}
}

func TestDPDPDeletionProcessor_StaffAndAdminExclusion(t *testing.T) {
	repo := NewMemoryRepository()
	reqSvc := NewDPDPRequestService(repo)
	proc := NewDPDPDeletionProcessor(repo, nil)

	ctx := context.Background()

	// 1. ADMIN self-service deletion request -> REJECTED
	_, err := reqSvc.CreateRequest(ctx, "admin-usr-1", "ADMIN", "DELETION", "Self-service deletion")
	if err == nil || !strings.Contains(err.Error(), "ADMIN_DELETION_EXCLUDED") {
		t.Fatalf("Expected ADMIN_DELETION_EXCLUDED error for admin self-service deletion, got: %v", err)
	}

	// 2. STAFF deletion request -> Expected services are retailer-auth-service + notification-service ONLY
	staffReq, err := reqSvc.CreateRequest(ctx, "staff-usr-1", "STAFF", "DELETION", "Staff deletion request")
	if err != nil {
		t.Fatalf("CreateRequest for STAFF failed: %v", err)
	}

	_, err = proc.ProcessDeletionRequest(ctx, staffReq.ID, "admin-mgr-1", true)
	if err != nil {
		t.Fatalf("ProcessDeletionRequest for STAFF failed: %v", err)
	}

	// First service (retailer-auth-service) reports -> Still IN_PROGRESS
	payload1, _ := json.Marshal(DPDPDeletionCompletedPayload{
		UserID:           "staff-usr-1",
		DPDPRequestID:    staffReq.ID,
		ServiceName:      "retailer-auth-service",
		TablesAffected:   []string{"staff_members"},
		RowsDeletedCount: 1,
	})
	_ = proc.HandleDeletionCompletedConsumer(ctx, payload1)

	r1, _ := repo.GetDPDPRequestByID(ctx, staffReq.ID)
	if r1.Status != DPDPStatusInProgress {
		t.Fatalf("Expected IN_PROGRESS after 1 of 2 services reported, got %v", r1.Status)
	}

	// Second service (notification-service) reports -> Should reach COMPLETED
	payload2, _ := json.Marshal(DPDPDeletionCompletedPayload{
		UserID:           "staff-usr-1",
		DPDPRequestID:    staffReq.ID,
		ServiceName:      "notification-service",
		TablesAffected:   []string{"notification_preferences"},
		RowsDeletedCount: 1,
	})
	_ = proc.HandleDeletionCompletedConsumer(ctx, payload2)

	r2, _ := repo.GetDPDPRequestByID(ctx, staffReq.ID)
	if r2.Status != DPDPStatusCompleted {
		t.Fatalf("Expected COMPLETED after retailer-auth-service and notification-service reported, got %v", r2.Status)
	}
}
