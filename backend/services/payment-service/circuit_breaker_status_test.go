package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/jwt"
)

func TestCircuitBreakerStatus_ExposesStateWhenForcedOpen(t *testing.T) {
	cb := NewRollingCircuitBreaker(0.05, 30*time.Second)

	// Simulate failures to force circuit breaker open
	for i := 0; i < 6; i++ {
		cb.RecordResult(errors.New("Razorpay API Timeout"))
	}

	if cb.State() != StateOpen {
		t.Fatalf("Expected circuit breaker to be StateOpen after 6 errors")
	}

	handler := &PaymentHandler{
		circuitBreaker: cb,
		jwtSecret:      "zippyra-dev-jwt-secret-key-32bytes",
	}

	router := mux.NewRouter()
	RegisterRoutes(router, handler, NewOutboxRelay(nil, nil, 1*time.Second))

	// Generate ADMIN JWT token
	token, _ := jwt.GenerateToken(&jwt.Claims{UserID: "admin-1", Role: "ADMIN"}, "zippyra-dev-jwt-secret-key-32bytes", 30*time.Minute)

	req, _ := http.NewRequest("GET", "/v1/payment/internal/circuit-breaker-status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var status map[string]interface{}
	_ = json.NewDecoder(rr.Body).Decode(&status)

	if status["gateway"] != "razorpay" {
		t.Errorf("Expected gateway razorpay, got %v", status["gateway"])
	}
	if status["state"] != "OPEN" {
		t.Errorf("Expected state OPEN, got %v", status["state"])
	}
	if status["opened_at"] == "" {
		t.Errorf("Expected opened_at timestamp when state is OPEN")
	}
	if status["will_retry_at"] == "" {
		t.Errorf("Expected will_retry_at timestamp when state is OPEN")
	}
}
