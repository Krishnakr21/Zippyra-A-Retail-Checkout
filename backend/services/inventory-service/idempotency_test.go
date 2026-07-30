package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIdempotency_DuplicateOrderCompleted(t *testing.T) {
	db, repo, engine := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	consumer := NewEventConsumer(engine)

	storeID := "store-idemp-1"
	barcode := "8901111111111"
	orderID := "order-dup-100"

	// 1. Initial stock of 50
	_, _, err := engine.ApplyMovement(ctx, nil, storeID, barcode, MovementGRNReceived, 50, RefGRN, "grn-1", nil, nil, true)
	if err != nil {
		t.Fatalf("GRN failed: %v", err)
	}

	payload := OrderCompletedKafkaPayload{
		OrderID: orderID,
		StoreID: storeID,
		Items: []struct {
			Barcode string `json:"barcode"`
			Qty     int64  `json:"qty"`
		}{
			{Barcode: barcode, Qty: 5},
		},
	}
	val, _ := json.Marshal(payload)

	// First event processing -> Qty decremented to 45
	if err := consumer.ProcessOrderCompleted(ctx, val); err != nil {
		t.Fatalf("First ProcessOrderCompleted failed: %v", err)
	}

	sl1, _ := repo.GetStockLevel(ctx, storeID, barcode)
	if sl1.OnHandQty != 45 {
		t.Errorf("Expected on_hand_qty 45 after first sale, got %d", sl1.OnHandQty)
	}

	// Second event processing (DUPLICATE) -> Qty stays 45
	if err := consumer.ProcessOrderCompleted(ctx, val); err != nil {
		t.Fatalf("Second ProcessOrderCompleted failed: %v", err)
	}

	sl2, _ := repo.GetStockLevel(ctx, storeID, barcode)
	if sl2.OnHandQty != 45 {
		t.Errorf("Expected on_hand_qty to REMAIN 45 after duplicate sale event, got %d", sl2.OnHandQty)
	}
}

func TestIdempotency_DuplicateApplyGRN(t *testing.T) {
	db, repo, engine := setupTestDB(t)
	defer db.Close()

	internalHandler := NewInternalHandler(repo, engine)
	storeID := "store-idemp-2"
	grnID := "grn-dup-200"

	reqBody := ApplyGRNRequest{
		StoreID: storeID,
		GRNID:   grnID,
		Items: []GRNItem{
			{Barcode: "8902222222222", QtyReceived: 20, UnitCostPaise: 1000},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Call 1
	req1 := httptest.NewRequest(http.MethodPost, "/v1/inventory/internal/apply-grn", bytes.NewReader(bodyBytes))
	rec1 := httptest.NewRecorder()
	internalHandler.ApplyGRNHandler(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("First GRN application failed: %d", rec1.Code)
	}

	var resp1 map[string]interface{}
	_ = json.Unmarshal(rec1.Body.Bytes(), &resp1)
	if resp1["items_applied"].(float64) != 1 {
		t.Errorf("Expected 1 item applied on first GRN call, got %v", resp1["items_applied"])
	}

	sl1, _ := repo.GetStockLevel(context.Background(), storeID, "8902222222222")
	if sl1.OnHandQty != 20 {
		t.Errorf("Expected on_hand_qty 20 after first GRN call, got %d", sl1.OnHandQty)
	}

	// Call 2 (DUPLICATE GRN)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/inventory/internal/apply-grn", bytes.NewReader(bodyBytes))
	rec2 := httptest.NewRecorder()
	internalHandler.ApplyGRNHandler(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Second GRN application failed: %d", rec2.Code)
	}

	var resp2 map[string]interface{}
	_ = json.Unmarshal(rec2.Body.Bytes(), &resp2)
	if resp2["items_applied"].(float64) != 0 {
		t.Errorf("Expected 0 items applied on duplicate GRN call, got %v", resp2["items_applied"])
	}

	sl2, _ := repo.GetStockLevel(context.Background(), storeID, "8902222222222")
	if sl2.OnHandQty != 20 {
		t.Errorf("Expected on_hand_qty to REMAIN 20 after duplicate GRN call, got %d", sl2.OnHandQty)
	}
}

func TestTransferOut_InsufficientStockRollback(t *testing.T) {
	db, repo, engine := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	internalHandler := NewInternalHandler(repo, engine)
	storeID := "store-idemp-3"

	// Stock item 1 = 10, Item 2 = 10, Item 3 = 2
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, "item-1", MovementGRNReceived, 10, RefGRN, "g1", nil, nil, true)
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, "item-2", MovementGRNReceived, 10, RefGRN, "g2", nil, nil, true)
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, "item-3", MovementGRNReceived, 2, RefGRN, "g3", nil, nil, true)

	// Attempt transfer out: Item 1 = 5, Item 2 = 5, Item 3 = 10 (Item 3 will fail!)
	transferReq := ApplyTransferRequest{
		StoreID:    storeID,
		TransferID: "tr-fail-1",
		Items: []TransferItem{
			{Barcode: "item-1", Qty: 5},
			{Barcode: "item-2", Qty: 5},
			{Barcode: "item-3", Qty: 10},
		},
	}
	bodyBytes, _ := json.Marshal(transferReq)

	req := httptest.NewRequest(http.MethodPost, "/v1/inventory/internal/apply-transfer-out", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	internalHandler.ApplyTransferOutHandler(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("Expected 409 Conflict status on insufficient stock for transfer, got %d", rec.Code)
	}

	// Verify ENTIRE call rolled back (item-1 and item-2 stock levels remain unchanged!)
	sl1, _ := repo.GetStockLevel(ctx, storeID, "item-1")
	if sl1.OnHandQty != 10 {
		t.Errorf("Expected item-1 stock to remain 10 after rollback, got %d", sl1.OnHandQty)
	}

	sl3, _ := repo.GetStockLevel(ctx, storeID, "item-3")
	if sl3.OnHandQty != 2 {
		t.Errorf("Expected item-3 stock to remain 2 after rollback, got %d", sl3.OnHandQty)
	}
}
