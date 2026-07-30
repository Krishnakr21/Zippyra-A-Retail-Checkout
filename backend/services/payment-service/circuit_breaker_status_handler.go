package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

func (h *PaymentHandler) HandleCircuitBreakerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims.Role != "ADMIN" {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Admin role required", nil)
		return
	}

	var status map[string]interface{}
	if h.circuitBreaker != nil {
		status = h.circuitBreaker.GetStatus()
	} else {
		status = map[string]interface{}{
			"gateway":                 "razorpay",
			"state":                   "CLOSED",
			"error_rate_rolling_1min": 0.0,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}
