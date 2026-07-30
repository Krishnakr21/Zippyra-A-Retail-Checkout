package main

import (
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/health"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, h *OrderHandler, relay *OutboxRelay) {
	// Health endpoints
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		// Custom readiness check: verify outbox relay worker poll health (within last 10s)
		lastPoll := relay.LastPollTime()
		if time.Since(lastPoll) > 10*time.Second {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"UNHEALTHY","reason":"Outbox relay worker stuck"}`))
			return
		}
		health.ReadyHandler(w, r)
	}).Methods(http.MethodGet)

	// Customer Order Routes
	r.HandleFunc("/v1/order/history", h.GetOrderHistoryHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/order/exit-token", h.GetExitTokenHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/order/{id}", h.GetOrderDetailHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/order/{id}/return", h.CreateReturnRequestHandler).Methods(http.MethodPost)

	// Staff / Retailer Order Routes
	r.HandleFunc("/v1/order/store", h.GetStoreOrdersHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/order/chain", h.GetChainOrdersHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/order/{id}/return/accept", h.AcceptReturnHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/order/{id}/return/reject", h.RejectReturnHandler).Methods(http.MethodPost)

	// Internal Service Routes
	r.HandleFunc("/v1/order/internal/lookup-by-phone-last4", h.HandleLookupByPhoneLast4).Methods(http.MethodGet)
	r.HandleFunc("/v1/order/internal/{id}", h.GetInternalOrderDetailHandler).Methods(http.MethodGet)
}
