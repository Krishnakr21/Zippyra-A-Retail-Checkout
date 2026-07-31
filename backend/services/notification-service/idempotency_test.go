package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDuplicateOrderCompleted_OnlyOneNotificationSentAndLogged(t *testing.T) {
	repo := NewMemoryRepository()
	fcmClient := NewMockFCMClient()
	waClient := NewMockWhatsAppClient()
	opsAlerts := NewOpsAlertDispatcher(repo)

	// Approve whatsapp template
	_ = repo.UpsertWhatsAppTemplateConfig(context.Background(), &WhatsAppTemplateConfig{
		TemplateKey:          "order_confirmation",
		WhatsAppTemplateName: "order_receipt_v1",
		IsApproved:           true,
		Language:             "en",
	})

	// Add device token for customer
	customerID := "cust-100"
	_ = repo.UpsertDeviceToken(context.Background(), &DeviceToken{
		UserID:   customerID,
		UserType: UserTypeCustomer,
		FCMToken: "fcm_token_cust_100",
		Platform: PlatformAndroid,
		DeviceID: "device_100",
	})

	engine := NewNotificationEngine(repo, fcmClient, waClient)
	consumer := NewEventConsumer(engine, opsAlerts)

	eventPayload := map[string]interface{}{
		"event_id":    "evt-ord-completed-001",
		"order_id":    "ord-500",
		"customer_id": customerID,
		"total_paise": 15000,
	}
	bytes, _ := json.Marshal(eventPayload)

	// 1. Process event first time
	err1 := consumer.ProcessOrderCompleted(context.Background(), bytes)
	if err1 != nil {
		t.Fatalf("First delivery failed: %v", err1)
	}

	if len(fcmClient.SentPushes) != 1 {
		t.Errorf("Expected 1 FCM push on first delivery, got %d", len(fcmClient.SentPushes))
	}
	if len(waClient.SentMessages) != 1 {
		t.Errorf("Expected 1 WhatsApp message on first delivery, got %d", len(waClient.SentMessages))
	}

	// 2. Process exact same event second time (Kafka redelivery)
	err2 := consumer.ProcessOrderCompleted(context.Background(), bytes)
	if err2 != nil {
		t.Fatalf("Second delivery failed: %v", err2)
	}

	// Should still be exactly 1 call each (idempotency guard)
	if len(fcmClient.SentPushes) != 1 {
		t.Errorf("Expected exactly 1 FCM push after redelivery, got %d", len(fcmClient.SentPushes))
	}
	if len(waClient.SentMessages) != 1 {
		t.Errorf("Expected exactly 1 WhatsApp message after redelivery, got %d", len(waClient.SentMessages))
	}

	logs, _ := repo.ListUserInbox(context.Background(), customerID, 1, 10)
	if len(logs) != 1 {
		t.Errorf("Expected exactly 1 inbox log row, got %d", len(logs))
	}
}
