package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/validator"
)

type AdminCatalogHandler struct {
	repo          Repository
	cacheMgr      CacheManager
	searchEngine  SearchEngine
	importWorker  *ImportWorker
	kafkaProducer *kafka.Producer
	jwtSecret     string
}

func NewAdminCatalogHandler(repo Repository, cacheMgr CacheManager, searchEngine SearchEngine, importWorker *ImportWorker, producer *kafka.Producer, jwtSecret string) *AdminCatalogHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &AdminCatalogHandler{
		repo:          repo,
		cacheMgr:      cacheMgr,
		searchEngine:  searchEngine,
		importWorker:  importWorker,
		kafkaProducer: producer,
		jwtSecret:     jwtSecret,
	}
}

// POST /v1/catalog/admin/products
func (h *AdminCatalogHandler) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	adminChainID := h.getChainIDFromContext(r)
	if adminChainID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Admin authentication required", nil)
		return
	}

	var req AdminProductCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	// Chain isolation enforcement
	if req.ChainID != "" && req.ChainID != adminChainID {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Chain isolation violation: cannot create product for another chain", nil)
		return
	}

	// Validate Barcode checksum
	if !validator.ValidateBarcode(req.Barcode) {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeBarcodeInvalid, "Invalid EAN-13 or UPC-A barcode checksum", nil)
		return
	}

	// Validate HSN Code
	hsnRate, err := h.repo.GetHSNRate(r.Context(), req.HSNCode)
	if err != nil || hsnRate == nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeHSNCodeNotFound, "Invalid or unknown HSN code", nil)
		return
	}

	if req.PricePaise <= 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "price_paise must be a positive integer", nil)
		return
	}

	mrpPaise := req.MRPPaise
	if mrpPaise < req.PricePaise {
		mrpPaise = req.PricePaise
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	isReturnable := true
	if req.IsReturnable != nil {
		isReturnable = *req.IsReturnable
	}

	imageStatus := "PROCESSED"
	thumbnailURL := req.ThumbnailURL
	if req.ImageURL != "" && (strings.Contains(req.ImageURL, "raw/") || strings.HasPrefix(req.ImageURL, "raw/")) {
		imageStatus = "PENDING"
		thumbnailURL = ""
	}

	product := &Product{
		ID:                    uuid.New().String(),
		StoreID:               req.StoreID,
		ChainID:               adminChainID,
		Barcode:               req.Barcode,
		Name:                  req.Name,
		Description:           req.Description,
		CategoryID:            req.CategoryID,
		PricePaise:            req.PricePaise,
		MRPPaise:              mrpPaise,
		HSNCode:               req.HSNCode,
		GSTRatePercent:        hsnRate.GSTRatePercent,
		IsActive:              isActive,
		IsReturnable:          isReturnable,
		ImageURL:              req.ImageURL,
		ThumbnailURL:          thumbnailURL,
		ImageProcessingStatus: imageStatus,
	}

	if err := h.repo.CreateProduct(r.Context(), product); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create product", nil)
		return
	}

	// Write-through cache update
	_ = h.cacheMgr.SetSKU(r.Context(), product.StoreID, product.Barcode, product)

	// Async ES index
	if h.searchEngine != nil {
		_ = h.searchEngine.IndexProduct(r.Context(), product)
	}

	// Publish catalog.updated Kafka event
	h.publishCatalogUpdated(r.Context(), product.StoreID, adminChainID, product.SyncSeq, 1)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(product)
}

// PUT /v1/catalog/admin/products/{id}
func (h *AdminCatalogHandler) HandleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	adminChainID := h.getChainIDFromContext(r)
	if adminChainID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Admin authentication required", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var productID string
	for i, p := range parts {
		if p == "products" && i+1 < len(parts) {
			productID = parts[i+1]
			break
		}
	}
	if productID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Missing product id in path", nil)
		return
	}

	existing, err := h.repo.GetProductByID(r.Context(), productID)
	if err != nil || existing == nil || existing.DeletedAt != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeProductNotFound, "Product not found", nil)
		return
	}

	// Chain isolation check
	if existing.ChainID != adminChainID {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Chain isolation violation: product belongs to another chain", nil)
		return
	}

	var req AdminProductUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.CategoryID != nil {
		existing.CategoryID = req.CategoryID
	}
	if req.PricePaise != nil && *req.PricePaise > 0 {
		existing.PricePaise = *req.PricePaise
	}
	if req.MRPPaise != nil {
		existing.MRPPaise = *req.MRPPaise
	}
	if req.HSNCode != "" {
		hsnRate, err := h.repo.GetHSNRate(r.Context(), req.HSNCode)
		if err == nil && hsnRate != nil {
			existing.HSNCode = req.HSNCode
			existing.GSTRatePercent = hsnRate.GSTRatePercent
		}
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.IsReturnable != nil {
		existing.IsReturnable = *req.IsReturnable
	}
	if req.ImageURL != "" {
		existing.ImageURL = req.ImageURL
	}
	if req.ThumbnailURL != "" {
		existing.ThumbnailURL = req.ThumbnailURL
	}

	if err := h.repo.UpdateProduct(r.Context(), existing); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update product", nil)
		return
	}

	// Write-through cache update
	_ = h.cacheMgr.SetSKU(r.Context(), existing.StoreID, existing.Barcode, existing)

	// Async ES Index
	if h.searchEngine != nil {
		_ = h.searchEngine.IndexProduct(r.Context(), existing)
	}

	// Publish catalog.updated
	h.publishCatalogUpdated(r.Context(), existing.StoreID, existing.ChainID, existing.SyncSeq, 1)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(existing)
}

