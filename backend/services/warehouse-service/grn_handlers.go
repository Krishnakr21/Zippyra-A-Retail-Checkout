package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/featureflags"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type GRNHandler struct {
	repo            Repository
	inventoryClient InventoryClient
	qcClient        QCClient
	producer        *kafka.Producer
}

func NewGRNHandler(repo Repository, inventoryClient InventoryClient, qcClient QCClient, producer *kafka.Producer) *GRNHandler {
	return &GRNHandler{
		repo:            repo,
		inventoryClient: inventoryClient,
		qcClient:        qcClient,
		producer:        producer,
	}
}

func (h *GRNHandler) CreateGRNHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, _ := r.Context().Value("user_claims").(*jwt.SessionClaims)

	var req CreateGRNRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.StoreID == "" || len(req.Items) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id and items are required", nil)
		return
	}

	receivedBy := "SYSTEM"
	if claims != nil {
		receivedBy = claims.UserID
	}

	// Validate PO if linked
	var po *PurchaseOrder
	if req.POID != nil && *req.POID != "" {
		p, err := h.repo.GetPOByID(ctx, *req.POID)
		if err != nil {
			logger.Error("Failed to fetch PO %s: %v", *req.POID, err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to validate Purchase Order", nil)
			return
		}
		if p == nil || p.StoreID != req.StoreID || (p.Status != POStatusSubmitted && p.Status != POStatusPartiallyReceived) {
			errors.WriteError(w, http.StatusConflict, errors.CodePONotReceivable, "Purchase Order is not in receivable status or belongs to another store", nil)
			return
		}
		po = p
	}

	// Determine initial status via feature flags (default true for qc_required)
	qcRequired := featureflags.IsEnabled(ctx, nil, nil, "qc_required", req.StoreID)
	initialStatus := GRNStatusDraft
	if qcRequired {
		initialStatus = GRNStatusQCPending
	}

	grn := &GoodsReceivedNote{
		POID:             req.POID,
		StoreID:          req.StoreID,
		ReceivedBy:       receivedBy,
		VendorInvoiceRef: req.VendorInvoiceRef,
		Status:           initialStatus,
	}

	var items []GRNLineItem
	for _, itemReq := range req.Items {
		var qtyExpected *int
		if po != nil {
			for _, poItem := range po.LineItems {
				if poItem.Barcode == itemReq.Barcode {
					val := poItem.QtyOrdered
					qtyExpected = &val
					break
				}
			}
		}

		items = append(items, GRNLineItem{
			Barcode:       itemReq.Barcode,
			QtyExpected:   qtyExpected,
			QtyReceived:   itemReq.QtyReceived,
			UnitCostPaise: itemReq.UnitCostPaise,
			QCStatus:      QCStatusPending,
		})
	}

	if err := h.repo.CreateGRN(ctx, grn, items); err != nil {
		logger.Error("Failed to create GRN: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create Goods Received Note", nil)
		return
	}

	grn.LineItems = items

	// Initialize QC review in qc-service if qcRequired
	if qcRequired && h.qcClient != nil {
		var qcPayloads []QCLineItemCreatePayload
		for _, item := range grn.LineItems {
			qcPayloads = append(qcPayloads, QCLineItemCreatePayload{
				GRNLineItemID: item.ID,
				Barcode:       item.Barcode,
				QtyReceived:   item.QtyReceived,
				QCStatus:      QCStatusPending,
			})
		}
		// Non-blocking warning if qc-service call fails during GRN creation
		if _, err := h.qcClient.CreateReview(ctx, grn.ID, grn.StoreID, qcPayloads); err != nil {
			logger.Warn("[GRNHandler] Failed to initialize QC review in qc-service for grn %s: %v (GRN created successfully)", grn.ID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(grn)
}

func (h *GRNHandler) UpdateQCHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	grnID := vars["id"]

	var req QCUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	grn, err := h.repo.GetGRNByID(ctx, grnID)
	if err != nil || grn == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "GRN not found", nil)
		return
	}

	if grn.Status == GRNStatusCompleted {
		errors.WriteError(w, http.StatusConflict, errors.CodeGRNAlreadyCompleted, "GRN is already completed and QC cannot be modified", nil)
		return
	}

	if h.qcClient != nil {
		var updates []QCLineItemUpdatePayload
		for _, u := range req.LineItemUpdates {
			updates = append(updates, QCLineItemUpdatePayload{
				GRNLineItemID: u.GRNLineItemID,
				QCStatus:      u.QCStatus,
				QCNote:        u.QCNote,
			})
		}
		if _, err := h.qcClient.UpdateReview(ctx, grnID, updates); err != nil {
			logger.Error("Failed to update QC review in qc-service: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update QC status", nil)
			return
		}
	} else {
		if err := h.repo.UpdateGRNQC(ctx, grnID, req.LineItemUpdates); err != nil {
			logger.Error("Failed to update GRN QC: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update QC status", nil)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"grn_id":  grnID,
		"updated": len(req.LineItemUpdates),
	})
}

