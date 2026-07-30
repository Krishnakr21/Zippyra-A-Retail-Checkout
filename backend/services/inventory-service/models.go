package main

import (
	"time"
)

const (
	MovementGRNReceived = "GRN_RECEIVED"
	MovementSale        = "SALE"
	MovementReturn      = "RETURN"
	MovementTransferOut = "TRANSFER_OUT"
	MovementTransferIn  = "TRANSFER_IN"
	MovementAdjustment  = "ADJUSTMENT"

	RefOrder      = "ORDER"
	RefGRN        = "GRN"
	RefTransfer   = "TRANSFER"
	RefManual     = "MANUAL"
	RefStockCount = "STOCK_COUNT"

	TopicStockUpdated   = "inventory.stock_updated"
	TopicLowStock       = "inventory.low_stock"
	TopicShrinkageAlert = "inventory.shrinkage_alert"
)

type StockLevel struct {
	StoreID          string    `json:"store_id"`
	Barcode          string    `json:"barcode"`
	OnHandQty        int64     `json:"on_hand_qty"`
	ReorderPoint     int       `json:"reorder_point"`
	ReorderQty       int       `json:"reorder_qty"`
	LowStockAlerted  bool      `json:"low_stock_alerted"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type StockMovement struct {
	ID            string    `json:"id"`
	StoreID       string    `json:"store_id"`
	Barcode       string    `json:"barcode"`
	MovementType  string    `json:"movement_type"`
	QtyDelta      int64     `json:"qty_delta"`
	ReferenceType *string   `json:"reference_type,omitempty"`
	ReferenceID   *string   `json:"reference_id,omitempty"`
	Note          *string   `json:"note,omitempty"`
	CreatedBy     *string   `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type StockCount struct {
	ID          string    `json:"id"`
	StoreID     string    `json:"store_id"`
	Barcode     string    `json:"barcode"`
	ExpectedQty int64     `json:"expected_qty"`
	CountedQty  int64     `json:"counted_qty"`
	VarianceQty int64     `json:"variance_qty"`
	CountedBy   string    `json:"counted_by"`
	CountedAt   time.Time `json:"counted_at"`
}

type ShrinkageDaily struct {
	ID               string    `json:"id"`
	StoreID          string    `json:"store_id"`
	Date             string    `json:"date"`
	TotalVarianceQty int64     `json:"total_variance_qty"`
	TotalExpectedQty int64     `json:"total_expected_qty"`
	ShrinkagePercent float64   `json:"shrinkage_percent"`
	ItemCount        int       `json:"item_count"`
	CreatedAt        time.Time `json:"created_at"`
}

// Request / Response DTOs
type AdjustStockRequest struct {
	StoreID  string `json:"store_id"`
	Barcode  string `json:"barcode"`
	QtyDelta int64  `json:"qty_delta"`
	Reason   string `json:"reason"` // DAMAGE | THEFT | ADMIN_ERROR | FOUND | OTHER
	Note     string `json:"note"`
}

type StockCountEntry struct {
	Barcode    string `json:"barcode"`
	CountedQty int64  `json:"counted_qty"`
}

type StockCountRequest struct {
	StoreID string            `json:"store_id"`
	Entries []StockCountEntry `json:"entries"`
}

type StockCountEntryResult struct {
	Barcode     string `json:"barcode"`
	ExpectedQty int64  `json:"expected_qty"`
	CountedQty  int64  `json:"counted_qty"`
	VarianceQty int64  `json:"variance_qty"`
}

type StockCountResponse struct {
	TotalCounted       int                     `json:"total_counted"`
	DiscrepanciesFound int                     `json:"discrepancies_found"`
	Results            []StockCountEntryResult `json:"results"`
}

type GRNItem struct {
	Barcode        string `json:"barcode"`
	QtyReceived    int64  `json:"qty_received"`
	UnitCostPaise  int64  `json:"unit_cost_paise"`
}

type ApplyGRNRequest struct {
	StoreID string    `json:"store_id"`
	GRNID   string    `json:"grn_id"`
	Items   []GRNItem `json:"items"`
}

type TransferItem struct {
	Barcode string `json:"barcode"`
	Qty     int64  `json:"qty"`
}

type ApplyTransferRequest struct {
	StoreID    string         `json:"store_id"`
	TransferID string         `json:"transfer_id"`
	Items      []TransferItem `json:"items"`
}

type LowStockItemResponse struct {
	Barcode      string `json:"barcode"`
	ProductName  string `json:"product_name"`
	OnHandQty    int64  `json:"on_hand_qty"`
	ReorderPoint int    `json:"reorder_point"`
	ReorderQty   int    `json:"reorder_qty"`
}

// Kafka Event Payloads
type InventoryStockUpdatedPayload struct {
	StoreID      string    `json:"store_id"`
	Barcode      string    `json:"barcode"`
	AvailableQty int64     `json:"available_qty"`
	Timestamp    time.Time `json:"ts"`
}

type LowStockPayload struct {
	StoreID      string    `json:"store_id"`
	Barcode      string    `json:"barcode"`
	CurrentQty   int64     `json:"current_qty"`
	ReorderPoint int       `json:"reorder_point"`
	ReorderQty   int       `json:"reorder_qty"`
	Timestamp    time.Time `json:"ts"`
}

type ShrinkageAlertPayload struct {
	StoreID          string    `json:"store_id"`
	Date             string    `json:"date"`
	ShrinkagePercent float64   `json:"shrinkage_percent"`
	Timestamp        time.Time `json:"ts"`
}

type OrderCompletedKafkaPayload struct {
	OrderID string `json:"order_id"`
	StoreID string `json:"store_id"`
	Items   []struct {
		Barcode string `json:"barcode"`
		Qty     int64  `json:"qty"`
	} `json:"items"`
	Timestamp time.Time `json:"ts"`
}

type OrderReturnedKafkaPayload struct {
	OrderID string `json:"order_id"`
	StoreID string `json:"store_id"`
	Items   []struct {
		Barcode string `json:"barcode"`
		Qty     int64  `json:"qty"`
	} `json:"items"`
	Timestamp time.Time `json:"ts"`
}
