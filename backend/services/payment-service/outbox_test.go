package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open memory db: %v", err)
	}

	queries := []string{
		`CREATE TABLE payments (
			id TEXT PRIMARY KEY,
			checkout_session_id TEXT UNIQUE NOT NULL,
			user_id TEXT NOT NULL,
			store_id TEXT NOT NULL,
			amount_paise INTEGER NOT NULL,
			loyalty_points_used INTEGER NOT NULL DEFAULT 0,
			loyalty_discount_paise INTEGER NOT NULL DEFAULT 0,
			payable_amount_paise INTEGER NOT NULL,
			payment_method TEXT NOT NULL,
			gateway TEXT NOT NULL,
			gateway_order_id TEXT NULL,
			gateway_payment_id TEXT NULL,
			status TEXT NOT NULL DEFAULT 'INITIATED',
			failure_reason TEXT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE payment_outbox (
			id TEXT PRIMARY KEY,
			topic TEXT NOT NULL,
			payload TEXT NOT NULL,
			published_at DATETIME NULL,
			retry_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE payment_webhook_events (
			id TEXT PRIMARY KEY,
			gateway TEXT NOT NULL,
			gateway_event_id TEXT UNIQUE NOT NULL,
			event_type TEXT NOT NULL,
			raw_payload TEXT NOT NULL,
			processed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("Failed to execute schema query: %v", err)
		}
	}

	return db
}

func TestOutboxPattern_TransactionalPublish(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert initial payment row
	p := &Payment{
		CheckoutSessionID:    "sess-outbox-1",
		UserID:               "user-1",
		StoreID:              "store-1",
		AmountPaise:          10000,
		PayableAmountPaise:   10000,
		PaymentMethod:        MethodUPI,
		Gateway:              GatewayRazorpay,
	}
	inserted, newRow, err := repo.InitiatePaymentOnConflict(ctx, p)
	if err != nil || !newRow {
		t.Fatalf("Failed to initiate payment: %v", err)
	}

	// Single SQL transaction: Capture payment and insert outbox row
	payload := []byte(`{"payment_id":"` + inserted.ID + `","status":"CAPTURED"}`)
	err = repo.CapturePaymentAndOutboxTx(ctx, inserted.ID, "pay_rzp_123", "payment.confirmed", payload)
	if err != nil {
		t.Fatalf("CapturePaymentAndOutboxTx failed: %v", err)
	}

	// Verify payment status updated to CAPTURED
	updatedPayment, err := repo.GetPaymentByID(ctx, inserted.ID)
	if err != nil || updatedPayment.Status != StatusCaptured {
		t.Fatalf("Expected status CAPTURED, got %v", updatedPayment.Status)
	}

	// Verify outbox row inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM payment_outbox WHERE topic = 'payment.confirmed' AND published_at IS NULL").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Expected 1 unpublished outbox row, got %d, err: %v", count, err)
	}

	// Run OutboxRelay processing
	producer := kafka.NewProducer("localhost:9092")
	relay := NewOutboxRelay(db, producer, 50*time.Millisecond)
	relay.processBatch()

	// Verify outbox row marked published
	err = db.QueryRow("SELECT COUNT(*) FROM payment_outbox WHERE published_at IS NOT NULL").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Expected 1 published outbox row after relay run, got %d", count)
	}
}
