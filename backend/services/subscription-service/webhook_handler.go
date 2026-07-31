package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
)

// POST /v1/subscription/webhook/razorpay
func (h *SubscriptionHandler) HandleRazorpayWebhook(w http.ResponseWriter, r *http.Request) {
	// 1. Read raw body BEFORE JSON parsing (HMAC requirement)
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Failed to read request body", nil)
		return
	}

	// 2. Verify HMAC Signature
	signature := r.Header.Get("X-Razorpay-Signature")
	if !h.verifyWebhookSignature(rawBody, signature) {
		logger.Warn("Razorpay subscription webhook HMAC verification failed")
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid HMAC signature", nil)
		return
	}

	// 3. Decode payload
	var payload RazorpaySubWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	rzpSubID := payload.Payload.Subscription.Entity.ID
	if rzpSubID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Missing subscription entity ID in payload", nil)
		return
	}

	eventID := payload.EventID
	if eventID == "" {
		eventID = rzpSubID + "_" + payload.Event
	}

	// 4. Map event status
	status := "ACTIVE"
	switch payload.Event {
	case "subscription.activated", "subscription.charged":
		status = "ACTIVE"
	case "subscription.cancelled":
		status = "CANCELLED"
	case "subscription.halted", "subscription.paused":
		status = "PAST_DUE"
	}

	var periodEnd *time.Time
	if payload.Payload.Subscription.Entity.CurrentEnd > 0 {
		t := time.Unix(payload.Payload.Subscription.Entity.CurrentEnd, 0)
		periodEnd = &t
	}

	// 5. Idempotently update database
	processed, err := h.repo.ProcessWebhookEventIdempotent(r.Context(), eventID, payload.Event, rzpSubID, status, periodEnd)
	if err != nil {
		logger.Error("Failed to process subscription webhook event %s: %v", eventID, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Database error processing webhook", nil)
		return
	}

	if !processed {
		logger.Info("Subscription webhook event %s already processed (idempotent skip)", eventID)
	} else {
		logger.Info("Subscription webhook processed event %s: subscription %s -> status %s", eventID, rzpSubID, status)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *SubscriptionHandler) verifyWebhookSignature(body []byte, signature string) bool {
	if h.webhookSecret == "" {
		// Accept signature in dev/test if secret is not set
		return true
	}
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(expectedMAC), []byte(signature)) == 1
}
