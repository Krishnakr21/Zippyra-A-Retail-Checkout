package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
)

// POST /v1/catalog/internal/image-processed
func (h *AdminCatalogHandler) HandleImageProcessedWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	secret := os.Getenv("LAMBDA_WEBHOOK_SHARED_SECRET")
	if secret == "" {
		secret = "zippyra-lambda-webhook-secret-32bytes"
	}

	providedSig := r.Header.Get("X-Signature")
	if providedSig == "" {
		providedSig = r.Header.Get("X-Webhook-Signature")
	}

	if providedSig == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Missing webhook signature header", nil)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Failed to read request body", nil)
		return
	}

	// Verify HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(bodyBytes)
	computedSig := hex.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(strings.ToLower(providedSig)), []byte(strings.ToLower(computedSig))) != 1 {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid webhook HMAC signature", nil)
		return
	}

	var req ImageProcessedWebhookRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid webhook JSON payload", nil)
		return
	}

	req.S3RawKey = strings.TrimSpace(req.S3RawKey)
	if req.S3RawKey == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "s3_raw_key is required", nil)
		return
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != "PROCESSED" && status != "FAILED" {
		status = "PROCESSED"
	}

	if status == "FAILED" {
		logger.Warn("Image processing failed for key %s: %s", req.S3RawKey, req.ErrorMessage)
	}

	product, err := h.repo.UpdateProductImageStatus(r.Context(), req.S3RawKey, req.FullURL, req.ThumbnailURL, status)
	if err != nil {
		logger.Error("Failed to update product image status for %s: %v", req.S3RawKey, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update product image status", nil)
		return
	}

	// Evict cache if product was updated
	if product != nil && h.cacheMgr != nil {
		_ = h.cacheMgr.DeleteSKU(r.Context(), product.StoreID, product.Barcode)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":                  "success",
		"s3_raw_key":              req.S3RawKey,
		"image_processing_status": status,
		"product_id":              func() string { if product != nil { return product.ID }; return "" }(),
	})
}
