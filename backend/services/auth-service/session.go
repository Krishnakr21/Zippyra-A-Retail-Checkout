package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/jwt"
)

// issueSession creates a new JWT access token + refresh token and stores session in repository.
func issueSession(ctx context.Context, repo Repository, user *User, deviceID, deviceLabel, jwtSecret string) (string, string, error) {
	if deviceID == "" {
		deviceID = "unknown-device"
	}
	if deviceLabel == "" {
		deviceLabel = "Mobile Device"
	}
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}

	sessionID := uuid.New().String()
	now := time.Now()

	// 1. Generate Access Token (15m TTL) containing sessionID
	accessToken, err := jwt.GenerateAccessTokenWithSession(user.ID, deviceID, sessionID, jwtSecret, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	// 2. Generate Refresh Token
	refreshTokenRaw, refreshTokenHash, err := jwt.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// 3. Save session in auth_sessions table
	session := &AuthSession{
		ID:               sessionID,
		DeviceID:         deviceID,
		DeviceLabel:      &deviceLabel,
		RefreshTokenHash: refreshTokenHash,
		UserID:           user.ID,
		CreatedAt:        now,
		LastUsedAt:       &now,
	}

	if err := repo.CreateSession(ctx, session); err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenRaw, nil
}
