package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateOrderAndOutboxTx(ctx context.Context, order *Order, flags []OrderItemReturnableFlag, exitSvc ExitTokenService, topic string, payload []byte) (bool, error)
	CreateOrderFailureOutboxTx(ctx context.Context, paymentID, userID, storeID string, amountPaise int64, reason string) error
	GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) ([]OrderSummary, error)
	GetOrderByID(ctx context.Context, orderID string) (*Order, error)
	GetOrderByPaymentID(ctx context.Context, paymentID string) (*Order, error)
	GetReturnableFlags(ctx context.Context, orderID string) ([]OrderItemReturnableFlag, error)
	GetOrdersByStoreID(ctx context.Context, storeID string, page, pageSize int) ([]OrderSummary, error)
	UpdateOrderStatusAndPublishOutbox(ctx context.Context, orderID, status string, topic string, payload []byte) error
	CreateReturnRequestTx(ctx context.Context, req *ReturnRequest, flagUpdates map[string]int, updateOrderStatus bool) error
	UpdateOrderInvoiceAndIRN(ctx context.Context, orderID string, invoiceKey, irn, irnAckNo *string, irnAckDate *time.Time, irnQRCode *string) error
	ListChainOrders(ctx context.Context, chainID string, page, pageSize int) ([]*Order, int, error)
	AnonymizeUserOrders(ctx context.Context, userID string) (int, error)
	LookupCustomerByPhoneLast4(ctx context.Context, storeID, phoneLast4 string, timeWindow time.Duration) ([]CustomerLookupMatch, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateOrderAndOutboxTx(ctx context.Context, order *Order, flags []OrderItemReturnableFlag, exitSvc ExitTokenService, topic string, payload []byte) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	now := time.Now()
	order.CreatedAt = now
	order.CompletedAt = &now
	order.Status = StatusCompleted

	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return false, fmt.Errorf("failed to marshal items: %w", err)
	}

	orderQuery := `
		INSERT INTO orders (
			id, payment_id, user_id, store_id, items,
			subtotal_paise, discount_paise, cgst_paise, sgst_paise, igst_paise, total_paise,
			loyalty_points_used, payment_method, supply_type, status, created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (payment_id) DO NOTHING
	`
	res, err := tx.ExecContext(ctx, orderQuery,
		order.ID, order.PaymentID, order.UserID, order.StoreID, itemsJSON,
		order.SubtotalPaise, order.DiscountPaise, order.CGSTPaise, order.SGSTPaise, order.IGSTPaise, order.TotalPaise,
		order.LoyaltyPointsUsed, order.PaymentMethod, order.SupplyType, order.Status, order.CreatedAt, order.CompletedAt,
	)
	if err != nil {
		return false, fmt.Errorf("failed to insert order: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		// Idempotency conflict hit: order already created for this payment_id
		return false, nil
	}

	// Insert order items returnable flags
	flagQuery := `
		INSERT INTO order_items_returnable_flags (order_id, barcode, is_returnable, returned_qty)
		VALUES ($1, $2, $3, $4)
	`
	for _, flag := range flags {
		_, err := tx.ExecContext(ctx, flagQuery, order.ID, flag.Barcode, flag.IsReturnable, flag.ReturnedQty)
		if err != nil {
			return false, fmt.Errorf("failed to insert returnable flag for barcode %s: %w", flag.Barcode, err)
		}
	}

	// Issue exit token before committing transaction
	if exitSvc != nil {
		_, exitErr := exitSvc.IssueAndStoreExitToken(ctx, order.ID, order.UserID, order.StoreID, order.SessionID)
		if exitErr != nil {
			return false, fmt.Errorf("exit token issuance failed: %w", exitErr)
		}
	} else {
		return false, fmt.Errorf("exit token service unavailable")
	}

	// Insert order_creation_outbox row
	outboxID := uuid.New().String()
	outboxQuery := `
		INSERT INTO order_creation_outbox (id, topic, payload, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.ExecContext(ctx, outboxQuery, outboxID, topic, payload, now)
	if err != nil {
		return false, fmt.Errorf("failed to insert order outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("failed to commit order creation tx: %w", err)
	}

	return true, nil
}

func (r *PostgresRepository) CreateOrderFailureOutboxTx(ctx context.Context, paymentID, userID, storeID string, amountPaise int64, reason string) error {
	payload := OrderCreationFailedPayload{
		PaymentID:   paymentID,
		UserID:      userID,
		StoreID:     storeID,
		AmountPaise: amountPaise,
		Reason:      reason,
		Timestamp:   time.Now(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal order creation failure payload: %w", err)
	}

	outboxID := uuid.New().String()
	outboxQuery := `
		INSERT INTO order_creation_outbox (id, topic, payload, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`
	_, err = r.db.ExecContext(ctx, outboxQuery, outboxID, TopicOrderCreationFailed, payloadBytes)
	if err != nil {
		return fmt.Errorf("failed to insert compensating order.creation_failed outbox event: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) ([]OrderSummary, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT id, store_id, total_paise, jsonb_array_length(items) as item_count, status, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query user orders: %w", err)
	}
	defer rows.Close()

	var summaries []OrderSummary
	for rows.Next() {
		var s OrderSummary
		if err := rows.Scan(&s.ID, &s.StoreID, &s.TotalPaise, &s.ItemCount, &s.Status, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order summary row: %w", err)
		}
		s.StoreName = "Zippyra Store " + s.StoreID
		summaries = append(summaries, s)
	}

	return summaries, nil
}

func (r *PostgresRepository) GetOrderByID(ctx context.Context, orderID string) (*Order, error) {
	query := `
		SELECT id, payment_id, user_id, store_id, items, subtotal_paise, discount_paise,
		       cgst_paise, sgst_paise, igst_paise, total_paise, loyalty_points_used,
		       payment_method, supply_type, status, invoice_s3_key, irn, irn_ack_no,
		       irn_ack_date, irn_qr_code, created_at, completed_at
		FROM orders
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, orderID)

	var o Order
	var itemsJSON []byte
	err := row.Scan(
		&o.ID, &o.PaymentID, &o.UserID, &o.StoreID, &itemsJSON, &o.SubtotalPaise, &o.DiscountPaise,
		&o.CGSTPaise, &o.SGSTPaise, &o.IGSTPaise, &o.TotalPaise, &o.LoyaltyPointsUsed,
		&o.PaymentMethod, &o.SupplyType, &o.Status, &o.InvoiceS3Key, &o.IRN, &o.IRNAckNo,
		&o.IRNAckDate, &o.IRNQRCode, &o.CreatedAt, &o.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to scan order row: %w", err)
	}

	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order items: %w", err)
	}

	return &o, nil
}

