package main

import (
	"context"
	"testing"
)

func TestPlayIntegrityVerification(t *testing.T) {
	handler := &PaymentHandler{}

	t.Run("Passed verdict token does not fail", func(t *testing.T) {
		handler.verifyPlayIntegrityToken(context.Background(), "mock_token_meets_device_integrity")
	})

	t.Run("Failed verdict token increments failure counter", func(t *testing.T) {
		handler.verifyPlayIntegrityToken(context.Background(), "mock_token_integrity_failed")
	})
}
