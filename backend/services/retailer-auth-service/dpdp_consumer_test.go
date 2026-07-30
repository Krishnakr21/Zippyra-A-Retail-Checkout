package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDPDPDeletionConsumer_AnonymizesStaffMember(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// 1. Create a staff member
	staff := &StaffMember{
		ID:       "staff-dpdp-001",
		StoreID:  "store-001",
		Name:     "Rahul Sharma",
		Phone:    "+919876543210",
		Role:     "CASHIER",
		IsActive: true,
	}
	_ = repo.CreateStaffMember(ctx, staff)

	consumer := NewDPDPDeletionConsumer(repo, nil)

	// 2. Process DPDP deletion request for STAFF
	payload, _ := json.Marshal(DPDPDeletionRequestedPayload{
		UserID:        "staff-dpdp-001",
		UserType:      "STAFF",
		DPDPRequestID: "req-dpdp-staff-1",
	})

	if err := consumer.ProcessDeletionRequest(ctx, payload); err != nil {
		t.Fatalf("ProcessDeletionRequest failed: %v", err)
	}

	// 3. Verify staff member anonymized
	updatedStaff, err := repo.GetStaffByID(ctx, "staff-dpdp-001")
	if err != nil {
		t.Fatalf("GetStaffByID failed: %v", err)
	}

	if updatedStaff.Name != "Anonymized Staff Member" {
		t.Errorf("Expected name 'Anonymized Staff Member', got '%s'", updatedStaff.Name)
	}

	if updatedStaff.IsActive != false {
		t.Errorf("Expected IsActive = false, got true")
	}

	if updatedStaff.Phone != "deleted_staff-dpdp-001" {
		t.Errorf("Expected tombstone phone 'deleted_staff-dpdp-001', got '%s'", updatedStaff.Phone)
	}
}
