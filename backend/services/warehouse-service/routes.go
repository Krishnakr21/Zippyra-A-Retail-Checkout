package main

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/middleware"
)

func RegisterRoutes(
	r *mux.Router,
	poHandler *POHandler,
	grnHandler *GRNHandler,
	transferHandler *TransferHandler,
	db *sql.DB,
	jwtSecret string,
) {
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

	// API Subrouter (JWT Protected)
	api := r.PathPrefix("/v1/warehouse").Subrouter()
	api.Use(middleware.AuthMiddleware(jwtSecret))

	// Purchase Orders
	api.HandleFunc("/po", poHandler.CreatePOHandler).Methods(http.MethodPost)
	api.HandleFunc("/po", poHandler.ListPOsHandler).Methods(http.MethodGet)
	api.HandleFunc("/po/{id}", poHandler.GetPOHandler).Methods(http.MethodGet)
	api.HandleFunc("/po/{id}/submit", poHandler.SubmitPOHandler).Methods(http.MethodPut)

	// GRN & QC
	api.HandleFunc("/grn", grnHandler.CreateGRNHandler).Methods(http.MethodPost)
	api.HandleFunc("/grn/{id}/qc", grnHandler.UpdateQCHandler).Methods(http.MethodPut)
	api.HandleFunc("/grn/{id}/complete", grnHandler.CompleteGRNHandler).Methods(http.MethodPost)

	// Transfers
	api.HandleFunc("/transfers", transferHandler.ListChainTransfersHandler).Methods(http.MethodGet)
	api.HandleFunc("/transfer", transferHandler.CreateTransferHandler).Methods(http.MethodPost)
	api.HandleFunc("/transfer/{id}/approve", transferHandler.ApproveTransferHandler).Methods(http.MethodPut)
	api.HandleFunc("/transfer/{id}/reject", transferHandler.RejectTransferHandler).Methods(http.MethodPut)
	api.HandleFunc("/transfer/{id}/ship", transferHandler.ShipTransferHandler).Methods(http.MethodPut)
	api.HandleFunc("/transfer/{id}/receive", transferHandler.ReceiveTransferHandler).Methods(http.MethodPut)
}
