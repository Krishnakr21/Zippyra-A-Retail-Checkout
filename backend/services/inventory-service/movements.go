package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type MovementEngine struct {
	db       *sql.DB
	producer *kafka.Producer
	isSQLite bool
}

func NewMovementEngine(db *sql.DB, producer *kafka.Producer, isSQLite bool) *MovementEngine {
	return &MovementEngine{
		db:       db,
		producer: producer,
		isSQLite: isSQLite,
	}
}

func (e *MovementEngine) ApplyMovement(
	ctx context.Context,
	tx *sql.Tx,
	storeID string,
	barcode string,
	movementType string,
	qtyDelta int64,
	refType string,
	refID string,
	createdBy *string,
	note *string,
	allowNegative bool,
) (applied bool, newOnHandQty int64, err error) {

	ownTx := false
	if tx == nil {
		tx, err = e.db.BeginTx(ctx, nil)
		if err != nil {
			return false, 0, fmt.Errorf("failed to begin tx: %w", err)
		}
		ownTx = true
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			}
		}()
	}

	// 1. Idempotency Check: INSERT INTO stock_movements ON CONFLICT DO NOTHING
	movementID := uuid.New().String()
	now := time.Now()

	var insertMovementQuery string
	if e.isSQLite {
		insertMovementQuery = `
			INSERT OR IGNORE INTO stock_movements (id, store_id, barcode, movement_type, qty_delta, reference_type, reference_id, note, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
	} else {
		insertMovementQuery = `
			INSERT INTO stock_movements (id, store_id, barcode, movement_type, qty_delta, reference_type, reference_id, note, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (store_id, barcode, reference_type, reference_id, movement_type) DO NOTHING
		`
	}

	res, err := tx.ExecContext(ctx, insertMovementQuery, movementID, storeID, barcode, movementType, qtyDelta, refType, refID, note, createdBy, now)
	if err != nil {
		return false, 0, fmt.Errorf("failed to insert stock movement: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, 0, fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Idempotent duplicate: fetch current on_hand_qty
		var currentQty int64
		_ = tx.QueryRowContext(ctx, `SELECT on_hand_qty FROM stock_levels WHERE store_id = $1 AND barcode = $2`, storeID, barcode).Scan(&currentQty)
		if ownTx {
			_ = tx.Commit()
		}
		return false, currentQty, nil
	}

	// 2. Fetch current stock level
	var currentOnHand int64 = 0
	var reorderPoint int = 10
	var reorderQty int = 50
	var lowStockAlerted bool = false

	_ = tx.QueryRowContext(ctx, `
		SELECT on_hand_qty, reorder_point, reorder_qty, low_stock_alerted
		FROM stock_levels
		WHERE store_id = $1 AND barcode = $2
	`, storeID, barcode).Scan(&currentOnHand, &reorderPoint, &reorderQty, &lowStockAlerted)

	// 3. Negative stock guard BEFORE modifying stock levels
	if !allowNegative && (currentOnHand+qtyDelta < 0) {
		return false, currentOnHand, fmt.Errorf("INSUFFICIENT_STOCK")
	}

	rawNewQty := currentOnHand + qtyDelta
	targetOnHand := rawNewQty
	if targetOnHand < 0 {
		targetOnHand = 0 // Floor at 0 for DB schema constraint check
	}

	// 4. UPSERT stock_levels
	upsertQuery := `
		INSERT INTO stock_levels (store_id, barcode, on_hand_qty, reorder_point, reorder_qty, low_stock_alerted, updated_at)
		VALUES ($1, $2, $3, 10, 50, false, $4)
		ON CONFLICT (store_id, barcode) DO UPDATE SET
			on_hand_qty = $3,
			updated_at = EXCLUDED.updated_at
	`
	_, err = tx.ExecContext(ctx, upsertQuery, storeID, barcode, targetOnHand, now)
	if err != nil {
		return false, 0, fmt.Errorf("failed to upsert stock levels: %w", err)
	}

	// 5. Low-Stock Alerting & Reset Logic
	var shouldAlertLowStock bool
	var shouldResetLowStock bool

	if qtyDelta < 0 && targetOnHand <= int64(reorderPoint) && !lowStockAlerted {
		shouldAlertLowStock = true
		_, _ = tx.ExecContext(ctx, `UPDATE stock_levels SET low_stock_alerted = true WHERE store_id = $1 AND barcode = $2`, storeID, barcode)
	} else if qtyDelta > 0 && targetOnHand > int64(reorderPoint) && lowStockAlerted {
		shouldResetLowStock = true
		_, _ = tx.ExecContext(ctx, `UPDATE stock_levels SET low_stock_alerted = false WHERE store_id = $1 AND barcode = $2`, storeID, barcode)
	}

	if ownTx {
		if err := tx.Commit(); err != nil {
			return false, 0, fmt.Errorf("failed to commit movement tx: %w", err)
		}
	}

	// 6. Publish Kafka Events (After successful commit)
	stockUpdatedMsg := InventoryStockUpdatedPayload{
		StoreID:      storeID,
		Barcode:      barcode,
		AvailableQty: targetOnHand,
		Timestamp:    now,
	}
	_ = e.producer.PublishEvent(ctx, TopicStockUpdated, fmt.Sprintf("%s:%s", storeID, barcode), stockUpdatedMsg)

	if shouldAlertLowStock {
		lowStockMsg := LowStockPayload{
			StoreID:      storeID,
			Barcode:      barcode,
			CurrentQty:   targetOnHand,
			ReorderPoint: reorderPoint,
			ReorderQty:   reorderQty,
			Timestamp:    now,
		}
		_ = e.producer.PublishEvent(ctx, TopicLowStock, fmt.Sprintf("%s:%s", storeID, barcode), lowStockMsg)
		logger.Warn("[LOW STOCK ALERT] Store: %s | Barcode: %s | On Hand: %d <= Reorder Point: %d", storeID, barcode, targetOnHand, reorderPoint)
	}

	if shouldResetLowStock {
		logger.Info("[LOW STOCK RESET] Store: %s | Barcode: %s | On Hand: %d > Reorder Point: %d", storeID, barcode, targetOnHand, reorderPoint)
	}

	return true, targetOnHand, nil
}
