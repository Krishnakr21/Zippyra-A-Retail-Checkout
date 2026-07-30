package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/zippyra/backend/shared/jwt"
)

type ExitTokenData struct {
	OrderID   string    `json:"order_id"`
	ExitToken string    `json:"exit_token"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ExitTokenService interface {
	IssueAndStoreExitToken(ctx context.Context, orderID, userID, storeID, sessionID string) (*ExitTokenData, error)
	GetExitToken(ctx context.Context, userID, storeID string) (*ExitTokenData, error)
}

type MockRedisExitTokenService struct {
	jwtSecret string
	mu        sync.RWMutex
	store     map[string][]byte
}

func NewMockRedisExitTokenService(jwtSecret string) *MockRedisExitTokenService {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &MockRedisExitTokenService{
		jwtSecret: jwtSecret,
		store:     make(map[string][]byte),
	}
}

func (s *MockRedisExitTokenService) key(userID, storeID string) string {
	return fmt.Sprintf("exit_preauth:%s:%s", userID, storeID)
}

func (s *MockRedisExitTokenService) IssueAndStoreExitToken(ctx context.Context, orderID, userID, storeID, sessionID string) (*ExitTokenData, error) {
	now := time.Now()
	ttl := 10 * time.Minute
	expiresAt := now.Add(ttl)

	tokenStr, err := jwt.GenerateExitToken(orderID, userID, storeID, sessionID, s.jwtSecret, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to generate exit token JWT: %w", err)
	}

	data := &ExitTokenData{
		OrderID:   orderID,
		ExitToken: tokenStr,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal exit token data: %w", err)
	}

	s.mu.Lock()
	s.store[s.key(userID, storeID)] = dataBytes
	s.mu.Unlock()

	return data, nil
}

func (s *MockRedisExitTokenService) GetExitToken(ctx context.Context, userID, storeID string) (*ExitTokenData, error) {
	s.mu.RLock()
	dataBytes, exists := s.store[s.key(userID, storeID)]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no pending exit token found")
	}

	var data ExitTokenData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exit token data: %w", err)
	}

	if time.Now().After(data.ExpiresAt) {
		s.mu.Lock()
		delete(s.store, s.key(userID, storeID))
		s.mu.Unlock()
		return nil, fmt.Errorf("exit token expired")
	}

	return &data, nil
}
