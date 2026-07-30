package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sharedErrors "github.com/zippyra/backend/shared/errors"
)

type MockGoogleVerifier struct {
	ShouldFail bool
	Sub        string
	Email      string
}

func (m *MockGoogleVerifier) VerifyIDToken(ctx context.Context, token, clientID string) (*GoogleTokenPayload, error) {
	if m.ShouldFail || token == "invalid-token" {
		return nil, errors.New("invalid google token")
	}
	return &GoogleTokenPayload{
		Sub:           m.Sub,
		Email:         m.Email,
		EmailVerified: true,
		Name:          "Test User",
	}, nil
}

func TestGoogleOAuth_AccountLinking(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	email := "existing_user@example.com"
	googleSub := "google-sub-12345"

	// 1. User signs up via Email OTP first
	emailUser, err := repo.CreateUserWithEmail(ctx, email)
	if err != nil {
		t.Fatalf("failed to create email user: %v", err)
	}

	if emailUser.GoogleSub != nil {
		t.Fatalf("expected initial google_sub to be nil")
	}

	// 2. Later, user signs in with Google using the same email
	mockVerifier := &MockGoogleVerifier{
		ShouldFail: false,
		Sub:        googleSub,
		Email:      email,
	}

	handler := NewAuthHandler(repo, NewDefaultOTPManager(nil, &LogSmsSender{}, NewGmailEmailSender()), mockVerifier)

	bodyBytes, _ := json.Marshal(GoogleOAuthRequest{
		IDToken:  "valid-google-token",
		DeviceID: "device-xyz",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/oauth/google", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	handler.HandleGoogleOAuth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got: %d", rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.IsNewUser {
		t.Fatalf("expected is_new_user to be false for account linking")
	}

	if resp.User.ID != emailUser.ID {
		t.Fatalf("expected user ID to match existing user %s, got %s", emailUser.ID, resp.User.ID)
	}

	if resp.User.GoogleSub == nil || *resp.User.GoogleSub != googleSub {
		t.Fatalf("expected google_sub to be linked to %s", googleSub)
	}

	// Assert single row in users map
	updatedUser, _ := repo.GetUserByID(ctx, emailUser.ID)
	if updatedUser.GoogleSub == nil || *updatedUser.GoogleSub != googleSub {
		t.Fatalf("assert failed: google_sub not updated in repository")
	}
}

func TestGoogleOAuth_InvalidToken_Rejection(t *testing.T) {
	repo := NewMemoryRepository()

	mockVerifier := &MockGoogleVerifier{
		ShouldFail: true,
	}

	handler := NewAuthHandler(repo, NewDefaultOTPManager(nil, &LogSmsSender{}, NewGmailEmailSender()), mockVerifier)

	bodyBytes, _ := json.Marshal(GoogleOAuthRequest{
		IDToken:  "invalid-token",
		DeviceID: "device-xyz",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/oauth/google", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	handler.HandleGoogleOAuth(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got: %d", rec.Code)
	}

	var errResp sharedErrors.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error.Code != sharedErrors.CodeGoogleTokenInvalid {
		t.Fatalf("expected error code GOOGLE_TOKEN_INVALID, got: %s", errResp.Error.Code)
	}

	// Assert no user was created
	usersMap := repo.users
	if len(usersMap) != 0 {
		t.Fatalf("expected 0 users created on invalid token, found %d", len(usersMap))
	}
}
