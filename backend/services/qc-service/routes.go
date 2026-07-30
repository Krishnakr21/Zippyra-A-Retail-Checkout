package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

func NewRouter(handler *ReviewHandler, jwtSecret string) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	internal := r.PathPrefix("/v1/qc/internal").Subrouter()
	internal.Use(systemJWTAuthMiddleware(jwtSecret))

	internal.HandleFunc("/reviews", handler.CreateReviewHandler).Methods(http.MethodPost)
	internal.HandleFunc("/reviews/{grn_id}", handler.GetReviewHandler).Methods(http.MethodGet)
	internal.HandleFunc("/reviews/{grn_id}", handler.UpdateReviewHandler).Methods(http.MethodPut)
	internal.HandleFunc("/reviews/{grn_id}/is-complete", handler.IsCompleteHandler).Methods(http.MethodGet)

	return r
}

func systemJWTAuthMiddleware(jwtSecret string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				// Header fallback for testing and inter-service calls
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
