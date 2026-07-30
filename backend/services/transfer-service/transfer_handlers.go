package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type TransferHandler struct {
	repo            Repository
	inventoryClient InventoryClient
	producer        *kafka.Producer
}

func NewTransferHandler(repo Repository, inventoryClient InventoryClient, producer *kafka.Producer) *TransferHandler {
	return &TransferHandler{
		repo:            repo,
		inventoryClient: inventoryClient,
		producer:        producer,
	}
}

func (h *TransferHandler) getClaims(r *http.Request) *jwt.Claims {
	if val := r.Context().Value("user_claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	if val := r.Context().Value("claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	return nil
}

func (h *TransferHandler) CreateTransferHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := h.getClaims(r)

	var req CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.SourceStoreID == "" || req.DestStoreID == "" || len(req.Items) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "source_store_id, dest_store_id, and items are required", nil)
		return
	}

	requestedBy := "SYSTEM"
	chainID := req.ChainID
	if chainID == "" {
		chainID = "chain-default-1"
	}

	if claims != nil {
		if claims.UserID != "" {
			requestedBy = claims.UserID
		}
		if claims.ChainID != "" {
			chainID = claims.ChainID
		}
	}
	if req.RequestedBy != "" {
		requestedBy = req.RequestedBy
	}

	// Cross-Chain transfer guard: source & dest stores must belong to caller's chain
	if claims != nil && claims.UserType != "SYSTEM" && claims.StoreID != "" {
		if claims.StoreID != req.SourceStoreID && claims.StoreID != req.DestStoreID {
			errors.WriteError(w, http.StatusForbidden, errors.CodeCrossChainTransferDenied, "Cannot create transfer between stores outside your assigned scope", nil)
			return
		}
	}

	transfer := &TransferOrder{
		SourceStoreID: req.SourceStoreID,
		DestStoreID:   req.DestStoreID,
		ChainID:       chainID,
		Status:        TransferStatusRequested,
		RequestedBy:   requestedBy,
	}

	var items []TransferLineItem
	for _, itemReq := range req.Items {
		items = append(items, TransferLineItem{
			Barcode:      itemReq.Barcode,
			QtyRequested: itemReq.QtyRequested,
		})
	}

	if err := h.repo.CreateTransfer(ctx, transfer, items); err != nil {
		logger.Error("Failed to create transfer: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create transfer order", nil)
		return
	}

	transfer.LineItems = items

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transfer)
}

func (h *TransferHandler) ListTransfersHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sourceID := q.Get("source_store_id")
	destID := q.Get("dest_store_id")
	chainID := q.Get("chain_id")
	statusFilter := q.Get("status")

	transfers, err := h.repo.ListTransfers(r.Context(), sourceID, destID, chainID, statusFilter)
	if err != nil {
		transfers = []*TransferOrder{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"transfers": transfers,
		"chain_id":  chainID,
	})
}

func (h *TransferHandler) GetTransferHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transferID := vars["id"]

	transfer, err := h.repo.GetTransferByID(r.Context(), transferID)
	if err != nil || transfer == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Transfer order not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transfer)
}

func (h *TransferHandler) ApproveTransferHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	transferID := vars["id"]

	transfer, err := h.repo.GetTransferByID(ctx, transferID)
	if err != nil || transfer == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Transfer order not found", nil)
		return
	}

	if transfer.Status != TransferStatusRequested {
		errors.WriteError(w, http.StatusConflict, errors.CodeInvalidRequest, "Transfer order is not in REQUESTED status", nil)
		return
	}

	if err := h.repo.UpdateTransferStatus(ctx, transferID, TransferStatusApproved, nil); err != nil {
		logger.Error("Failed to approve transfer: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to approve transfer order", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transfer_id": transferID,
		"status":      TransferStatusApproved,
	})
}

func (h *TransferHandler) RejectTransferHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	transferID := vars["id"]

	var req RejectTransferRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	transfer, err := h.repo.GetTransferByID(ctx, transferID)
	if err != nil || transfer == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Transfer order not found", nil)
		return
	}

	if transfer.Status != TransferStatusRequested {
		errors.WriteError(w, http.StatusConflict, errors.CodeInvalidRequest, "Transfer order is not in REQUESTED status", nil)
		return
	}

	reasonStr := req.Reason
	if err := h.repo.UpdateTransferStatus(ctx, transferID, TransferStatusRejected, &reasonStr); err != nil {
		logger.Error("Failed to reject transfer: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to reject transfer order", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transfer_id": transferID,
		"status":      TransferStatusRejected,
	})
}

