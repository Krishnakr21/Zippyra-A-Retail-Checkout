package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/zippyra/backend/shared/logger"
)

var ErrCannotDisableMandatory = errors.New("CANNOT_DISABLE_MANDATORY_NOTIFICATION")

type NotificationEngine struct {
	repo           NotificationRepository
	fcmClient      FCMClient
	whatsAppClient WhatsAppClient
}

func NewNotificationEngine(repo NotificationRepository, fcmClient FCMClient, whatsAppClient WhatsAppClient) *NotificationEngine {
	return &NotificationEngine{
		repo:           repo,
		fcmClient:      fcmClient,
		whatsAppClient: whatsAppClient,
	}
}

func (e *NotificationEngine) Notify(
	ctx context.Context,
	userID string,
	userType UserType,
	notificationType NotificationType,
	title string,
	body string,
	deepLink string,
	sourceEventType string,
	sourceEventID string,
	templateKey string,
	templateParams []map[string]interface{},
) error {
	// 1. Idempotency Guard: ON CONFLICT DO NOTHING against notification_log FIRST
	nLog := &NotificationLog{
		UserID:          userID,
		UserType:        userType,
		NotificationType: notificationType,
		ChannelSent:     string(ChannelNone),
		Title:           title,
		Body:            body,
		DeepLink:        deepLink,
		SourceEventType: sourceEventType,
		SourceEventID:   sourceEventID,
	}

	created, err := e.repo.CreateNotificationLog(ctx, nLog)
	if err != nil {
		return fmt.Errorf("failed to create notification log: %w", err)
	}

	if !created {
		// Already sent (duplicate event delivery), return immediately
		logger.Info("[NotificationEngine] Duplicate event delivery ignored (%s, %s, %s)", sourceEventType, sourceEventID, userID)
		return nil
	}

	// 2. Preference & Regulatory Category Check
	messageCategory := "TRANSACTIONAL"
	if notificationType == NotificationTypeMarketing {
		messageCategory = "PROMOTIONAL"
	}

	preferredChannel := ChannelBoth
	if pref, err := e.repo.GetPreference(ctx, userID, notificationType); err == nil && pref != nil {
		preferredChannel = pref.Channel
	} else if notificationType == NotificationTypeMarketing {
		// Statutory Opt-In Rule: MARKETING notification_preferences default to NONE for new users
		preferredChannel = ChannelNone
	}

	// Mandatory Types Override
	isMandatory := MandatoryNotificationTypes[notificationType]
	if isMandatory && preferredChannel == ChannelNone {
		preferredChannel = ChannelBoth // Enforce minimum transactional tier
	}

	if preferredChannel == ChannelNone {
		_ = e.repo.UpdateNotificationLogChannel(ctx, nLog.ID, string(ChannelNone))
		logger.Info("[NotificationEngine] Notification (%s - %s) suppressed by preference/DND rule for user %s", notificationType, messageCategory, userID)
		return nil
	}

	var pushSuccess, waSuccess bool

	// 3. PUSH Send via FCM
	if preferredChannel == ChannelPush || preferredChannel == ChannelBoth {
		tokens, err := e.repo.GetActiveDeviceTokens(ctx, userID)
		if err == nil && len(tokens) > 0 {
			count, _ := e.fcmClient.SendMulticastPush(ctx, tokens, title, body, deepLink, func(tokenID string) {
				_ = e.repo.MarkDeviceTokenInactive(ctx, tokenID)
			})
			if count > 0 {
				pushSuccess = true
			}
		}
	}

	// 4. WHATSAPP Send via Meta API
	if (preferredChannel == ChannelWhatsApp || preferredChannel == ChannelBoth) && templateKey != "" {
		waConfig, err := e.repo.GetWhatsAppTemplateConfig(ctx, templateKey, "en")
		if err != nil || !waConfig.IsApproved {
			logger.Warn("[NotificationEngine] WhatsApp template %s not approved (or missing). Skipping WhatsApp send.", templateKey)
		} else if waConfig.MetaCategory == "MARKETING" && messageCategory != "PROMOTIONAL" {
			logger.Warn("[NotificationEngine] Meta category mismatch for template %s (requires PROMOTIONAL message category). Skipping WhatsApp send.", templateKey)
		} else {
			// In production, fetch customer phone number from auth-service.
			recipientPhone := userID // Mock/ID recipient
			err := e.whatsAppClient.SendTemplateMessage(ctx, recipientPhone, waConfig.WhatsAppTemplateName, waConfig.Language, messageCategory, templateParams)
			if err == nil {
				waSuccess = true
			}
		}
	}

	// 5. Update notification_log channel_sent
	channelSent := string(ChannelNone)
	if pushSuccess && waSuccess {
		channelSent = string(ChannelBoth)
	} else if pushSuccess {
		channelSent = string(ChannelPush)
	} else if waSuccess {
		channelSent = string(ChannelWhatsApp)
	}

	_ = e.repo.UpdateNotificationLogChannel(ctx, nLog.ID, channelSent)
	return nil
}
