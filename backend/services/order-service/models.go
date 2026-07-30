package main

import (
	"time"
)

const (
	StatusCreated        = "CREATED"
	StatusCompleted      = "COMPLETED"
	StatusCreationFailed = "CREATION_FAILED"
	StatusReturnRequested = "RETURN_REQUESTED"
	StatusReturned       = "RETURNED"
	StatusReturnRejected = "RETURN_REJECTED"

	TopicOrderCompleted      = "order.completed"
	TopicOrderCreationFailed = "order.creation_failed"
	TopicOrderReturned       = "order.returned"
	TopicOrderReturnRejected = "order.return_rejected"
)

type OrderItem struct {
	Barcode      string `json:"barcode"`
	Name         string `json:"name"`
	Qty          int    `json:"qty"`
	PricePaise   int64  `json:"price_paise"`
	HSNCode      string `json:"hsn_code"`
	IsReturnable bool   `json:"is_returnable"`
}

type Order struct {
	ID                 string      `json:"id"`
	PaymentID          string      `json:"payment_id"`
	SessionID          string      `json:"session_id,omitempty"`
	ChainID            string      `json:"chain_id,omitempty"`
	UserID             string      `json:"user_id"`
	StoreID            string      `json:"store_id"`
	Items              []OrderItem `json:"items"`
	SubtotalPaise      int64       `json:"subtotal_paise"`
	DiscountPaise      int64       `json:"discount_paise"`
	CGSTPaise          int64       `json:"cgst_paise"`
	SGSTPaise          int64       `json:"sgst_paise"`
	IGSTPaise          int64       `json:"igst_paise"`
	TotalPaise         int64       `json:"total_paise"`
	LoyaltyPointsUsed  int64       `json:"loyalty_points_used"`
	PaymentMethod      string      `json:"payment_method"`
	SupplyType         string      `json:"supply_type"`
	Status             string      `json:"status"`
	InvoiceS3Key       *string     `json:"invoice_s3_key,omitempty"`
	IRN                *string     `json:"irn,omitempty"`
	IRNAckNo           *string     `json:"irn_ack_no,omitempty"`
	IRNAckDate         *time.Time  `json:"irn_ack_date,omitempty"`
	IRNQRCode          *string     `json:"irn_qr_code,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	CompletedAt        *time.Time  `json:"completed_at,omitempty"`
}

type CustomerLookupMatch struct {
	CustomerID        string    `json:"customer_id"`
	FirstName         string    `json:"first_name"`
	PhoneMasked       string    `json:"phone_masked"`
	StoreID           string    `json:"store_id"`
	HasActiveSession  bool      `json:"has_active_session"`
	SessionID         string    `json:"session_id,omitempty"`
	ActiveOrderID     string    `json:"active_order_id,omitempty"`
	ActiveOrderStatus string    `json:"active_order_status,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type CustomerLookupResponse struct {
	MatchType  string                `json:"match_type"` // NONE, SINGLE, MULTIPLE
	Customer   *CustomerLookupMatch  `json:"customer,omitempty"`
	Candidates []CustomerLookupMatch `json:"candidates,omitempty"`
}

type OrderSummary struct {
	ID            string    `json:"id"`
	PaymentID     string    `json:"payment_id"`
	UserID        string    `json:"user_id"`
	StoreID       string    `json:"store_id"`
	StoreName     string    `json:"store_name"`
	TotalPaise    int64     `json:"total_paise"`
	PaymentMethod string    `json:"payment_method"`
	ItemCount     int       `json:"item_count"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type OrderItemReturnableFlag struct {
	OrderID      string `json:"order_id"`
	Barcode      string `json:"barcode"`
	IsReturnable bool   `json:"is_returnable"`
	ReturnedQty  int    `json:"returned_qty"`
}

type ReturnItem struct {
	Barcode string `json:"barcode"`
	Qty     int    `json:"qty"`
	Reason  string `json:"reason"`
}

type CreateReturnRequestInput struct {
	ItemBarcodes []string `json:"item_barcodes"`
	Reason       string   `json:"reason"`
}

type ReturnRequest struct {
	ID        string       `json:"id"`
	OrderID   string       `json:"order_id"`
	UserID    string       `json:"user_id"`
	StoreID   string       `json:"store_id"`
	Items     []ReturnItem `json:"items"`
	Status    string       `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
}

// PaymentConfirmedPayload receives event payload from Kafka topic payment.confirmed
type PaymentConfirmedPayload struct {
	PaymentID          string      `json:"payment_id"`
	CheckoutSessionID  string      `json:"checkout_session_id"`
	SessionID          string      `json:"session_id,omitempty"`
	ChainID            string      `json:"chain_id,omitempty"`
	UserID             string      `json:"user_id"`
	StoreID            string      `json:"store_id"`
	AmountPaise        int64       `json:"amount_paise"`
	PayableAmountPaise int64       `json:"payable_amount_paise"`
	LoyaltyPointsUsed  int64       `json:"loyalty_points_used"`
	PaymentMethod      string      `json:"payment_method"`
	Items              []OrderItem `json:"items,omitempty"`
	Timestamp          time.Time   `json:"timestamp"`
}

// OrderCompletedPayload published to Kafka topic order.completed
type OrderCompletedPayload struct {
	OrderID           string      `json:"order_id"`
	SessionID         string      `json:"session_id,omitempty"`
	ChainID           string      `json:"chain_id,omitempty"`
	UserID            string      `json:"user_id"`
	StoreID           string      `json:"store_id"`
	TotalPaise        int64       `json:"total_paise"`
	Items             []OrderItem `json:"items"`
	LoyaltyPointsUsed int64       `json:"loyalty_points_used"`
	PaymentMethod     string      `json:"payment_method"`
	Timestamp         time.Time   `json:"timestamp"`
}

// OrderCreationFailedPayload published to Kafka topic order.creation_failed (triggers payment refund)
type OrderCreationFailedPayload struct {
	PaymentID   string    `json:"payment_id"`
	UserID      string    `json:"user_id"`
	StoreID     string    `json:"store_id"`
	AmountPaise int64     `json:"amount_paise"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

type ExitTokenResponse struct {
	OrderID   string    `json:"order_id"`
	ExitToken string    `json:"exit_token"`
	ExpiresAt time.Time `json:"expires_at"`
}
