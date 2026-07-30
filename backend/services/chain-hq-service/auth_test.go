package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zippyra/backend/shared/errors"
)

func TestOTPSend_UnregisteredPhone_Returns403UserNotRegistered(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAuthHandler(repo)

	body, _ := json.Marshal(SendOTPRequest{Phone: "+919999999999"})
	req := httptest.NewRequest("POST", "/v1/chain-hq/otp/send", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleSendOTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var resp errors.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error.Code != CodeChainHQUserNotRegistered {
		t.Fatalf("expected error code CHAIN_HQ_USER_NOT_REGISTERED, got %s", resp.Error.Code)
	}
}

func TestOTPVerify_ValidCredentials_ReturnsTokens(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAuthHandler(repo)

	user := &ChainHQUser{
		ChainID:  "chain-001",
		Phone:    "+919876543210",
		Name:     "Chain Owner",
		Role:     RoleOwner,
		IsActive: true,
	}
	_ = repo.CreateUser(context.Background(), user)

	body, _ := json.Marshal(VerifyOTPRequest{Phone: "+919876543210", OTP: "123456"})
	req := httptest.NewRequest("POST", "/v1/chain-hq/otp/verify", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleVerifyOTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["access_token"] == nil || resp["access_token"] == "" {
		t.Fatalf("expected valid access_token in response")
	}
}
