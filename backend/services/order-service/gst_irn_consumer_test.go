package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestGSTIRNConsumer_HandleIRNIssued_UpdatesOrder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewPostgresRepository(db)

	ctx := context.Background()

	// 1. Create an initial order
	order := &Order{
		ID:            "ord-irn-test-1",
		PaymentID:     "pay-irn-test-1",
		UserID:        "user-100",
		StoreID:       "store-1",
		Items:         []OrderItem{},
		SubtotalPaise: 10000,
		TotalPaise:    10000,
		PaymentMethod: "CASH",
		Status:        StatusCompleted,
	}
	flags := []OrderItemReturnableFlag{}
	exitSvc := NewMockRedisExitTokenService("sec")
	_, err := repo.CreateOrderAndOutboxTx(ctx, order, flags, exitSvc, TopicOrderCompleted, []byte(`{}`))
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	consumer := NewGSTIRNConsumer(repo, nil)

	now := time.Now().UTC()
	payloadBytes, _ := json.Marshal(IRNIssuedPayload{
		OrderID:      "ord-irn-test-1",
		IRN:          "irn_hash_64char_abc1234567890abcdef1234567890abcdef1234567890abcdef",
		AckNo:        "122607310099",
		AckDate:      &now,
		SignedQRCode: "data:image/png;base64,signed_qr_code_bytes",
	})

	// 2. Consume IRN issued event
	if err := consumer.HandleIRNIssued(ctx, payloadBytes); err != nil {
		t.Fatalf("HandleIRNIssued failed: %v", err)
	}

	// 3. Verify order was updated in DB
	fetched, err := repo.GetOrderByID(ctx, "ord-irn-test-1")
	if err != nil || fetched == nil {
		t.Fatalf("Failed to fetch order: %v", err)
	}

	if fetched.IRN == nil || *fetched.IRN != "irn_hash_64char_abc1234567890abcdef1234567890abcdef1234567890abcdef" {
		t.Fatalf("Expected IRN hash to be updated, got %v", fetched.IRN)
	}
	if fetched.IRNAckNo == nil || *fetched.IRNAckNo != "122607310099" {
		t.Fatalf("Expected IRNAckNo 122607310099, got %v", fetched.IRNAckNo)
	}
}

func TestGSTIRNConsumer_HandleIRNFailed_LeavesIRNFieldsNull(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewPostgresRepository(db)

	ctx := context.Background()

	order := &Order{
		ID:            "ord-irn-test-2",
		PaymentID:     "pay-irn-test-2",
		UserID:        "user-100",
		StoreID:       "store-1",
		Items:         []OrderItem{},
		SubtotalPaise: 5000,
		TotalPaise:    5000,
		PaymentMethod: "CASH",
		Status:        StatusCompleted,
	}
	flags := []OrderItemReturnableFlag{}
	exitSvc := NewMockRedisExitTokenService("sec")
	_, err := repo.CreateOrderAndOutboxTx(ctx, order, flags, exitSvc, TopicOrderCompleted, []byte(`{}`))
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	consumer := NewGSTIRNConsumer(repo, nil)

	payloadBytes, _ := json.Marshal(IRNFailedPayload{
		OrderID: "ord-irn-test-2",
		Reason:  "IRP Portal Timeout",
	})

	// Consume IRN failed event
	if err := consumer.HandleIRNFailed(ctx, payloadBytes); err != nil {
		t.Fatalf("HandleIRNFailed failed: %v", err)
	}

	// Verify IRN remains NULL
	fetched, err := repo.GetOrderByID(ctx, "ord-irn-test-2")
	if err != nil || fetched == nil {
		t.Fatalf("Failed to fetch order: %v", err)
	}

	if fetched.IRN != nil {
		t.Fatalf("Expected IRN to remain NULL on failure, got %s", *fetched.IRN)
	}
}
