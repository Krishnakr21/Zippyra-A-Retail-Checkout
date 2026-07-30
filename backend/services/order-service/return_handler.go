package main

import (
	"encoding/json"
	"net/http"
	"time"

	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/gorilla/mux"
)

func (h *OrderHandler) CreateReturnRequestHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	orderID := vars["id"]
	if orderID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "order id is required", nil)
		return
	}

	var input CreateReturnRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if len(input.ItemBarcodes) == 0 {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "item_barcodes cannot be empty", nil)
		return
	}

	ctx := r.Context()

	// 1. Fetch Order
	order, err := h.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeOrderNotFound, "Order not found", nil)
		return
	}

	if order.UserID != claims.UserID && claims.Role != RoleSystem && claims.Role != RoleAdmin {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeOrderNotFound, "Order not found", nil)
		return
	}

	// 2. Validate 24h Return Window Server-side
	if time.Since(order.CreatedAt) > 24*time.Hour {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeReturnWindowExpired, "Return window of 24 hours has expired for this order", nil)
		return
	}

	// 3. Query Returnable Flags
	flags, err := h.repo.GetReturnableFlags(ctx, orderID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to query item returnable flags", nil)
		return
	}

	flagMap := make(map[string]OrderItemReturnableFlag)
	for _, f := range flags {
		flagMap[f.Barcode] = f
	}

	orderItemMap := make(map[string]OrderItem)
	for _, item := range order.Items {
		orderItemMap[item.Barcode] = item
	}

	// Count return frequency per barcode requested
	requestedCount := make(map[string]int)
	for _, barcode := range input.ItemBarcodes {
		requestedCount[barcode]++
	}

	var returnItems []ReturnItem
	flagUpdates := make(map[string]int)

	for barcode, reqQty := range requestedCount {
		item, exists := orderItemMap[barcode]
		if !exists {
			sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Barcode "+barcode+" is not part of this order", nil)
			return
		}

		flag, flagExists := flagMap[barcode]
		if !flagExists || !flag.IsReturnable {
			sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeItemNotReturnable, "Item "+item.Name+" ("+barcode+") is not returnable", nil)
			return
		}

		if flag.ReturnedQty+reqQty > item.Qty {
			sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeReturnQtyExceeded, "Requested return quantity exceeds original purchased quantity for "+item.Name, nil)
			return
		}

		returnItems = append(returnItems, ReturnItem{
			Barcode: barcode,
			Qty:     reqQty,
			Reason:  input.Reason,
		})
		flagUpdates[barcode] = reqQty
	}

	returnReq := &ReturnRequest{
		OrderID: order.ID,
		UserID:  claims.UserID,
		StoreID: order.StoreID,
		Items:   returnItems,
		Status:  "PENDING_STAFF_REVIEW",
	}

	// Single SQL Transaction: insert return_requests, update returnable flags, update order status to RETURN_REQUESTED if first return
	err = h.repo.CreateReturnRequestTx(ctx, returnReq, flagUpdates, order.Status == StatusCompleted)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to record return request", nil)
		return
	}

	resp := map[string]interface{}{
		"return_request_id": returnReq.ID,
		"order_id":          order.ID,
		"status":            returnReq.Status,
		"message":           "Return request submitted for staff review",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
