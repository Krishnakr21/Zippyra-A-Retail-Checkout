package main

import (
	"encoding/json"
	"net/http"
	"time"

	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/gorilla/mux"
)

func (h *OrderHandler) GetInternalOrderDetailHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || (claims.Role != RoleSystem && claims.Role != RoleAdmin) {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "System authentication required", nil)
		return
	}

	vars := mux.Vars(r)
	orderID := vars["id"]
	if orderID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "order id is required", nil)
		return
	}

	order, err := h.repo.GetOrderByID(r.Context(), orderID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeOrderNotFound, "Order not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) HandleLookupByPhoneLast4(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || (claims.UserType != "STAFF" && claims.Role != RoleAdmin && claims.Role != RoleSystem) {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "STAFF JWT required for customer lookup", nil)
		return
	}

	storeID := r.URL.Query().Get("store_id")
	phoneLast4 := r.URL.Query().Get("phone_last4")

	if storeID == "" || len(phoneLast4) != 4 {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "store_id and 4-digit phone_last4 are required", nil)
		return
	}

	if claims.UserType == "STAFF" && claims.StoreID != "" && claims.StoreID != storeID {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Access denied for requested store_id", nil)
		return
	}

	matches, err := h.repo.LookupCustomerByPhoneLast4(r.Context(), storeID, phoneLast4, 2*time.Hour)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to query customer records", nil)
		return
	}

	resp := CustomerLookupResponse{
		MatchType: "NONE",
	}

	if len(matches) == 1 {
		resp.MatchType = "SINGLE"
		match := matches[0]
		resp.Customer = &match
	} else if len(matches) > 1 {
		resp.MatchType = "MULTIPLE"
		resp.Candidates = matches
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
