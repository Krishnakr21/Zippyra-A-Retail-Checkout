package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/jwt"
)

func TestRequireStepUp_FreshToken_AllowsAccess(t *testing.T) {
	middleware := RequireStepUp(10 * time.Minute)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	claims := &jwt.Claims{
		AdminID:  "admin-123",
		UserType: "ADMIN",
		StepUpAt: time.Now().Unix(),
	}

	req := httptest.NewRequest("POST", "/protected", nil)
	ctx := context.WithValue(req.Context(), "user_claims", claims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestRequireStepUp_StaleToken_Returns403StepUpRequired(t *testing.T) {
	middleware := RequireStepUp(10 * time.Minute)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Token from 15 minutes ago
	claims := &jwt.Claims{
		AdminID:  "admin-123",
		UserType: "ADMIN",
		StepUpAt: time.Now().Add(-15 * time.Minute).Unix(),
	}

	req := httptest.NewRequest("POST", "/protected", nil)
	ctx := context.WithValue(req.Context(), "user_claims", claims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestRequireStepUp_NonAdminUserType_Returns403(t *testing.T) {
	middleware := RequireStepUp(10 * time.Minute)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	claims := &jwt.Claims{
		UserID:   "user-123",
		UserType: "STAFF",
		StepUpAt: time.Now().Unix(),
	}

	req := httptest.NewRequest("POST", "/protected", nil)
	ctx := context.WithValue(req.Context(), "user_claims", claims)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}
