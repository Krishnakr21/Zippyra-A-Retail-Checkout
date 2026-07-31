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

func SetupRoutes(authH *AuthHandler, userH *UserManagementHandler, dashH *DashboardHandler, bulkH *BulkImportHandler) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/healthz/live", health.LiveHandler).Methods("GET")
	r.HandleFunc("/healthz/ready", health.ReadyHandler).Methods("GET")
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods("GET")

	// OTP Auth (Public)
	r.HandleFunc("/v1/chain-hq/otp/send", authH.HandleSendOTP).Methods("POST")
	r.HandleFunc("/v1/chain-hq/otp/verify", authH.HandleVerifyOTP).Methods("POST")

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

	// Auth & Profile
	r.HandleFunc("/v1/chain-hq/refresh", tokenMiddleware(authH.HandleRefresh)).Methods("POST")
	r.HandleFunc("/v1/chain-hq/logout", tokenMiddleware(authH.HandleLogout)).Methods("POST")
	r.HandleFunc("/v1/chain-hq/me", tokenMiddleware(authH.HandleMe)).Methods("GET")

	// User Management
	r.HandleFunc("/v1/chain-hq/users", tokenMiddleware(userH.HandleInviteUser)).Methods("POST")
	r.HandleFunc("/v1/chain-hq/users", tokenMiddleware(userH.HandleListUsers)).Methods("GET")
	r.HandleFunc("/v1/chain-hq/users/{id}", tokenMiddleware(userH.HandleDeactivateUser)).Methods("DELETE")

	// Admin Provisioning (Admin JWT + StepUp required)
	r.HandleFunc("/v1/chain-hq/admin/owner", tokenMiddleware(stepUpMiddleware(http.HandlerFunc(userH.HandleAdminProvisionOwner)).ServeHTTP)).Methods("POST")

	// Dashboard & Proxies
	r.HandleFunc("/v1/chain-hq/dashboard", tokenMiddleware(dashH.HandleDashboard)).Methods("GET")
	r.HandleFunc("/v1/chain-hq/stores", tokenMiddleware(dashH.HandleStoresProxy)).Methods("GET")
	r.HandleFunc("/v1/chain-hq/orders", tokenMiddleware(dashH.HandleOrdersProxy)).Methods("GET")
	r.HandleFunc("/v1/chain-hq/transfers", tokenMiddleware(dashH.HandleTransfersProxy)).Methods("GET")
	r.HandleFunc("/v1/chain-hq/catalog", tokenMiddleware(dashH.HandleCatalogProxy)).Methods("GET")
	r.HandleFunc("/v1/chain-hq/analytics/sales", tokenMiddleware(dashH.HandleAnalyticsSalesProxy)).Methods("GET")
	r.HandleFunc("/v1/chain-hq/analytics/chain-summary", tokenMiddleware(dashH.HandleAnalyticsChainSummaryProxy)).Methods("GET")

	// Bulk Catalog Import
	r.HandleFunc("/v1/chain-hq/catalog/bulk-import", tokenMiddleware(bulkH.HandleBulkImport)).Methods("POST")
	r.HandleFunc("/v1/chain-hq/catalog/bulk-import/{id}/status", tokenMiddleware(bulkH.HandleGetBulkImportStatus)).Methods("GET")

	h := middleware.MaxBytesMiddleware(1048576)(r)
	h = middleware.CORSMiddleware(h)
	h = middleware.RecoverMiddleware(h)

	return h
}
