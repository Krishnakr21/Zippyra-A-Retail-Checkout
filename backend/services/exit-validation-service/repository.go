package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateExitAttempt(ctx context.Context, attempt *ExitAttempt) error
	CreateStaffOverride(ctx context.Context, override *StaffOverride) error
	GetLatestExitAttemptByOrderID(ctx context.Context, orderID string) (*ExitAttempt, error)
	GetRecentExitAttempts(ctx context.Context, storeID string, limit int) ([]*ExitAttempt, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateExitAttempt(ctx context.Context, attempt *ExitAttempt) error {
	if attempt.ID == "" {
		attempt.ID = uuid.New().String()
	}
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO exit_attempts (id, order_id, user_id, store_id, gate_id, result, is_alarm, rfid_tag_ids, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		attempt.ID, attempt.OrderID, attempt.UserID, attempt.StoreID, attempt.GateID,
		attempt.Result, attempt.IsAlarm, attempt.RFIDTagIDs, attempt.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert exit attempt: %w", err)
	}
	return nil
}

func (r *PostgresRepository) CreateStaffOverride(ctx context.Context, override *StaffOverride) error {
	if override.ID == "" {
		override.ID = uuid.New().String()
	}
	if override.CreatedAt.IsZero() {
		override.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO staff_overrides (id, order_id, store_id, gate_id, staff_user_id, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		override.ID, override.OrderID, override.StoreID, override.GateID,
		override.StaffUserID, override.Reason, override.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert staff override: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetLatestExitAttemptByOrderID(ctx context.Context, orderID string) (*ExitAttempt, error) {
	query := `
		SELECT id, order_id, user_id, store_id, gate_id, result, is_alarm, rfid_tag_ids, created_at
		FROM exit_attempts
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, orderID)

	var attempt ExitAttempt
	var tagIDsJSON []byte
	err := row.Scan(
		&attempt.ID, &attempt.OrderID, &attempt.UserID, &attempt.StoreID, &attempt.GateID,
		&attempt.Result, &attempt.IsAlarm, &tagIDsJSON, &attempt.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan exit attempt: %w", err)
	}

	if len(tagIDsJSON) > 0 {
		attempt.RFIDTagIDs = json.RawMessage(tagIDsJSON)
	}

	return &attempt, nil
}

func (r *PostgresRepository) GetRecentExitAttempts(ctx context.Context, storeID string, limit int) ([]*ExitAttempt, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT id, order_id, user_id, store_id, gate_id, result, is_alarm, rfid_tag_ids, created_at
		FROM exit_attempts
		WHERE store_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, storeID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent exit attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*ExitAttempt
	for rows.Next() {
		var a ExitAttempt
		var tagIDsJSON []byte
		if err := rows.Scan(&a.ID, &a.OrderID, &a.UserID, &a.StoreID, &a.GateID, &a.Result, &a.IsAlarm, &tagIDsJSON, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan exit attempt row: %w", err)
		}
		if len(tagIDsJSON) > 0 {
			a.RFIDTagIDs = json.RawMessage(tagIDsJSON)
		}
		attempts = append(attempts, &a)
	}
	return attempts, nil
}

