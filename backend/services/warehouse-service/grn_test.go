package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/kafka"
)

type MockInventoryClient struct {
	applyGRNCallCount int
	lastGRNItems      []GRNItemPayload
	transferOutErr    error
	transferInErr     error
}

func (m *MockInventoryClient) ApplyGRN(ctx context.Context, storeID, grnID string, items []GRNItemPayload) (*ApplyGRNClientResponse, error) {
	m.applyGRNCallCount++
	m.lastGRNItems = items
	return &ApplyGRNClientResponse{
		Applied:        true,
		ItemsRequested: len(items),
		ItemsApplied:   len(items),
	}, nil
}

func (m *MockInventoryClient) ApplyTransferOut(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error {
	return m.transferOutErr
}

func (m *MockInventoryClient) ApplyTransferIn(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error {
	return m.transferInErr
}

type MockQCClient struct {
	reviews map[string]*QCReviewResponse
}

func NewMockQCClient() *MockQCClient {
	return &MockQCClient{reviews: make(map[string]*QCReviewResponse)}
}

func (m *MockQCClient) CreateReview(ctx context.Context, grnID, storeID string, items []QCLineItemCreatePayload) (*QCReviewResponse, error) {
	var snapshots []QCLineItemResponse
	for _, item := range items {
		snapshots = append(snapshots, QCLineItemResponse{
			GRNLineItemID: item.GRNLineItemID,
			Barcode:       item.Barcode,
			QtyReceived:   item.QtyReceived,
			QCStatus:      item.QCStatus,
		})
	}
	res := &QCReviewResponse{
		ID:            "rev-" + grnID,
		GRNID:         grnID,
		StoreID:       storeID,
		LineItems:     snapshots,
		OverallStatus: "PENDING",
	}
	m.reviews[grnID] = res
	return res, nil
}

func (m *MockQCClient) GetReview(ctx context.Context, grnID string) (*QCReviewResponse, error) {
	return m.reviews[grnID], nil
}

func (m *MockQCClient) UpdateReview(ctx context.Context, grnID string, updates []QCLineItemUpdatePayload) (*QCReviewResponse, error) {
	rev := m.reviews[grnID]
	if rev == nil {
		return nil, nil
	}
	allComplete := true
	for i := range rev.LineItems {
		item := &rev.LineItems[i]
		for _, u := range updates {
			if u.GRNLineItemID == item.GRNLineItemID {
				item.QCStatus = u.QCStatus
				item.QCNote = u.QCNote
			}
		}
		if item.QCStatus == "PENDING" {
			allComplete = false
		}
	}
	if allComplete {
		rev.OverallStatus = "COMPLETE"
	}
	return rev, nil
}

func (m *MockQCClient) IsReviewComplete(ctx context.Context, grnID string) (bool, error) {
	rev := m.reviews[grnID]
	if rev == nil {
		return false, nil
	}
	return rev.OverallStatus == "COMPLETE", nil
}

func TestGRNCompletion_IdempotencyAndQC(t *testing.T) {
	db, repo := setupWarehouseTestDB(t)
	defer db.Close()

	mockClient := &MockInventoryClient{}
	mockQCClient := NewMockQCClient()
	producer := kafka.NewProducer("localhost:9092")
	grnHandler := NewGRNHandler(repo, mockClient, mockQCClient, producer)

	storeID := "store-grn-1"

	// 1. Create a GRN
	createReq := CreateGRNRequest{
		StoreID: storeID,
		Items: []GRNItemRequest{
			{Barcode: "item-passed", QtyReceived: 10, UnitCostPaise: 100},
			{Barcode: "item-pending", QtyReceived: 5, UnitCostPaise: 100},
		},
	}
	bodyBytes, _ := json.Marshal(createReq)
	reqC := httptest.NewRequest(http.MethodPost, "/v1/warehouse/grn", bytes.NewReader(bodyBytes))
	recC := httptest.NewRecorder()
	grnHandler.CreateGRNHandler(recC, reqC)

	if recC.Code != http.StatusCreated {
		t.Fatalf("Create GRN failed: %d", recC.Code)
	}

	var grn GoodsReceivedNote
	_ = json.Unmarshal(recC.Body.Bytes(), &grn)

	// 2. Attempt Complete with PENDING QC items -> 409 QC_INCOMPLETE
	reqComp1 := httptest.NewRequest(http.MethodPost, "/v1/warehouse/grn/"+grn.ID+"/complete", nil)
	reqComp1 = mux.SetURLVars(reqComp1, map[string]string{"id": grn.ID})
	recComp1 := httptest.NewRecorder()
	grnHandler.CompleteGRNHandler(recComp1, reqComp1)

	if recComp1.Code != http.StatusConflict {
		t.Fatalf("Expected 409 Conflict on incomplete QC, got %d", recComp1.Code)
	}
	if mockClient.applyGRNCallCount != 0 {
		t.Fatalf("inventory-service apply-grn should NOT be called when QC incomplete")
	}

	// 3. Perform QC: item-passed -> PASSED, item-pending -> REJECTED
	qcReq := QCUpdateRequest{
		LineItemUpdates: []QCUpdateItem{
			{GRNLineItemID: grn.LineItems[0].ID, QCStatus: QCStatusPassed},
			{GRNLineItemID: grn.LineItems[1].ID, QCStatus: QCStatusRejected},
		},
	}
	qcBytes, _ := json.Marshal(qcReq)
	reqQC := httptest.NewRequest(http.MethodPut, "/v1/warehouse/grn/"+grn.ID+"/qc", bytes.NewReader(qcBytes))
	reqQC = mux.SetURLVars(reqQC, map[string]string{"id": grn.ID})
	recQC := httptest.NewRecorder()
	grnHandler.UpdateQCHandler(recQC, reqQC)

	if recQC.Code != http.StatusOK {
		t.Fatalf("Update QC failed: %d", recQC.Code)
	}

	// 4. Complete GRN -> Success! Only PASSED items sent to inventory-service
	reqComp2 := httptest.NewRequest(http.MethodPost, "/v1/warehouse/grn/"+grn.ID+"/complete", nil)
	reqComp2 = mux.SetURLVars(reqComp2, map[string]string{"id": grn.ID})
	recComp2 := httptest.NewRecorder()
	grnHandler.CompleteGRNHandler(recComp2, reqComp2)

	if recComp2.Code != http.StatusOK {
		t.Fatalf("Complete GRN failed: %d", recComp2.Code)
	}
	if mockClient.applyGRNCallCount != 1 {
		t.Fatalf("Expected 1 applyGRN call, got %d", mockClient.applyGRNCallCount)
	}
	if len(mockClient.lastGRNItems) != 1 || mockClient.lastGRNItems[0].Barcode != "item-passed" {
		t.Errorf("Expected ONLY passed items sent to inventory-service, got %v", mockClient.lastGRNItems)
	}

	// 5. Repeat Complete GRN (IDEMPOTENT) -> 200 OK, applyGRN NOT called a 2nd time!
	reqComp3 := httptest.NewRequest(http.MethodPost, "/v1/warehouse/grn/"+grn.ID+"/complete", nil)
	reqComp3 = mux.SetURLVars(reqComp3, map[string]string{"id": grn.ID})
	recComp3 := httptest.NewRecorder()
	grnHandler.CompleteGRNHandler(recComp3, reqComp3)

	if recComp3.Code != http.StatusOK {
		t.Fatalf("Duplicate Complete GRN failed: %d", recComp3.Code)
	}
	if mockClient.applyGRNCallCount != 1 {
		t.Fatalf("applyGRN should NOT be called a 2nd time on duplicate completion (call count: %d)", mockClient.applyGRNCallCount)
	}
}

func TestPOStatusTransition_PartiallyToFullyReceived(t *testing.T) {
	db, repo := setupWarehouseTestDB(t)
	defer db.Close()

	mockClient := &MockInventoryClient{}
	mockQCClient := NewMockQCClient()
	producer := kafka.NewProducer("localhost:9092")
	poHandler := NewPOHandler(repo)
	grnHandler := NewGRNHandler(repo, mockClient, mockQCClient, producer)

	storeID := "store-po-trans"

	// 1. Create & Submit PO of 20 units of item-x
	createPOReq := CreatePORequest{
		StoreID:    storeID,
		VendorName: "Vendor X",
		Items: []POLineItemRequest{
			{Barcode: "item-x", QtyOrdered: 20, UnitCostPaise: 100},
		},
	}
	poBytes, _ := json.Marshal(createPOReq)
	reqPO := httptest.NewRequest(http.MethodPost, "/v1/warehouse/po", bytes.NewReader(poBytes))
	recPO := httptest.NewRecorder()
	poHandler.CreatePOHandler(recPO, reqPO)

	var po PurchaseOrder
	_ = json.Unmarshal(recPO.Body.Bytes(), &po)

	// Submit PO
	reqSub := httptest.NewRequest(http.MethodPut, "/v1/warehouse/po/"+po.ID+"/submit", nil)
	reqSub = mux.SetURLVars(reqSub, map[string]string{"id": po.ID})
	recSub := httptest.NewRecorder()
	poHandler.SubmitPOHandler(recSub, reqSub)

	// 2. GRN #1 for 10 units -> PO status PARTIALLY_RECEIVED
	grnReq1 := CreateGRNRequest{
		StoreID: storeID,
		POID:    &po.ID,
		Items:   []GRNItemRequest{{Barcode: "item-x", QtyReceived: 10, UnitCostPaise: 100}},
	}
	b1, _ := json.Marshal(grnReq1)
	reqG1 := httptest.NewRequest(http.MethodPost, "/v1/warehouse/grn", bytes.NewReader(b1))
	recG1 := httptest.NewRecorder()
	grnHandler.CreateGRNHandler(recG1, reqG1)

	var grn1 GoodsReceivedNote
	_ = json.Unmarshal(recG1.Body.Bytes(), &grn1)

	// Set QC PASSED & Complete GRN #1
	qc1Body, _ := json.Marshal(QCUpdateRequest{LineItemUpdates: []QCUpdateItem{{GRNLineItemID: grn1.LineItems[0].ID, QCStatus: QCStatusPassed}}})
	reqQC1 := httptest.NewRequest(http.MethodPut, "/v1/warehouse/grn/"+grn1.ID+"/qc", bytes.NewReader(qc1Body))
	reqQC1 = mux.SetURLVars(reqQC1, map[string]string{"id": grn1.ID})
	recQC1 := httptest.NewRecorder()
	grnHandler.UpdateQCHandler(recQC1, reqQC1)

	reqComp1 := httptest.NewRequest(http.MethodPost, "/v1/warehouse/grn/"+grn1.ID+"/complete", nil)
	reqComp1 = mux.SetURLVars(reqComp1, map[string]string{"id": grn1.ID})
	recComp1 := httptest.NewRecorder()
	grnHandler.CompleteGRNHandler(recComp1, reqComp1)

	poState1, _ := repo.GetPOByID(context.Background(), po.ID)
	if poState1.Status != POStatusPartiallyReceived {
		t.Errorf("Expected PO status PARTIALLY_RECEIVED after 1st GRN, got %s", poState1.Status)
	}

	// 3. GRN #2 for remaining 10 units -> PO status RECEIVED
	grnReq2 := CreateGRNRequest{
		StoreID: storeID,
		POID:    &po.ID,
		Items:   []GRNItemRequest{{Barcode: "item-x", QtyReceived: 10, UnitCostPaise: 100}},
	}
	b2, _ := json.Marshal(grnReq2)
	reqG2 := httptest.NewRequest(http.MethodPost, "/v1/warehouse/grn", bytes.NewReader(b2))
	recG2 := httptest.NewRecorder()
	grnHandler.CreateGRNHandler(recG2, reqG2)

	var grn2 GoodsReceivedNote
	_ = json.Unmarshal(recG2.Body.Bytes(), &grn2)

	// Set QC PASSED & Complete GRN #2
	qc2Body, _ := json.Marshal(QCUpdateRequest{LineItemUpdates: []QCUpdateItem{{GRNLineItemID: grn2.LineItems[0].ID, QCStatus: QCStatusPassed}}})
	reqQC2 := httptest.NewRequest(http.MethodPut, "/v1/warehouse/grn/"+grn2.ID+"/qc", bytes.NewReader(qc2Body))
	reqQC2 = mux.SetURLVars(reqQC2, map[string]string{"id": grn2.ID})
	recQC2 := httptest.NewRecorder()
	grnHandler.UpdateQCHandler(recQC2, reqQC2)
	reqComp2 := httptest.NewRequest(http.MethodPost, "/v1/warehouse/grn/"+grn2.ID+"/complete", nil)
	reqComp2 = mux.SetURLVars(reqComp2, map[string]string{"id": grn2.ID})
	recComp2 := httptest.NewRecorder()
	grnHandler.CompleteGRNHandler(recComp2, reqComp2)

	poState2, _ := repo.GetPOByID(context.Background(), po.ID)
	if poState2.Status != POStatusReceived {
		t.Errorf("Expected PO status RECEIVED after full quantity received, got %s", poState2.Status)
	}
}
