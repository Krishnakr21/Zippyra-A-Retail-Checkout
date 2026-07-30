package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/kafka"
)

func TestStockCount_ZeroVarianceNoMovementRow(t *testing.T) {
	db, repo, engine := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	handler := NewInventoryHandler(repo, engine, nil)
	storeID := "store-shrink-1"
	barcode := "8904444444444"

	// 1. Initial stock of 10
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, barcode, MovementGRNReceived, 10, RefGRN, "g1", nil, nil, true)

	// 2. Submit stock count where counted_qty == expected_qty (10 == 10, variance = 0)
	countReq := StockCountRequest{
		StoreID: storeID,
		Entries: []StockCountEntry{
			{Barcode: barcode, CountedQty: 10},
		},
	}
	bodyBytes, _ := json.Marshal(countReq)

	req := httptest.NewRequest(http.MethodPost, "/v1/inventory/stock-count", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	handler.StockCountHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Stock count handler failed: %d", rec.Code)
	}

	// 3. Verify stock_counts row WAS written
	var countRows int
	_ = db.QueryRow(`SELECT COUNT(*) FROM stock_counts WHERE store_id = $1 AND barcode = $2`, storeID, barcode).Scan(&countRows)
	if countRows != 1 {
		t.Errorf("Expected 1 stock_counts row, got %d", countRows)
	}

	// 4. Verify NO STOCK_COUNT movement row was written
	var movementRows int
	_ = db.QueryRow(`SELECT COUNT(*) FROM stock_movements WHERE store_id = $1 AND barcode = $2 AND reference_type = $3`, storeID, barcode, RefStockCount).Scan(&movementRows)
	if movementRows != 0 {
		t.Errorf("Expected 0 STOCK_COUNT movement rows for 0 variance count, got %d", movementRows)
	}
}

func TestManualAdjustment_NegativeStockRejected(t *testing.T) {
	db, repo, engine := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	handler := NewInventoryHandler(repo, engine, nil)
	storeID := "store-shrink-2"
	barcode := "8905555555555"

	// Initial stock of 5
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, barcode, MovementGRNReceived, 5, RefGRN, "g1", nil, nil, true)

	// Manual adjustment of -10 -> would result in on_hand_qty = -5
	adjustReq := AdjustStockRequest{
		StoreID:  storeID,
		Barcode:  barcode,
		QtyDelta: -10,
		Reason:   "DAMAGE",
	}
	bodyBytes, _ := json.Marshal(adjustReq)

	req := httptest.NewRequest(http.MethodPost, "/v1/inventory/adjust", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	handler.AdjustStockHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request on negative stock adjustment, got %d", rec.Code)
	}

	// Verify stock level remains 5
	sl, _ := repo.GetStockLevel(ctx, storeID, barcode)
	if sl.OnHandQty != 5 {
		t.Errorf("Expected stock level to remain 5, got %d", sl.OnHandQty)
	}
}

func TestShrinkageRollupJob_CalculationAndAlert(t *testing.T) {
	db, repo, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	producer := kafka.NewProducer("localhost:9092")
	job := NewShrinkageRollupJob(db, repo, producer)

	storeID := "store-shrink-3"
	today := time.Now().Format("2006-01-02")

	// Insert stock_counts with total expected = 1000, counted = 990 (variance = -10, shrinkage = 1.0% > 0.5%)
	_, err := db.Exec(`
		INSERT INTO stock_counts (id, store_id, barcode, expected_qty, counted_qty, variance_qty, counted_by, counted_at)
		VALUES ('sc-1', $1, 'item-a', 1000, 990, -10, 'staff-1', $2)
	`, storeID, today)
	if err != nil {
		t.Fatalf("Failed to insert stock count fixture: %v", err)
	}

	if err := job.Run(ctx, today); err != nil {
		t.Fatalf("Shrinkage job failed: %v", err)
	}

	// Verify shrinkage_daily row
	report, percent, err := repo.GetShrinkageReport(ctx, storeID, today, today)
	if err != nil || len(report) != 1 {
		t.Fatalf("Failed to get shrinkage report: %v", err)
	}

	if percent < 0.99 || percent > 1.01 {
		t.Errorf("Expected shrinkage percent ~1.00%%, got %.2f%%", percent)
	}
}
