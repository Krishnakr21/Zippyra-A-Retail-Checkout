package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func setupWarehouseTestDB(t *testing.T) (*sql.DB, *PostgresRepository) {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open sqlite memory DB: %v", err)
	}

	schema := `
		CREATE TABLE purchase_orders (
			id TEXT PRIMARY KEY,
			store_id TEXT NOT NULL,
			chain_id TEXT NOT NULL,
			vendor_name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'DRAFT',
			source TEXT NOT NULL DEFAULT 'MANUAL',
			created_by TEXT,
			auto_reorder_item_barcode TEXT,
			auto_reorder_date TEXT,
			expected_delivery_date TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			submitted_at TIMESTAMP,
			completed_at TIMESTAMP,
			CONSTRAINT uq_po_auto_reorder UNIQUE (store_id, auto_reorder_item_barcode, auto_reorder_date)
		);

		CREATE TABLE po_line_items (
			id TEXT PRIMARY KEY,
			po_id TEXT NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
			barcode TEXT NOT NULL,
			qty_ordered INTEGER NOT NULL,
			unit_cost_paise INTEGER NOT NULL,
			qty_received INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE goods_received_notes (
			id TEXT PRIMARY KEY,
			po_id TEXT REFERENCES purchase_orders(id) ON DELETE SET NULL,
			store_id TEXT NOT NULL,
			received_by TEXT NOT NULL,
			vendor_invoice_ref TEXT,
			status TEXT NOT NULL DEFAULT 'DRAFT',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP
		);

		CREATE TABLE grn_line_items (
			id TEXT PRIMARY KEY,
			grn_id TEXT NOT NULL REFERENCES goods_received_notes(id) ON DELETE CASCADE,
			barcode TEXT NOT NULL,
			qty_expected INTEGER,
			qty_received INTEGER NOT NULL,
			unit_cost_paise INTEGER NOT NULL,
			qc_status TEXT NOT NULL DEFAULT 'PENDING',
			qc_note TEXT
		);

		CREATE TABLE transfer_orders (
			id TEXT PRIMARY KEY,
			source_store_id TEXT NOT NULL,
			dest_store_id TEXT NOT NULL,
			chain_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'REQUESTED',
			requested_by TEXT NOT NULL,
			rejection_reason TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			approved_at TIMESTAMP,
			shipped_at TIMESTAMP,
			received_at TIMESTAMP
		);

		CREATE TABLE transfer_line_items (
			id TEXT PRIMARY KEY,
			transfer_id TEXT NOT NULL REFERENCES transfer_orders(id) ON DELETE CASCADE,
			barcode TEXT NOT NULL,
			qty_requested INTEGER NOT NULL,
			qty_shipped INTEGER,
			qty_received INTEGER
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create warehouse test schema: %v", err)
	}

	repo := NewPostgresRepository(db)
	return db, repo
}

func TestPOCreationAndSubmission(t *testing.T) {
	db, repo := setupWarehouseTestDB(t)
	defer db.Close()

	poHandler := NewPOHandler(repo)
	storeID := "store-po-1"

	// 1. Create PO
	reqBody := CreatePORequest{
		StoreID:    storeID,
		VendorName: "Acme Supplies",
		Items: []POLineItemRequest{
			{Barcode: "890111", QtyOrdered: 100, UnitCostPaise: 5000},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/warehouse/po", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	poHandler.CreatePOHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create PO failed: %d - %s", rec.Code, rec.Body.String())
	}

	var createdPO PurchaseOrder
	_ = json.Unmarshal(rec.Body.Bytes(), &createdPO)
	if createdPO.Status != POStatusDraft {
		t.Errorf("Expected DRAFT status on creation, got %s", createdPO.Status)
	}

	// 2. Submit PO
	reqSub := httptest.NewRequest(http.MethodPut, "/v1/warehouse/po/"+createdPO.ID+"/submit", nil)
	reqSub = mux.SetURLVars(reqSub, map[string]string{"id": createdPO.ID})
	recSub := httptest.NewRecorder()
	poHandler.SubmitPOHandler(recSub, reqSub)

	if recSub.Code != http.StatusOK {
		t.Fatalf("Submit PO failed: %d - %s", recSub.Code, recSub.Body.String())
	}

	// 3. Re-submit PO -> 409 PO_ALREADY_SUBMITTED
	reqSub2 := httptest.NewRequest(http.MethodPut, "/v1/warehouse/po/"+createdPO.ID+"/submit", nil)
	reqSub2 = mux.SetURLVars(reqSub2, map[string]string{"id": createdPO.ID})
	recSub2 := httptest.NewRecorder()
	poHandler.SubmitPOHandler(recSub2, reqSub2)

	if recSub2.Code != http.StatusConflict {
		t.Fatalf("Expected 409 Conflict on duplicate submit, got %d", recSub2.Code)
	}
}
