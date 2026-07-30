package main

import (
	"time"
)

const (
	TransferStatusRequested = "REQUESTED"
	TransferStatusApproved  = "APPROVED"
	TransferStatusRejected  = "REJECTED"
	TransferStatusInTransit = "IN_TRANSIT"
	TransferStatusReceived  = "RECEIVED"

	TopicTransferDiscrepancy = "warehouse.transfer_discrepancy"
)

type TransferLineItem struct {
	ID           string `json:"id"`
	TransferID   string `json:"transfer_id"`
	Barcode      string `json:"barcode"`
	QtyRequested int    `json:"qty_requested"`
	QtyShipped   *int   `json:"qty_shipped,omitempty"`
	QtyReceived  *int   `json:"qty_received,omitempty"`
}

type TransferOrder struct {
	ID              string             `json:"id"`
	SourceStoreID   string             `json:"source_store_id"`
	DestStoreID     string             `json:"dest_store_id"`
	ChainID         string             `json:"chain_id"`
	Status          string             `json:"status"`
	RequestedBy     string             `json:"requested_by"`
	RejectionReason *string            `json:"rejection_reason,omitempty"`
	ShippedAt       *time.Time         `json:"shipped_at,omitempty"`
	ReceivedAt      *time.Time         `json:"received_at,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	LineItems       []TransferLineItem `json:"line_items"`
}

type TransferItemRequest struct {
	Barcode      string `json:"barcode"`
	QtyRequested int    `json:"qty_requested"`
}

type CreateTransferRequest struct {
	SourceStoreID string                `json:"source_store_id"`
	DestStoreID   string                `json:"dest_store_id"`
	ChainID       string                `json:"chain_id"`
	RequestedBy   string                `json:"requested_by,omitempty"`
	Items         []TransferItemRequest `json:"items"`
}

type ShipTransferItemRequest struct {
	Barcode    string `json:"barcode"`
	QtyShipped int    `json:"qty_shipped"`
}

type ShipTransferRequest struct {
	Items []ShipTransferItemRequest `json:"items"`
}

type ReceiveTransferItemRequest struct {
	Barcode     string `json:"barcode"`
	QtyReceived int    `json:"qty_received"`
}

type ReceiveTransferRequest struct {
	Items []ReceiveTransferItemRequest `json:"items"`
}

type RejectTransferRequest struct {
	Reason string `json:"reason"`
}

type TransferDiscrepancyPayload struct {
	TransferID  string    `json:"transfer_id"`
	Barcode     string    `json:"barcode"`
	QtyShipped  int       `json:"qty_shipped"`
	QtyReceived int       `json:"qty_received"`
	Timestamp   time.Time `json:"ts"`
}
