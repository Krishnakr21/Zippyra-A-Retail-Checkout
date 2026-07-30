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
)

func setupTestDB(t *testing.T) (*sql.DB, Repository) {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open sqlite memory DB: %v", err)
	}

	_, err = db.Exec(`
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
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db, NewPostgresRepository(db)
}

func TestIdempotency_DuplicatePost_DoesNotCreateDuplicateRow(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	handler := NewReviewHandler(repo)
	router := NewRouter(handler, "secret")

	grnID := uuid.New().String()
	storeID := uuid.New().String()

	reqBody := CreateReviewRequest{
		GRNID:   grnID,
		StoreID: storeID,
		LineItems: []CreateReviewItemRequest{
			{GRNLineItemID: uuid.New().String(), Barcode: "8901001", QtyReceived: 10},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Call 1
	req1 := httptest.NewRequest(http.MethodPost, "/v1/qc/internal/reviews", bytes.NewReader(bodyBytes))
	req1.Header.Set("X-User-Type", "SYSTEM")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created on first call, got %d: %s", rec1.Code, rec1.Body.String())
	}

	// Call 2 (Duplicate retry)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/qc/internal/reviews", bytes.NewReader(bodyBytes))
	req2.Header.Set("X-User-Type", "SYSTEM")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created on duplicate call, got %d", rec2.Code)
	}

	// Assert exactly 1 row exists in DB
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM qc_reviews WHERE grn_id = ?", grnID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 review row, found %d", count)
	}
}
