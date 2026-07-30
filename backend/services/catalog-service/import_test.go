package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSVImport_WorkerBatchValidation(t *testing.T) {
	repo := NewMemoryRepository()
	cacheMgr := NewMemoryCacheManager()
	worker := NewImportWorker(repo, cacheMgr, nil, nil)
	ctx := context.Background()

	storeID := "store-import-1"
	chainID := "chain-1"

	job := &CatalogImportJob{
		ID:        "job-100",
		StoreID:   storeID,
		ChainID:   chainID,
		Status:    "PENDING",
		ErrorRows: []*ImportRowError{},
	}
	_ = repo.CreateImportJob(ctx, job)

	// CSV file with 3 good rows and 2 bad rows:
	// Row 1: Valid (EAN-13 8901030300011, HSN 0901)
	// Row 2: Invalid barcode checksum (8901030300019)
	// Row 3: Valid (UPC-A 012345678905, HSN 1905)
	// Row 4: Unknown HSN code (4006381333931, HSN 9999)
	// Row 5: Valid (EAN-13 8901030300028, HSN 0901)
	csvData := `barcode,name,description,category_id,price_paise,mrp_paise,hsn_code,is_returnable
8901030300011,Coffee Pack 100g,Fresh Blend,,25000,28000,0901,true
8901030300019,Bad Barcode Item,Invalid Checksum,,15000,18000,0901,true
012345678905,Biscuits Pack,Crispy,,10000,12000,1905,true
4006381333931,Unknown Tax Item,Invalid HSN,,30000,35000,9999,true
8901030300028,Valid High-End Pen,Fine Nib,,45000,50000,0901,true`

	worker.ProcessCSVImportJob(ctx, job.ID, strings.NewReader(csvData))

	completedJob, err := repo.GetImportJob(ctx, job.ID)
	if err != nil || completedJob == nil {
		t.Fatalf("Failed to fetch completed import job: %v", err)
	}

	if completedJob.Status != "COMPLETED" {
		t.Errorf("Expected status COMPLETED, got %s", completedJob.Status)
	}

	if completedJob.ProcessedRows != 3 {
		t.Errorf("Expected 3 processed rows, got %d", completedJob.ProcessedRows)
	}

	if len(completedJob.ErrorRows) != 2 {
		t.Errorf("Expected exactly 2 error rows, got %d", len(completedJob.ErrorRows))
	}
}

func TestAdmin_ChainIsolation_Forbidden(t *testing.T) {
	repo := NewMemoryRepository()
	cacheMgr := NewMemoryCacheManager()
	adminHandler := NewAdminCatalogHandler(repo, cacheMgr, nil, nil, nil, "test-secret")
	ctx := context.Background()

	// Seed product in Chain A
	productChainA := &Product{
		ID:         "p-chain-A",
		StoreID:    "store-A",
		ChainID:    "chain-A",
		Barcode:    "8901030300011",
		Name:       "Chain A Product",
		PricePaise: 10000,
		IsActive:   true,
	}
	_ = repo.CreateProduct(ctx, productChainA)

	// Admin belonging to Chain B attempts to update product in Chain A
	reqBody := `{"name":"Hacked Name","price_paise":1}`
	req := httptest.NewRequest(http.MethodPut, "/v1/catalog/admin/products/p-chain-A", strings.NewReader(reqBody))
	req.Header.Set("X-Chain-ID", "chain-B") // Admin is for Chain B

	w := httptest.NewRecorder()
	adminHandler.HandleUpdateProduct(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 Forbidden for cross-chain product update, got %d", w.Code)
	}
}
