package main

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/middleware"
)

func RegisterRoutes(r *mux.Router, handler *InventoryHandler, internalHandler *InternalHandler, db *sql.DB, jwtSecret string) {
	// Liveness check
	r.HandleFunc("/healthz/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	// Readiness check
	r.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("DB Unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	// Startup check
	r.HandleFunc("/healthz/startup", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	// Staff / Manager / HQ endpoints (JWT protected)
	apiSub := r.PathPrefix("/v1/inventory").Subrouter()
	apiSub.Use(middleware.AuthMiddleware(jwtSecret))

	apiSub.HandleFunc("/stock", handler.GetStockHandler).Methods(http.MethodGet)
	apiSub.HandleFunc("/low-stock", handler.GetLowStockHandler).Methods(http.MethodGet)
	apiSub.HandleFunc("/adjust", handler.AdjustStockHandler).Methods(http.MethodPost)
	apiSub.HandleFunc("/stock-count", handler.StockCountHandler).Methods(http.MethodPost)
	apiSub.HandleFunc("/shrinkage-report", handler.GetShrinkageReportHandler).Methods(http.MethodGet)

	// Internal SYSTEM JWT endpoints (called by warehouse-service)
	internalSub := r.PathPrefix("/v1/inventory/internal").Subrouter()
	internalSub.Use(middleware.AuthMiddleware(jwtSecret))

	internalSub.HandleFunc("/apply-grn", internalHandler.ApplyGRNHandler).Methods(http.MethodPost)
	internalSub.HandleFunc("/apply-transfer-out", internalHandler.ApplyTransferOutHandler).Methods(http.MethodPost)
	internalSub.HandleFunc("/apply-transfer-in", internalHandler.ApplyTransferInHandler).Methods(http.MethodPost)
}
