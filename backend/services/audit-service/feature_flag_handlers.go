package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/audit"
	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/featureflags"
)

type CreateFlagRequest struct {
	FlagKey     string                 `json:"flag_key"`
	Description string                 `json:"description"`
	ScopeType   featureflags.ScopeType `json:"scope_type"`
}

type UpdateFlagRequest struct {
	EnabledGlobally *bool     `json:"enabled_globally,omitempty"`
	EnabledScopeIDs []string  `json:"enabled_scope_ids,omitempty"`
	UserPercentage  *int      `json:"user_percentage,omitempty"`
	Description     string    `json:"description,omitempty"`
	ScopeType       *featureflags.ScopeType `json:"scope_type,omitempty"`
}

var highRiskFlagPrefixes = []string{"payment.", "security.", "exit."}

func isHighRiskFlag(key string) bool {
	for _, prefix := range highRiskFlagPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// GET /v1/audit/feature-flags
func (h *AuditHandler) HandleListFeatureFlags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sharedErrors.WriteError(w, http.StatusMethodNotAllowed, sharedErrors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	_, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	flags, err := h.repo.ListFeatureFlags(r.Context())
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list feature flags", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"feature_flags": flags,
	})
}

// POST /v1/audit/feature-flags
func (h *AuditHandler) HandleCreateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sharedErrors.WriteError(w, http.StatusMethodNotAllowed, sharedErrors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req CreateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.FlagKey) == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid flag payload", nil)
		return
	}

	existing, _ := h.repo.GetFeatureFlag(r.Context(), req.FlagKey)
	if existing != nil {
		sharedErrors.WriteError(w, http.StatusConflict, "FLAG_ALREADY_EXISTS", "Feature flag already exists", nil)
		return
	}

	scopeType := req.ScopeType
	if scopeType == "" {
		scopeType = featureflags.ScopeGlobal
	}

	userID, _ := uuid.Parse(claims.UserID)
	flag := &featureflags.FeatureFlag{
		FlagKey:         req.FlagKey,
		Description:     req.Description,
		ScopeType:       scopeType,
		EnabledGlobally: false, // Disabled by default for safety
		EnabledScopeIDs: []string{},
		UpdatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = h.repo.SaveFeatureFlag(r.Context(), flag)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to create feature flag", nil)
		return
	}

	// Audit Logging
	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "audit.feature_flag_created",
			TargetType: "feature_flag",
			TargetID:   req.FlagKey,
			Payload:    map[string]interface{}{"scope_type": scopeType},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(flag)
}

// PUT /v1/audit/feature-flags/{key}
func (h *AuditHandler) HandleUpdateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sharedErrors.WriteError(w, http.StatusMethodNotAllowed, sharedErrors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/v1/audit/feature-flags/")
	if key == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Missing flag key", nil)
		return
	}

	// High-risk flag step-up check
	if isHighRiskFlag(key) {
		stepUpHeader := r.Header.Get("X-StepUp-Token")
		if stepUpHeader == "" {
			sharedErrors.WriteError(w, http.StatusForbidden, "STEP_UP_REQUIRED", "High-risk feature flag modification requires step-up authentication", nil)
			return
		}
	}

	flag, err := h.repo.GetFeatureFlag(r.Context(), key)
	if err != nil || flag == nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Feature flag not found", nil)
		return
	}

	var req UpdateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if req.EnabledGlobally != nil {
		flag.EnabledGlobally = *req.EnabledGlobally
	}
	if req.EnabledScopeIDs != nil {
		flag.EnabledScopeIDs = req.EnabledScopeIDs
	}
	if req.UserPercentage != nil {
		flag.UserPercentage = req.UserPercentage
	}
	if req.Description != "" {
		flag.Description = req.Description
	}
	if req.ScopeType != nil {
		flag.ScopeType = *req.ScopeType
	}

	userID, _ := uuid.Parse(claims.UserID)
	flag.UpdatedBy = userID
	flag.UpdatedAt = time.Now()

	err = h.repo.SaveFeatureFlag(r.Context(), flag)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to update feature flag", nil)
		return
	}

	// Audit Logging
	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "audit.feature_flag_updated",
			TargetType: "feature_flag",
			TargetID:   key,
			Payload: map[string]interface{}{
				"enabled_globally": flag.EnabledGlobally,
				"scope_type":       flag.ScopeType,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(flag)
}

// DELETE /v1/audit/feature-flags/{key}
func (h *AuditHandler) HandleDeleteFeatureFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sharedErrors.WriteError(w, http.StatusMethodNotAllowed, sharedErrors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/v1/audit/feature-flags/")
	if key == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Missing flag key", nil)
		return
	}

	flag, err := h.repo.GetFeatureFlag(r.Context(), key)
	if err != nil || flag == nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodeNotFound, "Feature flag not found", nil)
		return
	}

	err = h.repo.DeleteFeatureFlag(r.Context(), key)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to delete feature flag", nil)
		return
	}

	// Audit Logging
	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "audit.feature_flag_deleted",
			TargetType: "feature_flag",
			TargetID:   key,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "DELETED",
		"flag_key": key,
	})
}
