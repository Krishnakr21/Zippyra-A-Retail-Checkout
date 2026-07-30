package main

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_OpensOnErrors(t *testing.T) {
	cb := NewRollingCircuitBreaker(0.05, 50*time.Millisecond)

	// Record 5 successful requests
	for i := 0; i < 5; i++ {
		cb.RecordResult(nil)
	}

	if cb.ShouldFallbackToCashfree() {
		t.Fatalf("Circuit breaker should be CLOSED after 5 successful requests")
	}

	// Record 1 failure (1/6 = 16.6% error rate > 5% threshold)
	cb.RecordResult(errors.New("gateway error"))

	if !cb.ShouldFallbackToCashfree() {
		t.Fatalf("Circuit breaker should be OPEN and fallback to Cashfree after high error rate")
	}

	// Wait for openTimeout (50ms) to transition to HalfOpen
	time.Sleep(60 * time.Millisecond)

	// First trial request should attempt Razorpay (false for fallback)
	if cb.ShouldFallbackToCashfree() {
		t.Fatalf("Circuit breaker should be HALF-OPEN after timeout allowing 1 trial request")
	}

	// Record success for trial request
	cb.RecordResult(nil)

	if cb.ShouldFallbackToCashfree() {
		t.Fatalf("Circuit breaker should reset to CLOSED after successful trial request")
	}
}
