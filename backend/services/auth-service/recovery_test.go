package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockOTPManager struct{}

func (m *MockOTPManager) SendOTP(ctx context.Context, channel, identifier, ip string) (string, error) {
	return "123456", nil
}

func (m *MockOTPManager) VerifyOTP(ctx context.Context, channel, identifier, code string) error {
	if code == "123456" {
		return nil
	}
	return errors.New("invalid otp")
}

func TestAccountRecovery_NonexistentPhone_GenericResponseNoInfoLeak(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAuthHandler(repo, &MockOTPManager{}, &MockGoogleVerifier{})

	reqBody, _ := json.Marshal(map[string]string{
		"original_phone": "+919999999999",
		"new_phone":      "+918888888888",
	})

	req := httptest.NewRequest("POST", "/v1/auth/account-recovery/initiate", bytes.NewBuffer(reqBody))
	req.RemoteAddr = "192.168.1.100:1234"
	w := httptest.NewRecorder()

	handler.HandleInitiateAccountRecovery(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)

	if res["message"] != "If this matches an account, you will receive next steps." {
		t.Errorf("unexpected message: %s", res["message"])
	}
}

func TestAccountRecovery_RateLimiting_FourthAttemptReturns429(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAuthHandler(repo, &MockOTPManager{}, &MockGoogleVerifier{})

	ip := "203.0.113.45:5678"

	// First 3 attempts succeed
	for i := 1; i <= 3; i++ {
		reqBody, _ := json.Marshal(map[string]string{
			"original_phone": "+919000000001",
			"new_phone":      "+919000000002",
		})
		req := httptest.NewRequest("POST", "/v1/auth/account-recovery/initiate", bytes.NewBuffer(reqBody))
		req.RemoteAddr = ip
		w := httptest.NewRecorder()

		handler.HandleInitiateAccountRecovery(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("attempt %d expected 200 OK, got %d", i, w.Code)
		}
	}

	// 4th attempt from same IP must be rate limited (429)
	reqBody, _ := json.Marshal(map[string]string{
		"original_phone": "+919000000001",
		"new_phone":      "+919000000002",
	})
	req := httptest.NewRequest("POST", "/v1/auth/account-recovery/initiate", bytes.NewBuffer(reqBody))
	req.RemoteAddr = ip
	w := httptest.NewRecorder()

	handler.HandleInitiateAccountRecovery(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests on 4th attempt, got %d", w.Code)
	}
}

func TestAccountRecovery_Confirm_SuccessPhoneUpdatedSessionsRevoked(t *testing.T) {
	repo := NewMemoryRepository()
	otpMgr := &MockOTPManager{}
	handler := NewAuthHandler(repo, otpMgr, &MockGoogleVerifier{})

	// Create user with phone and verified recovery email
	u, _ := repo.CreateUserWithPhone(context.Background(), "+919111111111")
	_ = repo.SetRecoveryEmail(context.Background(), u.ID, "user@recovery.com")
	_ = repo.ConfirmRecoveryEmail(context.Background(), u.ID)

	// Create active session
	_ = repo.CreateSession(context.Background(), &AuthSession{
		ID:        "sess-001",
		UserID:    u.ID,
		DeviceID:  "dev-1",
		CreatedAt: time.Now(),
	})

	// Initiate recovery
	recReq := &AccountRecoveryRequest{
		ID:                 "rec-req-100",
		UserID:             u.ID,
		OriginalIdentifier: "+919111111111",
		NewIdentifier:      "+919222222222",
		VerificationMethod: "RECOVERY_EMAIL_OTP",
		Status:             "PENDING",
		CreatedAt:          time.Now(),
	}
	_ = repo.CreateRecoveryRequest(context.Background(), recReq)

	// Confirm recovery
	reqBody, _ := json.Marshal(map[string]string{
		"request_id": "rec-req-100",
		"otp":        "123456",
	})
	req := httptest.NewRequest("POST", "/v1/auth/account-recovery/confirm", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	handler.HandleConfirmAccountRecovery(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	// Verify phone was updated
	updatedUser, _ := repo.GetUserByID(context.Background(), u.ID)
	if updatedUser.Phone == nil || *updatedUser.Phone != "+919222222222" {
		t.Fatalf("expected phone to be updated to +919222222222, got %v", updatedUser.Phone)
	}

	// Verify session was revoked
	sessions, _ := repo.GetUserSessions(context.Background(), u.ID)
	if len(sessions) > 0 && sessions[0].RevokedAt == nil {
		t.Fatalf("expected active session to be revoked")
	}
}

func TestAccountRecovery_Confirm_TargetPhoneAlreadyInUse_Returns409(t *testing.T) {
	repo := NewMemoryRepository()
	otpMgr := &MockOTPManager{}
	handler := NewAuthHandler(repo, otpMgr, &MockGoogleVerifier{})

	// User 1 attempting recovery
	u1, _ := repo.CreateUserWithPhone(context.Background(), "+919111111111")
	_ = repo.SetRecoveryEmail(context.Background(), u1.ID, "user1@recovery.com")
	_ = repo.ConfirmRecoveryEmail(context.Background(), u1.ID)

	// User 2 already owns the target phone number
	_, _ = repo.CreateUserWithPhone(context.Background(), "+919333333333")

	recReq := &AccountRecoveryRequest{
		ID:                 "rec-req-200",
		UserID:             u1.ID,
		OriginalIdentifier: "+919111111111",
		NewIdentifier:      "+919333333333", // Conflicts with User 2
		VerificationMethod: "RECOVERY_EMAIL_OTP",
		Status:             "PENDING",
		CreatedAt:          time.Now(),
	}
	_ = repo.CreateRecoveryRequest(context.Background(), recReq)

	reqBody, _ := json.Marshal(map[string]string{
		"request_id": "rec-req-200",
		"otp":        "123456",
	})
	req := httptest.NewRequest("POST", "/v1/auth/account-recovery/confirm", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	handler.HandleConfirmAccountRecovery(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", w.Code)
	}
}
