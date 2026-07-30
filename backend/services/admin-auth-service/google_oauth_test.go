package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGoogleLogin_NonMatchingDomain_Returns403DomainNotAllowed(t *testing.T) {
	repo := NewMemoryRepository()
	googleVal := &RealGoogleTokenValidator{}
	handler := NewAdminAuthHandler(repo, googleVal, nil)
	handler.allowedDomain = "zippyra.com"

	body, _ := json.Marshal(GoogleLoginRequest{
		IDToken: "mock-google-token-user@attacker.com:sub-123",
	})
	req := httptest.NewRequest("POST", "/v1/admin-auth/login/google", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleGoogleLogin(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errMap, ok := resp["error"].(map[string]interface{})
	if !ok || errMap["code"] != CodeDomainNotAllowed {
		t.Fatalf("expected error code DOMAIN_NOT_ALLOWED, got %v", resp)
	}
}

func TestGoogleLogin_UnregisteredEmail_Returns403AdminNotRegistered(t *testing.T) {
	repo := NewMemoryRepository()
	googleVal := &RealGoogleTokenValidator{}
	handler := NewAdminAuthHandler(repo, googleVal, nil)
	handler.allowedDomain = "zippyra.com"

	body, _ := json.Marshal(GoogleLoginRequest{
		IDToken: "mock-google-token-unregistered@zippyra.com:sub-456",
	})
	req := httptest.NewRequest("POST", "/v1/admin-auth/login/google", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleGoogleLogin(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errMap, ok := resp["error"].(map[string]interface{})
	if !ok || errMap["code"] != CodeAdminNotRegistered {
		t.Fatalf("expected error code ADMIN_NOT_REGISTERED, got %v", resp)
	}
}
