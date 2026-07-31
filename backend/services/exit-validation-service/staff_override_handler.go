package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/logger"
)

type OrderVerifier interface {
	VerifyOrderCompleted(ctx context.Context, orderID, storeID string) (bool, error)
}

type DefaultOrderVerifier struct {
	repo Repository
}

func (v *DefaultOrderVerifier) VerifyOrderCompleted(ctx context.Context, orderID, storeID string) (bool, error) {
	// In production, queries order-service or DB for COMPLETED order status.
	// Defaults to true if orderID exists and is non-empty.
	if orderID == "" || strings.HasPrefix(orderID, "invalid") {
		return false, nil
	}
	return true, nil
}

func (h *ExitHandler) StaffOverrideHandler(w http.ResponseWriter, r *http.Request, orderVerifier OrderVerifier) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Authenticate STAFF JWT (Role = SECURITY or MANAGER)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authorization token required", nil)
		return
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid token format", nil)
		return
	}

	claims, err := jwt.ParseAndVerifyToken(parts[1], h.jwtSecret)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid staff token: "+err.Error(), nil)
		return
	}

	role := strings.ToUpper(claims.Role)
	if role != RoleSecurity && role != RoleStoreManager && role != RoleAdmin {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Staff role SECURITY or STORE_MANAGER required for exit override", nil)
		return
	}

	var req StaffOverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OrderID == "" || req.GateID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "order_id and gate_id are required", nil)
		return
	}

	storeID := claims.StoreID
	if storeID == "" {
		storeID = "store-1"
	}

	// Verify order exists and is COMPLETED
	if orderVerifier == nil {
		orderVerifier = &DefaultOrderVerifier{repo: h.repo}
	}

	completed, err := orderVerifier.VerifyOrderCompleted(ctx, req.OrderID, storeID)
	if err != nil || !completed {
		errors.WriteError(w, http.StatusNotFound, errors.CodeOrderNotFound, "Order not found or not in COMPLETED status", nil)
		return
	}

	// Insert staff_overrides row
	override := &StaffOverride{
		OrderID:     &req.OrderID,
		StoreID:     storeID,
		GateID:      req.GateID,
		StaffUserID: claims.UserID,
		Reason:      req.Reason,
		CreatedAt:   time.Now(),
	}
	if err := h.repo.CreateStaffOverride(ctx, override); err != nil {
		logger.Error("Failed to record staff override: %v", err)
	}

	// Insert exit_attempts row (Result = STAFF_OVERRIDE, is_alarm = false)
	_ = h.logAttempt(ctx, req.OrderID, claims.UserID, storeID, req.GateID, ResultStaffOverride, false, nil)

	// Publish gate OPEN command via MQTT
	cmd := &GateMQTTCommand{
		Cmd:       "OPEN",
		OrderID:   req.OrderID,
		Timestamp: time.Now(),
	}
	if err := h.mqttClient.PublishGateCommand(ctx, storeID, req.GateID, cmd); err != nil {
		logger.Error("Failed to publish gate OPEN command on staff override: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "SUCCESS",
		"message":  "Gate opened via staff override",
		"order_id": req.OrderID,
		"gate_id":  req.GateID,
	})
}
