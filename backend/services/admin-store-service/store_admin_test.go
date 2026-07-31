package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// mockStoreClient records calls and returns canned responses for store admin handler tests.
type mockStoreClient struct {
	createStoreResult *StoreResponse
	createStoreCalled bool
	updateCalled      map[string]bool
}

func newMockStoreClient() *mockStoreClient {
	return &mockStoreClient{
		createStoreResult: &StoreResponse{ID: "store-new-1", Status: "ACTIVE"},
		updateCalled:      make(map[string]bool),
	}
}

func (m *mockStoreClient) CreateStore(req *CreateStoreRequest) (*StoreResponse, error) {
	m.createStoreCalled = true
	r := *m.createStoreResult
	r.ChainID = req.ChainID
	r.Name = req.Name
	return &r, nil
}
func (m *mockStoreClient) UpdateGeofence(storeID string, req *UpdateGeofenceRequest) error {
	m.updateCalled["geofence"] = true
	return nil
}
func (m *mockStoreClient) UpdateHours(storeID string, req *UpdateHoursRequest) error {
	m.updateCalled["hours"] = true
	return nil
}
func (m *mockStoreClient) UpdateCapacity(storeID string, req *UpdateCapacityRequest) error {
	m.updateCalled["capacity"] = true
	return nil
}
func (m *mockStoreClient) UpdateStatus(storeID string, req *UpdateStatusRequest) error {
	m.updateCalled["status"] = true
	return nil
}
func (m *mockStoreClient) UpdatePaymentSetup(storeID string, req *UpdatePaymentSetupRequest) error {
	m.updateCalled["payment"] = true
	return nil
}
func (m *mockStoreClient) RotateQRTokens(storeID string, req *RotateQRTokensRequest) (map[string]interface{}, error) {
	m.updateCalled["qr-rotate"] = true
	return map[string]interface{}{"tokens": []interface{}{}}, nil
}
func (m *mockStoreClient) GetQRTokens(storeID string) (map[string]interface{}, error) {
	return map[string]interface{}{"tokens": []interface{}{}}, nil
}
func (m *mockStoreClient) ListStores(query string) (map[string]interface{}, error) {
	return map[string]interface{}{"stores": []interface{}{}, "total": 0}, nil
}
func (m *mockStoreClient) GetStoreByID(storeID string) (*StoreResponse, error) {
	return &StoreResponse{ID: storeID, Status: "ACTIVE"}, nil
}

// ── Test: store creation blocked by suspended chain ───────────────────────────