func (r *PostgresRepository) GetOrderByPaymentID(ctx context.Context, paymentID string) (*Order, error) {
	query := `
		SELECT id, payment_id, user_id, store_id, items, subtotal_paise, discount_paise,
		       cgst_paise, sgst_paise, igst_paise, total_paise, loyalty_points_used,
		       payment_method, supply_type, status, invoice_s3_key, irn, irn_ack_no,
		       irn_ack_date, irn_qr_code, created_at, completed_at
		FROM orders
		WHERE payment_id = $1
	`
	row := r.db.QueryRowContext(ctx, query, paymentID)

	var o Order
	var itemsJSON []byte
	err := row.Scan(
		&o.ID, &o.PaymentID, &o.UserID, &o.StoreID, &itemsJSON, &o.SubtotalPaise, &o.DiscountPaise,
		&o.CGSTPaise, &o.SGSTPaise, &o.IGSTPaise, &o.TotalPaise, &o.LoyaltyPointsUsed,
		&o.PaymentMethod, &o.SupplyType, &o.Status, &o.InvoiceS3Key, &o.IRN, &o.IRNAckNo,
		&o.IRNAckDate, &o.IRNQRCode, &o.CreatedAt, &o.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to scan order row: %w", err)
	}

	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order items: %w", err)
	}

	return &o, nil
}

