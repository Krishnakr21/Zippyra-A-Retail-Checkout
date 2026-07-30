package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReturnRequest_NonReturnableItem_Rejection(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	handler := NewOrderHandler(repo, NewMockRedisExitTokenService("sec"), NewMockInvoiceService(repo), "sec")

	// Seed order with non-returnable item
	order := &Order{
		ID:        "ord-non-ret",
		PaymentID: "pay-non-ret",
		UserID:    "user-ret-1",
		StoreID:   "store-1",
		Items: []OrderItem{
			{Barcode: "8901000000001", Name: "Fresh Milk", Qty: 1, PricePaise: 5000, HSNCode: "0401", IsReturnable: false},
		},
		SubtotalPaise: 5000, TotalPaise: 5000, PaymentMethod: "UPI", SupplyType: "INTRASTATE", Status: StatusCompleted,
	}
	flags := []OrderItemReturnableFlag{
		{OrderID: order.ID, Barcode: "8901000000001", IsReturnable: false, ReturnedQty: 0},
	}
	_, _ = repo.CreateOrderAndOutboxTx(context.Background(), order, flags, NewMockRedisExitTokenService("sec"), TopicOrderCompleted, []byte("{}"))

	body := []byte(`{"item_barcodes":["8901000000001"],"reason":"DAMAGED"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/order/ord-non-ret/return", bytes.NewBuffer(body))
	req.Header.Set("X-User-ID", "user-ret-1")

	rec := httptest.NewRecorder()
	handler.CreateReturnRequestHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for non-returnable item, got %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestReturnRequest_Expired24hWindow_Rejection(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	handler := NewOrderHandler(repo, NewMockRedisExitTokenService("sec"), NewMockInvoiceService(repo), "sec")

	// Seed order created 25 hours ago
	oldTime := time.Now().Add(-25 * time.Hour)
	_, _ = db.Exec(`
		INSERT INTO orders (id, payment_id, user_id, store_id, items, subtotal_paise, total_paise, payment_method, status, created_at)
		VALUES ('ord-old', 'pay-old', 'user-old', 'store-1', '[{"barcode":"8901000000002","name":"T-Shirt","qty":1,"price_paise":50000,"hsn_code":"6109","is_returnable":true}]', 50000, 50000, 'UPI', 'COMPLETED', ?)
	`, oldTime)
	_, _ = db.Exec(`INSERT INTO order_items_returnable_flags (order_id, barcode, is_returnable, returned_qty) VALUES ('ord-old', '8901000000002', 1, 0)`)

	body := []byte(`{"item_barcodes":["8901000000002"],"reason":"WRONG_ITEM"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/order/ord-old/return", bytes.NewBuffer(body))
	req.Header.Set("X-User-ID", "user-old")

	rec := httptest.NewRecorder()
	handler.CreateReturnRequestHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for expired 24h window, got %d, body: %s", rec.Code, rec.Body.String())
	}
}
