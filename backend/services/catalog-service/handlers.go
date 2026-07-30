package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/validator"
)

type CatalogHandler struct {
	repo         Repository
	cacheMgr     CacheManager
	searchEngine SearchEngine
	syncEngine   *SyncEngineService
	jwtSecret    string
}

func NewCatalogHandler(repo Repository, cacheMgr CacheManager, searchEngine SearchEngine, syncEngine *SyncEngineService, jwtSecret string) *CatalogHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &CatalogHandler{
		repo:         repo,
		cacheMgr:     cacheMgr,
		searchEngine: searchEngine,
		syncEngine:   syncEngine,
		jwtSecret:    jwtSecret,
	}
}

// HandleGetByBarcode retrieves item catalog details by barcode EAN-13/UPC-A
// @Summary Get Product by Barcode
// @Description Retrieve item pricing, tax rates, and availability by scanned barcode
// @Tags Catalog
// @Produce json
// @Param barcode path string true "Barcode EAN-13 / UPC-A"
// @Param store_id query string true "Store ID"
// @Success 200 {object} SKUDetailDTO
// @Failure 400 {object} errors.APIError
// @Failure 404 {object} errors.APIError
// @Router /v1/catalog/barcode/{barcode} [get]
func (h *CatalogHandler) HandleGetByBarcode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
	defer cancel()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[3] == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Missing barcode in path", nil)
		return
	}
	barcode := parts[3]
	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id parameter is required", nil)
		return
	}

	if !validator.ValidateBarcode(barcode) {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeBarcodeInvalid, "Invalid EAN-13 / UPC-A barcode checksum", nil)
		return
	}

	if cachedProduct, err := h.cacheMgr.GetSKU(ctx, storeID, barcode); err == nil && cachedProduct != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache-Hit", "true")
		_ = json.NewEncoder(w).Encode(cachedProduct)
		return
	}

	product, err := h.repo.GetProductByBarcode(ctx, storeID, barcode)
	if err != nil || product == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeProductNotFound, "Product not found", nil)
		return
	}

	_ = h.cacheMgr.SetSKU(ctx, storeID, barcode, product)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Hit", "false")
	_ = json.NewEncoder(w).Encode(product)
}

// GET /v1/catalog/search?q={query}&store_id={id}&category_id={cat}&page={n}
func (h *CatalogHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
	defer cancel()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	storeID := r.URL.Query().Get("store_id")
	categoryID := r.URL.Query().Get("category_id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id parameter is required", nil)
		return
	}

	if query == "" {
		products, total, err := h.repo.AdminListProducts(ctx, storeID, page, pageSize)
		if err != nil {
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list products", nil)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"products":  products,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
		return
	}

	res, err := h.searchEngine.Search(ctx, storeID, query, categoryID, page, pageSize)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Search operation failed", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GET /v1/catalog/categories?chain_id={id}
func (h *CatalogHandler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	chainID := r.URL.Query().Get("chain_id")
	if chainID == "" {
		chainID = "chain-hq-001"
	}

	categories, err := h.repo.GetCategoriesByChain(r.Context(), chainID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to fetch categories", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"categories": categories})
}

// GET /v1/catalog/sync?store_id={id}&since_seq={n}&limit={m}
func (h *CatalogHandler) HandleDeltaSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
	defer cancel()

	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id parameter is required", nil)
		return
	}

	sinceSeq, _ := strconv.ParseInt(r.URL.Query().Get("since_seq"), 10, 64)

	resp, err := h.syncEngine.PerformDeltaSync(ctx, &CatalogSyncRequest{
		StoreID:  storeID,
		SinceSeq: sinceSeq,
	})
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to perform delta sync", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GET /v1/catalog/admin/products?chain_id={id}&store_id={id}&page={n}
func (h *CatalogHandler) HandleAdminListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	chainID := r.URL.Query().Get("chain_id")
	storeID := r.URL.Query().Get("store_id")

	if chainID == "" && storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id or chain_id parameter is required", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	products, total, err := h.repo.AdminListProducts(r.Context(), storeID, page, pageSize)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list products", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"products":  products,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GET /v1/catalog/admin/hsn-check?store_id={id}
func (h *CatalogHandler) HandleAdminHSNCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id parameter is required", nil)
		return
	}

	totalHSN, missingHSN, isReady, err := h.repo.CheckStoreHSN(r.Context(), storeID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to run HSN check", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"total_hsn_codes":   totalHSN,
		"missing_hsn_codes": missingHSN,
		"is_ready":          isReady,
	})
}
