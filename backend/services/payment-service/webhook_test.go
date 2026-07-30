package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhook_DuplicateDeliveryIdempotency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	cartService := NewDefaultCartServiceClient("http://mock")
	loyaltyService := NewMockLoyaltyServiceClient()
	rzp := NewRazorpayClient("key", "secret", "")
	cf := NewCashfreeClient("app", "sec")
	cb := NewRollingCircuitBreaker(0.05, 30)

	handler := NewPaymentHandler(repo, cartService, loyaltyService, rzp, cf, cb, "pubkey", "secret")
	ctx := context.Background()

	// Seed payment row
	p := &Payment{
		CheckoutSessionID:  "sess-wh-1",
		UserID:             "user-wh-1",
		StoreID:            "store-1",
		AmountPaise:        5000,
		PayableAmountPaise: 5000,
		PaymentMethod:      MethodUPI,
		Gateway:            GatewayRazorpay,
	}
	inserted, _, _ := repo.InitiatePaymentOnConflict(ctx, p)

	rawPayload := []byte(`{
		"event": "payment.captured",
		"event_id": "evt_duplicate_123",
		"payload": {
			"payment": {
				"entity": {
					"id": "pay_test_123",
					"notes": {
						"payment_id": "` + inserted.ID + `"
					}
				}
			}
		}
	}`)

	// Delivery 1
	req1 := httptest.NewRequest(http.MethodPost, "/v1/payment/webhook/razorpay", bytes.NewBuffer(rawPayload))
	rec1 := httptest.NewRecorder()
	handler.RazorpayWebhookHandler(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on first delivery, got %d", rec1.Code)
	}

	// Delivery 2 (Duplicate)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/payment/webhook/razorpay", bytes.NewBuffer(rawPayload))
	rec2 := httptest.NewRecorder()
	handler.RazorpayWebhookHandler(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on duplicate delivery, got %d", rec2.Code)
	}

	// Assert exactly 1 outbox entry exists (not 2!)
	var outboxCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM payment_outbox WHERE topic = 'payment.confirmed'").Scan(&outboxCount)
	if outboxCount != 1 {
		t.Fatalf("Expected exactly 1 outbox event for duplicate webhooks, got %d", outboxCount)
	}
}
