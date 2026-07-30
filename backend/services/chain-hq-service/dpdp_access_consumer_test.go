package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDPDPAccessConsumer_ReportsChainHQUserData(t *testing.T) {
	repo := NewMemoryRepository()
	_ = repo.CreateUser(context.Background(), &ChainHQUser{
		ID:      "hq-usr-acc-001",
		Phone:   "+919876543212",
		Name:    "Anita Sharma",
		ChainID: "chain-001",
		Role:    "CHAIN_HQ_ANALYTICS",
	})

	consumer := NewDPDPAccessConsumer(repo, nil)

	evt := DPDPAccessRequestedEvent{
		DPDPRequestID: "req-hq-acc-001",
		UserID:        "hq-usr-acc-001",
		UserType:      "CHAIN_HQ",
	}

	payload, _ := json.Marshal(evt)
	err := consumer.HandleAccessRequested(context.Background(), payload)
	if err != nil {
		t.Fatalf("HandleAccessRequested failed: %v", err)
	}
}
