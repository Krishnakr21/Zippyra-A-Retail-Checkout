package main

import (
	"time"
)

// Payment Status constants
const (
	StatusInitiated       = "INITIATED"
	StatusPending         = "PENDING"
	StatusAuthorized      = "AUTHORIZED"
	StatusCaptured        = "CAPTURED"
	StatusFailed          = "FAILED"
	StatusRefundInitiated = "REFUND_INITIATED"
	StatusRefunded        = "REFUNDED"
)

// Gateways
const (
	GatewayRazorpay = "razorpay"
	GatewayCashfree = "cashfree"
	GatewayCash     = "cash"
)

// Payment Methods
const (
	MethodUPI        = "UPI"
	MethodCard       = "CARD"
	MethodNetbanking = "NETBANKING"
	MethodCash       = "CASH"
)

type Payment struct {
	ID                   string    `json:"id"`
	CheckoutSessionID    string    `json:"checkout_session_id"`
	SessionID            string    `json:"session_id,omitempty"`
	UserID               string    `json:"user_id"`
	StoreID              string    `json:"store_id"`
	AmountPaise          int64     `json:"amount_paise"`
	LoyaltyPointsUsed    int64     `json:"loyalty_points_used"`
	LoyaltyDiscountPaise int64     `json:"loyalty_discount_paise"`
	PayableAmountPaise   int64     `json:"payable_amount_paise"`
	PaymentMethod        string    `json:"payment_method"`
	Gateway              string    `json:"gateway"`
	GatewayOrderID       *string   `json:"gateway_order_id,omitempty"`
	GatewayPaymentID     *string   `json:"gateway_payment_id,omitempty"`
	Status               string    `json:"status"`
	FailureReason        *string   `json:"failure_reason,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type OutboxEvent struct {
	ID          string     `json:"id"`
	Topic       string     `json:"topic"`
	Payload     []byte     `json:"payload"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	RetryCount  int        `json:"retry_count"`
	CreatedAt   time.Time  `json:"created_at"`
}

type PaymentConfirmedPayload struct {
	PaymentID           string    `json:"payment_id"`
	CheckoutSessionID   string    `json:"checkout_session_id"`
	SessionID           string    `json:"session_id,omitempty"`
	UserID              string    `json:"user_id"`
	StoreID             string    `json:"store_id"`
	AmountPaise         int64     `json:"amount_paise"`
	PayableAmountPaise  int64     `json:"payable_amount_paise"`
	LoyaltyPointsUsed   int64     `json:"loyalty_points_used"`
	PaymentMethod       string    `json:"payment_method"`
	Timestamp           time.Time `json:"ts"`
}

type WebhookEvent struct {
	ID             string    `json:"id"`
	Gateway        string    `json:"gateway"`
	GatewayEventID string    `json:"gateway_event_id"`
	EventType      string    `json:"event_type"`
	RawPayload     []byte    `json:"raw_payload"`
	ProcessedAt    time.Time `json:"processed_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type Refund struct {
	ID              string     `json:"id"`
	PaymentID       string     `json:"payment_id"`
	AmountPaise     int64      `json:"amount_paise"`
	Reason          string     `json:"reason"`
	GatewayRefundID *string    `json:"gateway_refund_id,omitempty"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type InitiatePaymentRequest struct {
	CheckoutSessionID     string `json:"checkout_session_id"`
	PaymentMethod         string `json:"payment_method"`
	LoyaltyPointsToRedeem int64  `json:"loyalty_points_to_redeem"`
	PlayIntegrityToken    string `json:"play_integrity_token,omitempty"`
}

type InitiatePaymentResponse struct {
	PaymentID          string    `json:"payment_id"`
	Gateway            string    `json:"gateway"`
	GatewayOrderID     string    `json:"gateway_order_id"`
	GatewayKeyID       string    `json:"gateway_key_id"`
	PayableAmountPaise int64     `json:"payable_amount_paise"`
	ExpiresAt          time.Time `json:"expires_at"`
}

type CashPaymentRequest struct {
	CheckoutSessionID  string `json:"checkout_session_id"`
	CashCollectedPaise int64  `json:"cash_collected_paise"`
}

type CashPaymentResponse struct {
	PaymentID     string `json:"payment_id"`
	Status        string `json:"status"`
	ChangeDuePaise int64  `json:"change_due_paise"`
}

type SplitEstimateRequest struct {
	CheckoutSessionID     string `json:"checkout_session_id"`
	LoyaltyPointsToRedeem int64  `json:"loyalty_points_to_redeem"`
}

type SplitEstimateResponse struct {
	OriginalTotalPaise   int64 `json:"original_total_paise"`
	LoyaltyDiscountPaise int64 `json:"loyalty_discount_paise"`
	PayableAmountPaise   int64 `json:"payable_amount_paise"`
	PointsBalance        int64 `json:"points_balance"`
}

type InternalRefundRequest struct {
	PaymentID string `json:"payment_id"`
	Reason    string `json:"reason"`
}

type InternalCheckoutSessionResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	StoreID    string    `json:"store_id"`
	TotalPaise int64     `json:"total_paise"`
	ExpiresAt  time.Time `json:"expires_at"`
}
