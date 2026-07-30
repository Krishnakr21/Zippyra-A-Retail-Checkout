package main

import (
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

type RollingCircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	totalCount   int64
	errorCount   int64
	lastStateChange time.Time
	windowDuration  time.Duration
	openTimeout     time.Duration
	errorThreshold  float64 // 0.05 = 5%
	minRequests     int64
}

func NewRollingCircuitBreaker(errorThreshold float64, openTimeout time.Duration) *RollingCircuitBreaker {
	if errorThreshold <= 0 {
		errorThreshold = 0.05 // 5%
	}
	if openTimeout <= 0 {
		openTimeout = 30 * time.Second
	}
	return &RollingCircuitBreaker{
		state:           StateClosed,
		windowDuration:  1 * time.Minute,
		openTimeout:     openTimeout,
		errorThreshold:  errorThreshold,
		minRequests:     5,
		lastStateChange: time.Now(),
	}
}

func (cb *RollingCircuitBreaker) RecordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Reset window if 1 minute passed
	if time.Since(cb.lastStateChange) > cb.windowDuration && cb.state == StateClosed {
		cb.totalCount = 0
		cb.errorCount = 0
		cb.lastStateChange = time.Now()
	}

	cb.totalCount++
	if err != nil {
		cb.errorCount++
	}

	if cb.state == StateHalfOpen {
		if err != nil {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		} else {
			cb.state = StateClosed
			cb.totalCount = 0
			cb.errorCount = 0
			cb.lastStateChange = time.Now()
		}
		return
	}

	if cb.state == StateClosed && cb.totalCount >= cb.minRequests {
		rate := float64(cb.errorCount) / float64(cb.totalCount)
		if rate > cb.errorThreshold {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}
	}
}

func (cb *RollingCircuitBreaker) ShouldFallbackToCashfree() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen {
		if time.Since(cb.lastStateChange) > cb.openTimeout {
			cb.state = StateHalfOpen
			cb.lastStateChange = time.Now()
			return false // Allow 1 trial request through to probe Razorpay
		}
		return true // Fallback to Cashfree
	}
	return false
}

func (cb *RollingCircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *RollingCircuitBreaker) GetStatus() map[string]interface{} {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	stateStr := "CLOSED"
	if cb.state == StateOpen {
		stateStr = "OPEN"
	} else if cb.state == StateHalfOpen {
		stateStr = "HALF_OPEN"
	}

	var errorRate float64 = 0.0
	if cb.totalCount > 0 {
		errorRate = float64(cb.errorCount) / float64(cb.totalCount)
	}

	status := map[string]interface{}{
		"gateway":                 "razorpay",
		"state":                   stateStr,
		"error_rate_rolling_1min": errorRate,
	}

	if cb.state == StateOpen {
		status["opened_at"] = cb.lastStateChange.Format(time.RFC3339)
		status["will_retry_at"] = cb.lastStateChange.Add(cb.openTimeout).Format(time.RFC3339)
	}

	return status
}