func (h *TransferHandler) ShipTransferHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	transferID := vars["id"]

	var req ShipTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	transfer, err := h.repo.GetTransferByID(ctx, transferID)
	if err != nil || transfer == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Transfer order not found", nil)
		return
	}

	if transfer.Status != TransferStatusApproved {
		errors.WriteError(w, http.StatusConflict, errors.CodeInvalidRequest, "Transfer order must be APPROVED before shipping", nil)
		return
	}

	var itemsToShip []TransferItemPayload
	shippedQtys := make(map[string]int)

	for _, item := range req.Items {
		itemsToShip = append(itemsToShip, TransferItemPayload{
			Barcode: item.Barcode,
			Qty:     item.QtyShipped,
		})
		shippedQtys[item.Barcode] = item.QtyShipped
	}

	// Call inventory-service apply-transfer-out
	if h.inventoryClient != nil {
		err = h.inventoryClient.ApplyTransferOut(ctx, transfer.SourceStoreID, transferID, itemsToShip)
		if err != nil {
			if strings.Contains(err.Error(), "INSUFFICIENT_STOCK_FOR_TRANSFER") || strings.Contains(err.Error(), "409") {
				errors.WriteError(w, http.StatusConflict, errors.CodeInsufficientStockForTransfer, "Insufficient stock in source store for transfer: "+err.Error(), nil)
				return
			}
			logger.Error("Failed to apply transfer-out in inventory-service: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to apply transfer out: "+err.Error(), nil)
			return
		}
	}

	if err := h.repo.ShipTransfer(ctx, transferID, shippedQtys); err != nil {
		logger.Error("Failed to ship transfer in repo: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update transfer shipping status", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transfer_id": transferID,
		"status":      TransferStatusInTransit,
	})
}

func (h *TransferHandler) ReceiveTransferHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	transferID := vars["id"]

	var req ReceiveTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	transfer, err := h.repo.GetTransferByID(ctx, transferID)
	if err != nil || transfer == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Transfer order not found", nil)
		return
	}

	if transfer.Status != TransferStatusInTransit {
		errors.WriteError(w, http.StatusConflict, errors.CodeInvalidRequest, "Transfer order must be IN_TRANSIT before receiving", nil)
		return
	}

	var itemsToReceive []TransferItemPayload
	receivedQtys := make(map[string]int)

	for _, item := range req.Items {
		itemsToReceive = append(itemsToReceive, TransferItemPayload{
			Barcode: item.Barcode,
			Qty:     item.QtyReceived,
		})
		receivedQtys[item.Barcode] = item.QtyReceived
	}

	// Call inventory-service apply-transfer-in
	if h.inventoryClient != nil {
		err = h.inventoryClient.ApplyTransferIn(ctx, transfer.DestStoreID, transferID, itemsToReceive)
		if err != nil {
			logger.Error("Failed to apply transfer-in in inventory-service: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to apply transfer in: "+err.Error(), nil)
			return
		}
	}

	if err := h.repo.ReceiveTransfer(ctx, transferID, receivedQtys); err != nil {
		logger.Error("Failed to receive transfer in repo: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update transfer receiving status", nil)
		return
	}

	// Discrepancy Check: If any item qty_received < qty_shipped -> publish warehouse.transfer_discrepancy
	now := time.Now()
	for _, lineItem := range transfer.LineItems {
		shippedVal := 0
		if lineItem.QtyShipped != nil {
			shippedVal = *lineItem.QtyShipped
		}
		receivedVal := receivedQtys[lineItem.Barcode]
		if receivedVal < shippedVal {
			discPayload := TransferDiscrepancyPayload{
				TransferID:  transferID,
				Barcode:     lineItem.Barcode,
				QtyShipped:  shippedVal,
				QtyReceived: receivedVal,
				Timestamp:   now,
			}
			if h.producer != nil {
				_ = h.producer.PublishEvent(ctx, TopicTransferDiscrepancy, transferID, discPayload)
			}
			logger.Warn("[TRANSFER DISCREPANCY] Transfer %s Barcode %s: shipped %d > received %d", transferID, lineItem.Barcode, shippedVal, receivedVal)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transfer_id": transferID,
		"status":      TransferStatusReceived,
	})
}
