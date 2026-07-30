package main

import (
	"encoding/json"
	"time"
)

type Offer struct {
	ID                 string                 `json:"id"`
	ChainID            string                 `json:"chain_id"`
	StoreID            *string                `json:"store_id,omitempty"`
	Type               string                 `json:"type"`
	AppliesTo          string                 `json:"applies_to"`
	TargetIDs          []string               `json:"target_ids"`
	RuleConfig         map[string]interface{} `json:"rule_config"`
	MinCartValuePaise int64                  `json:"min_cart_value_paise"`
	MaxDiscountPaise  *int64                 `json:"max_discount_paise,omitempty"`
	Priority           int                    `json:"priority"`
	ActiveFrom         time.Time              `json:"active_from"`
	ActiveUntil        *time.Time             `json:"active_until,omitempty"`
	IsActive           bool                   `json:"is_active"`
	CreatedBy          string                 `json:"created_by"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type OfferResponse struct {
	Offer
	Scope string `json:"scope"` // "CHAIN_WIDE" | "STORE_SPECIFIC"
}

type CreateOfferRequest struct {
	ChainID            string                 `json:"chain_id"`
	StoreID            *string                `json:"store_id,omitempty"`
	Type               string                 `json:"type"`
	AppliesTo          string                 `json:"applies_to"`
	TargetIDs          []string               `json:"target_ids"`
	RuleConfig         map[string]interface{} `json:"rule_config"`
	MinCartValuePaise int64                  `json:"min_cart_value_paise"`
	MaxDiscountPaise  *int64                 `json:"max_discount_paise,omitempty"`
	Priority           *int                   `json:"priority,omitempty"`
	ActiveFrom         *time.Time             `json:"active_from,omitempty"`
	ActiveUntil        *time.Time             `json:"active_until,omitempty"`
}

type UpdateOfferRequest struct {
	Type               string                 `json:"type"`
	AppliesTo          string                 `json:"applies_to"`
	TargetIDs          []string               `json:"target_ids"`
	RuleConfig         map[string]interface{} `json:"rule_config"`
	MinCartValuePaise int64                  `json:"min_cart_value_paise"`
	MaxDiscountPaise  *int64                 `json:"max_discount_paise,omitempty"`
	Priority           int                    `json:"priority"`
	ActiveFrom         time.Time              `json:"active_from"`
	ActiveUntil        *time.Time             `json:"active_until,omitempty"`
	IsActive           *bool                  `json:"is_active,omitempty"`
}

type ToggleOfferRequest struct {
	IsActive bool `json:"is_active"`
}

type OfferRulesAuditRow struct {
	ID          string          `json:"id"`
	StoreID     string          `json:"store_id"`
	Ruleset     json.RawMessage `json:"ruleset"`
	ActivatedAt time.Time       `json:"activated_at"`
}
