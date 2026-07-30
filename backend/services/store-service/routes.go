package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
	"github.com/zippyra/backend/shared/middleware"
	"github.com/zippyra/backend/shared/versioning"
)

func SetupRoutes(customerHandler *StoreHandler, internalHandler *InternalAdminWriteHandler, selfManageHandler *SelfManageHandler) http.Handler {
	r := mux.NewRouter()

	// Health routes
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthz/ready", health.ReadyHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods(http.MethodGet)

	// ── Customer Endpoints (unchanged) ───────────────────────────────────────
	r.HandleFunc("/v1/store/home-banners", customerHandler.HandleGetHomeBanners).Methods(http.MethodGet)
	r.HandleFunc("/v1/store/nearby", customerHandler.HandleGetNearbyStores).Methods(http.MethodGet)
	r.HandleFunc("/v1/store/bind", customerHandler.HandleBindStore).Methods(http.MethodPost)
	r.HandleFunc("/v1/store/unbind", customerHandler.HandleUnbindStore).Methods(http.MethodPost)
	r.HandleFunc("/v1/store/session", customerHandler.HandleGetActiveSession).Methods(http.MethodGet)
	r.HandleFunc("/v1/store/{id}", customerHandler.HandleGetStoreDetail).Methods(http.MethodGet)

	// ── Internal Admin-Write Endpoints (SYSTEM JWT only — called by admin-store-service) ─
	r.HandleFunc("/v1/store/internal/admin-write/stores", internalHandler.HandleCreateStore).Methods(http.MethodPost)
	r.HandleFunc("/v1/store/internal/admin-write/stores", internalHandler.HandleListStores).Methods(http.MethodGet)
	r.HandleFunc("/v1/store/internal/admin-write/stores/{id}", internalHandler.HandleGetStore).Methods(http.MethodGet)
	r.HandleFunc("/v1/store/internal/admin-write/stores/{id}/geofence", internalHandler.HandleUpdateGeofence).Methods(http.MethodPut)
	r.HandleFunc("/v1/store/internal/admin-write/stores/{id}/hours", internalHandler.HandleUpdateHours).Methods(http.MethodPut)
	r.HandleFunc("/v1/store/internal/admin-write/stores/{id}/capacity", internalHandler.HandleUpdateCapacity).Methods(http.MethodPut)
	r.HandleFunc("/v1/store/internal/admin-write/stores/{id}/status", internalHandler.HandleUpdateStatus).Methods(http.MethodPut)
	r.HandleFunc("/v1/store/internal/admin-write/stores/{id}/payment-setup", internalHandler.HandleUpdatePaymentSetup).Methods(http.MethodPut)
	r.HandleFunc("/v1/store/internal/admin-write/stores/{id}/qr-tokens/rotate", internalHandler.HandleRotateQRTokens).Methods(http.MethodPost)
	r.HandleFunc("/v1/store/internal/admin-write/stores/{id}/qr-tokens", internalHandler.HandleGetQRTokens).Methods(http.MethodGet)

	// ── Self-Manage Endpoints (MANAGER JWT — store managers updating their OWN store) ─
	r.HandleFunc("/v1/store/self-manage/stores/{id}/hours", selfManageHandler.HandleUpdateHours).Methods(http.MethodPut)
	r.HandleFunc("/v1/store/self-manage/stores/{id}/capacity", selfManageHandler.HandleUpdateCapacity).Methods(http.MethodPut)

	// ── Legacy /v1/store/admin/* Retroactive Deprecation Shims ─────────────
	// 12-month sunset grace period for callers still referencing pre-extraction paths
	sunsetDate := time.Date(2027, 8, 2, 0, 0, 0, 0, time.UTC)
	migrationURL := "https://docs.zippyra.com/api/v2/admin-store-service"

	adminShim := versioning.Deprecated(sunsetDate, migrationURL)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMovedPermanently)
		_, _ = w.Write([]byte(`{"error":{"code":"ROUTE_DEPRECATED","message":"This route has been extracted to admin-store-service. Please use /v1/admin/stores endpoints."}}`))
	}))
	r.PathPrefix("/v1/store/admin/").Handler(adminShim)

	handler := middleware.MaxBytesMiddleware(1048576)(r)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.RecoverMiddleware(handler)

	return handler
}
