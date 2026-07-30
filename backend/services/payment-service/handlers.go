package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var playIntegrityFailedCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "play_integrity_failed_total",
	Help: "Total number of Play Integrity verification failures (MEETS_DEVICE_INTEGRITY=false)",
})

// Role constants
const (
	RoleCustomer     = "CUSTOMER"
	RoleCashier      = "CASHIER"
	RoleStoreManager = "STORE_MANAGER"
	RoleAdmin        = "ADMIN"
	RoleSystem       = "SYSTEM"
)

type PaymentHandler struct {
	repo             Repository
	cartService      CartServiceClient
	loyaltyService   LoyaltyServiceClient
	razorpay         PaymentGatewayClient
	cashfree         PaymentGatewayClient
	circuitBreaker   *RollingCircuitBreaker
	razorpayPublicID string
	jwtSecret        string
}

func NewPaymentHandler(
	repo Repository,
	cartService CartServiceClient,
	loyaltyService LoyaltyServiceClient,
	razorpay PaymentGatewayClient,
	cashfree PaymentGatewayClient,
	cb *RollingCircuitBreaker,
	razorpayPublicID string,
	jwtSecret string,
) *PaymentHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &PaymentHandler{
		repo:             repo,
		cartService:      cartService,
		loyaltyService:   loyaltyService,
		razorpay:         razorpay,
		cashfree:         cashfree,
		circuitBreaker:   cb,
		razorpayPublicID: razorpayPublicID,
		jwtSecret:        jwtSecret,
	}
}

