package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
)

type InternalHandler struct {
	repo   Repository
	engine *MovementEngine
}

func NewInternalHandler(repo Repository, engine *MovementEngine) *InternalHandler {
	return &InternalHandler{
		repo:   repo,
		engine: engine,
	}
}

func (h *InternalHandler) ApplyGRNHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req ApplyGRNRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.StoreID == "" || req.GRNID == "" || len(req.Items) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id, grn_id, and items are required", nil)
		return
	}

	tx, err := h.repo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to begin GRN tx", nil)
		return
	}
	defer tx.Rollback()

	itemsApplied := 0
	noteStr := "GRN Goods Received"

	for _, item := range req.Items {
		applied, _, err := h.engine.ApplyMovement(
			ctx,
			tx,
			req.StoreID,
			item.Barcode,
			MovementGRNReceived,
			item.QtyReceived,
			RefGRN,
			req.GRNID,
			nil,
			&noteStr,
			true,
		)
		if err != nil {
			logger.Error("Failed to apply GRN item %s: %v", item.Barcode, err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to apply GRN item", nil)
			return
		}
		if applied {
			itemsApplied++
		}
	}

	if err := tx.Commit(); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to commit GRN application", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"applied":         true,
		"items_requested": len(req.Items),
		"items_applied":   itemsApplied,
	})
}

func (h *InternalHandler) ApplyTransferOutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req ApplyTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.StoreID == "" || req.TransferID == "" || len(req.Items) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id, transfer_id, and items are required", nil)
		return
	}

	tx, err := h.repo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to begin transfer-out tx", nil)
		return
	}
	defer tx.Rollback()

	noteStr := "Transfer Out"
	var failingBarcodes []string

	for _, item := range req.Items {
		_, _, err := h.engine.ApplyMovement(
			ctx,
			tx,
			req.StoreID,
			item.Barcode,
			MovementTransferOut,
			-item.Qty,
			RefTransfer,
			req.TransferID,
			nil,
			&noteStr,
			false, // allowNegative = false! Rollback if resulting stock goes negative
		)
		if err != nil {
			if strings.Contains(err.Error(), "INSUFFICIENT_STOCK") {
				failingBarcodes = append(failingBarcodes, item.Barcode)
			} else {
				logger.Error("Failed to apply transfer-out for %s: %v", item.Barcode, err)
				errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to process transfer out", nil)
				return
			}
		}
	}

	if len(failingBarcodes) > 0 {
		// Entire call rolls back!
		_ = tx.Rollback()
		errors.WriteError(w, http.StatusConflict, errors.CodeInsufficientStockForTransfer, "Transfer out failed: insufficient stock for barcodes: "+strings.Join(failingBarcodes, ", "), map[string]interface{}{
			"failing_barcodes": failingBarcodes,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to commit transfer out tx", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"applied":     true,
		"transfer_id": req.TransferID,
	})
}

func (h *InternalHandler) ApplyTransferInHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req ApplyTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.StoreID == "" || req.TransferID == "" || len(req.Items) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id, transfer_id, and items are required", nil)
		return
	}

	tx, err := h.repo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to begin transfer-in tx", nil)
		return
	}
	defer tx.Rollback()

	noteStr := "Transfer In"

	for _, item := range req.Items {
		_, _, err := h.engine.ApplyMovement(
			ctx,
			tx,
			req.StoreID,
			item.Barcode,
			MovementTransferIn,
			item.Qty,
			RefTransfer,
			req.TransferID,
			nil,
			&noteStr,
			true,
		)
		if err != nil {
			logger.Error("Failed to apply transfer-in for %s: %v", item.Barcode, err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to process transfer in", nil)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to commit transfer in tx", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"applied":     true,
		"transfer_id": req.TransferID,
	})
}
