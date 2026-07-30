package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestOrderDPDPConsumer_AnonymizesUserOrders(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewPostgresRepository(db)

	ctx := context.Background()

	// Create order for user-dpdp-99
	order := &Order{
		ID:            "ord-dpdp-anon-1",
		PaymentID:     "pay-dpdp-anon-1",
		UserID:        "user-dpdp-99",
		StoreID:       "store-1",
		Items:         []OrderItem{},
		SubtotalPaise: 10000,
		TotalPaise:    10000,
		PaymentMethod: "CASH",
		Status:        StatusCompleted,
	}
	exitSvc := NewMockRedisExitTokenService("sec")
	_, err := repo.CreateOrderAndOutboxTx(ctx, order, []OrderItemReturnableFlag{}, exitSvc, TopicOrderCompleted, []byte(`{}`))
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	consumer := NewOrderDPDPConsumer(repo, nil)

	reqBytes, _ := json.Marshal(DPDPOrderDeletionRequestPayload{
		UserID:        "user-dpdp-99",
		DPDPRequestID: "req-dpdp-order-001",
	})

	if err := consumer.HandleUserDataDeletionRequested(ctx, reqBytes); err != nil {
		t.Fatalf("HandleUserDataDeletionRequested failed: %v", err)
	}

	// Verify order user_id was replaced with tombstone UUID
	fetched, err := repo.GetOrderByID(ctx, "ord-dpdp-anon-1")
	if err != nil || fetched == nil {
		t.Fatalf("Failed to fetch order: %v", err)
	}

	if fetched.UserID != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("Expected user_id to be tombstone 00000000-0000-0000-0000-000000000000, got %s", fetched.UserID)
	}
	// Invoice / order record itself remains intact
	if fetched.ID != "ord-dpdp-anon-1" {
		t.Fatalf("Expected order ID to remain intact, got %s", fetched.ID)
	}
}
