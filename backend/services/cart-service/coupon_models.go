package main

import (
	"time"
)

type Coupon struct {
	ID                  string     `json:"id"`
	ChainID             string     `json:"chain_id"`
	StoreID             *string    `json:"store_id,omitempty"` // NULL = chain-wide
	Code                string     `json:"code"`
	DiscountType        string     `json:"discount_type"`        // PERCENT_OFF | FLAT_OFF
	DiscountValue       float64    `json:"discount_value"`       // percent (1-90) or flat paise
	MinCartValuePaise   int64      `json:"min_cart_value_paise"`
	MaxUses             *int       `json:"max_uses,omitempty"`
	MaxUsesPerCustomer  int        `json:"max_uses_per_customer"`
	CurrentUseCount     int        `json:"current_use_count"`
	ActiveFrom          time.Time  `json:"active_from"`
	ActiveUntil         *time.Time `json:"active_until,omitempty"`
	IsActive            bool       `json:"is_active"`
	CreatedBy           string     `json:"created_by"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type CouponRedemption struct {
	ID                string    `json:"id"`
	CouponID          string    `json:"coupon_id"`
	UserID            string    `json:"user_id"`
	CheckoutSessionID string    `json:"checkout_session_id"`
	RedeemedAt        time.Time `json:"redeemed_at"`
}

type CreateCouponRequest struct {
	ChainID            string     `json:"chain_id"`
	StoreID            *string    `json:"store_id,omitempty"`
	Code               string     `json:"code"`
	DiscountType       string     `json:"discount_type"`
	DiscountValue      float64    `json:"discount_value"`
	MinCartValuePaise  int64      `json:"min_cart_value_paise"`
	MaxUses            *int       `json:"max_uses,omitempty"`
	MaxUsesPerCustomer int        `json:"max_uses_per_customer"`
	ActiveFrom         *time.Time `json:"active_from,omitempty"`
	ActiveUntil        *time.Time `json:"active_until,omitempty"`
}

type UpdateCouponRequest struct {
	DiscountType       *string    `json:"discount_type,omitempty"`
	DiscountValue      *float64   `json:"discount_value,omitempty"`
	MinCartValuePaise  *int64     `json:"min_cart_value_paise,omitempty"`
	MaxUses            *int       `json:"max_uses,omitempty"`
	MaxUsesPerCustomer *int       `json:"max_uses_per_customer,omitempty"`
	ActiveFrom         *time.Time `json:"active_from,omitempty"`
	ActiveUntil        *time.Time `json:"active_until,omitempty"`
}

type CouponConfigJSON struct {
	ID                 string     `json:"id"`
	ChainID            string     `json:"chain_id"`
	StoreID            *string    `json:"store_id,omitempty"`
	Code               string     `json:"code"`
	DiscountType       string     `json:"discount_type"`
	DiscountValue      float64    `json:"discount_value"`
	MinCartValuePaise  int64      `json:"min_cart_value_paise"`
	MaxUses            *int       `json:"max_uses,omitempty"`
	MaxUsesPerCustomer int        `json:"max_uses_per_customer"`
	CurrentUseCount    int        `json:"current_use_count"`
	ActiveFrom         time.Time  `json:"active_from"`
	ActiveUntil        *time.Time `json:"active_until,omitempty"`
	IsActive           bool       `json:"is_active"`
}
