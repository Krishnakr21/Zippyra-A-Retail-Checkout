package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var ErrTokenInvalid = errors.New("fcm token invalid")

type FCMClient interface {
	SendPushNotification(ctx context.Context, fcmToken, title, body, deepLink string) error
	SendMulticastPush(ctx context.Context, tokens []*DeviceToken, title, body, deepLink string, markInactive func(tokenID string)) (int, error)
}

type MockFCMClient struct {
	mu           sync.Mutex
	SentPushes   []map[string]interface{}
	InvalidTokens map[string]bool
}

func NewMockFCMClient() *MockFCMClient {
	return &MockFCMClient{
		SentPushes:   []map[string]interface{}{},
		InvalidTokens: make(map[string]bool),
	}
}

func (m *MockFCMClient) SetTokenInvalid(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.InvalidTokens[token] = true
}

func (m *MockFCMClient) SendPushNotification(ctx context.Context, fcmToken, title, body, deepLink string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.InvalidTokens[fcmToken] {
		return ErrTokenInvalid
	}

	m.SentPushes = append(m.SentPushes, map[string]interface{}{
		"token":     fcmToken,
		"title":     title,
		"body":      body,
		"deep_link": deepLink,
	})
	log.Printf("[MockFCMClient] PUSH sent to token %s (title: %s, deep_link: %s)", fcmToken[:min(len(fcmToken), 10)]+"...", title, deepLink)
	return nil
}

func (m *MockFCMClient) SendMulticastPush(ctx context.Context, tokens []*DeviceToken, title, body, deepLink string, markInactive func(tokenID string)) (int, error) {
	if len(tokens) == 0 {
		return 0, nil
	}

	var wg sync.WaitGroup
	var successCount int32
	var mu sync.Mutex

	for _, tok := range tokens {
		wg.Add(1)
		go func(dt *DeviceToken) {
			defer wg.Done()

			// Individual 5s context timeout per token send
			tokenCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			err := m.SendPushNotification(tokenCtx, dt.FCMToken, title, body, deepLink)
			if err != nil {
				if errors.Is(err, ErrTokenInvalid) && markInactive != nil {
					markInactive(dt.ID)
				}
				log.Printf("[MockFCMClient] Failed to send push to token ID %s: %v", dt.ID, err)
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(tok)
	}

	wg.Wait()
	return int(successCount), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
