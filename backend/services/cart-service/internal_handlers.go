package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type InternalCartHandler struct {
	checkoutRepo CheckoutSessionRepository
	cartStore    CartStore
	holdManager  HoldManager
	lockManager  LockManager
	jwtSecret    string
}

func NewInternalCartHandler(
	checkoutRepo CheckoutSessionRepository,
	cartStore CartStore,
	holdManager HoldManager,
	lockManager LockManager,
	jwtSecret string,
) *InternalCartHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &InternalCartHandler{
		checkoutRepo: checkoutRepo,
		cartStore:    cartStore,
		holdManager:  holdManager,
		lockManager:  lockManager,
		jwtSecret:    jwtSecret,
	}
}

// GET /v1/cart/internal/checkout-session/{id} (SYSTEM JWT)
func (h *InternalCartHandler) HandleGetCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	if !h.verifySystemAuth(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "System authentication required", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var sessionID string
	for i, p := range parts {
		if p == "checkout-session" && i+1 < len(parts) {
			sessionID = parts[i+1]
			break
		}
	}

	if sessionID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Missing session ID in path", nil)
		return
	}

	session, err := h.checkoutRepo.GetCheckoutSession(r.Context(), sessionID)
	if err != nil || session == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeCheckoutSessionExpired, "Checkout session not found", nil)
		return
	}

	if session.Status != "PENDING" || time.Now().After(session.ExpiresAt) {
		errors.WriteError(w, http.StatusNotFound, errors.CodeCheckoutSessionExpired, "Checkout session has expired or been consumed", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(session)
}

// POST /v1/cart/internal/checkout-session/{id}/consume (SYSTEM JWT)
func (h *InternalCartHandler) HandleConsumeCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	if !h.verifySystemAuth(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "System authentication required", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var sessionID string
	for i, p := range parts {
		if p == "checkout-session" && i+1 < len(parts) && parts[i+1] != "consume" {
			sessionID = parts[i+1]
			break
		}
	}

	if sessionID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Missing session ID in path", nil)
		return
	}

	consumedSession, err := h.checkoutRepo.ConsumeCheckoutSession(r.Context(), sessionID)
	if err != nil || consumedSession == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeCheckoutSessionExpired, "Checkout session cannot be consumed", nil)
		return
	}

	// Release lock & clear cart + holds
	_ = h.lockManager.ReleaseCheckoutLock(r.Context(), consumedSession.UserID)
	_ = h.holdManager.ReleaseAllUserHolds(r.Context(), consumedSession.StoreID, consumedSession.UserID, consumedSession.Items)
	_ = h.cartStore.ClearCart(r.Context(), consumedSession.StoreID, consumedSession.UserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(consumedSession)
}

func (h *InternalCartHandler) verifySystemAuth(r *http.Request) bool {
	if apiKey := r.Header.Get("X-Internal-API-Key"); apiKey != "" && apiKey == "zippyra-internal-secret-key-32bytes" {
		return true
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		return false
	}
	return claims.Role == "system" || claims.Role == "admin"
}
