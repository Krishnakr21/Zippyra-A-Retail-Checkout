package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreatePO(ctx context.Context, po *PurchaseOrder, items []POLineItem) error
	GetPOByID(ctx context.Context, id string) (*PurchaseOrder, error)
	ListPOs(ctx context.Context, storeID, status string, limit, offset int) ([]PurchaseOrder, error)
	SubmitPO(ctx context.Context, id string) error

	CreateGRN(ctx context.Context, grn *GoodsReceivedNote, items []GRNLineItem) error
	GetGRNByID(ctx context.Context, id string) (*GoodsReceivedNote, error)
	UpdateGRNQC(ctx context.Context, grnID string, updates []QCUpdateItem) error
	CompleteGRN(ctx context.Context, grnID string, poID *string, receivedQtys map[string]int) error

	CreateTransfer(ctx context.Context, transfer *TransferOrder, items []TransferLineItem) error
	GetTransferByID(ctx context.Context, id string) (*TransferOrder, error)
	UpdateTransferStatus(ctx context.Context, id, status string, rejectionReason *string) error
	ShipTransfer(ctx context.Context, id string, shippedQtys map[string]int) error
	ReceiveTransfer(ctx context.Context, id string, receivedQtys map[string]int) error
	ListTransfersByChainID(ctx context.Context, chainID, status string) ([]*TransferOrder, error)

	CreateAutoPO(ctx context.Context, storeID, chainID, barcode string, reorderQty int) (*PurchaseOrder, error)

	GetDB() *sql.DB
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

