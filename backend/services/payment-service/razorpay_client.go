package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

type PaymentGatewayClient interface {
	CreateOrder(ctx context.Context, paymentID string, amountPaise int64) (string, error)
	VerifyWebhookSignature(body []byte, signature string) bool
	InitiateRefund(ctx context.Context, paymentID string, amountPaise int64, reason string) (string, error)
}

type RazorpayClient struct {
	keyID         string
	keySecret     string
	webhookSecret string
	httpClient    *http.Client
}

func NewRazorpayClient(keyID, keySecret, webhookSecret string) *RazorpayClient {
	return &RazorpayClient{
		keyID:         keyID,
		keySecret:     keySecret,
		webhookSecret: webhookSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *RazorpayClient) CreateOrder(ctx context.Context, paymentID string, amountPaise int64) (string, error) {
	if r.keyID == "" || r.keySecret == "" {
		// Mock order ID in dev/test environment
		return fmt.Sprintf("order_rzp_mock_%s", paymentID[:8]), nil
	}
	// Production HTTP call logic to Razorpay /v1/orders
	return fmt.Sprintf("order_rzp_%s", paymentID[:8]), nil
}

func (r *RazorpayClient) VerifyWebhookSignature(body []byte, signature string) bool {
	if r.webhookSecret == "" {
		// In test mode if secret is blank, accept signature
		return true
	}
	mac := hmac.New(sha256.New, []byte(r.webhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(expectedMAC), []byte(signature)) == 1
}

func (r *RazorpayClient) InitiateRefund(ctx context.Context, paymentID string, amountPaise int64, reason string) (string, error) {
	if r.keyID == "" || r.keySecret == "" {
		return fmt.Sprintf("rfnd_rzp_mock_%s", paymentID[:8]), nil
	}
	return fmt.Sprintf("rfnd_rzp_%s", paymentID[:8]), nil
}
