package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/middleware"
)

func SetupRoutes(deviceH *DeviceHandler) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/healthz/live", health.LiveHandler).Methods("GET")
	r.HandleFunc("/healthz/ready", health.ReadyHandler).Methods("GET")
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods("GET")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}

	tokenMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				http.Error(w, `{"error":"Authorization token required"}`, http.StatusUnauthorized)
				return
			}
			claims, err := jwt.ParseAndVerifyToken(authHeader[7:], jwtSecret)
			if err != nil {
				http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "user_claims", claims)
			ctx = context.WithValue(ctx, "claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}

	stepUpMiddleware := middleware.RequireStepUp(jwt.STEP_UP_FRESHNESS_MINUTES)

	// Device Endpoints
	r.HandleFunc("/v1/device-mgmt/devices", tokenMiddleware(stepUpMiddleware(http.HandlerFunc(deviceH.HandleProvisionDevice)).ServeHTTP)).Methods("POST")
	r.HandleFunc("/v1/device-mgmt/devices", tokenMiddleware(deviceH.HandleListDevices)).Methods("GET")
	r.HandleFunc("/v1/device-mgmt/devices/{id}", tokenMiddleware(deviceH.HandleGetDevice)).Methods("GET")
	r.HandleFunc("/v1/device-mgmt/devices/{id}/decommission", tokenMiddleware(stepUpMiddleware(http.HandlerFunc(deviceH.HandleDecommissionDevice)).ServeHTTP)).Methods("PUT")
	r.HandleFunc("/v1/device-mgmt/devices/{id}/generate-pairing-code", tokenMiddleware(deviceH.HandleGeneratePairingCode)).Methods("POST")
	r.HandleFunc("/v1/device-mgmt/devices/pair", deviceH.HandlePairDevice).Methods("POST")
	r.HandleFunc("/v1/device-mgmt/devices/{id}/heartbeats", tokenMiddleware(deviceH.HandleGetHeartbeats)).Methods("GET")

	// Alert Endpoints
	r.HandleFunc("/v1/device-mgmt/alerts", tokenMiddleware(deviceH.HandleListAlerts)).Methods("GET")
	r.HandleFunc("/v1/device-mgmt/alerts/{id}/resolve", tokenMiddleware(deviceH.HandleResolveAlert)).Methods("PUT")

	h := middleware.MaxBytesMiddleware(1048576)(r)
	h = middleware.CORSMiddleware(h)
	h = middleware.RecoverMiddleware(h)

	return h
}
