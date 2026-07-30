package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	sharedErrors "github.com/zippyra/backend/shared/errors"
)

func (h *PaymentHandler) CashPaymentHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	// Role check: CASHIER only
	if claims.Role != RoleCashier && claims.Role != RoleStoreManager && claims.Role != RoleAdmin {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Cash payment requires CASHIER role", nil)
		return
	}

	var req CashPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if req.CheckoutSessionID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "checkout_session_id is required", nil)
		return
	}

	ctx := r.Context()

	// 1. Fetch trusted checkout session
	sess, err := h.cartService.FetchCheckoutSession(ctx, req.CheckoutSessionID)
	if err != nil {
		var apiErr *sharedErrors.APIError
		if errors.As(err, &apiErr) {
			sharedErrors.WriteError(w, http.StatusConflict, apiErr.Code, apiErr.Message, nil)
			return
		}
		sharedErrors.WriteError(w, http.StatusConflict, sharedErrors.CodeCheckoutSessionExpired, "Checkout session expired or invalid", nil)
		return
	}

	// 2. Validate cash collected >= total_paise
	if req.CashCollectedPaise < sess.TotalPaise {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInsufficientCash, "Insufficient cash collected", map[string]interface{}{
			"total_paise":          sess.TotalPaise,
			"cash_collected_paise": req.CashCollectedPaise,
		})
		return
	}

	changeDuePaise := req.CashCollectedPaise - sess.TotalPaise

	// 3. Prepare payment struct
	p := &Payment{
		CheckoutSessionID:    req.CheckoutSessionID,
		UserID:               sess.UserID,
		StoreID:              sess.StoreID,
		AmountPaise:          sess.TotalPaise,
		LoyaltyPointsUsed:    0,
		LoyaltyDiscountPaise: 0,
		PayableAmountPaise:   sess.TotalPaise,
		PaymentMethod:        MethodCash,
		Gateway:              GatewayCash,
		Status:               StatusCaptured,
	}

	// Prepare payment.confirmed outbox payload (identical shape to digital path)
	confirmedPayload := PaymentConfirmedPayload{
		PaymentID:          p.ID,
		CheckoutSessionID:  p.CheckoutSessionID,
		SessionID:          p.SessionID,
		UserID:             p.UserID,
		StoreID:            p.StoreID,
		AmountPaise:        p.AmountPaise,
		PayableAmountPaise: p.PayableAmountPaise,
		LoyaltyPointsUsed:  0,
		PaymentMethod:      MethodCash,
		Timestamp:          time.Now(),
	}
	payloadBytes, _ := json.Marshal(confirmedPayload)

	// 4. Single SQL transaction: Insert payment + outbox event
	if err := h.repo.InsertCashPaymentTx(ctx, p, "payment.confirmed", payloadBytes); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to record cash payment", nil)
		return
	}

	resp := CashPaymentResponse{
		PaymentID:     p.ID,
		Status:        StatusCaptured,
		ChangeDuePaise: changeDuePaise,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
