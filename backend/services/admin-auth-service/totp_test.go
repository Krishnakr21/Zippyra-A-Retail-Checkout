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

func TestTOTPFlow_FirstLoginToFullToken_Success(t *testing.T) {
	repo := NewMemoryRepository()
	googleVal := &RealGoogleTokenValidator{}
	handler := NewAdminAuthHandler(repo, googleVal, nil)
	handler.allowedDomain = "zippyra.com"

	// 1. Pre-provision admin
	admin := &AdminUser{
		Email:    "newadmin@zippyra.com",
		Name:     "New Admin",
		Role:     RolePlatformAdmin,
		IsActive: true,
	}
	_ = repo.CreateAdmin(context.Background(), admin)

	// 2. Google Login -> setup_required: true
	body, _ := json.Marshal(GoogleLoginRequest{
		IDToken: "mock-google-token-newadmin@zippyra.com:sub-789",
	})
	req := httptest.NewRequest("POST", "/v1/admin-auth/login/google", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.HandleGoogleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var loginResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)
	if loginResp["setup_required"] != true {
		t.Fatalf("expected setup_required true, got %v", loginResp["setup_required"])
	}
	setupToken := loginResp["setup_token"].(string)

	// 3. Initiate TOTP Setup with setup_token
	setupClaims, _ := jwt.ParseAndVerifyToken(setupToken, handler.jwtSecret)
	reqSetup := httptest.NewRequest("POST", "/v1/admin-auth/totp/setup", nil)
	reqSetup = reqSetup.WithContext(context.WithValue(reqSetup.Context(), "user_claims", setupClaims))
	wSetup := httptest.NewRecorder()
	handler.HandleTOTPSetup(wSetup, reqSetup)

	if wSetup.Code != http.StatusOK {
		t.Fatalf("expected status 200 on setup, got %d", wSetup.Code)
	}

	var setupResp map[string]string
	_ = json.Unmarshal(wSetup.Body.Bytes(), &setupResp)
	secret := setupResp["secret_display"]
	if secret == "" {
		t.Fatalf("expected valid TOTP secret")
	}

	// 4. Generate valid TOTP code
	code, _ := totp.GenerateCode(secret, time.Now().UTC())

	// 5. Confirm TOTP
	bodyConfirm, _ := json.Marshal(TOTPConfirmRequest{Code: code})
	reqConfirm := httptest.NewRequest("POST", "/v1/admin-auth/totp/confirm", bytes.NewReader(bodyConfirm))
	reqConfirm = reqConfirm.WithContext(context.WithValue(reqConfirm.Context(), "user_claims", setupClaims))
	wConfirm := httptest.NewRecorder()
	handler.HandleTOTPConfirm(wConfirm, reqConfirm)

	if wConfirm.Code != http.StatusOK {
		t.Fatalf("expected status 200 on confirm, got %d", wConfirm.Code)
	}

	var confirmResp map[string]interface{}
	_ = json.Unmarshal(wConfirm.Body.Bytes(), &confirmResp)
	accessToken := confirmResp["access_token"].(string)
	if accessToken == "" {
		t.Fatalf("expected full access_token")
	}

	// Verify totp_enabled_at is set in repo
	updatedAdmin, _ := repo.GetAdminByEmail(context.Background(), "newadmin@zippyra.com")
	if updatedAdmin.TOTPEnabledAt == nil {
		t.Fatalf("expected TOTPEnabledAt to be set after confirm")
	}
}

func TestTOTPVerify_WrongCode5Times_LocksAccount(t *testing.T) {
	repo := NewMemoryRepository()
	googleVal := &RealGoogleTokenValidator{}
	handler := NewAdminAuthHandler(repo, googleVal, nil)

	secret, _, _ := GenerateTOTPKey("lockedadmin@zippyra.com")
	encrypted, _ := EncryptTOTPSecret(secret)
	now := time.Now().UTC()

	admin := &AdminUser{
		Email:               "lockedadmin@zippyra.com",
		Name:                "Locked Admin",
		Role:                RolePlatformAdmin,
		IsActive:            true,
		TOTPSecretEncrypted: encrypted,
		TOTPEnabledAt:       &now,
	}
	_ = repo.CreateAdmin(context.Background(), admin)

	verifyClaims := &jwt.Claims{
		AdminID:  admin.ID,
		Email:    admin.Email,
		Role:     admin.Role,
		UserType: "ADMIN_2FA_VERIFY",
	}

	// 4 wrong attempts -> 400 TOTP_INVALID
	for i := 0; i < 4; i++ {
		bodyWrong, _ := json.Marshal(TOTPVerifyRequest{Code: "000000"})
		req := httptest.NewRequest("POST", "/v1/admin-auth/totp/verify", bytes.NewReader(bodyWrong))
		req = req.WithContext(context.WithValue(req.Context(), "user_claims", verifyClaims))
		w := httptest.NewRecorder()
		handler.HandleTOTPVerify(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("attempt %d: expected 400, got %d", i+1, w.Code)
		}
	}

	// 5th wrong attempt -> 429 TOTP_LOCKED
	bodyFifth, _ := json.Marshal(TOTPVerifyRequest{Code: "000000"})
	reqFifth := httptest.NewRequest("POST", "/v1/admin-auth/totp/verify", bytes.NewReader(bodyFifth))
	reqFifth = reqFifth.WithContext(context.WithValue(reqFifth.Context(), "user_claims", verifyClaims))
	wFifth := httptest.NewRecorder()
	handler.HandleTOTPVerify(wFifth, reqFifth)

	if wFifth.Code != http.StatusTooManyRequests {
		t.Fatalf("5th attempt: expected status 429 TOTP_LOCKED, got %d", wFifth.Code)
	}

	// 6th attempt while locked -> 429 TOTP_LOCKED
	bodySixth, _ := json.Marshal(TOTPVerifyRequest{Code: "000000"})
	reqSixth := httptest.NewRequest("POST", "/v1/admin-auth/totp/verify", bytes.NewReader(bodySixth))
	reqSixth = reqSixth.WithContext(context.WithValue(reqSixth.Context(), "user_claims", verifyClaims))
	wSixth := httptest.NewRecorder()
	handler.HandleTOTPVerify(wSixth, reqSixth)

	if wSixth.Code != http.StatusTooManyRequests {
		t.Fatalf("6th attempt while locked: expected status 429, got %d", wSixth.Code)
	}
}
