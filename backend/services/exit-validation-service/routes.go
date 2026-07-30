package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, handler *ExitHandler, db *sql.DB, redisClient RedisClient, mqttClient MQTTClient) {
	// Liveness check
	r.HandleFunc("/healthz/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	// Readiness check (checks Postgres DB, Redis, AND active MQTT connection!)
	r.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "Database unavailable: "+err.Error(), http.StatusServiceUnavailable)
			return
		}

		if err := redisClient.Ping(ctx).Err(); err != nil {
			http.Error(w, "Redis unavailable: "+err.Error(), http.StatusServiceUnavailable)
			return
		}

		if mqttClient != nil && !mqttClient.IsConnected() {
			http.Error(w, "MQTT connection unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("READY"))
	}).Methods(http.MethodGet)

	// API Routes
	r.HandleFunc("/v1/exit/validate", handler.ValidateExitHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/exit/rfid-confirm", handler.RFIDConfirmHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/exit/status/{order_id}", handler.GetExitStatusHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/exit/staff-override", func(w http.ResponseWriter, r *http.Request) {
		handler.StaffOverrideHandler(w, r, nil)
	}).Methods(http.MethodPost)
	r.HandleFunc("/v1/exit/recent-attempts", handler.HandleGetRecentExitAttempts).Methods(http.MethodGet)
}
