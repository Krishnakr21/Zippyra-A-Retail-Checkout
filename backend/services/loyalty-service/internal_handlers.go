package main

import (
	"encoding/json"
	"net/http"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
)

type InternalHandler struct {
	repo Repository
}

func NewInternalHandler(repo Repository) *InternalHandler {
	return &InternalHandler{repo: repo}
}

func (h *InternalHandler) ReservePointsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req ReservePointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.UserID == "" || req.Points <= 0 || req.IdempotencyKey == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "user_id, positive points, and idempotency_key are required", nil)
		return
	}

	reserved, balanceAfter, err := h.repo.ReservePointsTx(ctx, req.UserID, req.Points, req.IdempotencyKey)
	if err != nil {
		if err.Error() == "INSUFFICIENT_LOYALTY_POINTS" {
			errors.WriteError(w, http.StatusBadRequest, errors.CodeInsufficientLoyaltyPoints, "Insufficient loyalty points balance", nil)
			return
		}
		logger.Error("Failed to reserve points for user %s: %v", req.UserID, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to reserve loyalty points", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ReservePointsResponse{
		Reserved:           reserved,
		PointsBalanceAfter: balanceAfter,
	})
}

func (h *InternalHandler) CommitPointsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CommitPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.UserID == "" || req.Points <= 0 || req.IdempotencyKey == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "user_id, positive points, and idempotency_key are required", nil)
		return
	}

	committed, err := h.repo.CommitPointsTx(ctx, req.UserID, req.Points, req.IdempotencyKey)
	if err != nil {
		logger.Error("Failed to commit points for user %s: %v", req.UserID, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to commit loyalty points", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CommitPointsResponse{
		Committed: committed,
	})
}

func (h *InternalHandler) ReleasePointsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req ReleasePointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	if req.UserID == "" || req.Points <= 0 || req.IdempotencyKey == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "user_id, positive points, and idempotency_key are required", nil)
		return
	}

	released, balanceAfter, err := h.repo.ReleasePointsTx(ctx, req.UserID, req.Points, req.IdempotencyKey)
	if err != nil {
		logger.Error("Failed to release points for user %s: %v", req.UserID, err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to release loyalty points", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ReleasePointsResponse{
		Released:           released,
		PointsBalanceAfter: balanceAfter,
	})
}
