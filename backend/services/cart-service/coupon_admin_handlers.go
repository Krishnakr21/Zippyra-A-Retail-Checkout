package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
)

type CouponAdminHandler struct {
	repo     OfferRepository
	compiler *CouponCompiler
}

func NewCouponAdminHandler(repo OfferRepository, compiler *CouponCompiler) *CouponAdminHandler {
	return &CouponAdminHandler{repo: repo, compiler: compiler}
}

func (h *CouponAdminHandler) validateCouponRules(discountType string, discountValue float64, maxUsesPerCust int, activeFrom *time.Time, activeUntil *time.Time) error {
	discTypeUpper := strings.ToUpper(discountType)
	if discTypeUpper != "PERCENT_OFF" && discTypeUpper != "FLAT_OFF" {
		return errors.NewAPIError(errors.CodeInvalidRequest, "discount_type must be PERCENT_OFF or FLAT_OFF", nil)
	}

	if discTypeUpper == "PERCENT_OFF" {
		if discountValue < 1 || discountValue > 90 {
			return errors.NewAPIError(errors.CodeInvalidRequest, "PERCENT_OFF discount_value must be between 1 and 90", nil)
		}
	} else if discTypeUpper == "FLAT_OFF" {
		if discountValue <= 0 {
			return errors.NewAPIError(errors.CodeInvalidRequest, "FLAT_OFF discount_value must be greater than 0", nil)
		}
	}

	if maxUsesPerCust < 1 {
		return errors.NewAPIError(errors.CodeInvalidRequest, "max_uses_per_customer must be at least 1", nil)
	}

	if activeFrom != nil && activeUntil != nil && activeUntil.Before(*activeFrom) {
		return errors.NewAPIError(errors.CodeInvalidRequest, "active_until must be after active_from", nil)
	}

	return nil
}

func (h *CouponAdminHandler) HandleCreateCoupon(w http.ResponseWriter, r *http.Request) {
	var req CreateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if req.ChainID == "" || req.Code == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "chain_id and code are required", nil)
		return
	}

	if err := h.validateCouponRules(req.DiscountType, req.DiscountValue, req.MaxUsesPerCustomer, req.ActiveFrom, req.ActiveUntil); err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			errors.WriteError(w, http.StatusBadRequest, apiErr.Code, apiErr.Message, nil)
		} else {
			errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		}
		return
	}

	now := time.Now().UTC()
	activeFrom := now
	if req.ActiveFrom != nil {
		activeFrom = *req.ActiveFrom
	}

	coupon := &Coupon{
		ID:                 uuid.New().String(),
		ChainID:            req.ChainID,
		StoreID:            req.StoreID,
		Code:               strings.ToUpper(strings.TrimSpace(req.Code)),
		DiscountType:       strings.ToUpper(req.DiscountType),
		DiscountValue:      req.DiscountValue,
		MinCartValuePaise:  req.MinCartValuePaise,
		MaxUses:            req.MaxUses,
		MaxUsesPerCustomer: req.MaxUsesPerCustomer,
		CurrentUseCount:    0,
		ActiveFrom:         activeFrom,
		ActiveUntil:        req.ActiveUntil,
		IsActive:           true,
		CreatedBy:          r.Header.Get("X-User-ID"),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := h.repo.CreateCoupon(r.Context(), coupon); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create coupon", nil)
		return
	}

	storeIDs, _ := h.repo.GetStoreIDsForChain(r.Context(), coupon.ChainID)
	_ = h.compiler.SyncCouponToRedis(r.Context(), coupon, storeIDs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(coupon)
}

func (h *CouponAdminHandler) HandleListCoupons(w http.ResponseWriter, r *http.Request) {
	chainID := r.URL.Query().Get("chain_id")
	if chainID == "" {
		chainID = "chain-default-001"
	}

	var storeIDPtr *string
	if storeID := r.URL.Query().Get("store_id"); storeID != "" {
		storeIDPtr = &storeID
	}
	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	coupons, err := h.repo.ListCoupons(r.Context(), chainID, storeIDPtr, includeInactive)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list coupons", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"coupons": coupons,
		"count":   len(coupons),
	})
}

func (h *CouponAdminHandler) HandleUpdateCoupon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	couponID := vars["id"]

	existing, err := h.repo.GetCouponByID(r.Context(), couponID)
	if err != nil || existing == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Coupon not found", nil)
		return
	}

	var req UpdateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	discType := existing.DiscountType
	if req.DiscountType != nil {
		discType = *req.DiscountType
	}
	discVal := existing.DiscountValue
	if req.DiscountValue != nil {
		discVal = *req.DiscountValue
	}
	maxCust := existing.MaxUsesPerCustomer
	if req.MaxUsesPerCustomer != nil {
		maxCust = *req.MaxUsesPerCustomer
	}

	if err := h.validateCouponRules(discType, discVal, maxCust, req.ActiveFrom, req.ActiveUntil); err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			errors.WriteError(w, http.StatusBadRequest, apiErr.Code, apiErr.Message, nil)
		} else {
			errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		}
		return
	}

	existing.DiscountType = strings.ToUpper(discType)
	existing.DiscountValue = discVal
	if req.MinCartValuePaise != nil {
		existing.MinCartValuePaise = *req.MinCartValuePaise
	}
	if req.MaxUses != nil {
		existing.MaxUses = req.MaxUses
	}
	existing.MaxUsesPerCustomer = maxCust
	if req.ActiveFrom != nil {
		existing.ActiveFrom = *req.ActiveFrom
	}
	if req.ActiveUntil != nil {
		existing.ActiveUntil = req.ActiveUntil
	}

	if err := h.repo.UpdateCoupon(r.Context(), existing); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update coupon", nil)
		return
	}

	storeIDs, _ := h.repo.GetStoreIDsForChain(r.Context(), existing.ChainID)
	_ = h.compiler.SyncCouponToRedis(r.Context(), existing, storeIDs)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(existing)
}

func (h *CouponAdminHandler) HandleDeleteCoupon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	couponID := vars["id"]

	existing, err := h.repo.GetCouponByID(r.Context(), couponID)
	if err != nil || existing == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Coupon not found", nil)
		return
	}

	if err := h.repo.SoftDeleteCoupon(r.Context(), couponID); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to delete coupon", nil)
		return
	}

	storeID := ""
	if existing.StoreID != nil {
		storeID = *existing.StoreID
	}
	storeIDs, _ := h.repo.GetStoreIDsForChain(r.Context(), existing.ChainID)
	_ = h.compiler.DeleteCouponFromRedis(r.Context(), storeID, existing.Code, storeIDs)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Coupon soft-deleted successfully",
	})
}

func (h *CouponAdminHandler) HandleToggleCoupon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	couponID := vars["id"]

	existing, err := h.repo.GetCouponByID(r.Context(), couponID)
	if err != nil || existing == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Coupon not found", nil)
		return
	}

	newStatus := !existing.IsActive
	if err := h.repo.ToggleCoupon(r.Context(), couponID, newStatus); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to toggle coupon status", nil)
		return
	}

	existing.IsActive = newStatus
	storeIDs, _ := h.repo.GetStoreIDsForChain(r.Context(), existing.ChainID)
	_ = h.compiler.SyncCouponToRedis(r.Context(), existing, storeIDs)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(existing)
}
