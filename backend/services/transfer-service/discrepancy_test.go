package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/kafka"
)

func TestDiscrepancy_PartialReceivePublishesEvent(t *testing.T) {
	db, repo := setupTransferTestDB(t)
	defer db.Close()

	producer := kafka.NewProducer("localhost:9092")
	mockInv := &MockInventoryClient{}

	handler := NewTransferHandler(repo, mockInv, producer)
	router := NewRouter(handler, "secret")

	// 1. Create Transfer
	sourceID := uuid.New().String()
	destID := uuid.New().String()

	createReq := CreateTransferRequest{
		SourceStoreID: sourceID,
		DestStoreID:   destID,
		ChainID:       "chain-1",
		Items:         []TransferItemRequest{{Barcode: "item-loss", QtyRequested: 10}},
	}
	bodyBytes, _ := json.Marshal(createReq)
	reqC := httptest.NewRequest(http.MethodPost, "/v1/transfer/internal/transfers", bytes.NewReader(bodyBytes))
	reqC.Header.Set("X-User-Type", "SYSTEM")
	recC := httptest.NewRecorder()
	router.ServeHTTP(recC, reqC)

	var transfer TransferOrder
	_ = json.Unmarshal(recC.Body.Bytes(), &transfer)

	// 2. Approve Transfer
	reqApp := httptest.NewRequest(http.MethodPut, "/v1/transfer/internal/transfers/"+transfer.ID+"/approve", nil)
	reqApp.Header.Set("X-User-Type", "SYSTEM")
	reqApp = mux.SetURLVars(reqApp, map[string]string{"id": transfer.ID})
	recApp := httptest.NewRecorder()
	router.ServeHTTP(recApp, reqApp)

	// 3. Ship Transfer (10 shipped)
	shipReq := ShipTransferRequest{
		Items: []ShipTransferItemRequest{{Barcode: "item-loss", QtyShipped: 10}},
	}
	shipBytes, _ := json.Marshal(shipReq)
	reqShip := httptest.NewRequest(http.MethodPut, "/v1/transfer/internal/transfers/"+transfer.ID+"/ship", bytes.NewReader(shipBytes))
	reqShip.Header.Set("X-User-Type", "SYSTEM")
	reqShip = mux.SetURLVars(reqShip, map[string]string{"id": transfer.ID})
	recShip := httptest.NewRecorder()
	router.ServeHTTP(recShip, reqShip)

	if recShip.Code != http.StatusOK {
		t.Fatalf("Ship failed: %d", recShip.Code)
	}

	// 4. Partial Receive (only 8 received < 10 shipped)
	recReq := ReceiveTransferRequest{
		Items: []ReceiveTransferItemRequest{{Barcode: "item-loss", QtyReceived: 8}},
	}
	recBytes, _ := json.Marshal(recReq)
	reqRec := httptest.NewRequest(http.MethodPut, "/v1/transfer/internal/transfers/"+transfer.ID+"/receive", bytes.NewReader(recBytes))
	reqRec.Header.Set("X-User-Type", "SYSTEM")
	reqRec = mux.SetURLVars(reqRec, map[string]string{"id": transfer.ID})
	recRec := httptest.NewRecorder()
	router.ServeHTTP(recRec, reqRec)

	if recRec.Code != http.StatusOK {
		t.Fatalf("Receive failed: %d", recRec.Code)
	}

	// Status updated to RECEIVED — read state before DB is closed
	tState, _ := repo.GetTransferByID(t.Context(), transfer.ID)
	if tState.Status != TransferStatusReceived {
		t.Errorf("Expected transfer status RECEIVED, got %s", tState.Status)
	}
	if tState.LineItems[0].QtyReceived == nil || *tState.LineItems[0].QtyReceived != 8 {
		t.Errorf("Expected qty_received = 8, got %v", tState.LineItems[0].QtyReceived)
	}
}
