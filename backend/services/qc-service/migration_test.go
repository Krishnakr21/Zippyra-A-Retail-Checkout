package main

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func TestMigration_PreservesStatusNotesAndParity(t *testing.T) {
	warehouseDB, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open warehouse DB: %v", err)
	}
	defer warehouseDB.Close()

	qcDB, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open QC DB: %v", err)
	}
	defer qcDB.Close()

	// 1. Setup Schemas
	_, _ = warehouseDB.Exec(`
		CREATE TABLE goods_received_notes (
			id TEXT PRIMARY KEY,
			store_id TEXT NOT NULL
		);
		CREATE TABLE grn_line_items (
			id TEXT PRIMARY KEY,
			grn_id TEXT NOT NULL,
			barcode TEXT NOT NULL,
			qty_received INTEGER NOT NULL,
			qc_status TEXT DEFAULT 'PENDING',
			qc_note TEXT NULL
		);
	`)

	_, _ = qcDB.Exec(`
		CREATE TABLE qc_reviews (
			id TEXT PRIMARY KEY,
			grn_id TEXT NOT NULL UNIQUE,
			store_id TEXT NOT NULL,
			line_items TEXT NOT NULL,
			overall_status TEXT NOT NULL DEFAULT 'PENDING',
			reviewed_by TEXT NULL,
			completed_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)

	// 2. Seed 3 GRNs in warehouse DB
	storeID := "store-mig-1"
	grn1 := uuid.New().String()
	grn2 := uuid.New().String()
	grn3 := uuid.New().String()

	_, _ = warehouseDB.Exec("INSERT INTO goods_received_notes (id, store_id) VALUES (?, ?), (?, ?), (?, ?)", grn1, storeID, grn2, storeID, grn3, storeID)

	// GRN 1: 2 items (1 PASSED, 1 REJECTED with note) -> COMPLETE
	item1_1 := uuid.New().String()
	item1_2 := uuid.New().String()
	_, _ = warehouseDB.Exec("INSERT INTO grn_line_items (id, grn_id, barcode, qty_received, qc_status, qc_note) VALUES (?, ?, ?, ?, ?, ?)", item1_1, grn1, "b1", 10, "PASSED", nil)
	_, _ = warehouseDB.Exec("INSERT INTO grn_line_items (id, grn_id, barcode, qty_received, qc_status, qc_note) VALUES (?, ?, ?, ?, ?, ?)", item1_2, grn1, "b2", 5, "REJECTED", "Box torn")

	// GRN 2: 2 items (1 PASSED, 1 PENDING) -> PENDING
	item2_1 := uuid.New().String()
	item2_2 := uuid.New().String()
	_, _ = warehouseDB.Exec("INSERT INTO grn_line_items (id, grn_id, barcode, qty_received, qc_status, qc_note) VALUES (?, ?, ?, ?, ?, ?)", item2_1, grn2, "b3", 20, "PASSED", nil)
	_, _ = warehouseDB.Exec("INSERT INTO grn_line_items (id, grn_id, barcode, qty_received, qc_status, qc_note) VALUES (?, ?, ?, ?, ?, ?)", item2_2, grn2, "b4", 15, "PENDING", nil)

	// GRN 3: 1 item (PASSED) -> COMPLETE
	item3_1 := uuid.New().String()
	_, _ = warehouseDB.Exec("INSERT INTO grn_line_items (id, grn_id, barcode, qty_received, qc_status, qc_note) VALUES (?, ?, ?, ?, ?, ?)", item3_1, grn3, "b5", 8, "PASSED", "All good")

	// 3. Perform Migration Logic
	query := `
		SELECT gli.id, gli.grn_id, grn.store_id, gli.barcode, gli.qty_received, COALESCE(gli.qc_status, 'PENDING') AS qc_status, gli.qc_note
		FROM grn_line_items gli
		JOIN goods_received_notes grn ON grn.id = gli.grn_id
		ORDER BY gli.grn_id, gli.id
	`
	rows, err := warehouseDB.Query(query)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	type LineItemRow struct {
		ID          string
		GRNID       string
		StoreID     string
		Barcode     string
		QtyReceived int
		QCStatus    string
		QCNote      sql.NullString
	}

	groupedItems := make(map[string][]LineItemRow)
	for rows.Next() {
		var item LineItemRow
		_ = rows.Scan(&item.ID, &item.GRNID, &item.StoreID, &item.Barcode, &item.QtyReceived, &item.QCStatus, &item.QCNote)
		groupedItems[item.GRNID] = append(groupedItems[item.GRNID], item)
	}

	now := time.Now()
	for gID, items := range groupedItems {
		stID := items[0].StoreID
		var snapshots []QCLineItemSnapshot
		allComplete := true

		for _, item := range items {
			var notePtr *string
			if item.QCNote.Valid && item.QCNote.String != "" {
				s := item.QCNote.String
				notePtr = &s
			}
			if item.QCStatus == "PENDING" {
				allComplete = false
			}
			snapshots = append(snapshots, QCLineItemSnapshot{
				GRNLineItemID: item.ID,
				Barcode:       item.Barcode,
				QtyReceived:   item.QtyReceived,
				QCStatus:      item.QCStatus,
				QCNote:        notePtr,
			})
		}

		overallStatus := "PENDING"
		var completedAt *time.Time
		if allComplete {
			overallStatus = "COMPLETE"
			completedAt = &now
		}

		snapshotsJSON, _ := json.Marshal(snapshots)
		insertQuery := `
			INSERT INTO qc_reviews (id, grn_id, store_id, line_items, overall_status, completed_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (grn_id) DO NOTHING
		`
		_, err := qcDB.Exec(insertQuery, uuid.New().String(), gID, stID, string(snapshotsJSON), overallStatus, completedAt, now, now)
		if err != nil {
			t.Fatalf("Insert failed for grn %s: %v", gID, err)
		}
	}

	// 4. Assertions
	var warehouseDistinctGRNs int
	_ = warehouseDB.QueryRow("SELECT COUNT(DISTINCT grn_id) FROM grn_line_items").Scan(&warehouseDistinctGRNs)

	var qcTotalReviews int
	_ = qcDB.QueryRow("SELECT COUNT(*) FROM qc_reviews").Scan(&qcTotalReviews)

	if warehouseDistinctGRNs != 3 || qcTotalReviews != 3 {
		t.Fatalf("Parity mismatch: expected 3, got warehouse=%d, qc=%d", warehouseDistinctGRNs, qcTotalReviews)
	}

	// Assert GRN 1 overall status COMPLETE and note preserved
	var rev1Overall string
	var lineItemsStr string
	_ = qcDB.QueryRow("SELECT overall_status, line_items FROM qc_reviews WHERE grn_id = ?", grn1).Scan(&rev1Overall, &lineItemsStr)

	if rev1Overall != "COMPLETE" {
		t.Errorf("Expected GRN 1 overall status COMPLETE, got %s", rev1Overall)
	}

	var snapshots1 []QCLineItemSnapshot
	_ = json.Unmarshal([]byte(lineItemsStr), &snapshots1)
	if len(snapshots1) != 2 {
		t.Fatalf("Expected 2 snapshots in GRN 1, got %d", len(snapshots1))
	}
	var foundNote *string
	for _, s := range snapshots1 {
		if s.Barcode == "b2" {
			foundNote = s.QCNote
		}
	}
	if foundNote == nil || *foundNote != "Box torn" {
		t.Errorf("Expected QC note 'Box torn' preserved for barcode b2, got %v", foundNote)
	}

	// Assert GRN 2 overall status PENDING
	var rev2Overall string
	_ = qcDB.QueryRow("SELECT overall_status FROM qc_reviews WHERE grn_id = ?", grn2).Scan(&rev2Overall)
	if rev2Overall != "PENDING" {
		t.Errorf("Expected GRN 2 overall status PENDING, got %s", rev2Overall)
	}
}
