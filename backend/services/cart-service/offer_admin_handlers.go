package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/validator"
)

type OfferAdminHandler struct {
	repo     OfferRepository
	compiler *OfferCompiler
	rdb      redis.Cmdable
}

func NewOfferAdminHandler(repo OfferRepository, compiler *OfferCompiler, rdb redis.Cmdable) *OfferAdminHandler {
	return &OfferAdminHandler{
		repo:     repo,
		compiler: compiler,
		rdb:      rdb,
	}
}

type UserClaims struct {
	UserID  string
	Role    string
	ChainID string
	StoreID string
}

func (h *OfferAdminHandler) extractClaims(r *http.Request) (*UserClaims, error) {
	// Standard JWT claims extraction mock/helper
	role := r.Header.Get("X-User-Role")
	if role == "" {
		role = "MANAGER"
	}
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "usr-admin-001"
	}
	chainID := r.Header.Get("X-Chain-ID")
	if chainID == "" {
		chainID = "chain-default-001"
	}
	storeID := r.Header.Get("X-Store-ID")

	return &UserClaims{
		UserID:  userID,
		Role:    role,
		ChainID: chainID,
		StoreID: storeID,
	}, nil
}

// 1. POST /v1/cart/admin/offers
func (h *OfferAdminHandler) CreateOfferHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req CreateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	// Default priority and active_from
	priority := 100
	if req.Priority != nil {
		priority = *req.Priority
	}

	activeFrom := time.Now().UTC()
	if req.ActiveFrom != nil {
		activeFrom = *req.ActiveFrom
	}

	// Scope & Role Checks
	isChainHQ := claims.Role == "OWNER" || claims.Role == "FINANCE" || claims.Role == "OPERATIONS" || claims.Role == "CHAIN_HQ"

	if req.StoreID == nil || *req.StoreID == "" {
		// Chain-wide offer creation attempt
		if claims.Role != "OWNER" {
			errors.WriteError(w, http.StatusForbidden, "CHAIN_WIDE_REQUIRES_OWNER", "Only OWNER role can create chain-wide offers", nil)
			return
		}
	} else {
		// Store-specific offer
		if !isChainHQ {
			if claims.StoreID == "" || *req.StoreID != claims.StoreID {
				errors.WriteError(w, http.StatusForbidden, "STORE_SCOPE_MISMATCH", "Cannot create offer for another store", nil)
				return
			}
		}
	}

	chainID := req.ChainID
	if chainID == "" {
		chainID = claims.ChainID
	}

	// Validate rule config using shared validator
	valRes := validator.ValidateOfferConfig(req.Type, req.AppliesTo, req.TargetIDs, req.RuleConfig, &activeFrom, req.ActiveUntil)
	if !valRes.IsValid {
		errors.WriteError(w, http.StatusBadRequest, valRes.ErrCode, valRes.ErrMsg, nil)
		return
	}

	offer := &Offer{
		ID:                 uuid.New().String(),
		ChainID:            chainID,
		StoreID:            req.StoreID,
		Type:               req.Type,
		AppliesTo:          req.AppliesTo,
		TargetIDs:          req.TargetIDs,
		RuleConfig:         req.RuleConfig,
		MinCartValuePaise: req.MinCartValuePaise,
		MaxDiscountPaise:  req.MaxDiscountPaise,
		Priority:           priority,
		ActiveFrom:         activeFrom,
		ActiveUntil:        req.ActiveUntil,
		IsActive:           true,
		CreatedBy:          claims.UserID,
	}

	if err := h.repo.CreateOffer(r.Context(), offer); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create offer", nil)
		return
	}

	// Trigger recompile for affected store(s)
	h.recompileAffected(r.Context(), offer)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"offer":    offer,
		"warnings": valRes.Warnings,
	})
}

// 2. GET /v1/cart/admin/offers
func (h *OfferAdminHandler) ListOffersHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	storeIDParam := r.URL.Query().Get("store_id")
	chainIDParam := r.URL.Query().Get("chain_id")
	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	chainID := chainIDParam
	if chainID == "" {
		chainID = claims.ChainID
	}

	var storeIDFilter *string
	if storeIDParam != "" {
		storeIDFilter = &storeIDParam
	} else if chainIDParam == "" && claims.StoreID != "" {
		storeIDFilter = &claims.StoreID
	}

	offers, err := h.repo.ListOffers(r.Context(), chainID, storeIDFilter, includeInactive)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list offers", nil)
		return
	}

	var result []*OfferResponse
	for _, o := range offers {
		scope := "STORE_SPECIFIC"
		if o.StoreID == nil || *o.StoreID == "" {
			scope = "CHAIN_WIDE"
		}
		result = append(result, &OfferResponse{
			Offer: *o,
			Scope: scope,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"offers": result})
}

// 3. GET /v1/cart/admin/offers/{id}
func (h *OfferAdminHandler) GetOfferHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	offer, err := h.repo.GetOffer(r.Context(), id)
	if err == ErrOfferNotFound {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Offer not found", nil)
		return
	} else if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to fetch offer", nil)
		return
	}

	scope := "STORE_SPECIFIC"
	if offer.StoreID == nil || *offer.StoreID == "" {
		scope = "CHAIN_WIDE"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&OfferResponse{Offer: *offer, Scope: scope})
}

