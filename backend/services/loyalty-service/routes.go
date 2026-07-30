package main

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/middleware"
)

func RegisterRoutes(r *mux.Router, customerHandler *CustomerHandler, internalHandler *InternalHandler, db *sql.DB, jwtSecret string) {
	// Liveness check
	r.HandleFunc("/healthz/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Readiness check
	r.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("DB Unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Startup check
	r.HandleFunc("/healthz/startup", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Public endpoints
	r.HandleFunc("/v1/loyalty/tiers", customerHandler.GetTiersHandler).Methods("GET")

	// Customer JWT protected endpoints
	customerSub := r.PathPrefix("/v1/loyalty").Subrouter()
	customerSub.Use(middleware.AuthMiddleware(jwtSecret))
	customerSub.HandleFunc("/balance", customerHandler.GetBalanceHandler).Methods("GET")
	customerSub.HandleFunc("/history", customerHandler.GetHistoryHandler).Methods("GET")
	customerSub.HandleFunc("/referral-code", customerHandler.HandleGetReferralCode).Methods("GET")
	customerSub.HandleFunc("/referral/apply", customerHandler.HandleApplyReferral).Methods("POST")

	// Internal SYSTEM JWT endpoints (called by payment-service)
	internalSub := r.PathPrefix("/v1/loyalty/internal").Subrouter()
	internalSub.Use(middleware.AuthMiddleware(jwtSecret))
	internalSub.HandleFunc("/reserve", internalHandler.ReservePointsHandler).Methods("POST")
	internalSub.HandleFunc("/commit", internalHandler.CommitPointsHandler).Methods("POST")
	internalSub.HandleFunc("/release", internalHandler.ReleasePointsHandler).Methods("POST")
}
