package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestComplianceReconciliationDiscrepancy_RoutesViaOpsAlertChannels_NoFCM(t *testing.T) {
	repo := NewMemoryRepository()
	fcmClient := NewMockFCMClient()
	waClient := NewMockWhatsAppClient()
	opsAlerts := NewOpsAlertDispatcher(repo)
	engine := NewNotificationEngine(repo, fcmClient, waClient)
	consumer := NewEventConsumer(engine, opsAlerts)

	// Configure active ops alert channel for reconciliation discrepancies
	_ = repo.CreateOpsAlertChannel(context.Background(), &OpsAlertChannel{
		ID:          "chan-slack-1",
		ChannelType: "SLACK",
		Target:      "https://hooks.slack.com/services/mock/ops/alert",
		AlertTypes:  []string{"compliance.reconciliation_discrepancy"},
		IsActive:    true,
	})

	eventPayload := map[string]interface{}{
		"event_id":       "evt-recon-disc-88",
		"date":           "2026-08-01",
		"discrepancies": 3,
	}
	bytes, _ := json.Marshal(eventPayload)

	err := consumer.ProcessComplianceReconciliationDiscrepancy(context.Background(), bytes)
	if err != nil {
		t.Fatalf("ProcessComplianceReconciliationDiscrepancy failed: %v", err)
	}

	// 0 FCM pushes or WhatsApp sends
	if len(fcmClient.SentPushes) != 0 {
		t.Errorf("Expected 0 FCM push sends for internal ops alert, got %d", len(fcmClient.SentPushes))
	}
	if len(waClient.SentMessages) != 0 {
		t.Errorf("Expected 0 WhatsApp sends for internal ops alert, got %d", len(waClient.SentMessages))
	}

	// Dispatched to Ops Alert channel
	dispatcher := opsAlerts
	if len(dispatcher.SentAlerts) != 1 {
		t.Fatalf("Expected 1 ops alert dispatched, got %d", len(dispatcher.SentAlerts))
	}
	if dispatcher.SentAlerts[0]["channel_type"] != "SLACK" {
		t.Errorf("Expected SLACK channel type, got %v", dispatcher.SentAlerts[0]["channel_type"])
	}
}
