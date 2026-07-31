package main

import (
	"context"
	"testing"
	"time"
)

func TestSalesAggregation_ReplacingMergeTreeDedup_CountsOrderOnce(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	now := time.Now().UTC()
	orderID := "ord-dedup-100"

	// First insert
	s1 := &SalesEvent{
		EventDate:     now,
		EventTime:     now,
		StoreID:       "store-100",
		ChainID:       "chain-100",
		OrderID:       orderID,
		TotalPaise:    500000, // ₹5,000
		DiscountPaise: 50000,
		PaymentMethod: "UPI",
	}
	_ = repo.InsertSalesEvent(ctx, s1)

	// Simulated duplicate delivery with slightly later event_time
	s2 := &SalesEvent{
		EventDate:     now,
		EventTime:     now.Add(1 * time.Second),
		StoreID:       "store-100",
		ChainID:       "chain-100",
		OrderID:       orderID,
		TotalPaise:    500000,
		DiscountPaise: 50000,
		PaymentMethod: "UPI",
	}
	_ = repo.InsertSalesEvent(ctx, s2)

	sales, err := repo.GetSales(ctx, "store-100", now.Format("2006-01-02"), now.Format("2006-01-02"), "day")
	if err != nil {
		t.Fatalf("unexpected error fetching sales: %v", err)
	}

	if len(sales) != 1 {
		t.Fatalf("expected 1 sales period row, got %d", len(sales))
	}

	if sales[0].OrderCount != 1 || sales[0].RevenuePaise != 500000 {
		t.Fatalf("expected 1 order and 500000 paise revenue, got count %d, revenue %d", sales[0].OrderCount, sales[0].RevenuePaise)
	}
}
