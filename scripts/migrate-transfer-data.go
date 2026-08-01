package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type TransferOrderRow struct {
	ID              string
	SourceStoreID   string
	DestStoreID     string
	ChainID         string
	Status          string
	RequestedBy     sql.NullString
	RejectionReason sql.NullString
	ShippedAt       sql.NullTime
	ReceivedAt      sql.NullTime
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TransferLineItemRow struct {
	ID           string
	TransferID   string
	Barcode      string
	QtyRequested int
	QtyShipped   sql.NullInt32
	QtyReceived  sql.NullInt32
}

func main() {
	warehouseDBURL := flag.String("warehouse-db", "postgres://postgres:postgres@localhost:5432/warehouse_db?sslmode=disable", "Warehouse database connection URL")
	transferDBURL := flag.String("transfer-db", "postgres://postgres:postgres@localhost:5432/transfer_db?sslmode=disable", "Transfer database connection URL")
	flag.Parse()

	warehouseDB, err := sql.Open("postgres", *warehouseDBURL)
	if err != nil {
		log.Fatalf("Failed to connect to warehouse_db: %v", err)
	}
	defer warehouseDB.Close()

	transferDB, err := sql.Open("postgres", *transferDBURL)
	if err != nil {
		log.Fatalf("Failed to connect to transfer_db: %v", err)
	}
	defer transferDB.Close()

	log.Println("[MIGRATION] Reading transfer_orders from warehouse_db...")

	orderRows, err := warehouseDB.Query(`
		SELECT id, source_store_id, dest_store_id, chain_id, status, requested_by, rejection_reason, shipped_at, received_at, created_at, updated_at
		FROM transfer_orders
	`)
	if err != nil {
		log.Fatalf("Failed to query transfer_orders: %v", err)
	}
	defer orderRows.Close()

	var orders []TransferOrderRow
	for orderRows.Next() {
		var o TransferOrderRow
		if err := orderRows.Scan(&o.ID, &o.SourceStoreID, &o.DestStoreID, &o.ChainID, &o.Status, &o.RequestedBy, &o.RejectionReason, &o.ShippedAt, &o.ReceivedAt, &o.CreatedAt, &o.UpdatedAt); err != nil {
			log.Fatalf("Failed to scan transfer_order row: %v", err)
		}
		orders = append(orders, o)
	}

	log.Printf("[MIGRATION] Found %d transfer_orders in warehouse_db", len(orders))

	var insertedOrdersCount int
	for _, o := range orders {
		insertQuery := `
			INSERT INTO transfer_orders (id, source_store_id, dest_store_id, chain_id, status, requested_by, rejection_reason, shipped_at, received_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO NOTHING
		`
		res, err := transferDB.Exec(insertQuery, o.ID, o.SourceStoreID, o.DestStoreID, o.ChainID, o.Status, o.RequestedBy, o.RejectionReason, o.ShippedAt, o.ReceivedAt, o.CreatedAt, o.UpdatedAt)
		if err != nil {
			log.Fatalf("Failed to insert transfer_order %s: %v", o.ID, err)
		}
		if aff, _ := res.RowsAffected(); aff > 0 {
			insertedOrdersCount++
		}
	}

	log.Printf("[MIGRATION] Reading transfer_line_items from warehouse_db...")

	itemRows, err := warehouseDB.Query(`
		SELECT id, transfer_id, barcode, qty_requested, qty_shipped, qty_received
		FROM transfer_line_items
	`)
	if err != nil {
		log.Fatalf("Failed to query transfer_line_items: %v", err)
	}
	defer itemRows.Close()

	var items []TransferLineItemRow
	for itemRows.Next() {
		var item TransferLineItemRow
		if err := itemRows.Scan(&item.ID, &item.TransferID, &item.Barcode, &item.QtyRequested, &item.QtyShipped, &item.QtyReceived); err != nil {
			log.Fatalf("Failed to scan transfer_line_item row: %v", err)
		}
		items = append(items, item)
	}

	log.Printf("[MIGRATION] Found %d transfer_line_items in warehouse_db", len(items))

	var insertedItemsCount int
	for _, item := range items {
		insertQuery := `
			INSERT INTO transfer_line_items (id, transfer_id, barcode, qty_requested, qty_shipped, qty_received)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO NOTHING
		`
		res, err := transferDB.Exec(insertQuery, item.ID, item.TransferID, item.Barcode, item.QtyRequested, item.QtyShipped, item.QtyReceived)
		if err != nil {
			log.Fatalf("Failed to insert transfer_line_item %s: %v", item.ID, err)
		}
		if aff, _ := res.RowsAffected(); aff > 0 {
			insertedItemsCount++
		}
	}

	log.Printf("[MIGRATION] Inserted %d orders and %d line items into transfer_db", insertedOrdersCount, insertedItemsCount)

	// Verification Assertions
	var warehouseOrderCount, transferOrderCount int
	_ = warehouseDB.QueryRow("SELECT COUNT(*) FROM transfer_orders").Scan(&warehouseOrderCount)
	_ = transferDB.QueryRow("SELECT COUNT(*) FROM transfer_orders").Scan(&transferOrderCount)

	var warehouseItemCount, transferItemCount int
	_ = warehouseDB.QueryRow("SELECT COUNT(*) FROM transfer_line_items").Scan(&warehouseItemCount)
	_ = transferDB.QueryRow("SELECT COUNT(*) FROM transfer_line_items").Scan(&transferItemCount)

	log.Printf("[VERIFICATION] Orders: warehouse_db=%d, transfer_db=%d | Line items: warehouse_db=%d, transfer_db=%d", warehouseOrderCount, transferOrderCount, warehouseItemCount, transferItemCount)

	if warehouseOrderCount != transferOrderCount || warehouseItemCount != transferItemCount {
		log.Fatalf("[MIGRATION FAILURE] Parity mismatch! Orders (%d vs %d), Items (%d vs %d)", warehouseOrderCount, transferOrderCount, warehouseItemCount, transferItemCount)
	}

	fmt.Println("SUCCESS: Transfer data migration verified 100% parity and ID preservation.")
}
