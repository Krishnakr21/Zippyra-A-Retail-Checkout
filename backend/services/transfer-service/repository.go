package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Repository interface {
	CreateTransfer(ctx context.Context, transfer *TransferOrder, items []TransferLineItem) error
	GetTransferByID(ctx context.Context, id string) (*TransferOrder, error)
	UpdateTransferStatus(ctx context.Context, id string, status string, rejectionReason *string) error
	ShipTransfer(ctx context.Context, id string, shippedQtys map[string]int) error
	ReceiveTransfer(ctx context.Context, id string, receivedQtys map[string]int) error
	ListTransfers(ctx context.Context, sourceStoreID, destStoreID, chainID, status string) ([]*TransferOrder, error)
}

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) CreateTransfer(ctx context.Context, transfer *TransferOrder, items []TransferLineItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if transfer.ID == "" {
		transfer.ID = uuid.New().String()
	}
	now := time.Now()
	if transfer.CreatedAt.IsZero() {
		transfer.CreatedAt = now
	}
	transfer.UpdatedAt = now

	insertOrderQuery := `
		INSERT INTO transfer_orders (id, source_store_id, dest_store_id, chain_id, status, requested_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.ExecContext(ctx, insertOrderQuery, transfer.ID, transfer.SourceStoreID, transfer.DestStoreID, transfer.ChainID, transfer.Status, transfer.RequestedBy, transfer.CreatedAt, transfer.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert transfer_order: %w", err)
	}

	insertItemQuery := `
		INSERT INTO transfer_line_items (id, transfer_id, barcode, qty_requested)
		VALUES ($1, $2, $3, $4)
	`
	for i := range items {
		if items[i].ID == "" {
			items[i].ID = uuid.New().String()
		}
		items[i].TransferID = transfer.ID
		_, err = tx.ExecContext(ctx, insertItemQuery, items[i].ID, transfer.ID, items[i].Barcode, items[i].QtyRequested)
		if err != nil {
			return fmt.Errorf("failed to insert transfer_line_item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *postgresRepo) GetTransferByID(ctx context.Context, id string) (*TransferOrder, error) {
	queryOrder := `
		SELECT id, source_store_id, dest_store_id, chain_id, status, requested_by, rejection_reason, shipped_at, received_at, created_at, updated_at
		FROM transfer_orders
		WHERE id = $1
	`
	var t TransferOrder
	var reqBy, rejReason sql.NullString
	var shippedAt, receivedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, queryOrder, id).Scan(
		&t.ID, &t.SourceStoreID, &t.DestStoreID, &t.ChainID, &t.Status, &reqBy, &rejReason, &shippedAt, &receivedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if reqBy.Valid {
		t.RequestedBy = reqBy.String
	}
	if rejReason.Valid {
		t.RejectionReason = &rejReason.String
	}
	if shippedAt.Valid {
		t.ShippedAt = &shippedAt.Time
	}
	if receivedAt.Valid {
		t.ReceivedAt = &receivedAt.Time
	}

	queryItems := `
		SELECT id, transfer_id, barcode, qty_requested, qty_shipped, qty_received
		FROM transfer_line_items
		WHERE transfer_id = $1
		ORDER BY id
	`
	rows, err := r.db.QueryContext(ctx, queryItems, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item TransferLineItem
		var qShipped, qReceived sql.NullInt32
		if err := rows.Scan(&item.ID, &item.TransferID, &item.Barcode, &item.QtyRequested, &qShipped, &qReceived); err != nil {
			return nil, err
		}
		if qShipped.Valid {
			v := int(qShipped.Int32)
			item.QtyShipped = &v
		}
		if qReceived.Valid {
			v := int(qReceived.Int32)
			item.QtyReceived = &v
		}
		t.LineItems = append(t.LineItems, item)
	}

	return &t, nil
}

func (r *postgresRepo) UpdateTransferStatus(ctx context.Context, id string, status string, rejectionReason *string) error {
	now := time.Now()
	query := `
		UPDATE transfer_orders
		SET status = $1, rejection_reason = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, status, rejectionReason, now, id)
	return err
}

func (r *postgresRepo) ShipTransfer(ctx context.Context, id string, shippedQtys map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.ExecContext(ctx, `UPDATE transfer_orders SET status = $1, shipped_at = $2, updated_at = $3 WHERE id = $4`, TransferStatusInTransit, now, now, id)
	if err != nil {
		return err
	}

	updateItemQuery := `UPDATE transfer_line_items SET qty_shipped = $1 WHERE transfer_id = $2 AND barcode = $3`
	for barcode, qty := range shippedQtys {
		_, err = tx.ExecContext(ctx, updateItemQuery, qty, id, barcode)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresRepo) ReceiveTransfer(ctx context.Context, id string, receivedQtys map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.ExecContext(ctx, `UPDATE transfer_orders SET status = $1, received_at = $2, updated_at = $3 WHERE id = $4`, TransferStatusReceived, now, now, id)
	if err != nil {
		return err
	}

	updateItemQuery := `UPDATE transfer_line_items SET qty_received = $1 WHERE transfer_id = $2 AND barcode = $3`
	for barcode, qty := range receivedQtys {
		_, err = tx.ExecContext(ctx, updateItemQuery, qty, id, barcode)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresRepo) ListTransfers(ctx context.Context, sourceStoreID, destStoreID, chainID, status string) ([]*TransferOrder, error) {
	query := `
		SELECT id, source_store_id, dest_store_id, chain_id, status, requested_by, rejection_reason, shipped_at, received_at, created_at, updated_at
		FROM transfer_orders
		WHERE ($1 = '' OR source_store_id = $1)
		  AND ($2 = '' OR dest_store_id = $2)
		  AND ($3 = '' OR chain_id = $3)
		  AND ($4 = '' OR status = $4)
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, sourceStoreID, destStoreID, chainID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*TransferOrder
	for rows.Next() {
		var t TransferOrder
		var reqBy, rejReason sql.NullString
		var shippedAt, receivedAt sql.NullTime

		if err := rows.Scan(&t.ID, &t.SourceStoreID, &t.DestStoreID, &t.ChainID, &t.Status, &reqBy, &rejReason, &shippedAt, &receivedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if reqBy.Valid {
			t.RequestedBy = reqBy.String
		}
		if rejReason.Valid {
			t.RejectionReason = &rejReason.String
		}
		if shippedAt.Valid {
			t.ShippedAt = &shippedAt.Time
		}
		if receivedAt.Valid {
			t.ReceivedAt = &receivedAt.Time
		}
		results = append(results, &t)
	}

	for _, t := range results {
		itemsQuery := `SELECT id, transfer_id, barcode, qty_requested, qty_shipped, qty_received FROM transfer_line_items WHERE transfer_id = $1 ORDER BY id`
		itemRows, err := r.db.QueryContext(ctx, itemsQuery, t.ID)
		if err == nil {
			for itemRows.Next() {
				var item TransferLineItem
				var qShipped, qReceived sql.NullInt32
				if err := itemRows.Scan(&item.ID, &item.TransferID, &item.Barcode, &item.QtyRequested, &qShipped, &qReceived); err == nil {
					if qShipped.Valid {
						v := int(qShipped.Int32)
						item.QtyShipped = &v
					}
					if qReceived.Valid {
						v := int(qReceived.Int32)
						item.QtyReceived = &v
					}
					t.LineItems = append(t.LineItems, item)
				}
			}
			itemRows.Close()
		}
	}

	return results, nil
}
