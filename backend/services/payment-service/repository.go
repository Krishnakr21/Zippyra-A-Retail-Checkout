package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	InitiatePaymentOnConflict(ctx context.Context, p *Payment) (*Payment, bool, error)
	GetPaymentByID(ctx context.Context, paymentID string) (*Payment, error)
	GetPaymentBySessionID(ctx context.Context, sessionID string) (*Payment, error)
	UpdatePaymentGatewayInfo(ctx context.Context, paymentID string, gateway string, gatewayOrderID string, status string) error
	RecordWebhookEventIdempotent(ctx context.Context, gateway, gatewayEventID, eventType string, rawPayload []byte) (bool, error)
	CapturePaymentAndOutboxTx(ctx context.Context, paymentID string, gatewayPaymentID string, topic string, payload []byte) error
	FailPaymentAndReleaseTx(ctx context.Context, paymentID string, failureReason string) error
	InsertCashPaymentTx(ctx context.Context, p *Payment, topic string, payload []byte) error
	InsertRefund(ctx context.Context, refund *Refund) error
	GetPendingTimeoutPayments(ctx context.Context, cutoff time.Duration) ([]*Payment, error)
	ResetFailedPayment(ctx context.Context, paymentID string, gateway string, payableAmount int64) error
	GetCapturedPaymentsByDate(ctx context.Context, dateStr string) ([]*Payment, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) InitiatePaymentOnConflict(ctx context.Context, p *Payment) (*Payment, bool, error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	query := `
		INSERT INTO payments (
			id, checkout_session_id, user_id, store_id, amount_paise,
			loyalty_points_used, loyalty_discount_paise, payable_amount_paise,
			payment_method, gateway, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (checkout_session_id) DO NOTHING
		RETURNING id, checkout_session_id, user_id, store_id, amount_paise,
		          loyalty_points_used, loyalty_discount_paise, payable_amount_paise,
		          payment_method, gateway, gateway_order_id, gateway_payment_id,
		          status, failure_reason, created_at, updated_at
	`
	row := r.db.QueryRowContext(ctx, query,
		p.ID, p.CheckoutSessionID, p.UserID, p.StoreID, p.AmountPaise,
		p.LoyaltyPointsUsed, p.LoyaltyDiscountPaise, p.PayableAmountPaise,
		p.PaymentMethod, p.Gateway, StatusInitiated, p.CreatedAt, p.UpdatedAt,
	)

	var res Payment
	err := row.Scan(
		&res.ID, &res.CheckoutSessionID, &res.UserID, &res.StoreID, &res.AmountPaise,
		&res.LoyaltyPointsUsed, &res.LoyaltyDiscountPaise, &res.PayableAmountPaise,
		&res.PaymentMethod, &res.Gateway, &res.GatewayOrderID, &res.GatewayPaymentID,
		&res.Status, &res.FailureReason, &res.CreatedAt, &res.UpdatedAt,
	)

	if err == nil {
		return &res, true, nil // Inserted new row
	}

	if errors.Is(err, sql.ErrNoRows) {
		// Conflict hit: retrieve existing row
		existing, getErr := r.GetPaymentBySessionID(ctx, p.CheckoutSessionID)
		if getErr != nil {
			return nil, false, getErr
		}
		return existing, false, nil
	}

	return nil, false, err
}

func (r *PostgresRepository) GetPaymentByID(ctx context.Context, paymentID string) (*Payment, error) {
	query := `
		SELECT id, checkout_session_id, user_id, store_id, amount_paise,
		       loyalty_points_used, loyalty_discount_paise, payable_amount_paise,
		       payment_method, gateway, gateway_order_id, gateway_payment_id,
		       status, failure_reason, created_at, updated_at
		FROM payments WHERE id = $1
	`
	var p Payment
	err := r.db.QueryRowContext(ctx, query, paymentID).Scan(
		&p.ID, &p.CheckoutSessionID, &p.UserID, &p.StoreID, &p.AmountPaise,
		&p.LoyaltyPointsUsed, &p.LoyaltyDiscountPaise, &p.PayableAmountPaise,
		&p.PaymentMethod, &p.Gateway, &p.GatewayOrderID, &p.GatewayPaymentID,
		&p.Status, &p.FailureReason, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) GetPaymentBySessionID(ctx context.Context, sessionID string) (*Payment, error) {
	query := `
		SELECT id, checkout_session_id, user_id, store_id, amount_paise,
		       loyalty_points_used, loyalty_discount_paise, payable_amount_paise,
		       payment_method, gateway, gateway_order_id, gateway_payment_id,
		       status, failure_reason, created_at, updated_at
		FROM payments WHERE checkout_session_id = $1
	`
	var p Payment
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&p.ID, &p.CheckoutSessionID, &p.UserID, &p.StoreID, &p.AmountPaise,
		&p.LoyaltyPointsUsed, &p.LoyaltyDiscountPaise, &p.PayableAmountPaise,
		&p.PaymentMethod, &p.Gateway, &p.GatewayOrderID, &p.GatewayPaymentID,
		&p.Status, &p.FailureReason, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) UpdatePaymentGatewayInfo(ctx context.Context, paymentID string, gateway string, gatewayOrderID string, status string) error {
	query := `
		UPDATE payments
		SET gateway = $1, gateway_order_id = $2, status = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, gateway, gatewayOrderID, status, paymentID)
	return err
}

func (r *PostgresRepository) ResetFailedPayment(ctx context.Context, paymentID string, gateway string, payableAmount int64) error {
	query := `
		UPDATE payments
		SET gateway = $1, status = $2, failure_reason = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, gateway, StatusInitiated, paymentID)
	return err
}

