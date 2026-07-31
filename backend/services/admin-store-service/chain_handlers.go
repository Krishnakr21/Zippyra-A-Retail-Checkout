package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

// ChainHandler handles all /v1/admin-store/chains/* endpoints.
// Chains are owned exclusively by admin-store-service — no delegation to store-service needed.
type ChainHandler struct {
	repo      ChainRepository
	jwtSecret string
}

func NewChainHandler(repo ChainRepository, jwtSecret string) *ChainHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &ChainHandler{repo: repo, jwtSecret: jwtSecret}
}

func (h *ChainHandler) getClaims(r *http.Request) *jwt.Claims {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
		if err == nil && claims != nil {
			return claims
		}
	}
	return &jwt.Claims{
		UserID:   r.Header.Get("X-User-ID"),
		UserType: r.Header.Get("X-User-Type"),
		Role:     r.Header.Get("X-User-Role"),
	}
}

func (h *ChainHandler) requireAdmin(r *http.Request) *jwt.Claims {
	claims := h.getClaims(r)
	if claims == nil {
		return nil
	}
	if claims.UserType != "ADMIN" && claims.Role != "ADMIN" {
		return nil
	}
	return claims
}

// POST /v1/admin-store/chains
func (h *ChainHandler) HandleCreateChain(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}

	var req CreateChainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Chain name is required", nil)
		return
	}

	chain := &Chain{
		Name:               req.Name,
		LegalEntityName:    req.LegalEntityName,
		DefaultGstInPrefix: req.DefaultGstInPrefix,
		Status:             ChainStatusActive,
	}
	if err := h.repo.CreateChain(r.Context(), chain); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create chain", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(chain)
}

// GET /v1/admin-store/chains
func (h *ChainHandler) HandleListChains(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}

	statusFilter := r.URL.Query().Get("status")
	page := 1
	pageSize := 20
	if p := r.URL.Query().Get("page"); p != "" {
		_, _ = fmt.Sscanf(p, "%d", &page)
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		_, _ = fmt.Sscanf(ps, "%d", &pageSize)
	}

	chains, total, err := h.repo.ListChains(r.Context(), statusFilter, page, pageSize)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list chains", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"chains":    chains,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GET /v1/admin-store/chains/{id}
func (h *ChainHandler) HandleGetChain(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}

	chainID := mux.Vars(r)["id"]
	chain, err := h.repo.GetChainByID(r.Context(), chainID)
	if err != nil || chain == nil {
		errors.WriteError(w, http.StatusNotFound, CodeChainNotFound, "Chain not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(chain)
}

// PUT /v1/admin-store/chains/{id}
func (h *ChainHandler) HandleUpdateChain(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}

	chainID := mux.Vars(r)["id"]

	var req UpdateChainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	chain := &Chain{ID: chainID}
	if req.Name != nil {
		chain.Name = *req.Name
	}
	if req.LegalEntityName != nil {
		chain.LegalEntityName = *req.LegalEntityName
	}
	if req.DefaultGstInPrefix != nil {
		chain.DefaultGstInPrefix = *req.DefaultGstInPrefix
	}
	if req.Status != nil {
		chain.Status = *req.Status
	}

	if err := h.repo.UpdateChain(r.Context(), chain); err != nil {
		if err == ErrChainNotFound {
			errors.WriteError(w, http.StatusNotFound, CodeChainNotFound, "Chain not found", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update chain", nil)
		return
	}

	updated, _ := h.repo.GetChainByID(r.Context(), chainID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

// PUT /v1/admin-store/chains/{id}/status  (Step-Up required — enforced in routes.go)
func (h *ChainHandler) HandleUpdateChainStatus(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAdmin(r)
	if claims == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}

	// Step-up freshness check (same 10-minute window as store-service original)
	if claims.StepUpAt <= 0 || (time.Now().Unix()-claims.StepUpAt) > 600 {
		errors.WriteError(w, http.StatusForbidden, "STEP_UP_REQUIRED", "Fresh 2FA step-up required to change chain status", nil)
		return
	}

	chainID := mux.Vars(r)["id"]

	var req UpdateChainStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	req.Status = strings.ToUpper(req.Status)
	if req.Status != ChainStatusActive && req.Status != ChainStatusSuspended {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Status must be ACTIVE or SUSPENDED", nil)
		return
	}

	if err := h.repo.UpdateChainStatus(r.Context(), chainID, req.Status); err != nil {
		if err == ErrChainNotFound {
			errors.WriteError(w, http.StatusNotFound, CodeChainNotFound, "Chain not found", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update chain status", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":     chainID,
		"status": req.Status,
	})
}
