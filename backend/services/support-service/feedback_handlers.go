package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

// POST /v1/support/feedback (Any authenticated JWT)
func (h *TicketHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value("user_claims").(*jwt.Claims)
	userID := "anonymous"
	userType := "CUSTOMER"
	if claims != nil && claims.UserID != "" {
		userID = claims.UserID
		if claims.Role != "" {
			userType = claims.Role
		}
	}

	var req CreateFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	sourceApp := strings.TrimSpace(req.SourceApp)
	if sourceApp == "" {
		sourceApp = "CUSTOMER_APP"
	}

	contextStr := strings.TrimSpace(req.Context)
	if contextStr == "" {
		contextStr = "general"
	}

	fb := &FeedbackSubmission{
		UserID:    userID,
		UserType:  userType,
		SourceApp: sourceApp,
		NPSScore:  req.NPSScore,
		Comment:   req.Comment,
		Context:   contextStr,
	}

	if err := h.repo.CreateFeedback(r.Context(), fb); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to submit feedback", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      fb.ID,
		"message": "Feedback submitted successfully",
	})
}

// GET /v1/support/feedback?source_app={f}&min_score={f}&page={p}&limit={l} (ADMIN/SYSTEM)
func (h *TicketHandler) ListFeedback(w http.ResponseWriter, r *http.Request) {
	sourceApp := r.URL.Query().Get("source_app")
	minScoreStr := r.URL.Query().Get("min_score")
	var minScore *int
	if minScoreStr != "" {
		if val, err := strconv.Atoi(minScoreStr); err == nil {
			minScore = &val
		}
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	list, total, err := h.repo.ListFeedback(r.Context(), sourceApp, minScore, page, limit)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve feedback list", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedbacks": list,
		"total":     total,
	})
}
