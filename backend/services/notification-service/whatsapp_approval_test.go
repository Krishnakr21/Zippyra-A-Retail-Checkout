package main

import (
	"context"
	"testing"
)

func TestUnapprovedWhatsAppTemplate_SkippedFallbackToPush(t *testing.T) {
	repo := NewMemoryRepository()
	fcmClient := NewMockFCMClient()
	waClient := NewMockWhatsAppClient()
	engine := NewNotificationEngine(repo, fcmClient, waClient)

	customerID := "cust-wa-unapproved"

	// Register device token for PUSH
	_ = repo.UpsertDeviceToken(context.Background(), &DeviceToken{
		UserID:   customerID,
		UserType: UserTypeCustomer,
		FCMToken: "fcm_token_wa_fallback",
		Platform: PlatformIOS,
		DeviceID: "device_wa_fallback",
	})

	// Template is NOT approved (is_approved = false)
	_ = repo.UpsertWhatsAppTemplateConfig(context.Background(), &WhatsAppTemplateConfig{
		TemplateKey:          "order_confirmation",
		WhatsAppTemplateName: "order_receipt_unapproved",
		IsApproved:           false,
		Language:             "en",
	})

	// User preference is BOTH (PUSH + WHATSAPP)
	_ = repo.UpsertPreference(context.Background(), &NotificationPreference{
		UserID:           customerID,
		UserType:         UserTypeCustomer,
		NotificationType: NotificationTypeOrderUpdates,
		Channel:          ChannelBoth,
	})

	err := engine.Notify(
		context.Background(),
		customerID,
		UserTypeCustomer,
		NotificationTypeOrderUpdates,
		"Order Confirmed",
		"Your order for ₹150.00 is confirmed!",
		"/orders/ord-999",
		"order.completed",
		"evt-ord-unapproved-1",
		"order_confirmation",
		nil,
	)
	if err != nil {
		t.Fatalf("Notify failed: %v", err)
	}

	// 0 WhatsApp calls (unapproved template skipped)
	if len(waClient.SentMessages) != 0 {
		t.Errorf("Expected 0 WhatsApp sends for unapproved template, got %d", len(waClient.SentMessages))
	}

	// 1 FCM push call (fallback to PUSH succeeds)
	if len(fcmClient.SentPushes) != 1 {
		t.Errorf("Expected 1 FCM push fallback, got %d", len(fcmClient.SentPushes))
	}

	// Notification log updated to channel_sent='PUSH'
	logs, _ := repo.ListUserInbox(context.Background(), customerID, 1, 10)
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log row, got %d", len(logs))
	}
	if logs[0].ChannelSent != string(ChannelPush) {
		t.Errorf("Expected channel_sent='PUSH', got '%s'", logs[0].ChannelSent)
	}
}
