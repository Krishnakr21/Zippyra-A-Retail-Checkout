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

func SetupRoutes(authHandler *AdminAuthHandler, mgmtHandler *AdminManagementHandler) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/healthz/live", health.LiveHandler).Methods("GET")
	r.HandleFunc("/healthz/ready", health.ReadyHandler).Methods("GET")
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods("GET")

	// Login with Google (Public)
	r.HandleFunc("/v1/admin-auth/login/google", authHandler.HandleGoogleLogin).Methods("POST")

	// Middleware helper for token validation
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

	// TOTP 2FA endpoints
	r.HandleFunc("/v1/admin-auth/totp/setup", tokenMiddleware(authHandler.HandleTOTPSetup)).Methods("POST")
	r.HandleFunc("/v1/admin-auth/totp/confirm", tokenMiddleware(authHandler.HandleTOTPConfirm)).Methods("POST")
	r.HandleFunc("/v1/admin-auth/totp/verify", tokenMiddleware(authHandler.HandleTOTPVerify)).Methods("POST")
	r.HandleFunc("/v1/admin-auth/step-up", tokenMiddleware(authHandler.HandleStepUp)).Methods("POST")
	r.HandleFunc("/v1/admin-auth/refresh", tokenMiddleware(authHandler.HandleRefresh)).Methods("POST")
	r.HandleFunc("/v1/admin-auth/logout", tokenMiddleware(authHandler.HandleLogout)).Methods("POST")
	r.HandleFunc("/v1/admin-auth/me", tokenMiddleware(authHandler.HandleMe)).Methods("GET")

	// Admin User Management endpoints
	r.HandleFunc("/v1/admin-auth/admin/users", tokenMiddleware(stepUpMiddleware(http.HandlerFunc(mgmtHandler.HandleCreateAdmin)).ServeHTTP)).Methods("POST")
	r.HandleFunc("/v1/admin-auth/admin/users", tokenMiddleware(mgmtHandler.HandleListAdmins)).Methods("GET")
	r.HandleFunc("/v1/admin-auth/admin/users/{id}/role", tokenMiddleware(stepUpMiddleware(http.HandlerFunc(mgmtHandler.HandleUpdateAdminRole)).ServeHTTP)).Methods("PUT")
	r.HandleFunc("/v1/admin-auth/admin/users/{id}", tokenMiddleware(stepUpMiddleware(http.HandlerFunc(mgmtHandler.HandleDeleteAdmin)).ServeHTTP)).Methods("DELETE")

	h := middleware.MaxBytesMiddleware(1048576)(r)
	h = middleware.CORSMiddleware(h)
	h = middleware.RecoverMiddleware(h)

	return h
}
