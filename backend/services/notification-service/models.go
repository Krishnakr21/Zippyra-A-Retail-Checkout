package main

import (
	"time"
)

type UserType string

const (
	UserTypeCustomer UserType = "CUSTOMER"
	UserTypeStaff    UserType = "STAFF"
)

type Platform string

const (
	PlatformAndroid Platform = "ANDROID"
	PlatformIOS     Platform = "IOS"
)

type Channel string

const (
	ChannelPush     Channel = "PUSH"
	ChannelWhatsApp Channel = "WHATSAPP"
	ChannelBoth     Channel = "BOTH"
	ChannelNone     Channel = "NONE"
)

type NotificationType string

const (
	NotificationTypeOrderUpdates   NotificationType = "ORDER_UPDATES"
	NotificationTypeLoyaltyUpdates NotificationType = "LOYALTY_UPDATES"
	NotificationTypeMarketing      NotificationType = "MARKETING"
	NotificationTypeSecurityAlerts NotificationType = "SECURITY_ALERTS"
	NotificationTypePaymentRefund  NotificationType = "PAYMENT_REFUND"
)

var MandatoryNotificationTypes = map[NotificationType]bool{
	NotificationTypeSecurityAlerts: true,
	NotificationTypePaymentRefund:  true,
	NotificationTypeOrderUpdates:   true,
}

type DeviceToken struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	UserType   UserType   `json:"user_type"`
	FCMToken   string     `json:"fcm_token"`
	Platform   Platform   `json:"platform"`
	DeviceID   string     `json:"device_id"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type NotificationPreference struct {
	UserID           string           `json:"user_id"`
	UserType         UserType         `json:"user_type"`
	NotificationType NotificationType `json:"notification_type"`
	Channel          Channel          `json:"channel"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type NotificationLog struct {
	ID              string           `json:"id"`
	UserID          string           `json:"user_id"`
	UserType        UserType         `json:"user_type"`
	NotificationType NotificationType `json:"notification_type"`
	ChannelSent     string           `json:"channel_sent"`
	Title           string           `json:"title"`
	Body            string           `json:"body"`
	DeepLink        string           `json:"deep_link"`
	SourceEventType string           `json:"source_event_type"`
	SourceEventID   string           `json:"source_event_id"`
	ReadAt          *time.Time       `json:"read_at,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
}

type WhatsAppTemplateConfig struct {
	TemplateKey          string    `json:"template_key"`
	WhatsAppTemplateName string    `json:"whatsapp_template_name"`
	IsApproved           bool      `json:"is_approved"`
	MetaCategory         string    `json:"meta_category"` // UTILITY | MARKETING | AUTHENTICATION
	Language             string    `json:"language"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type DNDRegistryCache struct {
	Phone           string    `json:"phone"`
	IsDNDRegistered bool      `json:"is_dnd_registered"`
	CheckedAt       time.Time `json:"checked_at"`
}

type OpsAlertChannel struct {
	ID          string    `json:"id"`
	ChannelType string    `json:"channel_type"` // SLACK | EMAIL
	Target      string    `json:"target"`       // Webhook URL or Email address
	AlertTypes  []string  `json:"alert_types"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type RegisterDeviceTokenRequest struct {
	FCMToken string   `json:"fcm_token"`
	Platform Platform `json:"platform"`
	DeviceID string   `json:"device_id"`
}

type UpdatePreferenceRequest struct {
	NotificationType NotificationType `json:"notification_type"`
	Channel          Channel          `json:"channel"`
}
