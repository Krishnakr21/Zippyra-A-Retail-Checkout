package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetStockLevel(ctx context.Context, storeID, barcode string) (*StockLevel, error)
	GetLowStockLevels(ctx context.Context, storeID string) ([]StockLevel, error)
	GetShrinkageReport(ctx context.Context, storeID, dateFrom, dateTo string) ([]ShrinkageDaily, float64, error)
	UpsertShrinkageDaily(ctx context.Context, storeID, date string, totalVariance, totalExpected int64, percent float64, itemCount int) error
	GetDB() *sql.DB
	IsSQLite() bool
}

type PostgresRepository struct {
	db       *sql.DB
	isSQLite bool
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	isSQLite := false
	if db != nil && fmt.Sprintf("%T", db.Driver()) == "*sqlite3.SQLiteDriver" {
		isSQLite = true
	}
	return &PostgresRepository{db: db, isSQLite: isSQLite}
}

func (r *PostgresRepository) GetDB() *sql.DB {
	return r.db
}

func (r *PostgresRepository) IsSQLite() bool {
	return r.isSQLite
}

func (r *PostgresRepository) lockClause() string {
	if r.isSQLite {
		return ""
	}
	return " FOR UPDATE"
}

func (r *PostgresRepository) GetStockLevel(ctx context.Context, storeID, barcode string) (*StockLevel, error) {
	query := `
		SELECT store_id, barcode, on_hand_qty, reorder_point, reorder_qty, low_stock_alerted, updated_at
		FROM stock_levels
		WHERE store_id = $1 AND barcode = $2
	`
	row := r.db.QueryRowContext(ctx, query, storeID, barcode)

	var sl StockLevel
	err := row.Scan(&sl.StoreID, &sl.Barcode, &sl.OnHandQty, &sl.ReorderPoint, &sl.ReorderQty, &sl.LowStockAlerted, &sl.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query stock level: %w", err)
	}
	return &sl, nil
}

func (r *PostgresRepository) GetLowStockLevels(ctx context.Context, storeID string) ([]StockLevel, error) {
	query := `
		SELECT store_id, barcode, on_hand_qty, reorder_point, reorder_qty, low_stock_alerted, updated_at
		FROM stock_levels
		WHERE store_id = $1 AND on_hand_qty <= reorder_point
		ORDER BY on_hand_qty ASC
	`
	rows, err := r.db.QueryContext(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query low stock levels: %w", err)
	}
	defer rows.Close()

	var results []StockLevel
	for rows.Next() {
		var sl StockLevel
		if err := rows.Scan(&sl.StoreID, &sl.Barcode, &sl.OnHandQty, &sl.ReorderPoint, &sl.ReorderQty, &sl.LowStockAlerted, &sl.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan low stock level: %w", err)
		}
		results = append(results, sl)
	}
	return results, nil
}

func (r *PostgresRepository) GetShrinkageReport(ctx context.Context, storeID, dateFrom, dateTo string) ([]ShrinkageDaily, float64, error) {
	query := `
		SELECT id, store_id, date, total_variance_qty, total_expected_qty, shrinkage_percent, item_count, created_at
		FROM shrinkage_daily
		WHERE store_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, storeID, dateFrom, dateTo)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query shrinkage report: %w", err)
	}
	defer rows.Close()

	var results []ShrinkageDaily
	var sumVariance, sumExpected int64

	for rows.Next() {
		var sd ShrinkageDaily
		if err := rows.Scan(&sd.ID, &sd.StoreID, &sd.Date, &sd.TotalVarianceQty, &sd.TotalExpectedQty, &sd.ShrinkagePercent, &sd.ItemCount, &sd.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan shrinkage daily: %w", err)
		}
		results = append(results, sd)
		sumVariance += sd.TotalVarianceQty
		sumExpected += sd.TotalExpectedQty
	}

	overallPercent := 0.0
	if sumExpected > 0 {
		// Abs of total variance
		absVar := sumVariance
		if absVar < 0 {
			absVar = -absVar
		}
		overallPercent = (float64(absVar) / float64(sumExpected)) * 100.0
	}

	return results, overallPercent, nil
}

func (r *PostgresRepository) UpsertShrinkageDaily(ctx context.Context, storeID, date string, totalVariance, totalExpected int64, percent float64, itemCount int) error {
	now := time.Now()
	id := uuid.New().String()
	query := `
		INSERT INTO shrinkage_daily (id, store_id, date, total_variance_qty, total_expected_qty, shrinkage_percent, item_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (store_id, date) DO UPDATE SET
			total_variance_qty = EXCLUDED.total_variance_qty,
			total_expected_qty = EXCLUDED.total_expected_qty,
			shrinkage_percent = EXCLUDED.shrinkage_percent,
			item_count = EXCLUDED.item_count
	`
	_, err := r.db.ExecContext(ctx, query, id, storeID, date, totalVariance, totalExpected, percent, itemCount, now)
	if err != nil {
		return fmt.Errorf("failed to upsert shrinkage daily: %w", err)
	}
	return nil
}
