package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageProcessedWebhook_InvalidHMAC_Rejected(t *testing.T) {
	os.Setenv("LAMBDA_WEBHOOK_SHARED_SECRET", "test-secret-key-32bytes-123456789")
	defer os.Unsetenv("LAMBDA_WEBHOOK_SHARED_SECRET")

	repo := NewMemoryRepository()
	// Create product with PENDING image status
	prod := &Product{
		ID:                    "prod-001",
		StoreID:               "store-001",
		ChainID:               "chain-001",
		Barcode:               "8901030000018",
		Name:                  "Test Product",
		HSNCode:               "1001",
		ImageURL:              "raw/prod_001.jpg",
		ImageProcessingStatus: "PENDING",
	}
	_ = repo.CreateProduct(nil, prod)

	adminHandler := NewAdminCatalogHandler(repo, nil, nil, nil, nil, "jwt-secret")
	customerHandler := NewCatalogHandler(repo, nil, nil, nil, "jwt-secret")
	router := SetupRoutes(customerHandler, adminHandler, nil)

	payload := ImageProcessedWebhookRequest{
		S3RawKey:     "raw/prod_001.jpg",
		ThumbnailURL: "https://cdn.zippyra.com/thumbnails/prod_001.webp",
		FullURL:      "https://cdn.zippyra.com/full/prod_001.webp",
		Status:       "PROCESSED",
	}
	jsonBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/catalog/internal/image-processed", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "invalid-hmac-signature-hex")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Verify product was NOT updated
	updatedProd, err := repo.GetProductByID(nil, "prod-001")
	assert.NoError(t, err)
	assert.Equal(t, "PENDING", updatedProd.ImageProcessingStatus)
	assert.Equal(t, "raw/prod_001.jpg", updatedProd.ImageURL)
}

func TestImageProcessedWebhook_ValidHMAC_SuccessProcessed(t *testing.T) {
	secret := "test-secret-key-32bytes-123456789"
	os.Setenv("LAMBDA_WEBHOOK_SHARED_SECRET", secret)
	defer os.Unsetenv("LAMBDA_WEBHOOK_SHARED_SECRET")

	repo := NewMemoryRepository()
	prod := &Product{
		ID:                    "prod-002",
		StoreID:               "store-001",
		ChainID:               "chain-001",
		Barcode:               "8901030000025",
		Name:                  "Test Organic Juice",
		HSNCode:               "2009",
		ImageURL:              "raw/juice_002.png",
		ImageProcessingStatus: "PENDING",
	}
	_ = repo.CreateProduct(nil, prod)

	adminHandler := NewAdminCatalogHandler(repo, nil, nil, nil, nil, "jwt-secret")
	customerHandler := NewCatalogHandler(repo, nil, nil, nil, "jwt-secret")
	router := SetupRoutes(customerHandler, adminHandler, nil)

	payload := ImageProcessedWebhookRequest{
		S3RawKey:     "raw/juice_002.png",
		ThumbnailURL: "https://cdn.zippyra.com/thumbnails/juice_002.webp",
		FullURL:      "https://cdn.zippyra.com/full/juice_002.webp",
		Status:       "PROCESSED",
	}
	jsonBytes, _ := json.Marshal(payload)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(jsonBytes)
	validSig := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/v1/catalog/internal/image-processed", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", validSig)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp["status"])
	assert.Equal(t, "PROCESSED", resp["image_processing_status"])

	// Verify product updated to PROCESSED with thumbnail & full image URLs
	updatedProd, err := repo.GetProductByID(nil, "prod-002")
	assert.NoError(t, err)
	assert.Equal(t, "PROCESSED", updatedProd.ImageProcessingStatus)
	assert.Equal(t, "https://cdn.zippyra.com/full/juice_002.webp", updatedProd.ImageURL)
	assert.Equal(t, "https://cdn.zippyra.com/thumbnails/juice_002.webp", updatedProd.ThumbnailURL)
}

func TestImageProcessedWebhook_CorruptImage_FailurePath(t *testing.T) {
	secret := "test-secret-key-32bytes-123456789"
	os.Setenv("LAMBDA_WEBHOOK_SHARED_SECRET", secret)
	defer os.Unsetenv("LAMBDA_WEBHOOK_SHARED_SECRET")

	repo := NewMemoryRepository()
	prod := &Product{
		ID:                    "prod-003",
		StoreID:               "store-001",
		ChainID:               "chain-001",
		Barcode:               "8901030000032",
		Name:                  "Test Corrupt Image Product",
		HSNCode:               "1001",
		ImageURL:              "raw/corrupt_file.xyz",
		ImageProcessingStatus: "PENDING",
	}
	_ = repo.CreateProduct(nil, prod)

	adminHandler := NewAdminCatalogHandler(repo, nil, nil, nil, nil, "jwt-secret")
	customerHandler := NewCatalogHandler(repo, nil, nil, nil, "jwt-secret")
	router := SetupRoutes(customerHandler, adminHandler, nil)

	payload := ImageProcessedWebhookRequest{
		S3RawKey:     "raw/corrupt_file.xyz",
		Status:       "FAILED",
		ErrorMessage: "Unsupported image format or corrupt binary payload",
	}
	jsonBytes, _ := json.Marshal(payload)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(jsonBytes)
	validSig := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/v1/catalog/internal/image-processed", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", validSig)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "FAILED", resp["image_processing_status"])

	// Verify product status is set to FAILED and not left stuck at PENDING
	updatedProd, err := repo.GetProductByID(nil, "prod-003")
	assert.NoError(t, err)
	assert.Equal(t, "FAILED", updatedProd.ImageProcessingStatus)
}
