package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
	"github.com/zippyra/backend/shared/middleware"
)

func SetupRoutes(handler *SubscriptionHandler) http.Handler {
	r := mux.NewRouter()

	// Health checks
	r.HandleFunc("/healthz/live", health.HealthHandler).Methods("GET")
	r.HandleFunc("/healthz/ready", health.HealthHandler).Methods("GET")

	// Public endpoints
	r.HandleFunc("/v1/subscription/plans", handler.HandleGetPlans).Methods("GET")
	r.HandleFunc("/v1/subscription/webhook/razorpay", handler.HandleRazorpayWebhook).Methods("POST")
	r.HandleFunc("/v1/subscription/internal/user-bonus", handler.HandleGetUserBonus).Methods("GET")

	// Protected customer endpoints
	r.HandleFunc("/v1/subscription/subscribe", handler.HandleSubscribe).Methods("POST")
	r.HandleFunc("/v1/subscription/mine", handler.HandleGetMine).Methods("GET")
	r.HandleFunc("/v1/subscription/cancel", handler.HandleCancel).Methods("POST")

	h := middleware.MaxBytesMiddleware(1024 * 1024)(r)
	h = middleware.CORSMiddleware(h)
	h = middleware.RecoverMiddleware(h)

	return h
}