func TestCreateStore_UnderSuspendedChain_Returns400ChainSuspended(t *testing.T) {
	repo := NewMemoryChainRepository()
	mock := newMockStoreClient()
	handler := NewStoreAdminHandler(repo, mock, "dev-secret")

	// Pre-create suspended chain
	_ = repo.CreateChain(context.Background(), &Chain{
		ID: "chain-suspended-1", Name: "Suspended Chain", Status: ChainStatusSuspended,
	})

	body, _ := json.Marshal(CreateStoreRequest{
		ChainID:     "chain-suspended-1",
		Name:        "Store Under Suspended",
		State:       "Karnataka",
		City:        "Bengaluru",
		Pincode:     "560001",
		CapacityMax: 50,
		OpeningTime: "08:00:00",
		ClosingTime: "22:00:00",
		Timezone:    "Asia/Kolkata",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/admin-store/stores", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "ADMIN")
	req.Header.Set("X-User-Role", "ADMIN")
	w := httptest.NewRecorder()
	handler.HandleCreateStore(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for suspended chain, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != CodeChainSuspended {
		t.Fatalf("expected code %s, got %v", CodeChainSuspended, errObj["code"])
	}
	// Crucially: store-service client must NOT be called
	if mock.createStoreCalled {
		t.Fatal("store-service client must NOT be called when chain is suspended")
	}
}

// ── Test: GSTIN checksum invalid → 400 GSTIN_CHECKSUM_INVALID ────────────────

func TestCreateStore_GSTINChecksumInvalid_Returns400(t *testing.T) {
	repo := NewMemoryChainRepository()
	_ = repo.CreateChain(context.Background(), &Chain{
		ID: "chain-gstin-1", Name: "GSTIN Test Chain", Status: ChainStatusActive,
	})

	mock := newMockStoreClient()
	handler := NewStoreAdminHandler(repo, mock, "dev-secret")

	body, _ := json.Marshal(CreateStoreRequest{
		ChainID: "chain-gstin-1",
		Name:    "Bad Checksum Store",
		State:   "Karnataka",
		GSTIN:   "29ABCDE1234F1ZX", // invalid checksum
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/admin-store/stores", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "ADMIN")
	req.Header.Set("X-User-Role", "ADMIN")
	w := httptest.NewRecorder()
	handler.HandleCreateStore(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad GSTIN checksum, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "GSTIN_CHECKSUM_INVALID" {
		t.Fatalf("expected GSTIN_CHECKSUM_INVALID, got %v", errObj["code"])
	}
}

// ── Test: GSTIN state mismatch → 400 GSTIN_STATE_MISMATCH ────────────────────

func TestCreateStore_GSTINStateMismatch_Returns400(t *testing.T) {
	repo := NewMemoryChainRepository()
	_ = repo.CreateChain(context.Background(), &Chain{
		ID: "chain-state-1", Name: "State Mismatch Chain", Status: ChainStatusActive,
	})

	mock := newMockStoreClient()
	handler := NewStoreAdminHandler(repo, mock, "dev-secret")

	body, _ := json.Marshal(CreateStoreRequest{
		ChainID: "chain-state-1",
		Name:    "State Mismatch Store",
		State:   "Karnataka", // prefix 29
		GSTIN:   "27AAPCU1427M1ZT", // prefix 27 = Maharashtra
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/admin-store/stores", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "ADMIN")
	req.Header.Set("X-User-Role", "ADMIN")
	w := httptest.NewRecorder()
	handler.HandleCreateStore(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for GSTIN state mismatch, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "GSTIN_STATE_MISMATCH" {
		t.Fatalf("expected GSTIN_STATE_MISMATCH, got %v", errObj["code"])
	}
}

// ── Test: valid store creation delegates to store-service client ─────────────

func TestCreateStore_ValidRequest_DelegatesToStoreService(t *testing.T) {
	repo := NewMemoryChainRepository()
	_ = repo.CreateChain(context.Background(), &Chain{
		ID: "chain-valid-1", Name: "Active Chain", Status: ChainStatusActive,
	})

	mock := newMockStoreClient()
	handler := NewStoreAdminHandler(repo, mock, "dev-secret")

	body, _ := json.Marshal(CreateStoreRequest{
		ChainID:     "chain-valid-1",
		Name:        "Bengaluru Flagship",
		Address:     "MG Road",
		City:        "Bengaluru",
		State:       "Karnataka",
		Pincode:     "560001",
		CapacityMax: 100,
		OpeningTime: "09:00:00",
		ClosingTime: "21:00:00",
		Timezone:    "Asia/Kolkata",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/admin-store/stores", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "ADMIN")
	req.Header.Set("X-User-Role", "ADMIN")
	w := httptest.NewRecorder()
	handler.HandleCreateStore(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}
	if !mock.createStoreCalled {
		t.Fatal("expected store-service client to be called for valid store creation")
	}
}

// ── Test: status change to INACTIVE requires step-up ─────────────────────────

func TestUpdateStoreStatus_Inactive_StepUpRequired(t *testing.T) {
	repo := NewMemoryChainRepository()
	mock := newMockStoreClient()
	handler := NewStoreAdminHandler(repo, mock, "dev-secret")

	body, _ := json.Marshal(UpdateStatusRequest{Status: "INACTIVE"})
	req := httptest.NewRequest(http.MethodPut, "/v1/admin-store/stores/store-1/status", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "ADMIN")
	req.Header.Set("X-User-Role", "ADMIN")
	// No StepUpAt in claims → should be rejected
	req = mux.SetURLVars(req, map[string]string{"id": "store-1"})
	w := httptest.NewRecorder()
	handler.HandleUpdateStatus(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 STEP_UP_REQUIRED for INACTIVE without step-up, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "STEP_UP_REQUIRED" {
		t.Fatalf("expected STEP_UP_REQUIRED, got %v", errObj["code"])
	}
}

// ── Test: non-admin cannot access admin-store endpoints ──────────────────────

func TestStoreAdmin_NonAdmin_Returns403(t *testing.T) {
	repo := NewMemoryChainRepository()
	mock := newMockStoreClient()
	handler := NewStoreAdminHandler(repo, mock, "dev-secret")

	body, _ := json.Marshal(CreateStoreRequest{ChainID: "c-1", Name: "Intruder Store"})
	req := httptest.NewRequest(http.MethodPost, "/v1/admin-store/stores", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "CHAIN_HQ") // not ADMIN
	w := httptest.NewRecorder()
	handler.HandleCreateStore(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin, got %d", w.Code)
	}
}
