package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/zippyra/backend/shared/jwt"
)

func TestStepUp_ValidCode_IssuesFreshStepUpAt(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAdminAuthHandler(repo, nil, nil)

	secret, _, _ := GenerateTOTPKey("stepup@zippyra.com")
	encrypted, _ := EncryptTOTPSecret(secret)
	now := time.Now().UTC()

	admin := &AdminUser{
		Email:               "stepup@zippyra.com",
		Name:                "StepUp Admin",
		Role:                RoleSuperAdmin,
		IsActive:            true,
		TOTPSecretEncrypted: encrypted,
		TOTPEnabledAt:       &now,
	}
	_ = repo.CreateAdmin(context.Background(), admin)

	oldStepUp := time.Now().Add(-20 * time.Minute).Unix()
	oldClaims := &jwt.Claims{
		AdminID:  admin.ID,
		Email:    admin.Email,
		Role:     admin.Role,
		UserType: "ADMIN",
		StepUpAt: oldStepUp,
	}

	code, _ := totp.GenerateCode(secret, time.Now().UTC())
	body, _ := json.Marshal(StepUpRequest{Code: code})

	req := httptest.NewRequest("POST", "/v1/admin-auth/step-up", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", oldClaims))
	w := httptest.NewRecorder()

	handler.HandleStepUp(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 on step-up, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	newAccessToken := resp["access_token"].(string)

	parsedClaims, err := jwt.ParseAndVerifyToken(newAccessToken, handler.jwtSecret)
	if err != nil {
		t.Fatalf("failed to parse step-up access token: %v", err)
	}

	if parsedClaims.StepUpAt <= oldStepUp {
		t.Fatalf("expected step_up_at to be renewed (greater than old %d), got %d", oldStepUp, parsedClaims.StepUpAt)
	}
}

func TestRefresh_PreservesOldStepUpAt(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAdminAuthHandler(repo, nil, nil)

	oldStepUp := time.Now().Add(-25 * time.Minute).Unix()
	claims := &jwt.Claims{
		AdminID:  "admin-999",
		Email:    "refresh@zippyra.com",
		Role:     RolePlatformAdmin,
		UserType: "ADMIN",
		StepUpAt: oldStepUp,
	}

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "mock-refresh-token"})
	req := httptest.NewRequest("POST", "/v1/admin-auth/refresh", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
	w := httptest.NewRecorder()

	handler.HandleRefresh(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	newAccessToken := resp["access_token"].(string)

	parsedClaims, err := jwt.ParseAndVerifyToken(newAccessToken, handler.jwtSecret)
	if err != nil {
		t.Fatalf("failed to parse refreshed token: %v", err)
	}

	if parsedClaims.StepUpAt != oldStepUp {
		t.Fatalf("expected refreshed token to preserve old step_up_at %d, got %d", oldStepUp, parsedClaims.StepUpAt)
	}
}