func (h *GRNHandler) CompleteGRNHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	grnID := vars["id"]

	grn, err := h.repo.GetGRNByID(ctx, grnID)
	if err != nil || grn == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "GRN not found", nil)
		return
	}

	// Idempotency: If already completed, return success immediately without re-calling inventory-service!
	if grn.Status == GRNStatusCompleted {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"grn_id":    grnID,
			"status":    GRNStatusCompleted,
			"duplicate": true,
		})
		return
	}

	qcRequired := featureflags.IsEnabled(ctx, nil, nil, "qc_required", grn.StoreID)

	var pendingBarcodes []string
	var itemsToApply []GRNItemPayload
	receivedQtysForPO := make(map[string]int)

	if qcRequired {
		if h.qcClient != nil {
			// Query qc-service for current QC review status
			qcReview, err := h.qcClient.GetReview(ctx, grnID)
			if err != nil || qcReview == nil {
				logger.Error("Failed to fetch QC review from qc-service for grn %s: %v", grnID, err)
				errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to verify QC completion status", nil)
				return
			}

			for _, item := range qcReview.LineItems {
				if item.QCStatus == QCStatusPending {
					pendingBarcodes = append(pendingBarcodes, item.Barcode)
				} else if item.QCStatus == QCStatusPassed {
					itemsToApply = append(itemsToApply, GRNItemPayload{
						Barcode:       item.Barcode,
						QtyReceived:   item.QtyReceived,
						UnitCostPaise: 0, // Filled below from GRN line items if present
					})
					receivedQtysForPO[item.Barcode] += item.QtyReceived
				}
			}

			// Map UnitCostPaise from grn line items
			for i := range itemsToApply {
				for _, lineItem := range grn.LineItems {
					if lineItem.Barcode == itemsToApply[i].Barcode {
						itemsToApply[i].UnitCostPaise = lineItem.UnitCostPaise
						break
					}
				}
			}
		} else {
			for _, item := range grn.LineItems {
				if item.QCStatus == QCStatusPending {
					pendingBarcodes = append(pendingBarcodes, item.Barcode)
				} else if item.QCStatus == QCStatusPassed {
					itemsToApply = append(itemsToApply, GRNItemPayload{
						Barcode:       item.Barcode,
						QtyReceived:   item.QtyReceived,
						UnitCostPaise: item.UnitCostPaise,
					})
					receivedQtysForPO[item.Barcode] += item.QtyReceived
				}
			}
		}

		if len(pendingBarcodes) > 0 {
			errors.WriteError(w, http.StatusConflict, errors.CodeQCIncomplete, "QC is incomplete for barcodes: "+strings.Join(pendingBarcodes, ", "), map[string]interface{}{
				"pending_barcodes": pendingBarcodes,
			})
			return
		}
	} else {
		// Explicit branch: qc_required = false -> treat all items as PASSED
		for _, item := range grn.LineItems {
			itemsToApply = append(itemsToApply, GRNItemPayload{
				Barcode:       item.Barcode,
				QtyReceived:   item.QtyReceived,
				UnitCostPaise: item.UnitCostPaise,
			})
			receivedQtysForPO[item.Barcode] += item.QtyReceived
		}
	}

	// Call inventory-service internal endpoint
	applyResp, err := h.inventoryClient.ApplyGRN(ctx, grn.StoreID, grn.ID, itemsToApply)
	if err != nil {
		logger.Error("Failed to apply GRN to inventory-service: %v", err)
		// Leave GRN status as QC_PENDING for manual retry!
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to apply stock to inventory service: "+err.Error(), nil)
		return
	}

	// Update local GRN and PO status
	if err := h.repo.CompleteGRN(ctx, grnID, grn.POID, receivedQtysForPO); err != nil {
		logger.Error("Failed to mark GRN completed in repo: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to finalize GRN completion", nil)
		return
	}

	// Publish warehouse.grn_completed Kafka event
	payload := GRNCompletedPayload{
		GRNID:     grnID,
		POID:      grn.POID,
		StoreID:   grn.StoreID,
		Timestamp: time.Now(),
	}
	for _, item := range itemsToApply {
		payload.Items = append(payload.Items, struct {
			Barcode     string `json:"barcode"`
			QtyReceived int    `json:"qty_received"`
		}{
			Barcode:     item.Barcode,
			QtyReceived: item.QtyReceived,
		})
	}
	_ = h.producer.PublishEvent(ctx, TopicGRNCompleted, grnID, payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"grn_id":         grnID,
		"status":         GRNStatusCompleted,
		"items_applied":  applyResp.ItemsApplied,
		"items_requested": applyResp.ItemsRequested,
	})
}