func (r *PostgresRepository) RecordWebhookEventIdempotent(ctx context.Context, gateway, gatewayEventID, eventType string, rawPayload []byte) (bool, error) {
	query := `
		INSERT INTO payment_webhook_events (id, gateway, gateway_event_id, event_type, raw_payload)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (gateway_event_id) DO NOTHING
	`
	res, err := r.db.ExecContext(ctx, query, uuid.New().String(), gateway, gatewayEventID, eventType, rawPayload)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *PostgresRepository) CapturePaymentAndOutboxTx(ctx context.Context, paymentID string, gatewayPaymentID string, topic string, payload []byte) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. UPDATE payments status to CAPTURED
	pQuery := `
		UPDATE payments
		SET status = $1, gateway_payment_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND status != $1
	`
	_, err = tx.ExecContext(ctx, pQuery, StatusCaptured, gatewayPaymentID, paymentID)
	if err != nil {
		return err
	}

	// 2. INSERT into payment_outbox
	oQuery := `
		INSERT INTO payment_outbox (id, topic, payload, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`
	_, err = tx.ExecContext(ctx, oQuery, uuid.New().String(), topic, payload)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) FailPaymentAndReleaseTx(ctx context.Context, paymentID string, failureReason string) error {
	query := `
		UPDATE payments
		SET status = $1, failure_reason = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, StatusFailed, failureReason, paymentID)
	return err
}

func (r *PostgresRepository) InsertCashPaymentTx(ctx context.Context, p *Payment, topic string, payload []byte) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now()

	pQuery := `
		INSERT INTO payments (
			id, checkout_session_id, user_id, store_id, amount_paise,
			loyalty_points_used, loyalty_discount_paise, payable_amount_paise,
			payment_method, gateway, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (checkout_session_id) DO NOTHING
	`
	res, err := tx.ExecContext(ctx, pQuery,
		p.ID, p.CheckoutSessionID, p.UserID, p.StoreID, p.AmountPaise,
		p.LoyaltyPointsUsed, p.LoyaltyDiscountPaise, p.PayableAmountPaise,
		MethodCash, GatewayCash, StatusCaptured, now, now,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return nil // Already existing row
	}

	// INSERT into payment_outbox
	oQuery := `
		INSERT INTO payment_outbox (id, topic, payload, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`
	_, err = tx.ExecContext(ctx, oQuery, uuid.New().String(), topic, payload)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) InsertRefund(ctx context.Context, refund *Refund) error {
	if refund.ID == "" {
		refund.ID = uuid.New().String()
	}
	query := `
		INSERT INTO refunds (id, payment_id, amount_paise, reason, gateway_refund_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`
	_, err := r.db.ExecContext(ctx, query, refund.ID, refund.PaymentID, refund.AmountPaise, refund.Reason, refund.GatewayRefundID, refund.Status)
	return err
}

func (r *PostgresRepository) GetPendingTimeoutPayments(ctx context.Context, cutoff time.Duration) ([]*Payment, error) {
	cutoffTime := time.Now().Add(-cutoff)
	query := `
		SELECT id, checkout_session_id, user_id, store_id, amount_paise,
		       loyalty_points_used, loyalty_discount_paise, payable_amount_paise,
		       payment_method, gateway, gateway_order_id, gateway_payment_id,
		       status, failure_reason, created_at, updated_at
		FROM payments
		WHERE status IN ('INITIATED', 'PENDING') AND created_at < $1
	`
	rows, err := r.db.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(
			&p.ID, &p.CheckoutSessionID, &p.UserID, &p.StoreID, &p.AmountPaise,
			&p.LoyaltyPointsUsed, &p.LoyaltyDiscountPaise, &p.PayableAmountPaise,
			&p.PaymentMethod, &p.Gateway, &p.GatewayOrderID, &p.GatewayPaymentID,
			&p.Status, &p.FailureReason, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payments = append(payments, &p)
	}
	return payments, nil
}

func (r *PostgresRepository) GetCapturedPaymentsByDate(ctx context.Context, dateStr string) ([]*Payment, error) {
	query := `
		SELECT id, checkout_session_id, user_id, store_id, amount_paise,
		       loyalty_points_used, loyalty_discount_paise, payable_amount_paise,
		       payment_method, gateway, gateway_order_id, gateway_payment_id,
		       status, failure_reason, created_at, updated_at
		FROM payments
		WHERE status IN ('CAPTURED', 'COMPLETED') AND DATE(created_at) = $1
	`
	rows, err := r.db.QueryContext(ctx, query, dateStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(
			&p.ID, &p.CheckoutSessionID, &p.UserID, &p.StoreID, &p.AmountPaise,
			&p.LoyaltyPointsUsed, &p.LoyaltyDiscountPaise, &p.PayableAmountPaise,
			&p.PaymentMethod, &p.Gateway, &p.GatewayOrderID, &p.GatewayPaymentID,
			&p.Status, &p.FailureReason, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payments = append(payments, &p)
	}
	return payments, nil
}
