package main

import (
	"context"
	"testing"
)

func TestMultiDeviceFCM_OneInvalidTokenCleanedUpOtherSucceeds(t *testing.T) {
	repo := NewMemoryRepository()
	fcmClient := NewMockFCMClient()
	waClient := NewMockWhatsAppClient()
	engine := NewNotificationEngine(repo, fcmClient, waClient)

	customerID := "cust-multi-dev"

	// Register 2 active tokens for user
	tok1 := &DeviceToken{
		ID:       "tok-id-1",
		UserID:   customerID,
		UserType: UserTypeCustomer,
		FCMToken: "fcm_token_valid_1",
		Platform: PlatformAndroid,
		DeviceID: "device_phone",
	}
	tok2 := &DeviceToken{
		ID:       "tok-id-2",
		UserID:   customerID,
		UserType: UserTypeCustomer,
		FCMToken: "fcm_token_invalid_2",
		Platform: PlatformIOS,
		DeviceID: "device_tablet",
	}

	_ = repo.UpsertDeviceToken(context.Background(), tok1)
	_ = repo.UpsertDeviceToken(context.Background(), tok2)

	// Set tok2 as invalid in FCM client mock
	fcmClient.SetTokenInvalid("fcm_token_invalid_2")

	err := engine.Notify(
		context.Background(),
		customerID,
		UserTypeCustomer,
		NotificationTypeOrderUpdates,
		"Order Shipping",
		"Your order has shipped!",
		"/orders/ord-shipping-1",
		"order.shipping",
		"evt-ship-001",
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("Notify should not fail overall when one token is invalid, got: %v", err)
	}

	// 1 push succeeded (valid token)
	if len(fcmClient.SentPushes) != 1 {
		t.Errorf("Expected 1 successful FCM push send, got %d", len(fcmClient.SentPushes))
	}

	// Token 2 should now be marked is_active=false in repo
	tokens, _ := repo.GetActiveDeviceTokens(context.Background(), customerID)
	if len(tokens) != 1 {
		t.Fatalf("Expected 1 active device token remaining, got %d", len(tokens))
	}
	if tokens[0].ID != "tok-id-1" {
		t.Errorf("Expected remaining active token to be tok-id-1, got %s", tokens[0].ID)
	}
}
