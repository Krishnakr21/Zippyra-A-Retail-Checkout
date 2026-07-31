package main

import (
	"context"
	"testing"
)

func TestLowStockAlertingLifecycle(t *testing.T) {
	db, repo, engine := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	storeID := "store-low-1"
	barcode := "8903333333333"

	// 1. Initial stock of 15 (reorder_point = 10)
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, barcode, MovementGRNReceived, 15, RefGRN, "g15", nil, nil, true)

	sl1, _ := repo.GetStockLevel(ctx, storeID, barcode)
	if sl1.LowStockAlerted {
		t.Errorf("Expected low_stock_alerted to be false initially")
	}

	// 2. Sale of 6 -> stock drops to 9 (<= 10 reorder_point) -> Alert fires & low_stock_alerted = true
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, barcode, MovementSale, -6, RefOrder, "o1", nil, nil, true)

	sl2, _ := repo.GetStockLevel(ctx, storeID, barcode)
	if !sl2.LowStockAlerted {
		t.Errorf("Expected low_stock_alerted to be true after dropping below threshold (9 <= 10)")
	}

	// 3. Second sale of 2 -> stock drops to 7 (still <= 10) -> Alert should NOT re-fire
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, barcode, MovementSale, -2, RefOrder, "o2", nil, nil, true)

	sl3, _ := repo.GetStockLevel(ctx, storeID, barcode)
	if !sl3.LowStockAlerted {
		t.Errorf("Expected low_stock_alerted to REMAIN true without re-firing alert")
	}

	// 4. Restock of 10 -> stock increases to 17 (> 10 reorder_point) -> low_stock_alerted resets to false!
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, barcode, MovementGRNReceived, 10, RefGRN, "g10", nil, nil, true)

	sl4, _ := repo.GetStockLevel(ctx, storeID, barcode)
	if sl4.LowStockAlerted {
		t.Errorf("Expected low_stock_alerted to RESET to false after restock above threshold (17 > 10)")
	}

	// 5. Subsequent sale of 8 -> stock drops to 9 (<= 10 reorder_point) -> FRESH Alert fires & low_stock_alerted = true again!
	_, _, _ = engine.ApplyMovement(ctx, nil, storeID, barcode, MovementSale, -8, RefOrder, "o3", nil, nil, true)

	sl5, _ := repo.GetStockLevel(ctx, storeID, barcode)
	if !sl5.LowStockAlerted {
		t.Errorf("Expected low_stock_alerted to be true again on second cycle crossing below threshold")
	}
}
