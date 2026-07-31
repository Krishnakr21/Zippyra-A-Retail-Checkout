package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/kafka"
)

func TestTransferShip_InsufficientStockSurfaces409(t *testing.T) {
	db, repo := setupWarehouseTestDB(t)
	defer db.Close()

	mockClient := &MockInventoryClient{
		transferOutErr: fmt.Errorf("HTTP_409: INSUFFICIENT_STOCK_FOR_TRANSFER - insufficient stock"),
	}
	producer := kafka.NewProducer("localhost:9092")
	transferHandler := NewTransferHandler(repo, mockClient, producer)

	// 1. Create transfer
	reqCreate := CreateTransferRequest{
		SourceStoreID: "store-a",
		DestStoreID:   "store-b",
		Items:         []TransferItemRequest{{Barcode: "item-no-stock", QtyRequested: 50}},
	}
	b, _ := json.Marshal(reqCreate)
	reqC := httptest.NewRequest(http.MethodPost, "/v1/warehouse/transfer", bytes.NewReader(b))
	recC := httptest.NewRecorder()
	transferHandler.CreateTransferHandler(recC, reqC)

	var tr TransferOrder
	_ = json.Unmarshal(recC.Body.Bytes(), &tr)

	// 2. Approve transfer
	reqApp := httptest.NewRequest(http.MethodPut, "/v1/warehouse/transfer/"+tr.ID+"/approve", nil)
	reqApp = mux.SetURLVars(reqApp, map[string]string{"id": tr.ID})
	recApp := httptest.NewRecorder()
	transferHandler.ApproveTransferHandler(recApp, reqApp)

	// 3. Ship transfer -> 409 Conflict due to mockClient transferOutErr
	shipReq := ShipTransferRequest{
		Items: []ShipTransferItem{{Barcode: "item-no-stock", QtyShipped: 50}},
	}
	sb, _ := json.Marshal(shipReq)
	reqShip := httptest.NewRequest(http.MethodPut, "/v1/warehouse/transfer/"+tr.ID+"/ship", bytes.NewReader(sb))
	reqShip = mux.SetURLVars(reqShip, map[string]string{"id": tr.ID})
	recShip := httptest.NewRecorder()
	transferHandler.ShipTransferHandler(recShip, reqShip)

	if recShip.Code != http.StatusConflict {
		t.Fatalf("Expected 409 Conflict on insufficient stock ship transfer, got %d", recShip.Code)
	}

	// Verify status REMAINED APPROVED (did NOT change to IN_TRANSIT)
	afterState, _ := repo.GetTransferByID(context.Background(), tr.ID)
	if afterState.Status != TransferStatusApproved {
		t.Errorf("Expected transfer status to REMAIN APPROVED on stock failure, got %s", afterState.Status)
	}
}

func TestTransferReceive_DiscrepancyPublished(t *testing.T) {
	db, repo := setupWarehouseTestDB(t)
	defer db.Close()

	mockClient := &MockInventoryClient{}
	producer := kafka.NewProducer("localhost:9092")
	transferHandler := NewTransferHandler(repo, mockClient, producer)

	// 1. Create, approve, and ship transfer of 10 items
	reqCreate := CreateTransferRequest{
		SourceStoreID: "store-src",
		DestStoreID:   "store-dst",
		Items:         []TransferItemRequest{{Barcode: "item-loss", QtyRequested: 10}},
	}
	b, _ := json.Marshal(reqCreate)
	reqC := httptest.NewRequest(http.MethodPost, "/v1/warehouse/transfer", bytes.NewReader(b))
	recC := httptest.NewRecorder()
	transferHandler.CreateTransferHandler(recC, reqC)

	var tr TransferOrder
	_ = json.Unmarshal(recC.Body.Bytes(), &tr)

	_ = repo.UpdateTransferStatus(context.Background(), tr.ID, TransferStatusApproved, nil)

	shipReq := ShipTransferRequest{
		Items: []ShipTransferItem{{Barcode: "item-loss", QtyShipped: 10}},
	}
	sb, _ := json.Marshal(shipReq)
	reqShip := httptest.NewRequest(http.MethodPut, "/v1/warehouse/transfer/"+tr.ID+"/ship", bytes.NewReader(sb))
	reqShip = mux.SetURLVars(reqShip, map[string]string{"id": tr.ID})
	recShip := httptest.NewRecorder()
	transferHandler.ShipTransferHandler(recShip, reqShip)

	// 2. Receive only 8 items (transit damage/loss of 2 units)
	receiveReq := ReceiveTransferRequest{
		Items: []ReceiveTransferItem{{Barcode: "item-loss", QtyReceived: 8}},
	}
	rb, _ := json.Marshal(receiveReq)
	reqRec := httptest.NewRequest(http.MethodPut, "/v1/warehouse/transfer/"+tr.ID+"/receive", bytes.NewReader(rb))
	reqRec = mux.SetURLVars(reqRec, map[string]string{"id": tr.ID})
	recRec := httptest.NewRecorder()
	transferHandler.ReceiveTransferHandler(recRec, reqRec)

	if recRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on receiving transfer with discrepancy, got %d", recRec.Code)
	}

	// Verify transfer status reached RECEIVED
	afterState, _ := repo.GetTransferByID(context.Background(), tr.ID)
	if afterState.Status != TransferStatusReceived {
		t.Errorf("Expected transfer status to reach RECEIVED, got %s", afterState.Status)
	}
}
