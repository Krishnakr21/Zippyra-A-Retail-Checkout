package main

import (
	"encoding/json"
	"time"
)

type SubscriptionPlan struct {
	ID              string          `json:"id"`
	ChainID         string          `json:"chain_id"`
	Name            string          `json:"name"`
	PricePaise      int64           `json:"price_paise"`
	BillingInterval string          `json:"billing_interval"` // MONTHLY | ANNUAL
	Benefits        json.RawMessage `json:"benefits"`         // e.g. {"loyalty_multiplier_bonus": 0.5}
	IsActive        bool            `json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
}

type BenefitsDTO struct {
	LoyaltyMultiplierBonus float64 `json:"loyalty_multiplier_bonus"`
	FreeDelivery           bool    `json:"free_delivery"`
}

type MemberSubscription struct {
	ID                     string     `json:"id"`
	UserID                 string     `json:"user_id"`
	PlanID                 string     `json:"plan_id"`
	Status                 string     `json:"status"` // PENDING, ACTIVE, CANCELLED, EXPIRED, PAST_DUE
	RazorpaySubscriptionID *string    `json:"razorpay_subscription_id,omitempty"`
	CurrentPeriodEnd       *time.Time `json:"current_period_end,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`

	Plan *SubscriptionPlan `json:"plan,omitempty"`
}

type SubscribeRequest struct {
	PlanID string `json:"plan_id"`
}

type SubscribeResponse struct {
	SubscriptionID         string     `json:"subscription_id"`
	PlanID                 string     `json:"plan_id"`
	Status                 string     `json:"status"`
	RazorpaySubscriptionID string     `json:"razorpay_subscription_id"`
	RazorpayKeyID          string     `json:"razorpay_key_id"`
	CurrentPeriodEnd       *time.Time `json:"current_period_end,omitempty"`
}

type UserBonusResponse struct {
	UserID                 string  `json:"user_id"`
	HasActiveSubscription  bool    `json:"has_active_subscription"`
	LoyaltyMultiplierBonus float64 `json:"loyalty_multiplier_bonus"`
}

type RazorpaySubWebhookPayload struct {
	Event     string `json:"event"`
	EventID   string `json:"event_id"`
	CreatedAt int64  `json:"created_at"`
	Payload   struct {
		Subscription struct {
			Entity struct {
				ID             string `json:"id"`
				PlanID         string `json:"plan_id"`
				Status         string `json:"status"`
				CurrentEnd     int64  `json:"current_end"`
				ChargeAt       int64  `json:"charge_at"`
				Notes          map[string]string `json:"notes"`
			} `json:"entity"`
		} `json:"subscription"`
	} `json:"payload"`
}
