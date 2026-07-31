package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type RedisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.BoolCmd
	Get(ctx context.Context, key string) *goredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *goredis.StatusCmd
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	Ping(ctx context.Context) *goredis.StatusCmd
}

type ExitHandler struct {
	repo       Repository
	redis      RedisClient
	producer   *kafka.Producer
	mqttClient MQTTClient
	verifier   *JWTVerifier
	metrics    *AlarmMetrics
	jwtSecret  string
}

func NewExitHandler(
	repo Repository,
	redis RedisClient,
	producer *kafka.Producer,
	mqttClient MQTTClient,
	verifier *JWTVerifier,
	metrics *AlarmMetrics,
	jwtSecret string,
) *ExitHandler {
	return &ExitHandler{
		repo:       repo,
		redis:      redis,
		producer:   producer,
		mqttClient: mqttClient,
		verifier:   verifier,
		metrics:    metrics,
		jwtSecret:  jwtSecret,
	}
}

func (h *ExitHandler) ValidateExitHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Authenticate DEVICE JWT
	deviceClaims, err := h.verifier.VerifyDeviceToken(r.Header.Get("Authorization"))
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid device credentials: "+err.Error(), nil)
		return
	}

	var req ValidateExitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ExitToken == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "exit_token is required", nil)
		return
	}

	// Step a: Parse & Verify Exit Token
	exitClaims, isExpired, err := h.verifier.VerifyExitToken(req.ExitToken)
	if err != nil {
		// Log attempt & alarm
		_ = h.logAttempt(ctx, "", "", deviceClaims.StoreID, deviceClaims.GateID, ResultInvalidToken, true, nil)
		h.metrics.Inc(deviceClaims.StoreID, ResultInvalidToken)
		h.publishExitDenied(ctx, deviceClaims.StoreID, deviceClaims.GateID, ResultInvalidToken)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ValidateExitResponse{Result: "DENY", Reason: ResultInvalidToken})
		return
	}

	// Step b: Check Token Expiration
	now := time.Now()
	expTime := exitClaims.ExpiresAt.Time
	if isExpired || now.After(expTime) {
		_ = h.logAttempt(ctx, exitClaims.OrderID, exitClaims.UserID, deviceClaims.StoreID, deviceClaims.GateID, ResultQRExpired, false, nil)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ValidateExitResponse{Result: "DENY", Reason: ResultQRExpired})
		return
	}

	// Step c: Store ID Mismatch Check
	if exitClaims.StoreID != deviceClaims.StoreID {
		_ = h.logAttempt(ctx, exitClaims.OrderID, exitClaims.UserID, deviceClaims.StoreID, deviceClaims.GateID, ResultWrongStore, true, nil)
		h.metrics.Inc(deviceClaims.StoreID, ResultWrongStore)
		h.publishExitDenied(ctx, deviceClaims.StoreID, deviceClaims.GateID, ResultWrongStore)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ValidateExitResponse{Result: "DENY", Reason: ResultWrongStore})
		return
	}

	// Step d: Redis SETNX exit_used:{order_id} (One-Time-Use Guard)
	remainingTTL := time.Until(expTime)
	if remainingTTL < time.Second {
		remainingTTL = 10 * time.Minute
	}

	usedKey := "exit_used:" + exitClaims.OrderID
	setOK, err := h.redis.SetNX(ctx, usedKey, "1", remainingTTL).Result()
	if err != nil || !setOK {
		// Key already existed -> QR_ALREADY_USED replay attack
		_ = h.logAttempt(ctx, exitClaims.OrderID, exitClaims.UserID, deviceClaims.StoreID, deviceClaims.GateID, ResultQRAlreadyUsed, true, nil)
		h.metrics.Inc(deviceClaims.StoreID, ResultQRAlreadyUsed)
		h.publishExitDenied(ctx, deviceClaims.StoreID, deviceClaims.GateID, ResultQRAlreadyUsed)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ValidateExitResponse{Result: "DENY", Reason: ResultQRAlreadyUsed})
		return
	}

	// Step e: Check RFID enabled flag
	rfidEnabled := h.isRFIDEnabledForStore(ctx, deviceClaims.StoreID)
	if rfidEnabled {
		// SET exit_awaiting_rfid:{order_id} (TTL 60s)
		rfidKey := "exit_awaiting_rfid:" + exitClaims.OrderID
		_ = h.redis.Set(ctx, rfidKey, deviceClaims.GateID, 60*time.Second).Err()

		_ = h.logAttempt(ctx, exitClaims.OrderID, exitClaims.UserID, deviceClaims.StoreID, deviceClaims.GateID, ResultAwaitingRFID, false, nil)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ValidateExitResponse{
			Result:  ResultAwaitingRFID,
			OrderID: exitClaims.OrderID,
		})
		return
	}

	// Step f: Publish gate OPEN command via MQTT
	cmd := &GateMQTTCommand{
		Cmd:       "OPEN",
		OrderID:   exitClaims.OrderID,
		Timestamp: time.Now(),
	}
	if err := h.mqttClient.PublishGateCommand(ctx, deviceClaims.StoreID, deviceClaims.GateID, cmd); err != nil {
		logger.Error("Failed to publish gate OPEN command to MQTT: %v", err)
	}

	// Step g: Log attempt OPENED & Publish Kafka exit.validated
	_ = h.logAttempt(ctx, exitClaims.OrderID, exitClaims.UserID, deviceClaims.StoreID, deviceClaims.GateID, ResultOpened, false, nil)
	h.publishExitValidated(ctx, exitClaims.OrderID, exitClaims.SessionID, exitClaims.UserID, deviceClaims.StoreID, deviceClaims.GateID)

	// Step h: Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidateExitResponse{
		Result:  "OPEN",
		OrderID: exitClaims.OrderID,
	})
}

