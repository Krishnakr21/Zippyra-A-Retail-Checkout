package main

import (
	"context"
	"testing"
)

func TestDNDCompliance_MarketingPassesPromotionalCategory(t *testing.T) {
	repo := NewMemoryRepository()
	fcm := NewMockFCMClient()
	wa := NewMockWhatsAppClient()

	engine := NewNotificationEngine(repo, fcm, wa)

	// Opt-in user to MARKETING via WhatsApp
	userID := "user-dnd-001"
	_ = repo.UpsertPreference(context.Background(), &NotificationPreference{
		UserID:           userID,
		UserType:         UserTypeCustomer,
		NotificationType: NotificationTypeMarketing,
		Channel:          ChannelWhatsApp,
	})

	// Add approved marketing template
	_ = repo.UpsertWhatsAppTemplateConfig(context.Background(), &WhatsAppTemplateConfig{
		TemplateKey:          "tmpl_promo_deal",
		WhatsAppTemplateName: "promotional_flash_sale",
		IsApproved:           true,
		MetaCategory:         "MARKETING",
		Language:             "en",
	})

	// Trigger MARKETING notification
	err := engine.Notify(
		context.Background(),
		userID,
		UserTypeCustomer,
		NotificationTypeMarketing,
		"Flash Sale 50% Off",
		"Huge savings on grocery items today!",
		"zippyra://promos/flash-sale",
		"promo.campaign",
		"camp-001",
		"tmpl_promo_deal",
		nil,
	)
	if err != nil {
		t.Fatalf("Notify failed: %v", err)
	}

	if len(wa.SentMessages) != 1 {
		t.Fatalf("Expected 1 WhatsApp message sent, got %d", len(wa.SentMessages))
	}

	msgCategory := wa.SentMessages[0]["message_category"]
	if msgCategory != "PROMOTIONAL" {
		t.Errorf("Expected WhatsApp message category to be PROMOTIONAL, got %v", msgCategory)
	}
}

func TestDNDCompliance_SuppressesPromotionalForOptedOutUsers(t *testing.T) {
	repo := NewMemoryRepository()
	fcm := NewMockFCMClient()
	wa := NewMockWhatsAppClient()

	engine := NewNotificationEngine(repo, fcm, wa)

	userID := "user-opt-out-002"
	// Explicitly set MARKETING preference to NONE (DND / Opt-out)
	_ = repo.UpsertPreference(context.Background(), &NotificationPreference{
		UserID:           userID,
		UserType:         UserTypeCustomer,
		NotificationType: NotificationTypeMarketing,
		Channel:          ChannelNone,
	})

	_ = repo.UpsertWhatsAppTemplateConfig(context.Background(), &WhatsAppTemplateConfig{
		TemplateKey:          "tmpl_promo_deal",
		WhatsAppTemplateName: "promotional_flash_sale",
		IsApproved:           true,
		MetaCategory:         "MARKETING",
		Language:             "en",
	})

	err := engine.Notify(
		context.Background(),
		userID,
		UserTypeCustomer,
		NotificationTypeMarketing,
		"Flash Sale",
		"Check out new arrivals!",
		"zippyra://promos",
		"promo.campaign",
		"camp-002",
		"tmpl_promo_deal",
		nil,
	)
	if err != nil {
		t.Fatalf("Notify failed: %v", err)
	}

	if len(wa.SentMessages) != 0 {
		t.Errorf("Expected 0 messages for opted-out user, got %d", len(wa.SentMessages))
	}
}

func TestDNDCompliance_NewUserMarketingDefaultsToNone(t *testing.T) {
	repo := NewMemoryRepository()
	fcm := NewMockFCMClient()
	wa := NewMockWhatsAppClient()

	engine := NewNotificationEngine(repo, fcm, wa)

	userID := "user-new-003" // No explicit preference set

	_ = repo.UpsertWhatsAppTemplateConfig(context.Background(), &WhatsAppTemplateConfig{
		TemplateKey:          "tmpl_promo_deal",
		WhatsAppTemplateName: "promotional_flash_sale",
		IsApproved:           true,
		MetaCategory:         "MARKETING",
		Language:             "en",
	})

	// Send promotional marketing notify
	err := engine.Notify(
		context.Background(),
		userID,
		UserTypeCustomer,
		NotificationTypeMarketing,
		"Welcome Offer",
		"Get 10% off your first purchase!",
		"zippyra://welcome",
		"promo.campaign",
		"camp-003",
		"tmpl_promo_deal",
		nil,
	)
	if err != nil {
		t.Fatalf("Notify failed: %v", err)
	}

	if len(wa.SentMessages) != 0 {
		t.Errorf("Expected 0 promotional messages for new user without explicit opt-in, got %d", len(wa.SentMessages))
	}
}

func TestDNDCompliance_RejectsMarketingTemplateUnderTransactionalCategory(t *testing.T) {
	repo := NewMemoryRepository()
	fcm := NewMockFCMClient()
	wa := NewMockWhatsAppClient()

	engine := NewNotificationEngine(repo, fcm, wa)

	userID := "user-trans-004"
	_ = repo.UpsertPreference(context.Background(), &NotificationPreference{
		UserID:           userID,
		UserType:         UserTypeCustomer,
		NotificationType: NotificationTypeOrderUpdates,
		Channel:          ChannelWhatsApp,
	})

	// Template marked as MARKETING category by Meta
	_ = repo.UpsertWhatsAppTemplateConfig(context.Background(), &WhatsAppTemplateConfig{
		TemplateKey:          "tmpl_mismatch",
		WhatsAppTemplateName: "marketing_template_name",
		IsApproved:           true,
		MetaCategory:         "MARKETING",
		Language:             "en",
	})

	// Trigger ORDER_UPDATES (TRANSACTIONAL category) using a MARKETING meta-category template
	err := engine.Notify(
		context.Background(),
		userID,
		UserTypeCustomer,
		NotificationTypeOrderUpdates,
		"Order Delivered",
		"Your order has been delivered",
		"zippyra://orders/123",
		"order.delivered",
		"ord-del-123",
		"tmpl_mismatch",
		nil,
	)
	if err != nil {
		t.Fatalf("Notify failed: %v", err)
	}

	if len(wa.SentMessages) != 0 {
		t.Errorf("Expected WhatsApp send to be rejected due to Meta category mismatch, got %d messages sent", len(wa.SentMessages))
	}
}
