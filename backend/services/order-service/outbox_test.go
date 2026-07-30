package main

import (
	"context"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/kafka"
)

func TestOutboxPattern_OrderServiceRelay(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	order := &Order{
		ID:            "ord-outbox-1",
		PaymentID:     "pay-outbox-1",
		UserID:        "user-outbox",
		StoreID:       "store-1",
		Items:         []OrderItem{{Barcode: "890100", Name: "Item 1", Qty: 1, PricePaise: 1000, HSNCode: "1000", IsReturnable: true}},
		SubtotalPaise: 1000, TotalPaise: 1000, PaymentMethod: "UPI", SupplyType: "INTRASTATE", Status: StatusCompleted,
	}
	flags := []OrderItemReturnableFlag{{OrderID: order.ID, Barcode: "890100", IsReturnable: true, ReturnedQty: 0}}

	inserted, err := repo.CreateOrderAndOutboxTx(context.Background(), order, flags, NewMockRedisExitTokenService("sec"), TopicOrderCompleted, []byte(`{"order_id":"ord-outbox-1"}`))
	if err != nil || !inserted {
		t.Fatalf("Failed to insert order and outbox: %v", err)
	}

	producer := kafka.NewProducer("localhost:9092")
	relay := NewOutboxRelay(db, producer, 50*time.Millisecond)
	relay.processBatch()

	var publishedCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM order_creation_outbox WHERE published_at IS NOT NULL").Scan(&publishedCount)
	if publishedCount != 1 {
		t.Fatalf("Expected 1 published outbox row after relay run, got %d", publishedCount)
	}
}
