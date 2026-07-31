package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type ReviewHandler struct {
	repo Repository
}

func NewReviewHandler(repo Repository) *ReviewHandler {
	return &ReviewHandler{repo: repo}
}

func (h *ReviewHandler) verifySystemJWT(r *http.Request) bool {
	claims, ok := r.Context().Value("claims").(*jwt.Claims)
	if !ok || claims == nil {
		claims, ok = r.Context().Value("user_claims").(*jwt.Claims)
	}
	if !ok || claims == nil {
		return false
	}
	return claims.Role == "SYSTEM" || claims.UserType == "SYSTEM"
}

func (h *ReviewHandler) CreateReviewHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GRNID == "" || req.StoreID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "grn_id and store_id are required", nil)
		return
	}

	var snapshots []QCLineItemSnapshot
	for _, item := range req.LineItems {
		status := item.QCStatus
		if status == "" {
			status = QCStatusPending
		}
		snapshots = append(snapshots, QCLineItemSnapshot{
			GRNLineItemID: item.GRNLineItemID,
			Barcode:       item.Barcode,
			QtyReceived:   item.QtyReceived,
			QCStatus:      status,
			QCNote:        item.QCNote,
		})
	}

	// Compute initial overall status
	allComplete := len(snapshots) > 0
	for _, item := range snapshots {
		if item.QCStatus == QCStatusPending {
			allComplete = false
			break
		}
	}
	initialOverall := OverallStatusPending
	if allComplete {
		initialOverall = OverallStatusComplete
	}

	review := &QCReview{
		GRNID:         req.GRNID,
		StoreID:       req.StoreID,
		LineItems:     snapshots,
		OverallStatus: initialOverall,
	}

	if err := h.repo.CreateReview(r.Context(), review); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create QC review: "+err.Error(), nil)
		return
	}

	// Fetch existing or newly inserted review (for ON CONFLICT idempotency)
	existing, err := h.repo.GetReviewByGRNID(r.Context(), req.GRNID)
	if err != nil || existing == nil {
		existing = review
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(existing)
}

func (h *ReviewHandler) GetReviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	grnID := vars["grn_id"]

	review, err := h.repo.GetReviewByGRNID(r.Context(), grnID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to fetch QC review: "+err.Error(), nil)
		return
	}
	if review == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "QC review not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(review)
}

func (h *ReviewHandler) UpdateReviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	grnID := vars["grn_id"]

	var req UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid json body", nil)
		return
	}

	reviewerID := "SYSTEM"
	if claims, ok := r.Context().Value("claims").(*jwt.Claims); ok && claims != nil {
		reviewerID = claims.UserID
	}

	updated, err := h.repo.UpdateReviewLineItems(r.Context(), grnID, req.LineItemUpdates, reviewerID)
	if err != nil {
		if err.Error() == "review not found for grn_id "+grnID {
			errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "QC review not found", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update QC review: "+err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *ReviewHandler) IsCompleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	grnID := vars["grn_id"]

	isComplete, err := h.repo.IsReviewComplete(r.Context(), grnID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to check completion: "+err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ReviewCompletionResponse{
		GRNID:      grnID,
		IsComplete: isComplete,
	})
}