func (r *PostgresRepository) GetReturnableFlags(ctx context.Context, orderID string) ([]OrderItemReturnableFlag, error) {
	query := `
		SELECT order_id, barcode, is_returnable, returned_qty
		FROM order_items_returnable_flags
		WHERE order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query returnable flags: %w", err)
	}
	defer rows.Close()

	var flags []OrderItemReturnableFlag
	for rows.Next() {
		var f OrderItemReturnableFlag
		if err := rows.Scan(&f.OrderID, &f.Barcode, &f.IsReturnable, &f.ReturnedQty); err != nil {
			return nil, fmt.Errorf("failed to scan returnable flag row: %w", err)
		}
		flags = append(flags, f)
	}

	return flags, nil
}

func (r *PostgresRepository) CreateReturnRequestTx(ctx context.Context, req *ReturnRequest, flagUpdates map[string]int, updateOrderStatus bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.CreatedAt = time.Now()

	itemsJSON, err := json.Marshal(req.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal return items: %w", err)
	}

	// Insert return request row
	insertQuery := `
		INSERT INTO return_requests (id, order_id, user_id, store_id, items, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, insertQuery, req.ID, req.OrderID, req.UserID, req.StoreID, itemsJSON, req.Status, req.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert return request: %w", err)
	}

	// Update returned_qty in order_items_returnable_flags
	updateFlagQuery := `
		UPDATE order_items_returnable_flags
		SET returned_qty = returned_qty + $1
		WHERE order_id = $2 AND barcode = $3
	`
	for barcode, qty := range flagUpdates {
		_, err := tx.ExecContext(ctx, updateFlagQuery, qty, req.OrderID, barcode)
		if err != nil {
			return fmt.Errorf("failed to update returned_qty for barcode %s: %w", barcode, err)
		}
	}

	// Update order status if requested
	if updateOrderStatus {
		updateOrderQuery := `UPDATE orders SET status = $1 WHERE id = $2 AND status = 'COMPLETED'`
		_, err = tx.ExecContext(ctx, updateOrderQuery, StatusReturnRequested, req.OrderID)
		if err != nil {
			return fmt.Errorf("failed to update order status to RETURN_REQUESTED: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) UpdateOrderInvoiceAndIRN(ctx context.Context, orderID string, invoiceKey, irn, irnAckNo *string, irnAckDate *time.Time, irnQRCode *string) error {
	query := `
		UPDATE orders
		SET invoice_s3_key = COALESCE($1, invoice_s3_key),
		    irn = COALESCE($2, irn),
		    irn_ack_no = COALESCE($3, irn_ack_no),
		    irn_ack_date = COALESCE($4, irn_ack_date),
		    irn_qr_code = COALESCE($5, irn_qr_code)
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query, invoiceKey, irn, irnAckNo, irnAckDate, irnQRCode, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order invoice/IRN fields: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetOrdersByStoreID(ctx context.Context, storeID string, page, pageSize int) ([]OrderSummary, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT id, payment_id, user_id, store_id, total_paise, payment_method, status, created_at
		FROM orders
		WHERE store_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, storeID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query store orders: %w", err)
	}
	defer rows.Close()

	var summaries []OrderSummary
	for rows.Next() {
		var s OrderSummary
		if err := rows.Scan(&s.ID, &s.PaymentID, &s.UserID, &s.StoreID, &s.TotalPaise, &s.PaymentMethod, &s.Status, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order summary: %w", err)
		}
		summaries = append(summaries, s)
	}

	if summaries == nil {
		summaries = []OrderSummary{}
	}

	return summaries, nil
}

func (r *PostgresRepository) UpdateOrderStatusAndPublishOutbox(ctx context.Context, orderID, status string, topic string, payload []byte) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	updateQuery := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if topic != "" && len(payload) > 0 {
		outboxQuery := `
			INSERT INTO outbox_events (id, topic, aggregate_id, payload, created_at, processed)
			VALUES ($1, $2, $3, $4, $5, FALSE)
		`
		outboxID := uuid.New().String()
		_, err = tx.ExecContext(ctx, outboxQuery, outboxID, topic, orderID, payload, time.Now())
		if err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) ListChainOrders(ctx context.Context, chainID string, page, pageSize int) ([]*Order, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT o.id, o.user_id, o.store_id, o.payment_id, o.status, o.total_amount_paise, o.created_at
		FROM orders o
		JOIN stores s ON o.store_id = s.id
		WHERE s.chain_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, chainID, pageSize, offset)
	if err != nil {
		return []*Order{}, 0, nil
	}
	defer rows.Close()

	var result []*Order
	for rows.Next() {
		o := &Order{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.StoreID, &o.PaymentID, &o.Status, &o.TotalPaise, &o.CreatedAt); err == nil {
			result = append(result, o)
		}
	}
	return result, len(result), nil
}

func (r *PostgresRepository) AnonymizeUserOrders(ctx context.Context, userID string) (int, error) {
	tombstone := "00000000-0000-0000-0000-000000000000"
	res, err := r.db.ExecContext(ctx, "UPDATE orders SET user_id = $1 WHERE user_id = $2", tombstone, userID)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return int(rows), nil
}

func (r *PostgresRepository) LookupCustomerByPhoneLast4(ctx context.Context, storeID, phoneLast4 string, timeWindow time.Duration) ([]CustomerLookupMatch, error) {
	cutoff := time.Now().Add(-timeWindow)

	query := `
		SELECT o.user_id, o.id, o.status, o.created_at
		FROM orders o
		WHERE o.store_id = $1 AND o.created_at >= $2
		ORDER BY o.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, storeID, cutoff)
	if err != nil {
		return []CustomerLookupMatch{}, nil
	}
	defer rows.Close()

	userMap := make(map[string]bool)
	var results []CustomerLookupMatch

	for rows.Next() {
		var userID, orderID, status string
		var createdAt time.Time
		if err := rows.Scan(&userID, &orderID, &status, &createdAt); err != nil {
			continue
		}

		if strings.HasSuffix(userID, phoneLast4) || (len(userID) >= 4 && userID[len(userID)-4:] == phoneLast4) {
			if !userMap[userID] {
				userMap[userID] = true
				firstName := "Customer"
				if len(userID) > 4 {
					firstName = "User-" + userID[:3]
				}
				results = append(results, CustomerLookupMatch{
					CustomerID:        userID,
					FirstName:         firstName,
					PhoneMasked:       "+91XXXXXX" + phoneLast4,
					StoreID:           storeID,
					HasActiveSession:  status == StatusCreated || status == StatusCompleted,
					SessionID:         "sess-" + userID[:min(len(userID), 6)],
					ActiveOrderID:     orderID,
					ActiveOrderStatus: status,
					CreatedAt:         createdAt,
				})
			}
		}
	}

	return results, nil
}
