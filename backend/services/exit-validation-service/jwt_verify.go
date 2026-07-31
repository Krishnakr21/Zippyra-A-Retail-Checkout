package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/zippyra/backend/shared/jwt"
	sharedRedis "github.com/zippyra/backend/shared/redis"
)

type RevocationChecker interface {
	IsDeviceRevoked(ctx context.Context, deviceID string) (bool, error)
}

type RedisRevocationChecker struct {
	client *sharedRedis.Client
}

func NewRedisRevocationChecker(client *sharedRedis.Client) *RedisRevocationChecker {
	return &RedisRevocationChecker{client: client}
}

func (r *RedisRevocationChecker) IsDeviceRevoked(ctx context.Context, deviceID string) (bool, error) {
	if r.client == nil || r.client.Client == nil {
		return false, nil
	}
	key := fmt.Sprintf("device_revoked:%s", deviceID)
	val, err := r.client.Get(ctx, key).Result()
	if err == nil && val == "true" {
		return true, nil
	}
	return false, nil
}

type MockRevocationChecker struct {
	revoked map[string]bool
}

func NewMockRevocationChecker() *MockRevocationChecker {
	return &MockRevocationChecker{revoked: make(map[string]bool)}
}

func (m *MockRevocationChecker) RevokeDevice(deviceID string) {
	m.revoked[deviceID] = true
}

func (m *MockRevocationChecker) IsDeviceRevoked(ctx context.Context, deviceID string) (bool, error) {
	return m.revoked[deviceID], nil
}

type JWTVerifier struct {
	secret  string
	checker RevocationChecker
}

func NewJWTVerifier(secret string, checker RevocationChecker) *JWTVerifier {
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &JWTVerifier{
		secret:  secret,
		checker: checker,
	}
}

func (v *JWTVerifier) VerifyDeviceToken(authHeader string) (*jwt.DeviceClaims, error) {
	if authHeader == "" {
		return nil, fmt.Errorf("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, fmt.Errorf("invalid Authorization header format")
	}

	claims, err := jwt.ParseAndVerifyDeviceToken(parts[1], v.secret)
	if err != nil {
		return nil, fmt.Errorf("invalid device token: %w", err)
	}

	if claims.UserType != "DEVICE" {
		return nil, fmt.Errorf("unauthorized user_type: expected DEVICE, got %s", claims.UserType)
	}

	deviceID := claims.DeviceID
	if deviceID == "" {
		deviceID = claims.Subject
	}

	if v.checker != nil && deviceID != "" {
		revoked, _ := v.checker.IsDeviceRevoked(context.Background(), deviceID)
		if revoked {
			return nil, fmt.Errorf("DEVICE_REVOKED: device credentials have been decommissioned")
		}
	}

	return claims, nil
}

func (v *JWTVerifier) VerifyExitToken(tokenStr string) (*jwt.ExitClaims, bool, error) {
	claims, isExpired, err := jwt.ParseAndVerifyExitToken(tokenStr, v.secret)
	if err != nil {
		return nil, false, fmt.Errorf("invalid exit token: %w", err)
	}
	return claims, isExpired, nil
}