// 4. PUT /v1/cart/admin/offers/{id}
func (h *OfferAdminHandler) UpdateOfferHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	existing, err := h.repo.GetOffer(r.Context(), id)
	if err == ErrOfferNotFound {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Offer not found", nil)
		return
	}

	// Permission Checks
	if existing.StoreID == nil || *existing.StoreID == "" {
		// Chain-wide offer edit attempt
		if claims.Role != "OWNER" {
			errors.WriteError(w, http.StatusForbidden, "CANNOT_EDIT_CHAIN_WIDE_OFFER", "Only OWNER role can edit chain-wide offers", nil)
			return
		}
	} else {
		if claims.Role != "OWNER" && claims.StoreID != "" && *existing.StoreID != claims.StoreID {
			errors.WriteError(w, http.StatusForbidden, "STORE_SCOPE_MISMATCH", "Cannot edit another store's offer", nil)
			return
		}
	}

	var req UpdateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	valRes := validator.ValidateOfferConfig(req.Type, req.AppliesTo, req.TargetIDs, req.RuleConfig, &req.ActiveFrom, req.ActiveUntil)
	if !valRes.IsValid {
		errors.WriteError(w, http.StatusBadRequest, valRes.ErrCode, valRes.ErrMsg, nil)
		return
	}

	existing.Type = req.Type
	existing.AppliesTo = req.AppliesTo
	existing.TargetIDs = req.TargetIDs
	existing.RuleConfig = req.RuleConfig
	existing.MinCartValuePaise = req.MinCartValuePaise
	existing.MaxDiscountPaise = req.MaxDiscountPaise
	existing.Priority = req.Priority
	existing.ActiveFrom = req.ActiveFrom
	existing.ActiveUntil = req.ActiveUntil
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.repo.UpdateOffer(r.Context(), existing); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update offer", nil)
		return
	}

	h.recompileAffected(r.Context(), existing)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(existing)
}

// 5. DELETE /v1/cart/admin/offers/{id}
func (h *OfferAdminHandler) DeleteOfferHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	existing, err := h.repo.GetOffer(r.Context(), id)
	if err == ErrOfferNotFound {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Offer not found", nil)
		return
	}

	if existing.StoreID == nil || *existing.StoreID == "" {
		if claims.Role != "OWNER" {
			errors.WriteError(w, http.StatusForbidden, "CANNOT_EDIT_CHAIN_WIDE_OFFER", "Only OWNER role can delete chain-wide offers", nil)
			return
		}
	} else if claims.Role != "OWNER" && claims.StoreID != "" && *existing.StoreID != claims.StoreID {
		errors.WriteError(w, http.StatusForbidden, "STORE_SCOPE_MISMATCH", "Cannot delete another store's offer", nil)
		return
	}

	if err := h.repo.DeleteOffer(r.Context(), id); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to delete offer", nil)
		return
	}

	h.recompileAffected(r.Context(), existing)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "deleted", "id": id})
}

// 6. POST /v1/cart/admin/offers/{id}/toggle
func (h *OfferAdminHandler) ToggleOfferHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	existing, err := h.repo.GetOffer(r.Context(), id)
	if err == ErrOfferNotFound {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Offer not found", nil)
		return
	}

	if existing.StoreID == nil || *existing.StoreID == "" {
		if claims.Role != "OWNER" {
			errors.WriteError(w, http.StatusForbidden, "CANNOT_EDIT_CHAIN_WIDE_OFFER", "Only OWNER role can toggle chain-wide offers", nil)
			return
		}
	} else if claims.Role != "OWNER" && claims.StoreID != "" && *existing.StoreID != claims.StoreID {
		errors.WriteError(w, http.StatusForbidden, "STORE_SCOPE_MISMATCH", "Cannot toggle another store's offer", nil)
		return
	}

	var req ToggleOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if err := h.repo.ToggleOffer(r.Context(), id, req.IsActive); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to toggle offer", nil)
		return
	}

	existing.IsActive = req.IsActive
	h.recompileAffected(r.Context(), existing)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "is_active": req.IsActive})
}

// 7. GET /v1/cart/admin/offers/{store_id}/preview
func (h *OfferAdminHandler) PreviewCompiledRulesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storeID := vars["store_id"]

	redisKey := fmt.Sprintf("offer_rules:%s", storeID)
	val := "[]"
	if h.rdb != nil {
		res, err := h.rdb.Get(r.Context(), redisKey).Result()
		if err == nil && strings.TrimSpace(res) != "" {
			val = res
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(val))
}

func (h *OfferAdminHandler) recompileAffected(ctx context.Context, offer *Offer) {
	if offer.StoreID != nil && *offer.StoreID != "" {
		_ = h.compiler.CompileAndPublish(ctx, *offer.StoreID)
		return
	}

	// Chain-wide offer -> recompile all stores in chain
	stores, err := h.repo.ListStoresForChain(ctx, offer.ChainID)
	if err != nil {
		return
	}
	for _, s := range stores {
		_ = h.compiler.CompileAndPublish(ctx, s)
	}
}
