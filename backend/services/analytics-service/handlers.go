package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type AnalyticsHandler struct {
	repo Repository
}

func NewAnalyticsHandler(repo Repository) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo}
}

func (h *AnalyticsHandler) getClaims(r *http.Request) *jwt.Claims {
	if val := r.Context().Value("user_claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	if val := r.Context().Value("claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	return nil
}

func (h *AnalyticsHandler) HandleGetSales(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	storeID := r.URL.Query().Get("store_id")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")
	granularity := r.URL.Query().Get("granularity")

	sales, err := h.repo.GetSales(ctx, storeID, dateFrom, dateTo, granularity)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to query sales metrics", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"sales": sales})
}

func (h *AnalyticsHandler) HandleGetTopProducts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	storeID := r.URL.Query().Get("store_id")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	topProducts, err := h.repo.GetTopProducts(ctx, storeID, dateFrom, dateTo, limit)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to query top products", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"products": topProducts})
}

func (h *AnalyticsHandler) HandleGetFunnel(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	storeID := r.URL.Query().Get("store_id")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	funnel, err := h.repo.GetFunnel(ctx, storeID, dateFrom, dateTo)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to query funnel analytics", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"stages": funnel})
}

func (h *AnalyticsHandler) HandleGetPeakHours(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	storeID := r.URL.Query().Get("store_id")
	weeksLookback, _ := strconv.Atoi(r.URL.Query().Get("weeks_lookback"))
	throughput := GetDefaultThroughputPerHour()

	grid, err := h.repo.GetPeakHours(ctx, storeID, weeksLookback, throughput)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to query peak hours", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"grid":                grid,
		"throughput_per_hour": throughput,
	})
}

func (h *AnalyticsHandler) HandleGetChainSummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	chainID := r.URL.Query().Get("chain_id")
	if chainID == "" {
		claims := h.getClaims(r)
		if claims != nil {
			chainID = claims.ChainID
		}
	}

	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	summary, err := h.repo.GetChainSummary(ctx, chainID, dateFrom, dateTo)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to query chain summary", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