func (h *ExitHandler) GetExitStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Authenticate Customer JWT
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
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid token: "+err.Error(), nil)
		return
	}

	vars := mux.Vars(r)
	orderID := vars["order_id"]
	if orderID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "order_id parameter required", nil)
		return
	}

	attempt, err := h.repo.GetLatestExitAttemptByOrderID(ctx, orderID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to query exit attempt", nil)
		return
	}

	if attempt == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ExitStatusResponse{Result: "PENDING"})
		return
	}

	// Verify order owning user matches requesting user
	if attempt.UserID != "" && attempt.UserID != claims.UserID {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Access denied to order exit status", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ExitStatusResponse{
		Result: attempt.Result,
		GateID: attempt.GateID,
	})
}

func (h *ExitHandler) logAttempt(ctx context.Context, orderID, userID, storeID, gateID, result string, isAlarm bool, tagIDs []string) error {
	var jsonTags json.RawMessage
	if len(tagIDs) > 0 {
		bytes, _ := json.Marshal(tagIDs)
		jsonTags = json.RawMessage(bytes)
	}

	attempt := &ExitAttempt{
		OrderID:    orderID,
		UserID:     userID,
		StoreID:    storeID,
		GateID:     gateID,
		Result:     result,
		IsAlarm:    isAlarm,
		RFIDTagIDs: jsonTags,
		CreatedAt:  time.Now(),
	}
	return h.repo.CreateExitAttempt(ctx, attempt)
}

func (h *ExitHandler) isRFIDEnabledForStore(ctx context.Context, storeID string) bool {
	// Read from Redis cache (5m TTL)
	key := "store_rfid_enabled:" + storeID
	val, err := h.redis.Get(ctx, key).Result()
	if err == nil && val != "" {
		return val == "true"
	}

	// Pilot stores default to false (QR_ONLY mode)
	_ = h.redis.Set(ctx, key, "false", 5*time.Minute).Err()
	return false
}

func (h *ExitHandler) publishExitValidated(ctx context.Context, orderID, sessionID, userID, storeID, gateID string) {
	if h.producer == nil {
		return
	}
	payload := ExitValidatedPayload{
		OrderID:   orderID,
		SessionID: sessionID,
		UserID:    userID,
		StoreID:   storeID,
		GateID:    gateID,
		Timestamp: time.Now(),
	}
	_ = h.producer.PublishEvent(ctx, TopicExitValidated, orderID, payload)
}

func (h *ExitHandler) publishExitDenied(ctx context.Context, storeID, gateID, reason string) {
	if h.producer == nil {
		return
	}
	payload := ExitDeniedPayload{
		StoreID:   storeID,
		GateID:    gateID,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	_ = h.producer.PublishEvent(ctx, TopicExitDenied, storeID, payload)
}
