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
	transferClient  TransferClient
	producer        *kafka.Producer
}

func NewTransferHandler(repo Repository, inventoryClient InventoryClient, producer *kafka.Producer) *TransferHandler {
	return &TransferHandler{
		repo:            repo,
		inventoryClient: inventoryClient,
		producer:        producer,
	}
}

func NewTransferHandlerWithClient(repo Repository, inventoryClient InventoryClient, transferClient TransferClient, producer *kafka.Producer) *TransferHandler {
	return &TransferHandler{
		repo:            repo,
		inventoryClient: inventoryClient,
		transferClient:  transferClient,
		producer:        producer,
	}
}

func (h *TransferHandler) CreateTransferHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, _ := r.Context().Value("user_claims").(*jwt.SessionClaims)

	var req CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.SourceStoreID == "" || req.DestStoreID == "" || len(req.Items) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "source_store_id, dest_store_id, and items are required", nil)
		return
	}

	// Cross-Chain transfer guard: source & dest stores must belong to caller's chain
	if claims != nil && claims.UserType != "SYSTEM" && claims.StoreID != "" {
		if claims.StoreID != req.SourceStoreID && claims.StoreID != req.DestStoreID {
			errors.WriteError(w, http.StatusForbidden, errors.CodeCrossChainTransferDenied, "Cannot create transfer between stores outside your assigned scope", nil)
			return
		}
	}

	if h.transferClient != nil {
		transfer, err := h.transferClient.CreateTransfer(ctx, req, claims)
		if err != nil {
			if strings.Contains(err.Error(), "CROSS_CHAIN_TRANSFER_DENIED") || strings.Contains(err.Error(), "403") {
				errors.WriteError(w, http.StatusForbidden, errors.CodeCrossChainTransferDenied, "Cannot create transfer between stores outside your assigned scope", nil)
				return
			}
			logger.Error("Failed to create transfer via transferClient: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create transfer order", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(transfer)
		return
	}

	// Fallback to local repo implementation if transferClient is nil
	requestedBy := "SYSTEM"
	chainID := "chain-default-1"
	if claims != nil {
		requestedBy = claims.UserID
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

func (h *TransferHandler) ApproveTransferHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	transferID := vars["id"]

	if h.transferClient != nil {
		res, err := h.transferClient.ApproveTransfer(ctx, transferID)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Transfer order not found", nil)
				return
			}
			if strings.Contains(err.Error(), "409") {
				errors.WriteError(w, http.StatusConflict, errors.CodeInvalidRequest, "Transfer order is not in REQUESTED status", nil)
				return
			}
			logger.Error("Failed to approve transfer via transferClient: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to approve transfer order", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

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

	if h.transferClient != nil {
		res, err := h.transferClient.RejectTransfer(ctx, transferID, req.Reason)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Transfer order not found", nil)
				return
			}
			if strings.Contains(err.Error(), "409") {
				errors.WriteError(w, http.StatusConflict, errors.CodeInvalidRequest, "Transfer order is not in REQUESTED status", nil)
				return
			}
			logger.Error("Failed to reject transfer via transferClient: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to reject transfer order", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

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

	if h.transferClient != nil {
		res, err := h.transferClient.ShipTransfer(ctx, transferID, req)
		if err != nil {
			if strings.Contains(err.Error(), "INSUFFICIENT_STOCK_FOR_TRANSFER") || strings.Contains(err.Error(), "409") {
				errors.WriteError(w, http.StatusConflict, errors.CodeInsufficientStockForTransfer, "Insufficient stock in source store for transfer: "+err.Error(), nil)
				return
			}
			if strings.Contains(err.Error(), "404") {
				errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Transfer order not found", nil)
				return
			}
			logger.Error("Failed to ship transfer via transferClient: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update transfer shipping status", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
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

	if h.transferClient != nil {
		res, err := h.transferClient.ReceiveTransfer(ctx, transferID, req)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Transfer order not found", nil)
				return
			}
			if strings.Contains(err.Error(), "409") {
				errors.WriteError(w, http.StatusConflict, errors.CodeInvalidRequest, "Transfer order must be IN_TRANSIT before receiving", nil)
				return
			}
			logger.Error("Failed to receive transfer via transferClient: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update transfer receiving status", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
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

func (h *TransferHandler) ListChainTransfersHandler(w http.ResponseWriter, r *http.Request) {
	chainID := r.URL.Query().Get("chain_id")
	statusFilter := r.URL.Query().Get("status")

	if h.transferClient != nil {
		transfers, err := h.transferClient.ListTransfers(r.Context(), chainID, statusFilter)
		if err != nil {
			transfers = []*TransferOrder{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"transfers": transfers,
			"chain_id":  chainID,
		})
		return
	}

	transfers, err := h.repo.ListTransfersByChainID(r.Context(), chainID, statusFilter)
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
