package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/redis"
)

type InventoryHandler struct {
	repo     Repository
	engine   *MovementEngine
	redisClient *redis.Client
}

func NewInventoryHandler(repo Repository, engine *MovementEngine, redisClient *redis.Client) *InventoryHandler {
	return &InventoryHandler{
		repo:     repo,
		engine:   engine,
		redisClient: redisClient,
	}
}

func (h *InventoryHandler) GetStockHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	storeID := r.URL.Query().Get("store_id")
	barcode := r.URL.Query().Get("barcode")

	if storeID == "" || barcode == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id and barcode parameters are required", nil)
		return
	}

	sl, err := h.repo.GetStockLevel(ctx, storeID, barcode)
	if err != nil {
		logger.Error("Failed to fetch stock level for %s / %s: %v", storeID, barcode, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve stock level", nil)
		return
	}

	if sl == nil {
		// Default zero stock level if never recorded
		sl = &StockLevel{
			StoreID:      storeID,
			Barcode:      barcode,
			OnHandQty:    0,
			ReorderPoint: 10,
			ReorderQty:   50,
			UpdatedAt:    time.Now(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sl)
}

func (h *InventoryHandler) GetLowStockHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id parameter is required", nil)
		return
	}

	levels, err := h.repo.GetLowStockLevels(ctx, storeID)
	if err != nil {
		logger.Error("Failed to fetch low stock levels for store %s: %v", storeID, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve low stock levels", nil)
		return
	}

	var responseItems []LowStockItemResponse
	for _, sl := range levels {
		productName := ""
		if h.redisClient != nil {
			// Try reading cached SKU name directly: sku:{store_id}:{barcode}
			val, err := h.redisClient.Get(ctx, "sku:"+storeID+":"+sl.Barcode).Result()
			if err == nil && val != "" {
				productName = val
			}
		}

		responseItems = append(responseItems, LowStockItemResponse{
			Barcode:      sl.Barcode,
			ProductName:  productName,
			OnHandQty:    sl.OnHandQty,
			ReorderPoint: sl.ReorderPoint,
			ReorderQty:   sl.ReorderQty,
		})
	}

	if responseItems == nil {
		responseItems = []LowStockItemResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": responseItems,
	})
}

func (h *InventoryHandler) AdjustStockHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, _ := r.Context().Value("user_claims").(*jwt.SessionClaims)
	
	var req AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.StoreID == "" || req.Barcode == "" || req.QtyDelta == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id, barcode, and non-zero qty_delta are required", nil)
		return
	}

	// Chain/store scoping guard for non-SYSTEM callers
	if claims != nil && claims.UserType != "SYSTEM" && claims.StoreID != "" && claims.StoreID != req.StoreID {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Cannot adjust stock for another store", nil)
		return
	}

	var createdBy *string
	if claims != nil {
		createdBy = &claims.UserID
	}

	refID := uuid.New().String()
	noteStr := req.Reason
	if req.Note != "" {
		noteStr = req.Reason + ": " + req.Note
	}

	applied, newQty, err := h.engine.ApplyMovement(
		ctx,
		nil,
		req.StoreID,
		req.Barcode,
		MovementAdjustment,
		req.QtyDelta,
		RefManual,
		refID,
		createdBy,
		&noteStr,
		false, // allowNegative = false
	)

	if err != nil {
		if strings.Contains(err.Error(), "INSUFFICIENT_STOCK") {
			errors.WriteError(w, http.StatusBadRequest, errors.CodeInsufficientStockForAdjustment, "Stock adjustment would result in negative on-hand quantity", nil)
			return
		}
		logger.Error("Failed to apply stock adjustment: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to apply stock adjustment", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"applied":       applied,
		"on_hand_qty":   newQty,
		"adjustment_id": refID,
	})
}

func (h *InventoryHandler) StockCountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, _ := r.Context().Value("user_claims").(*jwt.SessionClaims)

	var req StockCountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.StoreID == "" || len(req.Entries) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id and entries are required", nil)
		return
	}

	if claims != nil && claims.UserType != "SYSTEM" && claims.StoreID != "" && claims.StoreID != req.StoreID {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Cannot submit stock count for another store", nil)
		return
	}

	userID := "SYSTEM"
	if claims != nil {
		userID = claims.UserID
	}

	tx, err := h.repo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to begin stock count transaction", nil)
		return
	}
	defer tx.Rollback()

	var results []StockCountEntryResult
	discrepanciesCount := 0
	now := time.Now()

	for _, entry := range req.Entries {
		sl, err := h.repo.GetStockLevel(ctx, entry.Barcode, req.StoreID) // Note: order in query
		expectedQty := int64(0)
		if sl == nil {
			// Query directly in tx if needed
			var qty int64
			errQ := tx.QueryRowContext(ctx, `SELECT on_hand_qty FROM stock_levels WHERE store_id = $1 AND barcode = $2`, req.StoreID, entry.Barcode).Scan(&qty)
			if errQ == nil {
				expectedQty = qty
			}
		} else {
			expectedQty = sl.OnHandQty
		}

		varianceQty := entry.CountedQty - expectedQty
		countID := uuid.New().String()

		// Record stock_counts row
		_, err = tx.ExecContext(ctx, `
			INSERT INTO stock_counts (id, store_id, barcode, expected_qty, counted_qty, variance_qty, counted_by, counted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, countID, req.StoreID, entry.Barcode, expectedQty, entry.CountedQty, varianceQty, userID, now)
		if err != nil {
			logger.Error("Failed to insert stock count row: %v", err)
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to record stock count", nil)
			return
		}

		if varianceQty != 0 {
			discrepanciesCount++
			noteStr := "Cycle Count Reconciliation"
			_, _, err := h.engine.ApplyMovement(
				ctx,
				tx,
				req.StoreID,
				entry.Barcode,
				MovementAdjustment,
				varianceQty,
				RefStockCount,
				countID,
				&userID,
				&noteStr,
				true, // count reconciles as source of truth
			)
			if err != nil {
				logger.Error("Failed to apply movement for stock count variance: %v", err)
				errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to reconcile stock count variance", nil)
				return
			}
		}

		results = append(results, StockCountEntryResult{
			Barcode:     entry.Barcode,
			ExpectedQty: expectedQty,
			CountedQty:  entry.CountedQty,
			VarianceQty: varianceQty,
		})
	}

	if err := tx.Commit(); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to commit stock count submission", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StockCountResponse{
		TotalCounted:       len(req.Entries),
		DiscrepanciesFound: discrepanciesCount,
		Results:            results,
	})
}

func (h *InventoryHandler) GetShrinkageReportHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	storeID := r.URL.Query().Get("store_id")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	if storeID == "" || dateFrom == "" || dateTo == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id, date_from, and date_to are required", nil)
		return
	}

	report, overallPercent, err := h.repo.GetShrinkageReport(ctx, storeID, dateFrom, dateTo)
	if err != nil {
		logger.Error("Failed to fetch shrinkage report: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to generate shrinkage report", nil)
		return
	}

	if report == nil {
		report = []ShrinkageDaily{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"store_id":                 storeID,
		"date_from":                dateFrom,
		"date_to":                  dateTo,
		"overall_shrinkage_percent": overallPercent,
		"daily_records":            report,
	})
}
