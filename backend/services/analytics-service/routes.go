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

func SetupRoutes(h *AnalyticsHandler) http.Handler {
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

	// Analytics Endpoints
	r.HandleFunc("/v1/analytics/sales", tokenMiddleware(h.HandleGetSales)).Methods("GET")
	r.HandleFunc("/v1/analytics/top-products", tokenMiddleware(h.HandleGetTopProducts)).Methods("GET")
	r.HandleFunc("/v1/analytics/funnel", tokenMiddleware(h.HandleGetFunnel)).Methods("GET")
	r.HandleFunc("/v1/analytics/peak-hours", tokenMiddleware(h.HandleGetPeakHours)).Methods("GET")
	r.HandleFunc("/v1/analytics/chain-summary", tokenMiddleware(h.HandleGetChainSummary)).Methods("GET")

	handler := middleware.MaxBytesMiddleware(1048576)(r)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.RecoverMiddleware(handler)

	return handler
}
