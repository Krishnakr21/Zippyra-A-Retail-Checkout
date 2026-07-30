package main

import (
	"time"
)

type CartItem struct {
	Barcode            string `json:"barcode"`
	Name               string `json:"name"`
	Qty                int    `json:"qty"`
	PricePaiseSnapshot int64  `json:"price_paise_snapshot"`
	PricePaise         int64  `json:"price_paise"`
	LineTotalPaise     int64  `json:"line_total_paise"`
	HSNCode            string `json:"hsn_code"`
	CategoryID         string `json:"category_id,omitempty"`
}

type CartSummary struct {
	Items          []*CartItem `json:"items"`
	SubtotalPaise  int64       `json:"subtotal_paise"`
	DiscountPaise  int64       `json:"discount_paise"`
	AppliedOffers  []string    `json:"applied_offers"`
	CouponCode     string      `json:"coupon_code,omitempty"`
	CGSTPaise      int64       `json:"cgst_paise"`
	SGSTPaise      int64       `json:"sgst_paise"`
	IGSTPaise      int64       `json:"igst_paise"`
	TotalPaise     int64       `json:"total_paise"`
	ItemCount      int         `json:"item_count"`
}

type OfferRule struct {
	ID                 string     `json:"id"`
	Type               string     `json:"type"` // "PERCENT_OFF" | "FLAT_OFF" | "BOGO" | "CATEGORY_PERCENT_OFF"
	Value              float64    `json:"value"`
	AppliesTo          string     `json:"applies_to"` // "ALL" | "CATEGORY" | "BARCODE_LIST"
	TargetIDs          []string   `json:"target_ids,omitempty"`
	MinCartValuePaise int64      `json:"min_cart_value_paise"`
	MaxDiscountPaise  *int64     `json:"max_discount_paise,omitempty"`
	ActiveFrom         *time.Time `json:"active_from,omitempty"`
	ActiveUntil        *time.Time `json:"active_until,omitempty"`
}

type CouponRule struct {
	Code              string    `json:"code"`
	Type              string    `json:"type"` // "PERCENT_OFF" | "FLAT_OFF"
	Value             float64   `json:"value"`
	MinCartValuePaise int64     `json:"min_cart_value_paise"`
	MaxUses           int       `json:"max_uses"`
	CurrentUses       int       `json:"current_uses"`
	ExpiresAt         time.Time `json:"expires_at"`
}

type CheckoutSession struct {
	ID            string      `json:"id"`
	SessionID     string      `json:"session_id,omitempty"`
	UserID        string      `json:"user_id"`
	StoreID       string      `json:"store_id"`
	Items         []*CartItem `json:"items"`
	SubtotalPaise int64       `json:"subtotal_paise"`
	DiscountPaise int64       `json:"discount_paise"`
	CGSTPaise     int64       `json:"cgst_paise"`
	SGSTPaise     int64       `json:"sgst_paise"`
	IGSTPaise     int64       `json:"igst_paise"`
	TotalPaise    int64       `json:"total_paise"`
	CouponCode    string      `json:"coupon_code,omitempty"`
	SupplyType    string      `json:"supply_type"`
	Status        string      `json:"status"` // "PENDING" | "CONSUMED" | "EXPIRED" | "CANCELLED"
	CreatedAt     time.Time   `json:"created_at"`
	ExpiresAt     time.Time   `json:"expires_at"`
}

type ScanItemRequest struct {
	Barcode string `json:"barcode"`
	Qty     int    `json:"qty"`
}

type UpdateItemRequest struct {
	Qty int `json:"qty"`
}

type ApplyCouponRequest struct {
	Code string `json:"code"`
}
