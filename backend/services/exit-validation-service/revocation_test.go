package main

import (
	"strings"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/jwt"
)

func TestVerifyDeviceToken_RevokedInRedis_ReturnsDeviceRevokedError(t *testing.T) {
	mockChecker := NewMockRevocationChecker()
	secret := "zippyra-dev-jwt-secret-key-32bytes"
	verifier := NewJWTVerifier(secret, mockChecker)

	deviceID := "dev-decommissioned-999"

	tokenStr, err := jwt.GenerateDeviceToken(deviceID, "GATE_01", "store-100", secret, 365*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate device token: %v", err)
	}
	authHeader := "Bearer " + tokenStr

	// Unrevoked device token should verify successfully
	verifiedClaims, err := verifier.VerifyDeviceToken(authHeader)
	if err != nil || (verifiedClaims.DeviceID != deviceID && verifiedClaims.Subject != deviceID) {
		t.Fatalf("expected successful verification before revocation, got err: %v", err)
	}

	// Revoke device in mock checker
	mockChecker.RevokeDevice(deviceID)

	// Should now be rejected with DEVICE_REVOKED
	_, errRevoked := verifier.VerifyDeviceToken(authHeader)
	if errRevoked == nil || !strings.Contains(errRevoked.Error(), "DEVICE_REVOKED") {
		t.Fatalf("expected DEVICE_REVOKED error after revocation key set, got %v", errRevoked)
	}
}
