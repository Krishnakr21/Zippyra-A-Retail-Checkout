package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
)

type RazorpayWebhookPayload struct {
	Event     string `json:"event"`
	EventID   string `json:"event_id"`
	CreatedAt int64  `json:"created_at"`
	Contains  []string `json:"contains"`
	Payload   struct {
		Payment struct {
			Entity struct {
				ID          string `json:"id"`
				OrderID     string `json:"order_id"`
				Amount      int64  `json:"amount"`
				Status      string `json:"status"`
				Method      string `json:"method"`
				ErrorReason string `json:"error_reason"`
				Notes       map[string]string `json:"notes"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

func (h *PaymentHandler) RazorpayWebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Read raw body BEFORE JSON parsing (HMAC requirement)
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Failed to read request body", nil)
		return
	}

	// 2. HMAC Signature Validation
	signature := r.Header.Get("X-Razorpay-Signature")
	if !h.razorpay.VerifyWebhookSignature(rawBody, signature) {
		logger.Warn("Razorpay webhook HMAC signature verification failed")
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid HMAC signature", nil)
		return
	}

	// 3. Decode payload
	var payload RazorpayWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	eventID := payload.EventID
	if eventID == "" {
		eventID = payload.Payload.Payment.Entity.ID + "_" + payload.Event
	}

	// 4. Idempotency Check: INSERT INTO payment_webhook_events ON CONFLICT (gateway_event_id) DO NOTHING
	newInserted, err := h.repo.RecordWebhookEventIdempotent(ctx, GatewayRazorpay, eventID, payload.Event, rawBody)
	if err != nil {
		logger.Error("Failed to log webhook event: %v", err)
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Database error", nil)
		return
	}

	if !newInserted {
		// Event already processed! Return 200 OK immediately without repeating side-effects
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"already_processed"}`))
		return
	}

	// 5. Process event type
	orderID := payload.Payload.Payment.Entity.OrderID
	gatewayPaymentID := payload.Payload.Payment.Entity.ID

	if orderID == "" && payload.Payload.Payment.Entity.Notes != nil {
		orderID = payload.Payload.Payment.Entity.Notes["order_id"]
	}

	// Look up payment by session/order ID
	paymentID := payload.Payload.Payment.Entity.Notes["payment_id"]
	var payment *Payment

	if paymentID != "" {
		payment, _ = h.repo.GetPaymentByID(ctx, paymentID)
	}

	if payment == nil && orderID != "" {
		// Query payment by session ID or order ID
		payment, _ = h.repo.GetPaymentByID(ctx, orderID)
	}

	if payment != nil {
		switch payload.Event {
		case "payment.captured", "payment.authorized":
			// Commit reserved points if used
			if payment.LoyaltyPointsUsed > 0 {
				_ = h.loyaltyService.CommitReservedPoints(ctx, payment.UserID, payment.LoyaltyPointsUsed)
			}

			// Prepare payment.confirmed event payload
			confirmedPayload := PaymentConfirmedPayload{
				PaymentID:          payment.ID,
				CheckoutSessionID:  payment.CheckoutSessionID,
				UserID:             payment.UserID,
				StoreID:            payment.StoreID,
				AmountPaise:        payment.AmountPaise,
				PayableAmountPaise: payment.PayableAmountPaise,
				LoyaltyPointsUsed:  payment.LoyaltyPointsUsed,
				PaymentMethod:      payment.PaymentMethod,
				Timestamp:          time.Now(),
			}
			payloadBytes, _ := json.Marshal(confirmedPayload)

			// Single SQL transaction: UPDATE payments status + INSERT payment_outbox
			err := h.repo.CapturePaymentAndOutboxTx(ctx, payment.ID, gatewayPaymentID, "payment.confirmed", payloadBytes)
			if err != nil {
				logger.Error("Failed to capture payment in outbox transaction: %v", err)
			}

		case "payment.failed":
			_ = h.repo.FailPaymentAndReleaseTx(ctx, payment.ID, payload.Payload.Payment.Entity.ErrorReason)
			if payment.LoyaltyPointsUsed > 0 {
				_ = h.loyaltyService.ReleaseReservedPoints(ctx, payment.UserID, payment.LoyaltyPointsUsed)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
