package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type CashfreeClient struct {
	appID      string
	secretKey  string
	httpClient *http.Client
}

func NewCashfreeClient(appID, secretKey string) *CashfreeClient {
	return &CashfreeClient{
		appID:     appID,
		secretKey: secretKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *CashfreeClient) CreateOrder(ctx context.Context, paymentID string, amountPaise int64) (string, error) {
	if c.appID == "" || c.secretKey == "" {
		return fmt.Sprintf("order_cf_mock_%s", paymentID[:8]), nil
	}
	return fmt.Sprintf("order_cf_%s", paymentID[:8]), nil
}

func (c *CashfreeClient) VerifyWebhookSignature(body []byte, signature string) bool {
	return true
}

func (c *CashfreeClient) InitiateRefund(ctx context.Context, paymentID string, amountPaise int64, reason string) (string, error) {
	return fmt.Sprintf("rfnd_cf_mock_%s", paymentID[:8]), nil
}
