package main

import (
	"time"
)

const (
	QCStatusPending  = "PENDING"
	QCStatusPassed   = "PASSED"
	QCStatusRejected = "REJECTED"

	OverallStatusPending  = "PENDING"
	OverallStatusComplete = "COMPLETE"
)

type QCLineItemSnapshot struct {
	GRNLineItemID string  `json:"grn_line_item_id"`
	Barcode       string  `json:"barcode"`
	QtyReceived   int     `json:"qty_received"`
	QCStatus      string  `json:"qc_status"` // PENDING | PASSED | REJECTED
	QCNote        *string `json:"qc_note,omitempty"`
}

type QCReview struct {
	ID            string               `json:"id"`
	GRNID         string               `json:"grn_id"`
	StoreID       string               `json:"store_id"`
	LineItems     []QCLineItemSnapshot `json:"line_items"`
	OverallStatus string               `json:"overall_status"` // PENDING | COMPLETE
	ReviewedBy    *string              `json:"reviewed_by,omitempty"`
	CompletedAt   *time.Time           `json:"completed_at,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type CreateReviewItemRequest struct {
	GRNLineItemID string `json:"grn_line_item_id"`
	Barcode       string `json:"barcode"`
	QtyReceived   int    `json:"qty_received"`
	QCStatus      string `json:"qc_status,omitempty"`
	QCNote        *string `json:"qc_note,omitempty"`
}

type CreateReviewRequest struct {
	GRNID     string                    `json:"grn_id"`
	StoreID   string                    `json:"store_id"`
	LineItems []CreateReviewItemRequest `json:"line_items"`
}

type QCLineItemUpdate struct {
	GRNLineItemID string  `json:"grn_line_item_id"`
	QCStatus      string  `json:"qc_status"`
	QCNote        *string `json:"qc_note,omitempty"`
}

type UpdateReviewRequest struct {
	LineItemUpdates []QCLineItemUpdate `json:"line_item_updates"`
}

type ReviewCompletionResponse struct {
	GRNID      string `json:"grn_id"`
	IsComplete bool   `json:"is_complete"`
}
