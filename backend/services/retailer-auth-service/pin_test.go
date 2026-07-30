package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/jwt"
)

func TestPINSet_RequiresFreshOTPSession_OrReturns403StepUpRequired(t *testing.T) {
	repo := NewMemoryRepository()
	pinSvc := NewPINService(repo)
	authH := NewAuthHandler(repo, nil, pinSvc, "dev-secret")

	staff := &StaffMember{
		ID:      "staff-pin-1",
		StoreID: "store-1",
		ChainID: "chain-1",
		Phone:   "+919876500001",
		Name:    "PIN User",
		Role:    "CASHIER",
	}
	_ = repo.CreateStaffMember(context.Background(), staff)

	// Session 1: auth_method = PIN (Not OTP)
	sessPIN := &StaffSession{
		ID:         "sess-pin-1",
		StaffID:    staff.ID,
		AuthMethod: AuthMethodPIN,
	}
	_ = repo.CreateSession(context.Background(), sessPIN)

	token1, _ := jwt.GenerateToken(&jwt.Claims{UserID: staff.ID, SessionID: sessPIN.ID}, "dev-secret", time.Hour)

	body, _ := json.Marshal(map[string]string{"pin": "1234"})
	req1 := httptest.NewRequest("POST", "/v1/retailer-auth/pin/set", bytes.NewReader(body))
	req1.Header.Set("Authorization", "Bearer "+token1)
	w1 := httptest.NewRecorder()

	authH.HandleSetPin(w1, req1)

	if w1.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for PIN session, got %d", w1.Code)
	}

	var resp1 map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &resp1)
	errObj := resp1["error"].(map[string]interface{})
	if errObj["code"] != CodeStepUpRequired {
		t.Fatalf("expected code %s, got %v", CodeStepUpRequired, errObj["code"])
	}

	// Session 2: Old OTP session (> 10 mins old)
	sessOldOTP := &StaffSession{
		ID:         "sess-otp-old",
		StaffID:    staff.ID,
		AuthMethod: AuthMethodOTP,
		CreatedAt:  time.Now().Add(-15 * time.Minute),
	}
	_ = repo.CreateSession(context.Background(), sessOldOTP)

	token2, _ := jwt.GenerateToken(&jwt.Claims{UserID: staff.ID, SessionID: sessOldOTP.ID}, "dev-secret", time.Hour)

	req2 := httptest.NewRequest("POST", "/v1/retailer-auth/pin/set", bytes.NewReader(body))
	req2.Header.Set("Authorization", "Bearer "+token2)
	w2 := httptest.NewRecorder()

	authH.HandleSetPin(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for old OTP session, got %d", w2.Code)
	}
}

func TestPINLogin_LockoutAfter5FailedAttempts(t *testing.T) {
	repo := NewMemoryRepository()
	pinSvc := NewPINService(repo)
	authH := NewAuthHandler(repo, nil, pinSvc, "dev-secret")

	staff := &StaffMember{
		ID:      "staff-pin-lock",
		StoreID: "store-1",
		ChainID: "chain-1",
		Phone:   "+919876500002",
		Name:    "Lock User",
		Role:    "CASHIER",
	}
	_ = repo.CreateStaffMember(context.Background(), staff)

	// Set initial PIN "1234"
	freshOTP := &StaffSession{ID: "sess-fresh-1", StaffID: staff.ID, AuthMethod: AuthMethodOTP}
	_ = repo.CreateSession(context.Background(), freshOTP)
	_ = pinSvc.SetPin(context.Background(), staff.ID, "1234", freshOTP)

	// Fail PIN login 4 times
	for i := 1; i <= 4; i++ {
		body, _ := json.Marshal(map[string]string{"phone": staff.Phone, "pin": "9999"})
		req := httptest.NewRequest("POST", "/v1/retailer-auth/pin/login", bytes.NewReader(body))
		w := httptest.NewRecorder()
		authH.HandlePinLogin(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("attempt %d: expected 400 Bad Request, got %d", i, w.Code)
		}
	}

	// 5th failed attempt -> PIN_LOCKED (429)
	body5, _ := json.Marshal(map[string]string{"phone": staff.Phone, "pin": "9999"})
	req5 := httptest.NewRequest("POST", "/v1/retailer-auth/pin/login", bytes.NewReader(body5))
	w5 := httptest.NewRecorder()
	authH.HandlePinLogin(w5, req5)

	if w5.Code != http.StatusTooManyRequests {
		t.Fatalf("attempt 5: expected 429 Too Many Requests, got %d", w5.Code)
	}

	var resp5 map[string]interface{}
	_ = json.Unmarshal(w5.Body.Bytes(), &resp5)
	errObj := resp5["error"].(map[string]interface{})
	if errObj["code"] != CodePinLocked {
		t.Fatalf("expected code %s, got %v", CodePinLocked, errObj["code"])
	}

	// Correct PIN during lock window -> still rejected
	bodyCorrect, _ := json.Marshal(map[string]string{"phone": staff.Phone, "pin": "1234"})
	reqLock := httptest.NewRequest("POST", "/v1/retailer-auth/pin/login", bytes.NewReader(bodyCorrect))
	wLock := httptest.NewRecorder()
	authH.HandlePinLogin(wLock, reqLock)

	if wLock.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 during lock window, got %d", wLock.Code)
	}

	// Simulate lock expiry
	expiredLock := time.Now().Add(-1 * time.Minute)
	_ = repo.UpdatePinLockout(context.Background(), staff.ID, 0, &expiredLock)

	// Correct PIN after lock expires -> Success
	reqSuccess := httptest.NewRequest("POST", "/v1/retailer-auth/pin/login", bytes.NewReader(bodyCorrect))
	wSuccess := httptest.NewRecorder()
	authH.HandlePinLogin(wSuccess, reqSuccess)

	if wSuccess.Code != http.StatusOK {
		t.Fatalf("expected 200 OK after lock expires, got %d", wSuccess.Code)
	}
}
