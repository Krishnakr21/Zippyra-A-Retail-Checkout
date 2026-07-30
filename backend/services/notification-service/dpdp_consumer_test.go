package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNotificationDPDPConsumer_PurgesNotificationData(t *testing.T) {
	consumer := NewNotificationDPDPConsumer(nil)

	reqBytes, _ := json.Marshal(DPDPNotificationDeletionRequestPayload{
		UserID:        "user-notif-dpdp-1",
		DPDPRequestID: "req-dpdp-notif-001",
	})

	err := consumer.HandleUserDataDeletionRequested(context.Background(), reqBytes)
	if err != nil {
		t.Fatalf("HandleUserDataDeletionRequested failed: %v", err)
	}
}
