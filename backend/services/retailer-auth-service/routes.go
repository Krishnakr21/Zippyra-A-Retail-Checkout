package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
	"github.com/zippyra/backend/shared/middleware"
)

func RegisterRoutes(r *mux.Router, staffH *StaffHandler, authH *AuthHandler, shiftH *ShiftHandler, healthChecker func() bool) {
	// Health endpoints
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		if healthChecker != nil && !healthChecker() {
			http.Error(w, "NOT READY", http.StatusServiceUnavailable)
			return
		}
		health.ReadyHandler(w, r)
	}).Methods(http.MethodGet)
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods(http.MethodGet)

	// Auth Endpoints (Public / Gated)
	r.HandleFunc("/v1/retailer-auth/otp/send", authH.HandleSendOTP).Methods(http.MethodPost)
	r.HandleFunc("/v1/retailer-auth/otp/verify", authH.HandleVerifyOTP).Methods(http.MethodPost)
	r.HandleFunc("/v1/retailer-auth/pin/set", authH.HandleSetPin).Methods(http.MethodPost)
	r.HandleFunc("/v1/retailer-auth/pin/login", authH.HandlePinLogin).Methods(http.MethodPost)
	r.HandleFunc("/v1/retailer-auth/refresh", authH.HandleRefresh).Methods(http.MethodPost)
	r.HandleFunc("/v1/retailer-auth/logout", authH.HandleLogout).Methods(http.MethodPost)
	r.HandleFunc("/v1/retailer-auth/me", authH.HandleMe).Methods(http.MethodGet)
	r.HandleFunc("/v1/retailer-auth/recovery/request-manager-reset", staffH.HandleRequestManagerReset).Methods(http.MethodPost)

	// Staff Management Endpoints (MANAGER or ADMIN)
	r.HandleFunc("/v1/retailer-auth/staff", staffH.HandleCreateStaff).Methods(http.MethodPost)
	r.HandleFunc("/v1/retailer-auth/staff", staffH.HandleListStaff).Methods(http.MethodGet)
	r.HandleFunc("/v1/retailer-auth/staff/{id}", staffH.HandleUpdateStaff).Methods(http.MethodPut)
	r.HandleFunc("/v1/retailer-auth/staff/{id}", middleware.RequireStepUp(10*time.Minute)(http.HandlerFunc(staffH.HandleDeactivateStaff)).ServeHTTP).Methods(http.MethodDelete)

	// Shift Endpoints (STAFF / MANAGER)
	r.HandleFunc("/v1/retailer-auth/shift/start", shiftH.HandleStartShift).Methods(http.MethodPost)
	r.HandleFunc("/v1/retailer-auth/shift/end", shiftH.HandleEndShift).Methods(http.MethodPost)
	r.HandleFunc("/v1/retailer-auth/shift/current", shiftH.HandleGetCurrentShift).Methods(http.MethodGet)
	r.HandleFunc("/v1/retailer-auth/shift/history", shiftH.HandleGetShiftHistory).Methods(http.MethodGet)
}