func (r *PostgresRepository) CreatePO(ctx context.Context, po *PurchaseOrder, items []POLineItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if po.ID == "" {
		po.ID = uuid.New().String()
	}
	po.CreatedAt = time.Now()

	query := `
		INSERT INTO purchase_orders (id, store_id, chain_id, vendor_name, status, source, created_by, expected_delivery_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = tx.ExecContext(ctx, query, po.ID, po.StoreID, po.ChainID, po.VendorName, po.Status, po.Source, po.CreatedBy, po.ExpectedDeliveryDate, po.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert purchase order: %w", err)
	}

	itemQuery := `
		INSERT INTO po_line_items (id, po_id, barcode, qty_ordered, unit_cost_paise, qty_received)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for i := range items {
		items[i].ID = uuid.New().String()
		items[i].POID = po.ID
		_, err = tx.ExecContext(ctx, itemQuery, items[i].ID, po.ID, items[i].Barcode, items[i].QtyOrdered, items[i].UnitCostPaise, 0)
		if err != nil {
			return fmt.Errorf("failed to insert po line item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetPOByID(ctx context.Context, id string) (*PurchaseOrder, error) {
	poQuery := `
		SELECT id, store_id, chain_id, vendor_name, status, source, created_by, auto_reorder_item_barcode, auto_reorder_date, expected_delivery_date, created_at, submitted_at, completed_at
		FROM purchase_orders
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, poQuery, id)

	var po PurchaseOrder
	err := row.Scan(&po.ID, &po.StoreID, &po.ChainID, &po.VendorName, &po.Status, &po.Source, &po.CreatedBy, &po.AutoReorderItemBarcode, &po.AutoReorderDate, &po.ExpectedDeliveryDate, &po.CreatedAt, &po.SubmittedAt, &po.CompletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query purchase order: %w", err)
	}

	itemQuery := `
		SELECT id, po_id, barcode, qty_ordered, unit_cost_paise, qty_received
		FROM po_line_items
		WHERE po_id = $1
	`
	rows, err := r.db.QueryContext(ctx, itemQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query po line items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item POLineItem
		if err := rows.Scan(&item.ID, &item.POID, &item.Barcode, &item.QtyOrdered, &item.UnitCostPaise, &item.QtyReceived); err != nil {
			return nil, fmt.Errorf("failed to scan po line item: %w", err)
		}
		po.LineItems = append(po.LineItems, item)
	}

	return &po, nil
}

func (r *PostgresRepository) ListPOs(ctx context.Context, storeID, status string, limit, offset int) ([]PurchaseOrder, error) {
	query := `
		SELECT id, store_id, chain_id, vendor_name, status, source, created_by, auto_reorder_item_barcode, auto_reorder_date, expected_delivery_date, created_at, submitted_at, completed_at
		FROM purchase_orders
		WHERE store_id = $1
	`
	args := []interface{}{storeID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d OFFSET %d", limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchase orders: %w", err)
	}
	defer rows.Close()

	var pos []PurchaseOrder
	for rows.Next() {
		var po PurchaseOrder
		if err := rows.Scan(&po.ID, &po.StoreID, &po.ChainID, &po.VendorName, &po.Status, &po.Source, &po.CreatedBy, &po.AutoReorderItemBarcode, &po.AutoReorderDate, &po.ExpectedDeliveryDate, &po.CreatedAt, &po.SubmittedAt, &po.CompletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan purchase order: %w", err)
		}
		pos = append(pos, po)
	}

	return pos, nil
}

func (r *PostgresRepository) SubmitPO(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE purchase_orders
		SET status = $1, submitted_at = $2
		WHERE id = $3 AND status = $4
	`
	res, err := r.db.ExecContext(ctx, query, POStatusSubmitted, now, id, POStatusDraft)
	if err != nil {
		return fmt.Errorf("failed to submit purchase order: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("PO_ALREADY_SUBMITTED")
	}
	return nil
}

func (r *PostgresRepository) CreateGRN(ctx context.Context, grn *GoodsReceivedNote, items []GRNLineItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if grn.ID == "" {
		grn.ID = uuid.New().String()
	}
	grn.CreatedAt = time.Now()

	query := `
		INSERT INTO goods_received_notes (id, po_id, store_id, received_by, vendor_invoice_ref, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, query, grn.ID, grn.POID, grn.StoreID, grn.ReceivedBy, grn.VendorInvoiceRef, grn.Status, grn.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert GRN: %w", err)
	}

	itemQuery := `
		INSERT INTO grn_line_items (id, grn_id, barcode, qty_expected, qty_received, unit_cost_paise, qc_status, qc_note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	for i := range items {
		items[i].ID = uuid.New().String()
		items[i].GRNID = grn.ID
		_, err = tx.ExecContext(ctx, itemQuery, items[i].ID, grn.ID, items[i].Barcode, items[i].QtyExpected, items[i].QtyReceived, items[i].UnitCostPaise, items[i].QCStatus, items[i].QCNote)
		if err != nil {
			return fmt.Errorf("failed to insert GRN line item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetGRNByID(ctx context.Context, id string) (*GoodsReceivedNote, error) {
	grnQuery := `
		SELECT id, po_id, store_id, received_by, vendor_invoice_ref, status, created_at, completed_at
		FROM goods_received_notes
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, grnQuery, id)

	var grn GoodsReceivedNote
	err := row.Scan(&grn.ID, &grn.POID, &grn.StoreID, &grn.ReceivedBy, &grn.VendorInvoiceRef, &grn.Status, &grn.CreatedAt, &grn.CompletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query GRN: %w", err)
	}

	itemQuery := `
		SELECT id, grn_id, barcode, qty_expected, qty_received, unit_cost_paise, qc_status, qc_note
		FROM grn_line_items
		WHERE grn_id = $1
	`
	rows, err := r.db.QueryContext(ctx, itemQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query GRN line items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item GRNLineItem
		if err := rows.Scan(&item.ID, &item.GRNID, &item.Barcode, &item.QtyExpected, &item.QtyReceived, &item.UnitCostPaise, &item.QCStatus, &item.QCNote); err != nil {
			return nil, fmt.Errorf("failed to scan GRN line item: %w", err)
		}
		grn.LineItems = append(grn.LineItems, item)
	}

	return &grn, nil
}

func (r *PostgresRepository) UpdateGRNQC(ctx context.Context, grnID string, updates []QCUpdateItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE grn_line_items
		SET qc_status = $1, qc_note = $2
		WHERE id = $3 AND grn_id = $4
	`
	for _, update := range updates {
		_, err := tx.ExecContext(ctx, query, update.QCStatus, update.QCNote, update.GRNLineItemID, grnID)
		if err != nil {
			return fmt.Errorf("failed to update GRN line item QC: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) CompleteGRN(ctx context.Context, grnID string, poID *string, receivedQtys map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.ExecContext(ctx, `UPDATE goods_received_notes SET status = $1, completed_at = $2 WHERE id = $3`, GRNStatusCompleted, now, grnID)
	if err != nil {
		return fmt.Errorf("failed to set GRN status completed: %w", err)
	}

	if poID != nil && *poID != "" {
		for barcode, qty := range receivedQtys {
			_, err = tx.ExecContext(ctx, `UPDATE po_line_items SET qty_received = qty_received + $1 WHERE po_id = $2 AND barcode = $3`, qty, *poID, barcode)
			if err != nil {
				return fmt.Errorf("failed to update po_line_item qty_received: %w", err)
			}
		}

		// Recompute PO status
		var totalOrdered, totalReceived int
		err = tx.QueryRowContext(ctx, `SELECT SUM(qty_ordered), SUM(qty_received) FROM po_line_items WHERE po_id = $1`, *poID).Scan(&totalOrdered, &totalReceived)
		if err == nil {
			newPOStatus := POStatusPartiallyReceived
			if totalReceived >= totalOrdered {
				newPOStatus = POStatusReceived
			}
			_, _ = tx.ExecContext(ctx, `UPDATE purchase_orders SET status = $1, completed_at = CASE WHEN $1 = 'RECEIVED' THEN $2 ELSE completed_at END WHERE id = $3`, newPOStatus, now, *poID)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) CreateTransfer(ctx context.Context, transfer *TransferOrder, items []TransferLineItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if transfer.ID == "" {
		transfer.ID = uuid.New().String()
	}
	transfer.CreatedAt = time.Now()

	query := `
		INSERT INTO transfer_orders (id, source_store_id, dest_store_id, chain_id, status, requested_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, query, transfer.ID, transfer.SourceStoreID, transfer.DestStoreID, transfer.ChainID, transfer.Status, transfer.RequestedBy, transfer.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert transfer order: %w", err)
	}

	itemQuery := `
		INSERT INTO transfer_line_items (id, transfer_id, barcode, qty_requested)
		VALUES ($1, $2, $3, $4)
	`
	for i := range items {
		items[i].ID = uuid.New().String()
		items[i].TransferID = transfer.ID
		_, err = tx.ExecContext(ctx, itemQuery, items[i].ID, transfer.ID, items[i].Barcode, items[i].QtyRequested)
		if err != nil {
			return fmt.Errorf("failed to insert transfer line item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetTransferByID(ctx context.Context, id string) (*TransferOrder, error) {
	query := `
		SELECT id, source_store_id, dest_store_id, chain_id, status, requested_by, rejection_reason, created_at, approved_at, shipped_at, received_at
		FROM transfer_orders
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var t TransferOrder
	err := row.Scan(&t.ID, &t.SourceStoreID, &t.DestStoreID, &t.ChainID, &t.Status, &t.RequestedBy, &t.RejectionReason, &t.CreatedAt, &t.ApprovedAt, &t.ShippedAt, &t.ReceivedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query transfer order: %w", err)
	}

	itemQuery := `
		SELECT id, transfer_id, barcode, qty_requested, qty_shipped, qty_received
		FROM transfer_line_items
		WHERE transfer_id = $1
	`
	rows, err := r.db.QueryContext(ctx, itemQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query transfer line items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item TransferLineItem
		if err := rows.Scan(&item.ID, &item.TransferID, &item.Barcode, &item.QtyRequested, &item.QtyShipped, &item.QtyReceived); err != nil {
			return nil, fmt.Errorf("failed to scan transfer line item: %w", err)
		}
		t.LineItems = append(t.LineItems, item)
	}

	return &t, nil
}

func (r *PostgresRepository) UpdateTransferStatus(ctx context.Context, id, status string, rejectionReason *string) error {
	now := time.Now()
	var query string
	if status == TransferStatusApproved {
		query = `UPDATE transfer_orders SET status = $1, approved_at = $2 WHERE id = $3`
		_, err := r.db.ExecContext(ctx, query, status, now, id)
		return err
	} else if status == TransferStatusRejected {
		query = `UPDATE transfer_orders SET status = $1, rejection_reason = $2 WHERE id = $3`
		_, err := r.db.ExecContext(ctx, query, status, rejectionReason, id)
		return err
	}
	query = `UPDATE transfer_orders SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *PostgresRepository) ShipTransfer(ctx context.Context, id string, shippedQtys map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.ExecContext(ctx, `UPDATE transfer_orders SET status = $1, shipped_at = $2 WHERE id = $3`, TransferStatusInTransit, now, id)
	if err != nil {
		return fmt.Errorf("failed to update transfer status to IN_TRANSIT: %w", err)
	}

	for barcode, qty := range shippedQtys {
		_, err = tx.ExecContext(ctx, `UPDATE transfer_line_items SET qty_shipped = $1 WHERE transfer_id = $2 AND barcode = $3`, qty, id, barcode)
		if err != nil {
			return fmt.Errorf("failed to update transfer line item qty_shipped: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) ReceiveTransfer(ctx context.Context, id string, receivedQtys map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.ExecContext(ctx, `UPDATE transfer_orders SET status = $1, received_at = $2 WHERE id = $3`, TransferStatusReceived, now, id)
	if err != nil {
		return fmt.Errorf("failed to update transfer status to RECEIVED: %w", err)
	}

	for barcode, qty := range receivedQtys {
		_, err = tx.ExecContext(ctx, `UPDATE transfer_line_items SET qty_received = $1 WHERE transfer_id = $2 AND barcode = $3`, qty, id, barcode)
		if err != nil {
			return fmt.Errorf("failed to update transfer line item qty_received: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) CreateAutoPO(ctx context.Context, storeID, chainID, barcode string, reorderQty int) (*PurchaseOrder, error) {
	todayStr := time.Now().Format("2006-01-02")
	poID := uuid.New().String()
	now := time.Now()

	var query string
	if r.isSQLite {
		query = `
			INSERT OR IGNORE INTO purchase_orders (id, store_id, chain_id, vendor_name, status, source, auto_reorder_item_barcode, auto_reorder_date, created_at, submitted_at)
			VALUES ($1, $2, $3, 'Default Auto Vendor', 'SUBMITTED', 'AUTO_REORDER', $4, $5, $6, $6)
		`
	} else {
		query = `
			INSERT INTO purchase_orders (id, store_id, chain_id, vendor_name, status, source, auto_reorder_item_barcode, auto_reorder_date, created_at, submitted_at)
			VALUES ($1, $2, $3, 'Default Auto Vendor', 'SUBMITTED', 'AUTO_REORDER', $4, $5, $6, $6)
			ON CONFLICT (store_id, auto_reorder_item_barcode, auto_reorder_date) DO NOTHING
		`
	}

	res, err := r.db.ExecContext(ctx, query, poID, storeID, chainID, barcode, todayStr, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert auto PO: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		// Already created auto-PO for this store+barcode today!
		return nil, nil
	}

	// Insert line item
	lineID := uuid.New().String()
	itemQuery := `
		INSERT INTO po_line_items (id, po_id, barcode, qty_ordered, unit_cost_paise, qty_received)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = r.db.ExecContext(ctx, itemQuery, lineID, poID, barcode, reorderQty, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to insert auto PO line item: %w", err)
	}

	return &PurchaseOrder{
		ID:                     poID,
		StoreID:                storeID,
		ChainID:                chainID,
		VendorName:             "Default Auto Vendor",
		Status:                 POStatusSubmitted,
		Source:                 POSourceAutoReorder,
		AutoReorderItemBarcode: &barcode,
		AutoReorderDate:        &todayStr,
		CreatedAt:              now,
		SubmittedAt:            &now,
		LineItems: []POLineItem{
			{ID: lineID, POID: poID, Barcode: barcode, QtyOrdered: reorderQty, UnitCostPaise: 0, QtyReceived: 0},
		},
	}, nil
}

func (r *PostgresRepository) ListTransfersByChainID(ctx context.Context, chainID, status string) ([]*TransferOrder, error) {
	query := `SELECT id, source_store_id, dest_store_id, status, created_at FROM stock_transfers`
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []*TransferOrder{}, nil
	}
	defer rows.Close()

	var result []*TransferOrder
	for rows.Next() {
		t := &TransferOrder{}
		if err := rows.Scan(&t.ID, &t.SourceStoreID, &t.DestStoreID, &t.Status, &t.CreatedAt); err == nil {
			result = append(result, t)
		}
	}
	return result, nil
}
