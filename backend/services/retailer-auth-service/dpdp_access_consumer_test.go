package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDPDPAccessConsumer_ReportsStaffData(t *testing.T) {
	repo := NewMemoryRepository()
	_ = repo.CreateStaffMember(context.Background(), &StaffMember{
		ID:      "staff-acc-001",
		Phone:   "+919876543211",
		Name:    "Ramesh Kumar",
		StoreID: "store-001",
		Role:    "CASHIER",
	})

	consumer := NewDPDPAccessConsumer(repo, nil)

	evt := DPDPAccessRequestedEvent{
		DPDPRequestID: "req-staff-acc-001",
		UserID:        "staff-acc-001",
		UserType:      "STAFF",
	}

	payload, _ := json.Marshal(evt)
	err := consumer.HandleAccessRequested(context.Background(), payload)
	if err != nil {
		t.Fatalf("HandleAccessRequested failed: %v", err)
	}
}
