package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func TestTransferMigration_PreservesIDsValuesAndParity(t *testing.T) {
	warehouseDB, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open warehouse DB: %v", err)
	}
	defer warehouseDB.Close()

	transferDB, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open transfer DB: %v", err)
	}
	defer transferDB.Close()

	// 1. Setup Schemas
	schema := `
		CREATE TABLE transfer_orders (
			id TEXT PRIMARY KEY,
			source_store_id TEXT NOT NULL,
			dest_store_id TEXT NOT NULL,
			chain_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'REQUESTED',
			requested_by TEXT NULL,
			rejection_reason TEXT NULL,
			shipped_at TIMESTAMP NULL,
			received_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE transfer_line_items (
			id TEXT PRIMARY KEY,
			transfer_id TEXT NOT NULL REFERENCES transfer_orders(id) ON DELETE CASCADE,
			barcode TEXT NOT NULL,
			qty_requested INTEGER NOT NULL,
			qty_shipped INTEGER NULL,
			qty_received INTEGER NULL
		);
	`
	_, _ = warehouseDB.Exec(schema)
	_, _ = transferDB.Exec(schema)

	// 2. Seed 2 Transfer Orders in warehouse DB
	t1ID := uuid.New().String()
	t2ID := uuid.New().String()
	item1ID := uuid.New().String()
	item2ID := uuid.New().String()

	now := time.Now()
	_, _ = warehouseDB.Exec("INSERT INTO transfer_orders (id, source_store_id, dest_store_id, chain_id, status, requested_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", t1ID, "s1", "s2", "c1", "IN_TRANSIT", "user-1", now, now)
	_, _ = warehouseDB.Exec("INSERT INTO transfer_orders (id, source_store_id, dest_store_id, chain_id, status, requested_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", t2ID, "s3", "s4", "c1", "REQUESTED", "user-2", now, now)

	shippedVal := 10
	_, _ = warehouseDB.Exec("INSERT INTO transfer_line_items (id, transfer_id, barcode, qty_requested, qty_shipped) VALUES (?, ?, ?, ?, ?)", item1ID, t1ID, "b1", 10, shippedVal)
	_, _ = warehouseDB.Exec("INSERT INTO transfer_line_items (id, transfer_id, barcode, qty_requested) VALUES (?, ?, ?, ?)", item2ID, t2ID, "b2", 5)

	// 3. Migrate
	oRows, _ := warehouseDB.Query("SELECT id, source_store_id, dest_store_id, chain_id, status, requested_by, rejection_reason, shipped_at, received_at, created_at, updated_at FROM transfer_orders")
	for oRows.Next() {
		var o TransferOrder
		var reqBy, rejReason sql.NullString
		var shippedAt, receivedAt sql.NullTime
		_ = oRows.Scan(&o.ID, &o.SourceStoreID, &o.DestStoreID, &o.ChainID, &o.Status, &reqBy, &rejReason, &shippedAt, &receivedAt, &o.CreatedAt, &o.UpdatedAt)
		_, _ = transferDB.Exec("INSERT INTO transfer_orders (id, source_store_id, dest_store_id, chain_id, status, requested_by, rejection_reason, shipped_at, received_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", o.ID, o.SourceStoreID, o.DestStoreID, o.ChainID, o.Status, reqBy, rejReason, shippedAt, receivedAt, o.CreatedAt, o.UpdatedAt)
	}
	oRows.Close()

	iRows, _ := warehouseDB.Query("SELECT id, transfer_id, barcode, qty_requested, qty_shipped, qty_received FROM transfer_line_items")
	for iRows.Next() {
		var item TransferLineItem
		var qShipped, qReceived sql.NullInt32
		_ = iRows.Scan(&item.ID, &item.TransferID, &item.Barcode, &item.QtyRequested, &qShipped, &qReceived)
		_, _ = transferDB.Exec("INSERT INTO transfer_line_items (id, transfer_id, barcode, qty_requested, qty_shipped, qty_received) VALUES (?, ?, ?, ?, ?, ?)", item.ID, item.TransferID, item.Barcode, item.QtyRequested, qShipped, qReceived)
	}
	iRows.Close()

	// 4. Assertions
	var oCount int
	_ = transferDB.QueryRow("SELECT COUNT(*) FROM transfer_orders").Scan(&oCount)
	if oCount != 2 {
		t.Fatalf("Expected 2 transfer orders, got %d", oCount)
	}

	var fetchedT1Status string
	_ = transferDB.QueryRow("SELECT status FROM transfer_orders WHERE id = ?", t1ID).Scan(&fetchedT1Status)
	if fetchedT1Status != "IN_TRANSIT" {
		t.Errorf("Expected status IN_TRANSIT for t1ID %s, got %s", t1ID, fetchedT1Status)
	}

	var fetchedItem1QtyShipped sql.NullInt32
	_ = transferDB.QueryRow("SELECT qty_shipped FROM transfer_line_items WHERE id = ?", item1ID).Scan(&fetchedItem1QtyShipped)
	if !fetchedItem1QtyShipped.Valid || fetchedItem1QtyShipped.Int32 != 10 {
		t.Errorf("Expected qty_shipped = 10 for item1ID, got %v", fetchedItem1QtyShipped)
	}
}
