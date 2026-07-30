package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockCartClientForCash struct{}

func (m *MockCartClientForCash) FetchCheckoutSession(ctx context.Context, sessionID string) (*InternalCheckoutSessionResponse, error) {
	return &InternalCheckoutSessionResponse{
		ID:         sessionID,
		UserID:     "user-cash-1",
		StoreID:    "store-1",
		TotalPaise: 50000, // ₹500
		ExpiresAt:  time.Now().Add(10 * time.Minute),
	}, nil
}

func TestCashPayment_InsufficientCash(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	handler := NewPaymentHandler(repo, &MockCartClientForCash{}, NewMockLoyaltyServiceClient(), nil, nil, NewRollingCircuitBreaker(0.05, 30), "pubkey", "secret")

	body := []byte(`{"checkout_session_id":"sess-cash-1","cash_collected_paise":40000}`) // ₹400 collected for ₹500 total
	req := httptest.NewRequest(http.MethodPost, "/v1/payment/cash", bytes.NewBuffer(body))
	req.Header.Set("X-User-ID", "cashier-1")
	req.Header.Set("X-User-Role", RoleCashier)
	req.Header.Set("X-Store-ID", "store-1")

	rec := httptest.NewRecorder()
	handler.CashPaymentHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for insufficient cash, got %d", rec.Code)
	}

	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM payments").Scan(&count)
	if count != 0 {
		t.Fatalf("Expected 0 payments created on insufficient cash, got %d", count)
	}
}

func TestCashPayment_SufficientCash_CreatesOutboxConfirmed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	handler := NewPaymentHandler(repo, &MockCartClientForCash{}, NewMockLoyaltyServiceClient(), nil, nil, NewRollingCircuitBreaker(0.05, 30), "pubkey", "secret")

	body := []byte(`{"checkout_session_id":"sess-cash-2","cash_collected_paise":60000}`) // ₹600 collected for ₹500 total
	req := httptest.NewRequest(http.MethodPost, "/v1/payment/cash", bytes.NewBuffer(body))
	req.Header.Set("X-User-ID", "cashier-1")
	req.Header.Set("X-User-Role", RoleCashier)
	req.Header.Set("X-Store-ID", "store-1")

	rec := httptest.NewRecorder()
	handler.CashPaymentHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for sufficient cash, got %d, body: %s", rec.Code, rec.Body.String())
	}

	// Verify payment row created with status CAPTURED
	p, err := repo.GetPaymentBySessionID(context.Background(), "sess-cash-2")
	if err != nil || p.Status != StatusCaptured {
		t.Fatalf("Expected payment status CAPTURED, got %v, err: %v", p.Status, err)
	}

	// Verify outbox payment.confirmed event inserted
	var outboxCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM payment_outbox WHERE topic = 'payment.confirmed'").Scan(&outboxCount)
	if outboxCount != 1 {
		t.Fatalf("Expected 1 payment.confirmed outbox event for cash payment, got %d", outboxCount)
	}
}
