package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

func extractUserID(r *http.Request) string {
	if claims, ok := r.Context().Value("user_claims").(*jwt.Claims); ok && claims != nil && claims.UserID != "" {
		return claims.UserID
	}
	if claims, ok := r.Context().Value("user_claims").(*jwt.SessionClaims); ok && claims != nil && claims.UserID != "" {
		return claims.UserID
	}
	return ""
}

// GET /v1/loyalty/referral-code
func (h *CustomerHandler) HandleGetReferralCode(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	acc, err := h.repo.EnsureAccountExists(r.Context(), userID)
	if err != nil || acc == nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve loyalty account", nil)
		return
	}

	shareText := "Join me on Zippyra! Use my referral code " + acc.ReferralCode + " to get 50 bonus points on your first order. Download the app now!"

	resp := ReferralCodeResponse{
		ReferralCode:         acc.ReferralCode,
		ShareText:            shareText,
		ReferrerRewardPoints: 100,
		ReferredRewardPoints: 50,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /v1/loyalty/referral/apply
func (h *CustomerHandler) HandleApplyReferral(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req ApplyReferralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	req.ReferralCode = strings.ToUpper(strings.TrimSpace(req.ReferralCode))
	if req.ReferralCode == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "referral_code is required", nil)
		return
	}

	if err := h.repo.ApplyReferral(r.Context(), userID, req.ReferralCode); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "INVALID_REFERRAL_CODE") {
			errors.WriteError(w, http.StatusBadRequest, "INVALID_REFERRAL_CODE", "Invalid referral code", nil)
			return
		}
		if strings.Contains(errMsg, "CANNOT_REFER_SELF") {
			errors.WriteError(w, http.StatusBadRequest, "CANNOT_REFER_SELF", "You cannot use your own referral code", nil)
			return
		}
		if strings.Contains(errMsg, "REFERRAL_ALREADY_APPLIED") {
			errors.WriteError(w, http.StatusConflict, "REFERRAL_ALREADY_APPLIED", "Referral code already applied for this account", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to apply referral code", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Referral code applied successfully",
	})
}
