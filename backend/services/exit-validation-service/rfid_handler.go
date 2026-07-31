package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/logger"
)

func (h *ExitHandler) RFIDConfirmHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Authenticate DEVICE JWT
	deviceClaims, err := h.verifier.VerifyDeviceToken(r.Header.Get("Authorization"))
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid device credentials: "+err.Error(), nil)
		return
	}

	var req RFIDConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OrderID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "order_id is required", nil)
		return
	}

	// Check if order is awaiting RFID
	rfidKey := "exit_awaiting_rfid:" + req.OrderID
	gateVal, err := h.redis.Get(ctx, rfidKey).Result()
	if err != nil || gateVal == "" {
		// Not awaiting RFID! Log alarm & return 409
		_ = h.logAttempt(ctx, req.OrderID, "", deviceClaims.StoreID, deviceClaims.GateID, ResultNotAwaitingRFID, true, req.TagIDs)
		h.metrics.Inc(deviceClaims.StoreID, ResultNotAwaitingRFID)
		h.publishExitDenied(ctx, deviceClaims.StoreID, deviceClaims.GateID, ResultNotAwaitingRFID)

		errors.WriteError(w, http.StatusConflict, errors.CodeNotAwaitingRFID, "Order is not currently awaiting RFID confirmation", nil)
		return
	}

	// Delete awaiting RFID key
	_ = h.redis.Del(ctx, rfidKey).Err()

	if !req.DeactivationSuccess {
		// Deactivation failed or timed out -> RFID_TIMEOUT & Alarm
		_ = h.logAttempt(ctx, req.OrderID, "", deviceClaims.StoreID, deviceClaims.GateID, ResultRFIDTimeout, true, req.TagIDs)
		h.metrics.Inc(deviceClaims.StoreID, ResultRFIDTimeout)
		h.publishExitRFIDFailure(ctx, req.OrderID, deviceClaims.StoreID, deviceClaims.GateID, req.TagIDs)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ValidateExitResponse{
			Result:  "DENY",
			OrderID: req.OrderID,
			Reason:  ResultRFIDTimeout,
		})
		return
	}

	// Deactivation success -> RFID_CONFIRMED & OPEN gate!
	cmd := &GateMQTTCommand{
		Cmd:       "OPEN",
		OrderID:   req.OrderID,
		Timestamp: time.Now(),
	}
	if err := h.mqttClient.PublishGateCommand(ctx, deviceClaims.StoreID, deviceClaims.GateID, cmd); err != nil {
		logger.Error("Failed to publish gate OPEN command on RFID confirm: %v", err)
	}

	_ = h.logAttempt(ctx, req.OrderID, "", deviceClaims.StoreID, deviceClaims.GateID, ResultRFIDConfirmed, false, req.TagIDs)
	h.publishExitValidated(ctx, req.OrderID, "", "", deviceClaims.StoreID, deviceClaims.GateID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidateExitResponse{
		Result:  "OPEN",
		OrderID: req.OrderID,
	})
}

func (h *ExitHandler) publishExitRFIDFailure(ctx context.Context, orderID, storeID, gateID string, tagIDs []string) {
	if h.producer == nil {
		return
	}
	payload := ExitRFIDFailurePayload{
		OrderID:   orderID,
		StoreID:   storeID,
		GateID:    gateID,
		TagIDs:    tagIDs,
		Timestamp: time.Now(),
	}
	_ = h.producer.PublishEvent(ctx, TopicExitRFIDFailure, orderID, payload)
}
