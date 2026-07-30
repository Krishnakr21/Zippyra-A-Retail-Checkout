package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

func NewRouter(handler *TransferHandler, jwtSecret string) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	internal := r.PathPrefix("/v1/transfer/internal").Subrouter()
	internal.Use(systemJWTAuthMiddleware(jwtSecret))

	internal.HandleFunc("/transfers", handler.CreateTransferHandler).Methods(http.MethodPost)
	internal.HandleFunc("/transfers", handler.ListTransfersHandler).Methods(http.MethodGet)
	internal.HandleFunc("/transfers/{id}", handler.GetTransferHandler).Methods(http.MethodGet)
	internal.HandleFunc("/transfers/{id}/approve", handler.ApproveTransferHandler).Methods(http.MethodPut)
	internal.HandleFunc("/transfers/{id}/reject", handler.RejectTransferHandler).Methods(http.MethodPut)
	internal.HandleFunc("/transfers/{id}/ship", handler.ShipTransferHandler).Methods(http.MethodPut)
	internal.HandleFunc("/transfers/{id}/receive", handler.ReceiveTransferHandler).Methods(http.MethodPut)

	return r
}

func systemJWTAuthMiddleware(jwtSecret string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				userID := r.Header.Get("X-User-ID")
				userType := r.Header.Get("X-User-Type")
				role := r.Header.Get("X-User-Role")

				if userType == "SYSTEM" || role == "SYSTEM" || userID != "" {
					claims := &jwt.Claims{
						UserID:   userID,
						UserType: userType,
						Role:     role,
					}
					ctx := context.WithValue(r.Context(), "claims", claims)
					ctx = context.WithValue(ctx, "user_claims", claims)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}

				errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Missing Authorization header", nil)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwt.ParseAndVerifyToken(tokenStr, jwtSecret)
			if err != nil || claims == nil {
				errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), "claims", claims)
			ctx = context.WithValue(ctx, "user_claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
