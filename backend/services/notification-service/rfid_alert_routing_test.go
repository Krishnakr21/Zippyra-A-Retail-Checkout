package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestExitRFIDFailure_RoutesToSecurityStaffAtStoreOnly(t *testing.T) {
	repo := NewMemoryRepository()
	fcmClient := NewMockFCMClient()
	waClient := NewMockWhatsAppClient()
	opsAlerts := NewOpsAlertDispatcher(repo)
	engine := NewNotificationEngine(repo, fcmClient, waClient)
	consumer := NewEventConsumer(engine, opsAlerts)

	storeID := "store-mumbai-01"
	secStaffID := "staff-sec-99"

	// Register roster: store-mumbai-01 -> SECURITY -> [staff-sec-99]
	consumer.SetStaffRoster(storeID, "SECURITY", []string{secStaffID})

	// Add device token for security staff
	_ = repo.UpsertDeviceToken(context.Background(), &DeviceToken{
		UserID:   secStaffID,
		UserType: UserTypeStaff,
		FCMToken: "fcm_sec_staff_99",
		Platform: PlatformAndroid,
		DeviceID: "sec_handheld_device",
	})

	// Set staff preference to NONE (should be ignored because SECURITY_ALERTS is MANDATORY)
	_ = repo.UpsertPreference(context.Background(), &NotificationPreference{
		UserID:           secStaffID,
		UserType:         UserTypeStaff,
		NotificationType: NotificationTypeSecurityAlerts,
		Channel:          ChannelNone,
	})

	eventPayload := map[string]interface{}{
		"event_id": "evt-rfid-fail-77",
		"store_id": storeID,
		"tag_id":   "rfid_tag_xyz_123",
	}
	bytes, _ := json.Marshal(eventPayload)

	err := consumer.ProcessExitRFIDFailure(context.Background(), bytes)
	if err != nil {
		t.Fatalf("ProcessExitRFIDFailure failed: %v", err)
	}

	// 1 FCM push sent to security staff (mandatory alert ignored NONE preference)
	if len(fcmClient.SentPushes) != 1 {
		t.Fatalf("Expected 1 FCM push to security staff, got %d", len(fcmClient.SentPushes))
	}
	if fcmClient.SentPushes[0]["token"] != "fcm_sec_staff_99" {
		t.Errorf("Expected push token 'fcm_sec_staff_99', got %v", fcmClient.SentPushes[0]["token"])
	}

	logs, _ := repo.ListUserInbox(context.Background(), secStaffID, 1, 10)
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log row for security staff, got %d", len(logs))
	}
	if logs[0].NotificationType != NotificationTypeSecurityAlerts {
		t.Errorf("Expected SECURITY_ALERTS type, got %s", logs[0].NotificationType)
	}
}
