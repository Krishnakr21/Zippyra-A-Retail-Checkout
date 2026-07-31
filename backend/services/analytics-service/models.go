package main

import (
	"time"
)

const (
	StageSessionStarted   = "SESSION_STARTED"
	StageCheckoutInitiated = "CHECKOUT_INITIATED"
	StagePaymentConfirmed  = "PAYMENT_CONFIRMED"
	StageOrderCompleted    = "ORDER_COMPLETED"
	StageExitValidated     = "EXIT_VALIDATED"
)

var AllFunnelStages = []string{
	StageSessionStarted,
	StageCheckoutInitiated,
	StagePaymentConfirmed,
	StageOrderCompleted,
	StageExitValidated,
}

type SalesEvent struct {
	EventDate     time.Time `json:"event_date"`
	EventTime     time.Time `json:"event_time"`
	StoreID       string    `json:"store_id"`
	ChainID       string    `json:"chain_id"`
	OrderID       string    `json:"order_id"`
	TotalPaise    int64     `json:"total_paise"`
	DiscountPaise int64     `json:"discount_paise"`
	CGSTPaise     int64     `json:"cgst_paise"`
	SGSTPaise     int64     `json:"sgst_paise"`
	IGSTPaise     int64     `json:"igst_paise"`
	PaymentMethod string    `json:"payment_method"`
	ItemCount     uint16    `json:"item_count"`
}

type OrderItemEvent struct {
	EventDate      time.Time `json:"event_date"`
	OrderID        string    `json:"order_id"`
	StoreID        string    `json:"store_id"`
	ChainID        string    `json:"chain_id"`
	Barcode        string    `json:"barcode"`
	ProductName    string    `json:"product_name"`
	Qty            uint16    `json:"qty"`
	LineTotalPaise int64     `json:"line_total_paise"`
}

type FunnelEvent struct {
	EventDate time.Time `json:"event_date"`
	EventTime time.Time `json:"event_time"`
	StoreID   string    `json:"store_id"`
	SessionID string    `json:"session_id"`
	Stage     string    `json:"stage"`
}

type HourlyTransaction struct {
	EventDate        time.Time `json:"event_date"`
	Hour             uint8     `json:"hour"`
	DayOfWeek        uint8     `json:"day_of_week"` // 0=Sunday, 6=Saturday
	StoreID          string    `json:"store_id"`
	TransactionCount uint32    `json:"transaction_count"`
}

type SalesMetricPeriod struct {
	Period        string `json:"period"`
	RevenuePaise  int64  `json:"revenue_paise"`
	OrderCount    int64  `json:"order_count"`
	DiscountPaise int64  `json:"discount_paise"`
}

type TopProductItem struct {
	Barcode           string `json:"barcode"`
	ProductName       string `json:"product_name"`
	QtySold           int64  `json:"qty_sold"`
	TotalRevenuePaise int64  `json:"total_revenue_paise"`
}

type FunnelStageMetric struct {
	Stage                         string  `json:"stage"`
	SessionCount                  int64   `json:"session_count"`
	ConversionFromPreviousPercent float64 `json:"conversion_from_previous_percent"`
}

type PeakHourCell struct {
	DayOfWeek              uint8   `json:"day_of_week"`
	Hour                   uint8   `json:"hour"`
	AvgTransactionsPerWeek float64 `json:"avg_transactions_per_week"`
	RecommendedStaff       int     `json:"recommended_staff"`
}

type StoreSalesSummary struct {
	StoreID      string `json:"store_id"`
	StoreName    string `json:"store_name"`
	RevenuePaise int64  `json:"revenue_paise"`
	OrderCount   int64  `json:"order_count"`
}

type ChainSummaryResponse struct {
	TotalRevenuePaise int64               `json:"total_revenue_paise"`
	TotalOrders       int64               `json:"total_orders"`
	ByStore           []StoreSalesSummary `json:"by_store"`
}
