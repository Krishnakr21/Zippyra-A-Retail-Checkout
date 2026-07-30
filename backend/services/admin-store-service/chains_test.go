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

// ── Null StoreServiceClient for chain-only tests ──────────────────────────────

type nullStoreClient struct{}

func (n *nullStoreClient) CreateStore(req *CreateStoreRequest) (*StoreResponse, error) {
	return &StoreResponse{ID: "store-new-1", ChainID: req.ChainID, Name: req.Name, Status: "ACTIVE"}, nil
}
func (n *nullStoreClient) UpdateGeofence(storeID string, req *UpdateGeofenceRequest) error { return nil }
func (n *nullStoreClient) UpdateHours(storeID string, req *UpdateHoursRequest) error       { return nil }
func (n *nullStoreClient) UpdateCapacity(storeID string, req *UpdateCapacityRequest) error { return nil }
func (n *nullStoreClient) UpdateStatus(storeID string, req *UpdateStatusRequest) error     { return nil }
func (n *nullStoreClient) UpdatePaymentSetup(storeID string, req *UpdatePaymentSetupRequest) error {
	return nil
}
func (n *nullStoreClient) RotateQRTokens(storeID string, req *RotateQRTokensRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"tokens": []interface{}{}}, nil
}
func (n *nullStoreClient) GetQRTokens(storeID string) (map[string]interface{}, error) {
	return map[string]interface{}{"tokens": []interface{}{}}, nil
}

// ── Chain CRUD tests ──────────────────────────────────────────────────────────

func TestCreateChain_Success(t *testing.T) {
	repo := NewMemoryChainRepository()
	handler := NewChainHandler(repo, "dev-secret")

	body, _ := json.Marshal(CreateChainRequest{
		Name:               "Mega Retail Chain",
		LegalEntityName:    "Mega Retail Pvt Ltd",
		DefaultGstInPrefix: "29",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/admin-store/chains", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "ADMIN")
	req.Header.Set("X-User-Role", "ADMIN")
	w := httptest.NewRecorder()
	handler.HandleCreateChain(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}
	var chain Chain
	_ = json.Unmarshal(w.Body.Bytes(), &chain)
	if chain.ID == "" {
		t.Fatal("expected non-empty chain ID in response")
	}
	if chain.Status != ChainStatusActive {
		t.Fatalf("expected status ACTIVE, got %s", chain.Status)
	}
}

func TestChainStatusUpdate_StepUpRequired(t *testing.T) {
	repo := NewMemoryChainRepository()
	handler := NewChainHandler(repo, "dev-secret")

	// Pre-create chain
	_ = repo.CreateChain(context.Background(), &Chain{
		ID: "chain-stepup-1", Name: "Step-Up Chain", Status: ChainStatusActive,
	})

	// Attempt status change WITHOUT step-up claim — expect 403 STEP_UP_REQUIRED
	body, _ := json.Marshal(UpdateChainStatusRequest{Status: "SUSPENDED"})
	req := httptest.NewRequest(http.MethodPut, "/v1/admin-store/chains/chain-stepup-1/status", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "ADMIN")
	req.Header.Set("X-User-Role", "ADMIN")
	req = mux.SetURLVars(req, map[string]string{"id": "chain-stepup-1"})
	w := httptest.NewRecorder()
	handler.HandleUpdateChainStatus(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 STEP_UP_REQUIRED without step-up claim, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if errObj, ok := resp["error"].(map[string]interface{}); ok {
		if errObj["code"] != "STEP_UP_REQUIRED" {
			t.Fatalf("expected code STEP_UP_REQUIRED, got %v", errObj["code"])
		}
	}
}

func TestChainDetail_NotFound_Returns404(t *testing.T) {
	repo := NewMemoryChainRepository()
	handler := NewChainHandler(repo, "dev-secret")

	req := httptest.NewRequest(http.MethodGet, "/v1/admin-store/chains/nonexistent-id", nil)
	req.Header.Set("X-User-Type", "ADMIN")
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent-id"})
	w := httptest.NewRecorder()
	handler.HandleGetChain(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown chain, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if errObj, ok := resp["error"].(map[string]interface{}); ok {
		if errObj["code"] != CodeChainNotFound {
			t.Fatalf("expected code %s, got %v", CodeChainNotFound, errObj["code"])
		}
	}
}

func TestNonAdmin_Returns403(t *testing.T) {
	repo := NewMemoryChainRepository()
	handler := NewChainHandler(repo, "dev-secret")

	body, _ := json.Marshal(CreateChainRequest{Name: "Unauthorized Chain"})
	req := httptest.NewRequest(http.MethodPost, "/v1/admin-store/chains", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "STAFF") // not ADMIN
	w := httptest.NewRecorder()
	handler.HandleCreateChain(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-admin, got %d", w.Code)
	}
}
