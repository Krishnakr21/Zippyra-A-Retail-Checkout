package main

import (
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/health"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, h *PaymentHandler, relay *OutboxRelay) {
	// Health endpoints
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		// Custom readiness check: verify outbox relay poll health (within last 10s)
		lastPoll := relay.LastPollTime()
		if time.Since(lastPoll) > 10*time.Second {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"UNHEALTHY","reason":"Outbox relay worker stuck"}`))
			return
		}
		health.ReadyHandler(w, r)
	}).Methods(http.MethodGet)

	// Public Webhook route (No JWT, HMAC-authenticated)
	r.HandleFunc("/v1/payment/webhook/razorpay", h.RazorpayWebhookHandler).Methods(http.MethodPost)

	// Protected Customer / Staff Routes
	r.HandleFunc("/v1/payment/initiate", h.InitiatePaymentHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/payment/status/{payment_id}", h.GetPaymentStatusHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/payment/split/estimate", h.EstimateSplitPaymentHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/payment/cash", h.CashPaymentHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/payment/internal/refund", h.InternalRefundHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/payment/internal/captured", h.InternalGetCapturedPaymentsHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/payment/internal/circuit-breaker-status", h.HandleCircuitBreakerStatus).Methods(http.MethodGet)
}
