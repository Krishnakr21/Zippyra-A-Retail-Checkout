package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/logger"
)

type POHandler struct {
	repo Repository
}

func NewPOHandler(repo Repository) *POHandler {
	return &POHandler{repo: repo}
}

func (h *POHandler) CreatePOHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, _ := r.Context().Value("user_claims").(*jwt.SessionClaims)

	var req CreatePORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.StoreID == "" || req.VendorName == "" || len(req.Items) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id, vendor_name, and items are required", nil)
		return
	}

	// Chain / Store scoping check
	if claims != nil && claims.UserType != "SYSTEM" && claims.StoreID != "" && claims.StoreID != req.StoreID {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Cannot create PO for another store", nil)
		return
	}

	chainID := "chain-default-1"

	var createdBy *string
	if claims != nil {
		createdBy = &claims.UserID
	}

	po := &PurchaseOrder{
		StoreID:              req.StoreID,
		ChainID:              chainID,
		VendorName:           req.VendorName,
		Status:               POStatusDraft,
		Source:               POSourceManual,
		CreatedBy:            createdBy,
		ExpectedDeliveryDate: req.ExpectedDeliveryDate,
	}

	var items []POLineItem
	for _, itemReq := range req.Items {
		items = append(items, POLineItem{
			Barcode:       itemReq.Barcode,
			QtyOrdered:    itemReq.QtyOrdered,
			UnitCostPaise: itemReq.UnitCostPaise,
		})
	}

	if err := h.repo.CreatePO(ctx, po, items); err != nil {
		logger.Error("Failed to create PO: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create Purchase Order", nil)
		return
	}

	po.LineItems = items

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(po)
}

func (h *POHandler) SubmitPOHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	poID := vars["id"]

	if poID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "PO id is required", nil)
		return
	}

	existing, err := h.repo.GetPOByID(ctx, poID)
	if err != nil {
		logger.Error("Failed to get PO: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve PO", nil)
		return
	}

	if existing == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Purchase Order not found", nil)
		return
	}

	if existing.Status != POStatusDraft {
		errors.WriteError(w, http.StatusConflict, errors.CodePOAlreadySubmitted, "Purchase Order has already been submitted", nil)
		return
	}

	if err := h.repo.SubmitPO(ctx, poID); err != nil {
		if strings.Contains(err.Error(), "PO_ALREADY_SUBMITTED") {
			errors.WriteError(w, http.StatusConflict, errors.CodePOAlreadySubmitted, "Purchase Order has already been submitted", nil)
			return
		}
		logger.Error("Failed to submit PO: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to submit Purchase Order", nil)
		return
	}

	existing.Status = POStatusSubmitted

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"po_id":  poID,
		"status": POStatusSubmitted,
	})
}

func (h *POHandler) ListPOsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	storeID := r.URL.Query().Get("store_id")
	status := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id is required", nil)
		return
	}

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	pos, err := h.repo.ListPOs(ctx, storeID, status, pageSize, offset)
	if err != nil {
		logger.Error("Failed to list POs: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list Purchase Orders", nil)
		return
	}

	if pos == nil {
		pos = []PurchaseOrder{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items":     pos,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *POHandler) GetPOHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	poID := vars["id"]

	if poID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "PO id is required", nil)
		return
	}

	po, err := h.repo.GetPOByID(ctx, poID)
	if err != nil {
		logger.Error("Failed to get PO: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve PO", nil)
		return
	}

	if po == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeInvalidRequest, "Purchase Order not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(po)
}
