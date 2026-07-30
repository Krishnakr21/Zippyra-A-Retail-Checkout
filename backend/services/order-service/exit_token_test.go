package main

import (
	"context"
	"testing"
)

func TestExitToken_IssueAndRetrieve(t *testing.T) {
	svc := NewMockRedisExitTokenService("secret")
	ctx := context.Background()

	// Issue token
	data, err := svc.IssueAndStoreExitToken(ctx, "ord-100", "usr-100", "str-1", "sess-abc")
	if err != nil || data == nil {
		t.Fatalf("Failed to issue exit token: %v", err)
	}

	if data.OrderID != "ord-100" || data.ExitToken == "" {
		t.Fatalf("Invalid exit token data: %+v", data)
	}

	// Retrieve token
	retrieved, err := svc.GetExitToken(ctx, "usr-100", "str-1")
	if err != nil {
		t.Fatalf("Failed to retrieve exit token: %v", err)
	}

	if retrieved.OrderID != "ord-100" || retrieved.ExitToken != data.ExitToken {
		t.Fatalf("Retrieved token mismatch: expected %s, got %s", data.ExitToken, retrieved.ExitToken)
	}
}

func TestExitToken_ExpiredTokenReturnsError(t *testing.T) {
	svc := NewMockRedisExitTokenService("secret")
	ctx := context.Background()

	// Store an already-expired token directly
	svc.store["exit_preauth:usr-exp:str-1"] = []byte(`{
		"order_id": "ord-exp",
		"exit_token": "token_expired",
		"issued_at": "2026-07-31T10:00:00Z",
		"expires_at": "2026-07-31T10:10:00Z"
	}`)

	_, err := svc.GetExitToken(ctx, "usr-exp", "str-1")
	if err == nil {
		t.Fatalf("Expected error for expired exit token, got nil")
	}

	// Verify key was cleaned up
	_, exists := svc.store["exit_preauth:usr-exp:str-1"]
	if exists {
		t.Fatalf("Expected expired key to be deleted from store")
	}
}
