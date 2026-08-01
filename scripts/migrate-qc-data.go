package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type GRNLineItemRow struct {
	ID          string         `json:"id"`
	GRNID       string         `json:"grn_id"`
	StoreID     string         `json:"store_id"`
	Barcode     string         `json:"barcode"`
	QtyReceived int            `json:"qty_received"`
	QCStatus    string         `json:"qc_status"`
	QCNote      sql.NullString `json:"qc_note"`
}

type QCItemSnapshot struct {
	GRNLineItemID string  `json:"grn_line_item_id"`
	Barcode       string  `json:"barcode"`
	QtyReceived   int     `json:"qty_received"`
	QCStatus      string  `json:"qc_status"`
	QCNote        *string `json:"qc_note,omitempty"`
}

func main() {
	warehouseDBURL := flag.String("warehouse-db", "postgres://postgres:postgres@localhost:5432/warehouse_db?sslmode=disable", "Warehouse database connection URL")
	qcDBURL := flag.String("qc-db", "postgres://postgres:postgres@localhost:5432/qc_db?sslmode=disable", "QC database connection URL")
	flag.Parse()

	warehouseDB, err := sql.Open("postgres", *warehouseDBURL)
	if err != nil {
		log.Fatalf("Failed to connect to warehouse_db: %v", err)
	}
	defer warehouseDB.Close()

	qcDB, err := sql.Open("postgres", *qcDBURL)
	if err != nil {
		log.Fatalf("Failed to connect to qc_db: %v", err)
	}
	defer qcDB.Close()

	log.Println("[MIGRATION] Reading existing GRN line items from warehouse_db...")

	query := `
		SELECT gli.id, gli.grn_id, grn.store_id, gli.barcode, gli.qty_received, COALESCE(gli.qc_status, 'PENDING') AS qc_status, gli.qc_note
		FROM grn_line_items gli
		JOIN goods_received_notes grn ON grn.id = gli.grn_id
		ORDER BY gli.grn_id, gli.id
	`

	rows, err := warehouseDB.Query(query)
	if err != nil {
		log.Fatalf("Failed to query grn_line_items: %v", err)
	}
	defer rows.Close()

	groupedItems := make(map[string][]GRNLineItemRow)
	var totalLineItems int

	for rows.Next() {
		var item GRNLineItemRow
		if err := rows.Scan(&item.ID, &item.GRNID, &item.StoreID, &item.Barcode, &item.QtyReceived, &item.QCStatus, &item.QCNote); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		groupedItems[item.GRNID] = append(groupedItems[item.GRNID], item)
		totalLineItems++
	}

	log.Printf("[MIGRATION] Found %d total line items grouped across %d distinct GRNs", totalLineItems, len(groupedItems))

	var insertedCount int
	now := time.Now()

	for grnID, items := range groupedItems {
		storeID := items[0].StoreID

		var snapshots []QCItemSnapshot
		allComplete := true

		for _, item := range items {
			var notePtr *string
			if item.QCNote.Valid && item.QCNote.String != "" {
				s := item.QCNote.String
				notePtr = &s
			}

			status := item.QCStatus
			if status == "" {
				status = "PENDING"
			}
			if status == "PENDING" {
				allComplete = false
			}

			snapshots = append(snapshots, QCItemSnapshot{
				GRNLineItemID: item.ID,
				Barcode:       item.Barcode,
				QtyReceived:   item.QtyReceived,
				QCStatus:      status,
				QCNote:        notePtr,
			})
		}

		overallStatus := "PENDING"
		var completedAt *time.Time
		if allComplete {
			overallStatus = "COMPLETE"
			completedAt = &now
		}

		snapshotsJSON, err := json.Marshal(snapshots)
		if err != nil {
			log.Fatalf("Failed to marshal snapshots for grn %s: %v", grnID, err)
		}

		insertQuery := `
			INSERT INTO qc_reviews (id, grn_id, store_id, line_items, overall_status, completed_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (grn_id) DO NOTHING
		`

		res, err := qcDB.Exec(insertQuery, uuid.New().String(), grnID, storeID, snapshotsJSON, overallStatus, completedAt, now, now)
		if err != nil {
			log.Fatalf("Failed to insert qc_review for grn %s: %v", grnID, err)
		}

		affected, _ := res.RowsAffected()
		if affected > 0 {
			insertedCount++
		}
	}

	log.Printf("[MIGRATION] Migration complete. Inserted %d new qc_reviews rows.", insertedCount)

	// Verification Parity Assertion
	var warehouseDistinctGRNCount int
	err = warehouseDB.QueryRow("SELECT COUNT(DISTINCT grn_id) FROM grn_line_items").Scan(&warehouseDistinctGRNCount)
	if err != nil {
		log.Fatalf("Failed to verify warehouse GRN count: %v", err)
	}

	var qcTotalReviewsCount int
	err = qcDB.QueryRow("SELECT COUNT(*) FROM qc_reviews").Scan(&qcTotalReviewsCount)
	if err != nil {
		log.Fatalf("Failed to verify qc_reviews total count: %v", err)
	}

	log.Printf("[VERIFICATION] warehouse_db distinct grn_id count: %d | qc_db total reviews count: %d", warehouseDistinctGRNCount, qcTotalReviewsCount)

	if warehouseDistinctGRNCount != qcTotalReviewsCount {
		log.Fatalf("[MIGRATION FAILURE] Row count mismatch! Expected %d, got %d", warehouseDistinctGRNCount, qcTotalReviewsCount)
	}

	fmt.Println("SUCCESS: QC data migration verified 100% parity.")
}
