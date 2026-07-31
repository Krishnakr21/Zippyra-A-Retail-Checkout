package main

import (
	"time"
)

const (
	POStatusDraft             = "DRAFT"
	POStatusSubmitted         = "SUBMITTED"
	POStatusPartiallyReceived = "PARTIALLY_RECEIVED"
	POStatusReceived          = "RECEIVED"
	POStatusCancelled          = "CANCELLED"

	POSourceManual     = "MANUAL"
	POSourceAutoReorder = "AUTO_REORDER"

	GRNStatusDraft     = "DRAFT"
	GRNStatusQCPending = "QC_PENDING"
	GRNStatusCompleted = "COMPLETED"

	QCStatusPending  = "PENDING"
	QCStatusPassed   = "PASSED"
	QCStatusRejected = "REJECTED"

	TransferStatusRequested  = "REQUESTED"
	TransferStatusApproved   = "APPROVED"
	TransferStatusInTransit  = "IN_TRANSIT"
	TransferStatusReceived   = "RECEIVED"
	TransferStatusRejected   = "REJECTED"

	TopicGRNCompleted        = "warehouse.grn_completed"
	TopicTransferDiscrepancy = "warehouse.transfer_discrepancy"
	TopicPOAutoCreated       = "warehouse.po_auto_created"
	TopicLowStock            = "inventory.low_stock"
)

type PurchaseOrder struct {
	ID                       string       `json:"id"`
	StoreID                  string       `json:"store_id"`
	ChainID                  string       `json:"chain_id"`
	VendorName               string       `json:"vendor_name"`
	Status                   string       `json:"status"`
	Source                   string       `json:"source"`
	CreatedBy                *string      `json:"created_by,omitempty"`
	AutoReorderItemBarcode   *string      `json:"auto_reorder_item_barcode,omitempty"`
	AutoReorderDate          *string      `json:"auto_reorder_date,omitempty"`
	ExpectedDeliveryDate     *string      `json:"expected_delivery_date,omitempty"`
	CreatedAt                time.Time    `json:"created_at"`
	SubmittedAt              *time.Time   `json:"submitted_at,omitempty"`
	CompletedAt              *time.Time   `json:"completed_at,omitempty"`
	LineItems                []POLineItem `json:"line_items,omitempty"`
}

type POLineItem struct {
	ID            string `json:"id"`
	POID          string `json:"po_id"`
	Barcode       string `json:"barcode"`
	QtyOrdered    int    `json:"qty_ordered"`
	UnitCostPaise int64  `json:"unit_cost_paise"`
	QtyReceived   int    `json:"qty_received"`
}

type GoodsReceivedNote struct {
	ID               string        `json:"id"`
	POID             *string       `json:"po_id,omitempty"`
	StoreID          string        `json:"store_id"`
	ReceivedBy       string        `json:"received_by"`
	VendorInvoiceRef *string       `json:"vendor_invoice_ref,omitempty"`
	Status           string        `json:"status"`
	CreatedAt        time.Time     `json:"created_at"`
	CompletedAt      *time.Time    `json:"completed_at,omitempty"`
	LineItems        []GRNLineItem `json:"line_items,omitempty"`
}

type GRNLineItem struct {
	ID            string  `json:"id"`
	GRNID         string  `json:"grn_id"`
	Barcode       string  `json:"barcode"`
	QtyExpected   *int    `json:"qty_expected,omitempty"`
	QtyReceived   int     `json:"qty_received"`
	UnitCostPaise int64   `json:"unit_cost_paise"`
	QCStatus      string  `json:"qc_status"`
	QCNote        *string `json:"qc_note,omitempty"`
}

