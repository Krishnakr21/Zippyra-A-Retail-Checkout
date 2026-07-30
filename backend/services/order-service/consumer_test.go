package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Use a unique per-call DSN so tables don't collide when multiple tests share the same binary
	dbName := fmt.Sprintf("file:testdb_%s?mode=memory&cache=shared", t.Name())
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		t.Fatalf("Failed to open sqlite memory db: %v", err)
	}
	db.SetMaxOpenConns(1)

	schemas := []string{
		`CREATE TABLE orders (
			id TEXT PRIMARY KEY,
			payment_id TEXT UNIQUE NOT NULL,
			user_id TEXT NOT NULL,
			store_id TEXT NOT NULL,
			items TEXT NOT NULL,
			subtotal_paise INTEGER NOT NULL,
			discount_paise INTEGER NOT NULL DEFAULT 0,
			cgst_paise INTEGER NOT NULL DEFAULT 0,
			sgst_paise INTEGER NOT NULL DEFAULT 0,
			igst_paise INTEGER NOT NULL DEFAULT 0,
			total_paise INTEGER NOT NULL,
			loyalty_points_used INTEGER NOT NULL DEFAULT 0,
			payment_method TEXT NOT NULL,
			supply_type TEXT NOT NULL DEFAULT 'INTRASTATE',
			status TEXT NOT NULL DEFAULT 'CREATED',
			invoice_s3_key TEXT NULL,
			irn TEXT NULL,
			irn_ack_no TEXT NULL,
			irn_ack_date TIMESTAMP NULL,
			irn_qr_code TEXT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP NULL
		);`,
		`CREATE TABLE order_creation_outbox (
			id TEXT PRIMARY KEY,
			topic TEXT NOT NULL,
			payload TEXT NOT NULL,
			published_at TIMESTAMP NULL,
			retry_count INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE order_items_returnable_flags (
			order_id TEXT NOT NULL,
			barcode TEXT NOT NULL,
			is_returnable BOOLEAN NOT NULL DEFAULT 1,
			returned_qty INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (order_id, barcode)
		);`,
		`CREATE TABLE return_requests (
			id TEXT PRIMARY KEY,
			order_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			store_id TEXT NOT NULL,
			items TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING_STAFF_REVIEW',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, s := range schemas {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("Failed to execute schema: %v", err);
		}
	}

	return db
}

func TestConsumer_DuplicatePaymentConfirmed_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	exitTokenSvc := NewMockRedisExitTokenService("secret")
	invoiceSvc := NewMockInvoiceService(repo)

	consumer := NewPaymentConfirmedConsumer(repo, exitTokenSvc, invoiceSvc)

	payload := PaymentConfirmedPayload{
		PaymentID:          "pay-dupe-100",
		CheckoutSessionID:  "sess-100",
		UserID:             "user-1",
		StoreID:            "store-1",
		AmountPaise:        10000,
		PayableAmountPaise: 10000,
		LoyaltyPointsUsed:  0,
		PaymentMethod:      "UPI",
		Timestamp:          time.Now(),
	}

	msgBytes, _ := json.Marshal(payload)

	// First delivery
	if err := consumer.ProcessPaymentConfirmed(context.Background(), msgBytes); err != nil {
		t.Fatalf("First delivery failed: %v", err)
	}

	// Second delivery (duplicate event)
	if err := consumer.ProcessPaymentConfirmed(context.Background(), msgBytes); err != nil {
		t.Fatalf("Second delivery failed: %v", err)
	}

	// Verify exactly 1 orders row
	var orderCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM orders WHERE payment_id = 'pay-dupe-100'").Scan(&orderCount)
	if orderCount != 1 {
		t.Fatalf("Expected exactly 1 order row for duplicate payment, got %d", orderCount)
	}

	// Verify exactly 1 order.completed outbox event
	var outboxCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM order_creation_outbox WHERE topic = 'order.completed'").Scan(&outboxCount)
	if outboxCount != 1 {
		t.Fatalf("Expected exactly 1 order.completed outbox event for duplicate payment, got %d", outboxCount)
	}
}

func TestConsumer_CoreOrderFailure_PublishesCompensatingFailedEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	// Pass an uninitialized/nil exit token service so core order creation fails
	consumer := NewPaymentConfirmedConsumer(repo, nil, nil)

	payload := PaymentConfirmedPayload{
		PaymentID:          "pay-fail-999",
		CheckoutSessionID:  "sess-999",
		UserID:             "user-fail",
		StoreID:            "store-1",
		AmountPaise:        15000,
		PayableAmountPaise: 15000,
		LoyaltyPointsUsed:  0,
		PaymentMethod:      "UPI",
		Timestamp:          time.Now(),
	}

	msgBytes, _ := json.Marshal(payload)
	_ = consumer.ProcessPaymentConfirmed(context.Background(), msgBytes)

	// Verify NO orders row left in inconsistent state
	var orderCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM orders WHERE payment_id = 'pay-fail-999'").Scan(&orderCount)
	if orderCount != 0 {
		t.Fatalf("Expected 0 orders row on core failure, got %d", orderCount)
	}

	// Verify compensating order.creation_failed event inserted into outbox
	var failedOutboxCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM order_creation_outbox WHERE topic = 'order.creation_failed'").Scan(&failedOutboxCount)
	if failedOutboxCount != 1 {
		t.Fatalf("Expected 1 order.creation_failed outbox event for refund trigger, got %d", failedOutboxCount)
	}
}
