package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zippyra/backend/shared/jwt"
)

func setupTransferTestDB(t *testing.T) (*sql.DB, Repository) {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open sqlite memory DB: %v", err)
	}

	_, err = db.Exec(`
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
	`)
	if err != nil {
		t.Fatalf("Failed to create transfer tables: %v", err)
	}

	return db, NewPostgresRepository(db)
}

func TestCrossChainTransfer_DeniedWhenCallerStoreIDMismatch(t *testing.T) {
	db, repo := setupTransferTestDB(t)
	defer db.Close()

	handler := NewTransferHandler(repo, nil, nil)
	router := NewRouter(handler, "secret")

	sourceID := uuid.New().String()
	destID := uuid.New().String()
	unauthorizedCallerStoreID := uuid.New().String()

	reqBody := CreateTransferRequest{
		SourceStoreID: sourceID,
		DestStoreID:   destID,
		Items: []TransferItemRequest{
			{Barcode: "item-101", QtyRequested: 5},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Caller is bound to unauthorizedCallerStoreID which matches neither source nor dest store
	claims := &jwt.Claims{
		UserID:   "user-unauth-1",
		UserType: "STORE_STAFF",
		StoreID:  unauthorizedCallerStoreID,
		ChainID:  "chain-1",
	}
	token, _ := jwt.GenerateToken(claims, "secret", 1)

	req := httptest.NewRequest(http.MethodPost, "/v1/transfer/internal/transfers", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for cross-chain store mismatch, got %d: %s", rec.Code, rec.Body.String())
	}
}
