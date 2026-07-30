package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
	"github.com/zippyra/backend/shared/middleware"
)

func SetupRoutes(chainHandler *ChainHandler, storeHandler *StoreAdminHandler) http.Handler {
	r := mux.NewRouter()

	// Health
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthz/ready", health.ReadyHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods(http.MethodGet)

	// ── Chain Endpoints ──────────────────────────────────────────────────────
	r.HandleFunc("/v1/admin-store/chains", chainHandler.HandleCreateChain).Methods(http.MethodPost)
	r.HandleFunc("/v1/admin-store/chains", chainHandler.HandleListChains).Methods(http.MethodGet)
	r.HandleFunc("/v1/admin-store/chains/{id}", chainHandler.HandleGetChain).Methods(http.MethodGet)
	r.HandleFunc("/v1/admin-store/chains/{id}", chainHandler.HandleUpdateChain).Methods(http.MethodPut)
	// Step-up required for chain status changes (SUSPENDED transitions)
	r.Handle("/v1/admin-store/chains/{id}/status",
		middleware.RequireStepUp(10*time.Minute)(http.HandlerFunc(chainHandler.HandleUpdateChainStatus)),
	).Methods(http.MethodPut)

	// ── Store Admin Endpoints ────────────────────────────────────────────────
	r.HandleFunc("/v1/admin-store/stores", storeHandler.HandleCreateStore).Methods(http.MethodPost)
	r.HandleFunc("/v1/admin-store/stores", storeHandler.HandleListStores).Methods(http.MethodGet)
	r.HandleFunc("/v1/admin-store/stores/{id}", storeHandler.HandleGetStore).Methods(http.MethodGet)
	r.HandleFunc("/v1/admin-store/stores/{id}/geofence", storeHandler.HandleUpdateGeofence).Methods(http.MethodPut)
	r.HandleFunc("/v1/admin-store/stores/{id}/hours", storeHandler.HandleUpdateHours).Methods(http.MethodPut)
	r.HandleFunc("/v1/admin-store/stores/{id}/capacity", storeHandler.HandleUpdateCapacity).Methods(http.MethodPut)
	// Step-up required for INACTIVE / UNDER_MAINTENANCE status changes
	r.Handle("/v1/admin-store/stores/{id}/status",
		middleware.RequireStepUp(10*time.Minute)(http.HandlerFunc(storeHandler.HandleUpdateStatus)),
	).Methods(http.MethodPut)
	r.HandleFunc("/v1/admin-store/stores/{id}/payment-setup", storeHandler.HandleUpdatePaymentSetup).Methods(http.MethodPut)
	r.HandleFunc("/v1/admin-store/stores/{id}/qr-tokens/rotate", storeHandler.HandleRotateQRTokens).Methods(http.MethodPost)
	r.HandleFunc("/v1/admin-store/stores/{id}/qr-tokens", storeHandler.HandleGetQRTokens).Methods(http.MethodGet)

	handler := middleware.MaxBytesMiddleware(1048576)(r)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.RecoverMiddleware(handler)

	return handler
}
