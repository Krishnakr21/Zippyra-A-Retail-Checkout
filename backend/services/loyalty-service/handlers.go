package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/loyalty"
)

type CustomerHandler struct {
	repo Repository
}

func NewCustomerHandler(repo Repository) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

func (h *CustomerHandler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := r.Context().Value("user_claims").(*jwt.SessionClaims)
	if !ok || claims == nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	acc, err := h.repo.EnsureAccountExists(ctx, claims.UserID)
	if err != nil {
		logger.Error("Failed to get loyalty account for user %s: %v", claims.UserID, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve loyalty balance", nil)
		return
	}

	tierConfigs, err := h.repo.GetTierConfigs(ctx)
	if err != nil {
		tierConfigs = nil
	}

	sharedTiers := make([]loyalty.TierConfig, len(tierConfigs))
	for i, tc := range tierConfigs {
		sharedTiers[i] = loyalty.TierConfig{
			Tier:              tc.Tier,
			MinLifetimePoints: tc.MinLifetimePoints,
			EarnMultiplier:    tc.EarnMultiplier,
			DisplayName:       tc.DisplayName,
			DisplayOrder:      tc.DisplayOrder,
		}
	}

	currentTierConfig := loyalty.CalculateTier(acc.LifetimePointsEarned, sharedTiers)
	nextTierConfig := loyalty.CalculateNextTier(acc.LifetimePointsEarned, sharedTiers)

	var pointsToNextTier *int64
	var nextTierName *string

	if nextTierConfig != nil {
		needed := nextTierConfig.MinLifetimePoints - acc.LifetimePointsEarned
		if needed < 0 {
			needed = 0
		}
		pointsToNextTier = &needed
		tName := nextTierConfig.DisplayName
		nextTierName = &tName
	}

	resp := LoyaltyBalanceResponse{
		PointsBalance:        acc.PointsBalance,
		PointsReserved:       acc.PointsReserved,
		Tier:                 currentTierConfig.Tier,
		TierDisplayName:      currentTierConfig.DisplayName,
		LifetimePointsEarned: acc.LifetimePointsEarned,
		PointsToNextTier:     pointsToNextTier,
		NextTierName:         nextTierName,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *CustomerHandler) GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := r.Context().Value("user_claims").(*jwt.SessionClaims)
	if !ok || claims == nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, err := h.repo.GetLedgerHistory(ctx, claims.UserID, page, pageSize)
	if err != nil {
		logger.Error("Failed to fetch ledger history for user %s: %v", claims.UserID, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve points history", nil)
		return
	}

	if items == nil {
		items = []LedgerItemResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LedgerHistoryResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *CustomerHandler) GetTiersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tiers, err := h.repo.GetTierConfigs(ctx)
	if err != nil {
		logger.Error("Failed to fetch tier configs: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve tier configurations", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tiers": tiers,
	})
}