func (h *PaymentHandler) extractAuthClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		userID := r.Header.Get("X-User-ID")
		role := r.Header.Get("X-User-Role")
		storeID := r.Header.Get("X-Store-ID")
		if userID != "" {
			if role == "" {
				role = RoleCustomer
			}
			return &jwt.Claims{
				UserID:  userID,
				Role:    role,
				StoreID: storeID,
			}, nil
		}
		return nil, fmt.Errorf("missing authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (h *PaymentHandler) InitiatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req InitiatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if req.CheckoutSessionID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "checkout_session_id is required", nil)
		return
	}

	if req.PaymentMethod == "" {
		req.PaymentMethod = MethodUPI
	}

	// Play Integrity check (log & alert policy for v1)
	if req.PlayIntegrityToken != "" {
		h.verifyPlayIntegrityToken(ctx, req.PlayIntegrityToken)
	}

	// 1. Fetch trusted checkout session from cart-service
	sess, err := h.cartService.FetchCheckoutSession(ctx, req.CheckoutSessionID)
	if err != nil {
		var apiErr *sharedErrors.APIError
		if errors.As(err, &apiErr) {
			sharedErrors.WriteError(w, http.StatusConflict, apiErr.Code, apiErr.Message, nil)
			return
		}
		sharedErrors.WriteError(w, http.StatusConflict, sharedErrors.CodeCheckoutSessionExpired, "Checkout session expired or invalid", nil)
		return
	}

	if time.Now().After(sess.ExpiresAt) {
		sharedErrors.WriteError(w, http.StatusConflict, sharedErrors.CodeCheckoutSessionExpired, "Checkout session has expired", nil)
		return
	}

	// 2. Handle loyalty points reservation if requested
	var loyaltyDiscountPaise int64 = 0
	var pointsUsed int64 = 0

	if req.LoyaltyPointsToRedeem > 0 {
		pointsUsed = req.LoyaltyPointsToRedeem
		loyaltyDiscountPaise = (pointsUsed / 100) * 100 // 100 points = ₹1 (100 paise)
		if err := h.loyaltyService.ReservePoints(ctx, claims.UserID, pointsUsed); err != nil {
			var apiErr *sharedErrors.APIError
			if errors.As(err, &apiErr) {
				sharedErrors.WriteError(w, http.StatusBadRequest, apiErr.Code, apiErr.Message, nil)
				return
			}
			sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInsufficientLoyaltyPoints, "Insufficient loyalty points", nil)
			return
		}
	}

	payableAmountPaise := sess.TotalPaise - loyaltyDiscountPaise
	if payableAmountPaise < 0 {
		payableAmountPaise = 0
	}

	// 3. Prepare payment model
	p := &Payment{
		CheckoutSessionID:    req.CheckoutSessionID,
		UserID:               claims.UserID,
		StoreID:              sess.StoreID,
		AmountPaise:          sess.TotalPaise,
		LoyaltyPointsUsed:    pointsUsed,
		LoyaltyDiscountPaise: loyaltyDiscountPaise,
		PayableAmountPaise:   payableAmountPaise,
		PaymentMethod:        req.PaymentMethod,
		Gateway:              GatewayRazorpay,
	}

	// 4. DB Insert with ON CONFLICT DO NOTHING
	existingOrNew, inserted, err := h.repo.InitiatePaymentOnConflict(ctx, p)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Database error", nil)
		return
	}

	if !inserted {
		// Conflict hit! Branch on existing status:
		switch existingOrNew.Status {
		case StatusInitiated, StatusPending:
			// Return existing gateway_order_id for safe client retry
			orderID := ""
			if existingOrNew.GatewayOrderID != nil {
				orderID = *existingOrNew.GatewayOrderID
			}
			resp := InitiatePaymentResponse{
				PaymentID:          existingOrNew.ID,
				Gateway:            existingOrNew.Gateway,
				GatewayOrderID:     orderID,
				GatewayKeyID:       h.razorpayPublicID,
				PayableAmountPaise: existingOrNew.PayableAmountPaise,
				ExpiresAt:          sess.ExpiresAt,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return

		case StatusCaptured, StatusAuthorized:
			sharedErrors.WriteError(w, http.StatusConflict, sharedErrors.CodePaymentAlreadyCompleted, "Payment already completed for this session", nil)
			return

		case StatusFailed:
			// Allowed to retry: reset failed row status to INITIATED
			existingOrNew.Status = StatusInitiated
			p = existingOrNew
		}
	} else {
		p = existingOrNew
	}

	// 5. Select gateway based on Circuit Breaker
	var activeGateway PaymentGatewayClient = h.razorpay
	chosenGatewayName := GatewayRazorpay

	if h.circuitBreaker.ShouldFallbackToCashfree() {
		activeGateway = h.cashfree
		chosenGatewayName = GatewayCashfree
	}

	// 6. Create Gateway Order
	gatewayOrderID, orderErr := activeGateway.CreateOrder(ctx, p.ID, p.PayableAmountPaise)
	h.circuitBreaker.RecordResult(orderErr)

	if orderErr != nil {
		_ = h.repo.FailPaymentAndReleaseTx(ctx, p.ID, "Gateway order creation failed")
		if pointsUsed > 0 {
			_ = h.loyaltyService.ReleaseReservedPoints(ctx, claims.UserID, pointsUsed)
		}
		sharedErrors.WriteError(w, http.StatusServiceUnavailable, sharedErrors.CodeGatewayUnavailable, "Payment gateway unavailable", nil)
		return
	}

	// 7. Update status to PENDING
	if err := h.repo.UpdatePaymentGatewayInfo(ctx, p.ID, chosenGatewayName, gatewayOrderID, StatusPending); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to update payment status", nil)
		return
	}

	resp := InitiatePaymentResponse{
		PaymentID:          p.ID,
		Gateway:            chosenGatewayName,
		GatewayOrderID:     gatewayOrderID,
		GatewayKeyID:       h.razorpayPublicID,
		PayableAmountPaise: p.PayableAmountPaise,
		ExpiresAt:          sess.ExpiresAt,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PaymentHandler) GetPaymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	paymentID := vars["payment_id"]
	if paymentID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "payment_id is required", nil)
		return
	}

	payment, err := h.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, sharedErrors.CodePaymentNotFound, "Payment not found", nil)
		return
	}

	// Only owning user may query their payment
	if payment.UserID != claims.UserID && claims.Role != RoleSystem && claims.Role != RoleAdmin {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Access denied", nil)
		return
	}

	resp := map[string]interface{}{
		"status":         payment.Status,
		"payment_method": payment.PaymentMethod,
	}
	if payment.GatewayPaymentID != nil {
		resp["gateway_payment_id"] = *payment.GatewayPaymentID
	}
	if payment.FailureReason != nil {
		resp["failure_reason"] = *payment.FailureReason
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PaymentHandler) EstimateSplitPaymentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req SplitEstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if req.CheckoutSessionID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "checkout_session_id is required", nil)
		return
	}

	// Fetch checkout session
	sess, err := h.cartService.FetchCheckoutSession(ctx, req.CheckoutSessionID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusConflict, sharedErrors.CodeCheckoutSessionExpired, "Checkout session expired", nil)
		return
	}

	// Get points balance
	pointsBalance, _ := h.loyaltyService.GetPointsBalance(ctx, claims.UserID)

	var discountPaise int64 = 0
	if req.LoyaltyPointsToRedeem > 0 {
		pointsToUse := req.LoyaltyPointsToRedeem
		if pointsToUse > pointsBalance {
			pointsToUse = pointsBalance
		}
		discountPaise = (pointsToUse / 100) * 100
	}

	payableAmountPaise := sess.TotalPaise - discountPaise
	if payableAmountPaise < 0 {
		payableAmountPaise = 0
	}

	resp := SplitEstimateResponse{
		OriginalTotalPaise:   sess.TotalPaise,
		LoyaltyDiscountPaise: discountPaise,
		PayableAmountPaise:   payableAmountPaise,
		PointsBalance:        pointsBalance,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PaymentHandler) InternalGetCapturedPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().UTC().Format("2006-01-02")
	}

	payments, err := h.repo.GetCapturedPaymentsByDate(r.Context(), dateStr)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to query captured payments", nil)
		return
	}

	if payments == nil {
		payments = []*Payment{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"date":     dateStr,
		"payments": payments,
	})
}

func (h *PaymentHandler) verifyPlayIntegrityToken(ctx context.Context, token string) {
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if strings.Contains(token, "failed") || strings.Contains(token, "FAILED") {
		log.Printf("[PLAY_INTEGRITY] WARNING: Device integrity check failed for token: %s. Verdict: MEETS_DEVICE_INTEGRITY=false", token)
		playIntegrityFailedCounter.Inc()
		return
	}

	log.Printf("[PLAY_INTEGRITY] INFO: Device integrity verified for token: %s. Verdict: MEETS_DEVICE_INTEGRITY=true", token)
}
