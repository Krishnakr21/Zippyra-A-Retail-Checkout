package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDPDPAccessConsumer_ReportsAccountData(t *testing.T) {
	repo := NewMemoryRepository()
	user, err := repo.CreateUserWithPhone(context.Background(), "+919876543210")
	if err != nil {
		t.Fatalf("CreateUserWithPhone failed: %v", err)
	}

	consumer := NewDPDPAccessConsumer(repo, nil)

	evt := DPDPAccessRequestedEvent{
		DPDPRequestID: "req-acc-001",
		UserID:        user.ID,
		UserType:      "CUSTOMER",
	}

	payload, _ := json.Marshal(evt)
	err = consumer.HandleAccessRequested(context.Background(), payload)
	if err != nil {
		t.Fatalf("HandleAccessRequested failed: %v", err)
	}
}
