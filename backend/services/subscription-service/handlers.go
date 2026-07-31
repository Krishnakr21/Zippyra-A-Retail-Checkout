package main

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type SubscriptionHandler struct {
	repo          Repository
	jwtSecret     string
	razorpayKeyID string
	webhookSecret string
}

func NewSubscriptionHandler(repo Repository, jwtSecret string) *SubscriptionHandler {
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
		}
	}
	keyID := os.Getenv("RAZORPAY_KEY_ID")
	if keyID == "" {
		keyID = "rzp_test_key_12345"
	}
	whSecret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
	return &SubscriptionHandler{
		repo:          repo,
		jwtSecret:     jwtSecret,
		razorpayKeyID: keyID,
		webhookSecret: whSecret,
	}
}

// GET /v1/subscription/plans?chain_id={id}
func (h *SubscriptionHandler) HandleGetPlans(w http.ResponseWriter, r *http.Request) {
	chainID := r.URL.Query().Get("chain_id")
	if chainID == "" {
		chainID = "chain-hq-001"
	}

	plans, err := h.repo.GetActivePlansByChainID(r.Context(), chainID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve subscription plans", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plans": plans,
	})
}

// POST /v1/subscription/subscribe
func (h *SubscriptionHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if req.PlanID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "plan_id is required", nil)
		return
	}

	plan, err := h.repo.GetPlanByID(r.Context(), req.PlanID)
	if err != nil || plan == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Subscription plan not found", nil)
		return
	}

	now := time.Now()
	var periodEnd time.Time
	if plan.BillingInterval == "ANNUAL" {
		periodEnd = now.AddDate(1, 0, 0)
	} else {
		periodEnd = now.AddDate(0, 1, 0)
	}

	rzpSubID := fmt.Sprintf("sub_rzp_mock_%s_%d", claims.UserID[:6], now.Unix())

	sub := &MemberSubscription{
		UserID:                 claims.UserID,
		PlanID:                 plan.ID,
		Status:                 "ACTIVE", // Auto-activate in test/checkout mock
		RazorpaySubscriptionID: &rzpSubID,
		CurrentPeriodEnd:       &periodEnd,
		CreatedAt:              now,
	}

	if err := h.repo.CreateSubscription(r.Context(), sub); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create subscription", nil)
		return
	}

	resp := SubscribeResponse{
		SubscriptionID:         sub.ID,
		PlanID:                 plan.ID,
		Status:                 sub.Status,
		RazorpaySubscriptionID: rzpSubID,
		RazorpayKeyID:          h.razorpayKeyID,
		CurrentPeriodEnd:       sub.CurrentPeriodEnd,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /v1/subscription/mine
func (h *SubscriptionHandler) HandleGetMine(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	sub, err := h.repo.GetActiveUserSubscription(r.Context(), claims.UserID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve subscription", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if sub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"subscription": nil,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscription": sub,
	})
}

// POST /v1/subscription/cancel
func (h *SubscriptionHandler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	if err := h.repo.CancelSubscriptionByUserID(r.Context(), claims.UserID); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to cancel subscription", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Subscription cancelled successfully",
	})
}

// GET /v1/subscription/internal/user-bonus?user_id={id}
func (h *SubscriptionHandler) HandleGetUserBonus(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "user_id is required", nil)
		return
	}

	sub, err := h.repo.GetActiveUserSubscription(r.Context(), userID)
	if err != nil || sub == nil || sub.Plan == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserBonusResponse{
			UserID:                 userID,
			HasActiveSubscription:  false,
			LoyaltyMultiplierBonus: 0.0,
		})
		return
	}

	var benefits BenefitsDTO
	_ = json.Unmarshal(sub.Plan.Benefits, &benefits)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserBonusResponse{
		UserID:                 userID,
		HasActiveSubscription:  true,
		LoyaltyMultiplierBonus: benefits.LoyaltyMultiplierBonus,
	})
}

func (h *SubscriptionHandler) extractAuthClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, stdErrors.New("missing authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		return nil, stdErrors.New("invalid token")
	}
	return claims, nil
}
