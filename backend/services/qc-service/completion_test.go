package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestCompletion_TransitionsOnlyWhenAllItemsNonPending(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	handler := NewReviewHandler(repo)
	router := NewRouter(handler, "secret")

	grnID := uuid.New().String()
	storeID := uuid.New().String()
	item1ID := uuid.New().String()
	item2ID := uuid.New().String()

	// 1. Create Review with 2 line items
	createReq := CreateReviewRequest{
		GRNID:   grnID,
		StoreID: storeID,
		LineItems: []CreateReviewItemRequest{
			{GRNLineItemID: item1ID, Barcode: "8901001", QtyReceived: 10},
			{GRNLineItemID: item2ID, Barcode: "8901002", QtyReceived: 5},
		},
	}
	createBytes, _ := json.Marshal(createReq)
	reqC := httptest.NewRequest(http.MethodPost, "/v1/qc/internal/reviews", bytes.NewReader(createBytes))
	reqC.Header.Set("X-User-Type", "SYSTEM")
	recC := httptest.NewRecorder()
	router.ServeHTTP(recC, reqC)

	if recC.Code != http.StatusCreated {
		t.Fatalf("Create review failed: %d", recC.Code)
	}

	// 2. Partial Update: item1 -> PASSED, item2 remains PENDING
	updateReq1 := UpdateReviewRequest{
		LineItemUpdates: []QCLineItemUpdate{
			{GRNLineItemID: item1ID, QCStatus: QCStatusPassed},
		},
	}
	u1Bytes, _ := json.Marshal(updateReq1)
	reqU1 := httptest.NewRequest(http.MethodPut, "/v1/qc/internal/reviews/"+grnID, bytes.NewReader(u1Bytes))
	reqU1.Header.Set("X-User-Type", "SYSTEM")
	reqU1 = mux.SetURLVars(reqU1, map[string]string{"grn_id": grnID})
	recU1 := httptest.NewRecorder()
	router.ServeHTTP(recU1, reqU1)

	var rev1 QCReview
	_ = json.Unmarshal(recU1.Body.Bytes(), &rev1)
	if rev1.OverallStatus != OverallStatusPending {
		t.Errorf("Expected OverallStatus PENDING on partial completion, got %s", rev1.OverallStatus)
	}

	// Check IsComplete Endpoint
	reqComp1 := httptest.NewRequest(http.MethodGet, "/v1/qc/internal/reviews/"+grnID+"/is-complete", nil)
	reqComp1.Header.Set("X-User-Type", "SYSTEM")
	recComp1 := httptest.NewRecorder()
	router.ServeHTTP(recComp1, reqComp1)

	var compResp1 ReviewCompletionResponse
	_ = json.Unmarshal(recComp1.Body.Bytes(), &compResp1)
	if compResp1.IsComplete {
		t.Errorf("Expected IsComplete = false on partial completion, got true")
	}

	// 3. Complete Remaining Item: item2 -> REJECTED
	note := "Damaged packaging"
	updateReq2 := UpdateReviewRequest{
		LineItemUpdates: []QCLineItemUpdate{
			{GRNLineItemID: item2ID, QCStatus: QCStatusRejected, QCNote: &note},
		},
	}
	u2Bytes, _ := json.Marshal(updateReq2)
	reqU2 := httptest.NewRequest(http.MethodPut, "/v1/qc/internal/reviews/"+grnID, bytes.NewReader(u2Bytes))
	reqU2.Header.Set("X-User-Type", "SYSTEM")
	recU2 := httptest.NewRecorder()
	router.ServeHTTP(recU2, reqU2)

	var rev2 QCReview
	_ = json.Unmarshal(recU2.Body.Bytes(), &rev2)
	if rev2.OverallStatus != OverallStatusComplete {
		t.Errorf("Expected OverallStatus COMPLETE when all items non-pending, got %s", rev2.OverallStatus)
	}
	if rev2.CompletedAt == nil {
		t.Errorf("Expected CompletedAt timestamp to be set")
	}

	// Check IsComplete Endpoint again
	reqComp2 := httptest.NewRequest(http.MethodGet, "/v1/qc/internal/reviews/"+grnID+"/is-complete", nil)
	reqComp2.Header.Set("X-User-Type", "SYSTEM")
	recComp2 := httptest.NewRecorder()
	router.ServeHTTP(recComp2, reqComp2)

	var compResp2 ReviewCompletionResponse
	_ = json.Unmarshal(recComp2.Body.Bytes(), &compResp2)
	if !compResp2.IsComplete {
		t.Errorf("Expected IsComplete = true after all items set, got false")
	}
}
