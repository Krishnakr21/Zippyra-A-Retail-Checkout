package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zippyra/backend/shared/kafka"
)

func setupTestDB(t *testing.T) (*sql.DB, *PostgresRepository, *MovementEngine) {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open sqlite memory DB: %v", err)
	}

	schema := `
		CREATE TABLE stock_levels (
			store_id TEXT NOT NULL,
			barcode TEXT NOT NULL,
			on_hand_qty INTEGER NOT NULL DEFAULT 0 CHECK (on_hand_qty >= 0),
			reorder_point INTEGER NOT NULL DEFAULT 10,
			reorder_qty INTEGER NOT NULL DEFAULT 50,
			low_stock_alerted INTEGER NOT NULL DEFAULT 0,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (store_id, barcode)
		);

		CREATE TABLE stock_movements (
			id TEXT PRIMARY KEY,
			store_id TEXT NOT NULL,
			barcode TEXT NOT NULL,
			movement_type TEXT NOT NULL,
			qty_delta INTEGER NOT NULL,
			reference_type TEXT,
			reference_id TEXT,
			note TEXT,
			created_by TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (store_id, barcode, reference_type, reference_id, movement_type)
		);

		CREATE TABLE stock_counts (
			id TEXT PRIMARY KEY,
			store_id TEXT NOT NULL,
			barcode TEXT NOT NULL,
			expected_qty INTEGER NOT NULL,
			counted_qty INTEGER NOT NULL,
			variance_qty INTEGER NOT NULL,
			counted_by TEXT NOT NULL,
			counted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE shrinkage_daily (
			id TEXT PRIMARY KEY,
			store_id TEXT NOT NULL,
			date TEXT NOT NULL,
			total_variance_qty INTEGER NOT NULL,
			total_expected_qty INTEGER NOT NULL,
			shrinkage_percent REAL NOT NULL,
			item_count INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (store_id, date)
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	repo := NewPostgresRepository(db)
	producer := kafka.NewProducer("localhost:9092")
	engine := NewMovementEngine(db, producer, true)

	return db, repo, engine
}

func TestMovementInvariant_OnHandEqualsSumDeltas(t *testing.T) {
	db, repo, engine := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	storeID := "store-inv-1"
	barcode := "8901234567890"

	// 1. Initial GRN of 100
	_, _, err := engine.ApplyMovement(ctx, nil, storeID, barcode, MovementGRNReceived, 100, RefGRN, "grn-init", nil, nil, true)
	if err != nil {
		t.Fatalf("Initial GRN movement failed: %v", err)
	}

	movementTypes := []string{MovementGRNReceived, MovementSale, MovementReturn, MovementTransferIn, MovementTransferOut, MovementAdjustment}

	// 2. Perform 20 random valid movements
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 20; i++ {
		mType := movementTypes[r.Intn(len(movementTypes))]
		var delta int64

		if mType == MovementGRNReceived || mType == MovementReturn || mType == MovementTransferIn {
			delta = int64(r.Intn(20) + 1)
		} else if mType == MovementSale || mType == MovementTransferOut {
			delta = -int64(r.Intn(5) + 1)
		} else {
			delta = int64(r.Intn(10) - 5) // adjustment +/-
		}

		refID := fmt.Sprintf("ref-prop-%d", i)
		noteStr := "Property test"
		_, _, _ = engine.ApplyMovement(ctx, nil, storeID, barcode, mType, delta, RefManual, refID, nil, &noteStr, true)
	}

	// 3. Verify INVARIANT: stock_levels.on_hand_qty == SUM(stock_movements.qty_delta)
	sl, err := repo.GetStockLevel(ctx, storeID, barcode)
	if err != nil || sl == nil {
		t.Fatalf("Failed to fetch stock level: %v", err)
	}

	var sumDelta int64
	err = db.QueryRow(`SELECT SUM(qty_delta) FROM stock_movements WHERE store_id = $1 AND barcode = $2`, storeID, barcode).Scan(&sumDelta)
	if err != nil {
		t.Fatalf("Failed to query SUM(qty_delta): %v", err)
	}

	if sl.OnHandQty != sumDelta {
		t.Errorf("INVARIANT VIOLATION! stock_levels.on_hand_qty = %d, SUM(stock_movements.qty_delta) = %d", sl.OnHandQty, sumDelta)
	}
}
