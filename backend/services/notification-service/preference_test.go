package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestNotificationPreferenceGating(t *testing.T) {
	repo := NewMemoryRepository()
	fcmClient := NewMockFCMClient()
	waClient := NewMockWhatsAppClient()
	engine := NewNotificationEngine(repo, fcmClient, waClient)

	customerID := "cust-pref-1"

	// 1. User sets MARKETING preference to NONE
	_ = repo.UpsertPreference(context.Background(), &NotificationPreference{
		UserID:           customerID,
		UserType:         UserTypeCustomer,
		NotificationType: NotificationTypeMarketing,
		Channel:          ChannelNone,
	})

	// Trigger DPDP Marketing notification
	err := engine.Notify(
		context.Background(),
		customerID,
		UserTypeCustomer,
		NotificationTypeMarketing,
		"Data Request Received",
		"We received your data request.",
		"/settings/privacy",
		"dpdp.request_received",
		"evt-dpdp-001",
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("Notify failed: %v", err)
	}

	// 0 external sends
	if len(fcmClient.SentPushes) != 0 {
		t.Errorf("Expected 0 FCM pushes when preference is NONE, got %d", len(fcmClient.SentPushes))
	}

	// 1 notification_log row with channel_sent='NONE'
	logs, _ := repo.ListUserInbox(context.Background(), customerID, 1, 10)
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log row in inbox, got %d", len(logs))
	}
	if logs[0].ChannelSent != string(ChannelNone) {
		t.Errorf("Expected log channel_sent='NONE', got '%s'", logs[0].ChannelSent)
	}
}

func TestMandatoryNotification_CannotSetToNone(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewNotificationHandler(repo)
	r := mux.NewRouter()
	SetupRoutes(r, handler, NewAdminHandler(repo), nil)

	// Attempt to set PAYMENT_REFUND (Mandatory) to NONE
	payload := map[string]interface{}{
		"notification_type": "PAYMENT_REFUND",
		"channel":           "NONE",
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/v1/notification/preferences", bytes.NewBuffer(bodyBytes))
	req.Header.Set("X-User-ID", "cust-pref-2")
	req.Header.Set("X-User-Role", "CUSTOMER")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected HTTP 400 when attempting to set mandatory notification to NONE, got %d", rr.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	errObj, ok := resp["error"].(map[string]interface{})
	if !ok || errObj["code"] != "CANNOT_DISABLE_MANDATORY_NOTIFICATION" {
		t.Errorf("Expected code CANNOT_DISABLE_MANDATORY_NOTIFICATION, got %v", resp)
	}
}