type TransferOrder struct {
	ID              string             `json:"id"`
	SourceStoreID   string             `json:"source_store_id"`
	DestStoreID     string             `json:"dest_store_id"`
	ChainID         string             `json:"chain_id"`
	Status          string             `json:"status"`
	RequestedBy     string             `json:"requested_by"`
	RejectionReason *string            `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	ApprovedAt      *time.Time         `json:"approved_at,omitempty"`
	ShippedAt       *time.Time         `json:"shipped_at,omitempty"`
	ReceivedAt      *time.Time         `json:"received_at,omitempty"`
	LineItems       []TransferLineItem `json:"line_items,omitempty"`
}

type TransferLineItem struct {
	ID           string `json:"id"`
	TransferID   string `json:"transfer_id"`
	Barcode      string `json:"barcode"`
	QtyRequested int    `json:"qty_requested"`
	QtyShipped   *int   `json:"qty_shipped,omitempty"`
	QtyReceived  *int   `json:"qty_received,omitempty"`
}

// DTOs
type POLineItemRequest struct {
	Barcode       string `json:"barcode"`
	QtyOrdered    int    `json:"qty_ordered"`
	UnitCostPaise int64  `json:"unit_cost_paise"`
}

type CreatePORequest struct {
	StoreID              string              `json:"store_id"`
	VendorName           string              `json:"vendor_name"`
	ExpectedDeliveryDate *string             `json:"expected_delivery_date,omitempty"`
	Items                []POLineItemRequest `json:"items"`
}

type GRNItemRequest struct {
	Barcode       string `json:"barcode"`
	QtyReceived   int    `json:"qty_received"`
	UnitCostPaise int64  `json:"unit_cost_paise"`
}

type CreateGRNRequest struct {
	StoreID          string           `json:"store_id"`
	POID             *string          `json:"po_id,omitempty"`
	VendorInvoiceRef *string          `json:"vendor_invoice_ref,omitempty"`
	Items            []GRNItemRequest `json:"items"`
}

type QCUpdateItem struct {
	GRNLineItemID string  `json:"grn_line_item_id"`
	QCStatus      string  `json:"qc_status"` // PASSED | REJECTED
	QCNote        *string `json:"qc_note,omitempty"`
}

type QCUpdateRequest struct {
	LineItemUpdates []QCUpdateItem `json:"line_item_updates"`
}

type TransferItemRequest struct {
	Barcode      string `json:"barcode"`
	QtyRequested int    `json:"qty_requested"`
}

type CreateTransferRequest struct {
	SourceStoreID string                `json:"source_store_id"`
	DestStoreID   string                `json:"dest_store_id"`
	Items         []TransferItemRequest `json:"items"`
}

type RejectTransferRequest struct {
	Reason string `json:"reason"`
}

type ShipTransferItem struct {
	Barcode    string `json:"barcode"`
	QtyShipped int    `json:"qty_shipped"`
}

type ShipTransferRequest struct {
	Items []ShipTransferItem `json:"items"`
}

type ReceiveTransferItem struct {
	Barcode     string `json:"barcode"`
	QtyReceived int    `json:"qty_received"`
}

type ReceiveTransferRequest struct {
	Items []ReceiveTransferItem `json:"items"`
}

// Kafka Payloads
type GRNCompletedPayload struct {
	GRNID     string `json:"grn_id"`
	POID      *string `json:"po_id,omitempty"`
	StoreID   string `json:"store_id"`
	Items     []struct {
		Barcode     string `json:"barcode"`
		QtyReceived int    `json:"qty_received"`
	} `json:"items"`
	Timestamp time.Time `json:"ts"`
}

type TransferDiscrepancyPayload struct {
	TransferID  string    `json:"transfer_id"`
	Barcode     string    `json:"barcode"`
	QtyShipped  int       `json:"qty_shipped"`
	QtyReceived int       `json:"qty_received"`
	Timestamp   time.Time `json:"ts"`
}

type POAutoCreatedPayload struct {
	POID      string    `json:"po_id"`
	StoreID   string    `json:"store_id"`
	Barcode   string    `json:"barcode"`
	Qty       int       `json:"qty"`
	Timestamp time.Time `json:"ts"`
}

type LowStockKafkaPayload struct {
	StoreID      string    `json:"store_id"`
	Barcode      string    `json:"barcode"`
	CurrentQty   int64     `json:"current_qty"`
	ReorderPoint int       `json:"reorder_point"`
	ReorderQty   int       `json:"reorder_qty"`
	Timestamp    time.Time `json:"ts"`
}