// DELETE /v1/catalog/admin/products/{id}
func (h *AdminCatalogHandler) HandleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	adminChainID := h.getChainIDFromContext(r)
	if adminChainID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Admin authentication required", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var productID string
	for i, p := range parts {
		if p == "products" && i+1 < len(parts) {
			productID = parts[i+1]
			break
		}
	}
	if productID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Missing product id in path", nil)
		return
	}

	existing, err := h.repo.GetProductByID(r.Context(), productID)
	if err != nil || existing == nil || existing.DeletedAt != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeProductNotFound, "Product not found", nil)
		return
	}

	if existing.ChainID != adminChainID {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Chain isolation violation: product belongs to another chain", nil)
		return
	}

	deletedProduct, err := h.repo.SoftDeleteProduct(r.Context(), productID)
	if err != nil || deletedProduct == nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to delete product", nil)
		return
	}

	// Delete from Redis SKU cache
	_ = h.cacheMgr.DeleteSKU(r.Context(), deletedProduct.StoreID, deletedProduct.Barcode)

	// Delete from ES index
	if h.searchEngine != nil {
		_ = h.searchEngine.DeleteProductIndex(r.Context(), deletedProduct.StoreID, productID)
	}

	// Publish catalog.updated
	h.publishCatalogUpdated(r.Context(), deletedProduct.StoreID, deletedProduct.ChainID, deletedProduct.SyncSeq, 1)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "DELETED"})
}

// POST /v1/catalog/admin/import (Multipart CSV file upload, max 25MB)
func (h *AdminCatalogHandler) HandleCSVImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	adminChainID := h.getChainIDFromContext(r)
	if adminChainID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Admin authentication required", nil)
		return
	}

	// Limit multipart body to 25MB
	r.Body = http.MaxBytesReader(w, r.Body, 25*1024*1024)
	if err := r.ParseMultipartForm(25 * 1024 * 1024); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeImportFileInvalid, "File upload size exceeds 25MB limit", nil)
		return
	}

	storeID := r.FormValue("store_id")
	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id parameter is required", nil)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeImportFileInvalid, "Missing CSV file in form field 'file'", nil)
		return
	}
	defer file.Close()

	job := &CatalogImportJob{
		ID:        uuid.New().String(),
		StoreID:   storeID,
		ChainID:   adminChainID,
		Status:    "PENDING",
		ErrorRows: []*ImportRowError{},
	}

	if err := h.repo.CreateImportJob(r.Context(), job); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create import job", nil)
		return
	}

	// Trigger async CSV import worker
	if h.importWorker != nil {
		go h.importWorker.ProcessCSVImportJob(context.Background(), job.ID, file)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(job)
}

// GET /v1/catalog/admin/import/{job_id}/status
func (h *AdminCatalogHandler) HandleGetImportStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var jobID string
	for i, p := range parts {
		if p == "import" && i+1 < len(parts) && parts[i+1] != "status" {
			jobID = parts[i+1]
			break
		}
	}
	if jobID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Missing job_id in path", nil)
		return
	}

	job, err := h.repo.GetImportJob(r.Context(), jobID)
	if err != nil || job == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Import job not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(job)
}

// POST /v1/catalog/admin/reindex?store_id={id}
func (h *AdminCatalogHandler) HandleReindexES(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id parameter is required", nil)
		return
	}

	go func() {
		products, _, err := h.repo.SearchProductsPostgres(context.Background(), storeID, "", "", 1, 10000)
		if err == nil {
			for _, p := range products {
				if h.searchEngine != nil {
					_ = h.searchEngine.IndexProduct(context.Background(), p)
				}
			}
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "REINDEXING_TRIGGERED"})
}

func (h *AdminCatalogHandler) getChainIDFromContext(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		if chainID := r.Header.Get("X-Chain-ID"); chainID != "" {
			return chainID
		}
		return ""
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		return ""
	}
	if chainID := r.Header.Get("X-Chain-ID"); chainID != "" {
		return chainID
	}
	return "chain-hq-001"
}

func (h *AdminCatalogHandler) publishCatalogUpdated(ctx context.Context, storeID, chainID string, syncSeq int64, changedCount int) {
	if h.kafkaProducer == nil {
		return
	}
	payload := map[string]interface{}{
		"store_id":      storeID,
		"chain_id":      chainID,
		"sync_seq":      syncSeq,
		"changed_count": changedCount,
		"ts":            time.Now().Unix(),
	}
	_ = h.kafkaProducer.PublishEvent(ctx, "catalog.updated", storeID, payload)
}
