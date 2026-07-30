package main

import (
	"encoding/json"
	"net/http"

	sharedErrors "github.com/zippyra/backend/shared/errors"
)

func (h *PaymentHandler) InternalRefundHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.Role != RoleSystem {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "System authentication required", nil)
		return
	}

	var req InternalRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if req.PaymentID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "payment_id is required", nil)
		return
	}

	ctx := r.Context()

	payment, err := h.repo.GetPaymentByID(ctx, req.PaymentID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodePaymentNotFound, "Payment not found", nil)
		return
	}

	refundGateway := h.razorpay
	if payment.Gateway == GatewayCashfree {
		refundGateway = h.cashfree
	}

	gatewayRefundID, refundErr := refundGateway.InitiateRefund(ctx, payment.ID, payment.PayableAmountPaise, req.Reason)
	if refundErr != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Refund initiation failed at gateway", nil)
		return
	}

	refund := &Refund{
		PaymentID:       payment.ID,
		AmountPaise:     payment.PayableAmountPaise,
		Reason:          req.Reason,
		GatewayRefundID: &gatewayRefundID,
		Status:          StatusInitiated,
	}

	if err := h.repo.InsertRefund(ctx, refund); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to record refund", nil)
		return
	}

	resp := map[string]interface{}{
		"refund_id":         refund.ID,
		"payment_id":        payment.ID,
		"gateway_refund_id": gatewayRefundID,
		"status":            StatusInitiated,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
