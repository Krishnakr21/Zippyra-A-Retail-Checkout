package main

import (
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/health"
	"github.com/zippyra/backend/shared/middleware"
)

func SetupRoutes(handler *AuthHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/auth/otp/send", handler.HandleSendOTP)
	mux.HandleFunc("/v1/auth/otp/verify", handler.HandleVerifyOTP)
	mux.HandleFunc("/v1/auth/oauth/google", handler.HandleGoogleOAuth)
	mux.HandleFunc("/v1/auth/refresh", handler.HandleRefreshToken)
	mux.HandleFunc("/v1/auth/logout", handler.HandleLogout)
	mux.HandleFunc("/v1/auth/me", handler.HandleMe)
	mux.HandleFunc("/v1/auth/me/name", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			middleware.AuthMiddleware(handler.jwtSecret)(http.HandlerFunc(handler.HandleUpdateName)).ServeHTTP(w, r)
			return
		}
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
	})
	mux.HandleFunc("/v1/auth/recovery-email/set", handler.HandleSetRecoveryEmail)
	mux.HandleFunc("/v1/auth/recovery-email/confirm", handler.HandleConfirmRecoveryEmail)
	mux.HandleFunc("/v1/auth/account-recovery/initiate", handler.HandleInitiateAccountRecovery)
	mux.HandleFunc("/v1/auth/account-recovery/confirm", handler.HandleConfirmAccountRecovery)
	mux.HandleFunc("/v1/auth/version-check", handler.HandleVersionCheck)
	mux.HandleFunc("/v1/auth/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.HandleGetUserSessions(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			handler.HandleRevokeAllOtherSessions(w, r)
			return
		}
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
	})
	mux.HandleFunc("/v1/auth/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			handler.HandleRevokeSessionByID(w, r)
			return
		}
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
	})

	// Admin endpoints
	mux.HandleFunc("/v1/auth/admin/users", handler.HandleAdminListUsers)
	mux.HandleFunc("/v1/auth/admin/users/", handler.HandleAdminGetUserDetail)

	// Admin app-versions update endpoint with JWT + Step-Up requirement
	mux.HandleFunc("/v1/auth/admin/app-versions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			middleware.AuthMiddleware(handler.jwtSecret)(
				middleware.RequireStepUp(10 * time.Minute)(
					http.HandlerFunc(handler.HandleUpdateAppVersion),
				),
			).ServeHTTP(w, r)
			return
		}
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
	})

	mux.HandleFunc("/.well-known/jwks.json", handler.HandleJWKS)
	mux.HandleFunc("/healthz/live", health.HealthHandler)

	// Wrap mux with max body (1MB), CORS, and recovery middlewares
	h := middleware.MaxBytesMiddleware(1024 * 1024)(mux)
	h = middleware.CORSMiddleware(h)
	h = middleware.RecoverMiddleware(h)

	return h
}
