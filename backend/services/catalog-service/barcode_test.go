package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zippyra/backend/shared/validator"
)

func TestBarcodeValidator_EAN13_And_UPCA(t *testing.T) {
	// Valid EAN-13 barcodes
	validEAN13 := []string{
		"8901030300011", // Standard Indian FMCG product barcode (checksum 1)
		"4006381333931",
	}
	for _, bc := range validEAN13 {
		if !validator.ValidateEAN13(bc) {
			t.Errorf("Expected EAN-13 %s to be valid", bc)
		}
		if !validator.ValidateBarcode(bc) {
			t.Errorf("Expected barcode %s to be valid", bc)
		}
	}

	// Invalid EAN-13 (bad check digit)
	invalidEAN13 := "8901030300019" // Wrong checksum digit
	if validator.ValidateEAN13(invalidEAN13) {
		t.Errorf("Expected invalid EAN-13 checksum for %s to fail validation", invalidEAN13)
	}

	// Valid UPC-A barcodes
	validUPCA := []string{
		"012345678905",
		"036000291452",
	}
	for _, bc := range validUPCA {
		if !validator.ValidateUPCA(bc) {
			t.Errorf("Expected UPC-A %s to be valid", bc)
		}
		if !validator.ValidateBarcode(bc) {
			t.Errorf("Expected barcode %s to be valid", bc)
		}
	}

	// Invalid UPC-A
	invalidUPCA := "012345678909"
	if validator.ValidateUPCA(invalidUPCA) {
		t.Errorf("Expected invalid UPC-A checksum for %s to fail validation", invalidUPCA)
	}
}

type SpyRepository struct {
	*MemoryRepository
	GetProductByBarcodeCallCount int
}

func (s *SpyRepository) GetProductByBarcode(ctx context.Context, storeID, barcode string) (*Product, error) {
	s.GetProductByBarcodeCallCount++
	return s.MemoryRepository.GetProductByBarcode(ctx, storeID, barcode)
}

func TestBarcodeHandler_MalformedInput_RejectionBeforeDB(t *testing.T) {
	spyRepo := &SpyRepository{MemoryRepository: NewMemoryRepository().(*MemoryRepository)}
	cacheMgr := NewMemoryCacheManager()
	handler := NewCatalogHandler(spyRepo, cacheMgr, nil, nil, "test-secret")

	// Malformed barcode input
	req := httptest.NewRequest(http.MethodGet, "/v1/catalog/barcode/123456?store_id=store-1", nil)
	w := httptest.NewRecorder()

	handler.HandleGetByBarcode(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request for malformed barcode, got %d", w.Code)
	}

	// Assert repository was NEVER touched
	if spyRepo.GetProductByBarcodeCallCount != 0 {
		t.Errorf("Expected 0 DB queries for malformed barcode, got %d", spyRepo.GetProductByBarcodeCallCount)
	}
}

func TestBarcodeHandler_RedisHit_BypassesDB(t *testing.T) {
	spyRepo := &SpyRepository{MemoryRepository: NewMemoryRepository().(*MemoryRepository)}
	cacheMgr := NewMemoryCacheManager()
	handler := NewCatalogHandler(spyRepo, cacheMgr, nil, nil, "test-secret")

	ctx := context.Background()
	storeID := "store-1"
	barcode := "8901030300011"

	// Pre-populate Redis SKU cache directly
	cachedProduct := &Product{
		ID:         "p-cached-1",
		StoreID:    storeID,
		Barcode:    barcode,
		Name:       "Cached Coffee",
		PricePaise: 25000,
	}
	_ = cacheMgr.SetSKU(ctx, storeID, barcode, cachedProduct)

	req := httptest.NewRequest(http.MethodGet, "/v1/catalog/barcode/"+barcode+"?store_id="+storeID, nil)
	w := httptest.NewRecorder()

	handler.HandleGetByBarcode(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK on Redis hit, got %d", w.Code)
	}

	// Assert DB was NEVER queried on cache hit
	if spyRepo.GetProductByBarcodeCallCount != 0 {
		t.Errorf("Expected 0 DB calls on Redis SKU hit, got %d", spyRepo.GetProductByBarcodeCallCount)
	}
}

func TestBarcodeHandler_CacheMiss_BackfillsRedis(t *testing.T) {
	spyRepo := &SpyRepository{MemoryRepository: NewMemoryRepository().(*MemoryRepository)}
	cacheMgr := NewMemoryCacheManager()
	handler := NewCatalogHandler(spyRepo, cacheMgr, nil, nil, "test-secret")

	ctx := context.Background()
	storeID := "store-1"
	barcode := "8901030300011"

	// Seed product in Postgres repo only (not in Redis cache)
	dbProduct := &Product{
		ID:         "p-db-1",
		StoreID:    storeID,
		ChainID:    "chain-1",
		Barcode:    barcode,
		Name:       "DB Coffee",
		PricePaise: 30000,
		HSNCode:    "0901",
		IsActive:   true,
	}
	_ = spyRepo.CreateProduct(ctx, dbProduct)

	// First request: Cache Miss -> Queries DB
	req1 := httptest.NewRequest(http.MethodGet, "/v1/catalog/barcode/"+barcode+"?store_id="+storeID, nil)
	w1 := httptest.NewRecorder()
	handler.HandleGetByBarcode(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected 200 OK on first request, got %d", w1.Code)
	}
	if spyRepo.GetProductByBarcodeCallCount != 1 {
		t.Errorf("Expected 1 DB call on cache miss, got %d", spyRepo.GetProductByBarcodeCallCount)
	}

	// Second request: Cache Hit (Backfilled) -> Bypasses DB
	req2 := httptest.NewRequest(http.MethodGet, "/v1/catalog/barcode/"+barcode+"?store_id="+storeID, nil)
	w2 := httptest.NewRecorder()
	handler.HandleGetByBarcode(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected 200 OK on second request, got %d", w2.Code)
	}
	if spyRepo.GetProductByBarcodeCallCount != 1 {
		t.Errorf("Expected DB call count to remain 1 after cache backfill, got %d", spyRepo.GetProductByBarcodeCallCount)
	}
}
