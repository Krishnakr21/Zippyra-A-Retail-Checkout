package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockCacheManager struct{}

func (m *MockCacheManager) GetSKU(ctx context.Context, storeID, barcode string) (*Product, error) {
	return nil, nil
}
func (m *MockCacheManager) SetSKU(ctx context.Context, storeID, barcode string, product *Product) error {
	return nil
}
func (m *MockCacheManager) DeleteSKU(ctx context.Context, storeID, barcode string) error {
	return nil
}
func (m *MockCacheManager) GetCategoryTree(ctx context.Context, chainID string) ([]*Category, error) {
	return nil, nil
}
func (m *MockCacheManager) SetCategoryTree(ctx context.Context, chainID string, categories []*Category) error {
	return nil
}

type MockSearchEngine struct{}

func (m *MockSearchEngine) Search(ctx context.Context, storeID, query, categoryID string, page, pageSize int) (*SearchResponse, error) {
	return &SearchResponse{}, nil
}
func (m *MockSearchEngine) IndexProduct(ctx context.Context, p *Product) error {
	return nil
}
func (m *MockSearchEngine) DeleteProductIndex(ctx context.Context, storeID, productID string) error {
	return nil
}

func TestAdminHSNCheck_UnmappedCode_ReturnsIsReadyFalse(t *testing.T) {
	memRepo := NewMemoryRepository().(*MemoryRepository)
	handler := NewCatalogHandler(memRepo, &MockCacheManager{}, &MockSearchEngine{}, nil, "dev-secret")

	storeID := "store-hsn-1"

	memRepo.CreateHSNRate(context.Background(), &HsnGstRate{HSNCode: "0902", GSTRatePercent: 5.0})

	p1 := &Product{ID: "p1", StoreID: storeID, Barcode: "8901000000001", Name: "Tea", HSNCode: "0902"}
	p2 := &Product{ID: "p2", StoreID: storeID, Barcode: "8901000000002", Name: "Mystery Item", HSNCode: "9999"}
	_ = memRepo.CreateProduct(context.Background(), p1)
	_ = memRepo.CreateProduct(context.Background(), p2)

	req := httptest.NewRequest("GET", "/v1/catalog/admin/hsn-check?store_id=store-hsn-1", nil)
	w := httptest.NewRecorder()
	handler.HandleAdminHSNCheck(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["is_ready"] != false {
		t.Fatalf("expected is_ready false for unmapped HSN, got %v", resp["is_ready"])
	}

	missing := resp["missing_hsn_codes"].([]interface{})
	if len(missing) != 1 || missing[0] != "9999" {
		t.Fatalf("expected missing_hsn_codes ['9999'], got %v", missing)
	}
}

func TestAdminHSNCheck_FullyCoveredStore_ReturnsIsReadyTrue(t *testing.T) {
	memRepo := NewMemoryRepository().(*MemoryRepository)
	handler := NewCatalogHandler(memRepo, &MockCacheManager{}, &MockSearchEngine{}, nil, "dev-secret")

	storeID := "store-hsn-2"

	memRepo.CreateHSNRate(context.Background(), &HsnGstRate{HSNCode: "0902", GSTRatePercent: 5.0})
	p1 := &Product{ID: "p1", StoreID: storeID, Barcode: "8901000000003", Name: "Tea", HSNCode: "0902"}
	_ = memRepo.CreateProduct(context.Background(), p1)

	req := httptest.NewRequest("GET", "/v1/catalog/admin/hsn-check?store_id=store-hsn-2", nil)
	w := httptest.NewRecorder()
	handler.HandleAdminHSNCheck(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["is_ready"] != true {
		t.Fatalf("expected is_ready true for fully covered store, got %v", resp["is_ready"])
	}
}
