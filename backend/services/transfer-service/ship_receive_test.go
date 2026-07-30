package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type MockInventoryClient struct {
	shipErr    error
	receiveErr error
}

func (m *MockInventoryClient) ApplyTransferOut(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error {
	return m.shipErr
}

func (m *MockInventoryClient) ApplyTransferIn(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error {
	return m.receiveErr
}

func TestTransferShip_InsufficientStockSurfaces409(t *testing.T) {
	db, repo := setupTransferTestDB(t)
	defer db.Close()

	mockInv := &MockInventoryClient{
		shipErr: fmt.Errorf("HTTP 409: INSUFFICIENT_STOCK_FOR_TRANSFER"),
	}

	handler := NewTransferHandler(repo, mockInv, nil)
	router := NewRouter(handler, "secret")

	// 1. Create Transfer
	sourceID := uuid.New().String()
	destID := uuid.New().String()

	createReq := CreateTransferRequest{
		SourceStoreID: sourceID,
		DestStoreID:   destID,
		ChainID:       "chain-1",
		Items:         []TransferItemRequest{{Barcode: "item-out-of-stock", QtyRequested: 100}},
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

	// 3. Attempt Ship -> Returns 409 INSUFFICIENT_STOCK_FOR_TRANSFER
	shipReq := ShipTransferRequest{
		Items: []ShipTransferItemRequest{{Barcode: "item-out-of-stock", QtyShipped: 100}},
	}
	shipBytes, _ := json.Marshal(shipReq)
	reqShip := httptest.NewRequest(http.MethodPut, "/v1/transfer/internal/transfers/"+transfer.ID+"/ship", bytes.NewReader(shipBytes))
	reqShip.Header.Set("X-User-Type", "SYSTEM")
	reqShip = mux.SetURLVars(reqShip, map[string]string{"id": transfer.ID})
	recShip := httptest.NewRecorder()
	router.ServeHTTP(recShip, reqShip)

	if recShip.Code != http.StatusConflict {
		t.Fatalf("Expected 409 Conflict on insufficient stock, got %d: %s", recShip.Code, recShip.Body.String())
	}

	// Status remains APPROVED
	tState, _ := repo.GetTransferByID(context.Background(), transfer.ID)
	if tState.Status != TransferStatusApproved {
		t.Errorf("Expected transfer status to remain APPROVED on stock error, got %s", tState.Status)
	}
}
