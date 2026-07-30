package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/errors"
)

type CartServiceClient interface {
	FetchCheckoutSession(ctx context.Context, sessionID string) (*InternalCheckoutSessionResponse, error)
}

type DefaultCartServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewDefaultCartServiceClient(baseURL string) *DefaultCartServiceClient {
	if baseURL == "" {
		baseURL = "http://localhost:8084"
	}
	return &DefaultCartServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *DefaultCartServiceClient) FetchCheckoutSession(ctx context.Context, sessionID string) (*InternalCheckoutSessionResponse, error) {
	url := fmt.Sprintf("%s/v1/cart/internal/checkout-session/%s", c.baseURL, sessionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-System-Auth", "true")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return nil, errors.NewAPIError(errors.CodeCheckoutSessionExpired, "Checkout session expired or not found", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cart service returned status %d", resp.StatusCode)
	}

	var sess InternalCheckoutSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

type LoyaltyServiceClient interface {
	GetPointsBalance(ctx context.Context, userID string) (int64, error)
	ReservePoints(ctx context.Context, userID string, points int64) error
	CommitReservedPoints(ctx context.Context, userID string, points int64) error
	ReleaseReservedPoints(ctx context.Context, userID string, points int64) error
}

type MockLoyaltyServiceClient struct {
	Balances map[string]int64
	Reserved map[string]int64
}

func NewMockLoyaltyServiceClient() *MockLoyaltyServiceClient {
	return &MockLoyaltyServiceClient{
		Balances: make(map[string]int64),
		Reserved: make(map[string]int64),
	}
}

func (m *MockLoyaltyServiceClient) GetPointsBalance(ctx context.Context, userID string) (int64, error) {
	if bal, ok := m.Balances[userID]; ok {
		return bal, nil
	}
	return 500, nil // Default 500 points for test/mock
}

func (m *MockLoyaltyServiceClient) ReservePoints(ctx context.Context, userID string, points int64) error {
	bal, _ := m.GetPointsBalance(ctx, userID)
	if bal < points {
		return errors.NewAPIError(errors.CodeInsufficientLoyaltyPoints, "Insufficient loyalty points", nil)
	}
	m.Balances[userID] = bal - points
	m.Reserved[userID] += points
	return nil
}

func (m *MockLoyaltyServiceClient) CommitReservedPoints(ctx context.Context, userID string, points int64) error {
	m.Reserved[userID] -= points
	return nil
}

func (m *MockLoyaltyServiceClient) ReleaseReservedPoints(ctx context.Context, userID string, points int64) error {
	m.Reserved[userID] -= points
	m.Balances[userID] += points
	return nil
}
