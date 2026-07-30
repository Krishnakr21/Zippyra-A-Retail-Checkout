package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type SessionManager struct {
	repo          Repository
	capacityMgr   CapacityManager
	jwtSecret     string
	kafkaProducer *kafka.Producer
}

func NewSessionManager(repo Repository, capacityMgr CapacityManager, jwtSecret string, producer *kafka.Producer) *SessionManager {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &SessionManager{
		repo:          repo,
		capacityMgr:   capacityMgr,
		jwtSecret:     jwtSecret,
		kafkaProducer: producer,
	}
}

func (s *SessionManager) BindStore(ctx context.Context, userID, clientIP string, req *StoreBindRequest) (*StoreBindResponse, error) {
	// a. Rate limit check
	if err := s.capacityMgr.CheckRateLimits(ctx, userID, clientIP); err != nil {
		return nil, err
	}

	// b. Look up qr_token
	tokenStr := req.QRToken
	tokenLogMask := tokenStr
	if len(tokenStr) > 6 {
		tokenLogMask = tokenStr[:6] + "..."
	}

	qrToken, err := s.repo.GetActiveQRToken(ctx, tokenStr)
	if err != nil || qrToken == nil || !qrToken.IsActive {
		return nil, errors.NewAPIError(errors.CodeQRTokenInvalid, "QR token is invalid or inactive", nil)
	}
	if time.Now().After(qrToken.ExpiresAt) {
		return nil, errors.NewAPIError(errors.CodeQRTokenExpired, "QR token has expired", nil)
	}

	// Fetch target store
	store, err := s.repo.GetStoreByID(ctx, qrToken.StoreID)
	if err != nil || store == nil {
		return nil, errors.NewAPIError(errors.CodeStoreNotFound, "Store not found for QR token", nil)
	}

	// c. Store status and opening hours check
	if store.Status != "ACTIVE" || !IsStoreOpenNow(store, time.Now()) {
		return nil, errors.NewAPIError(errors.CodeStoreClosed, "Store is closed or under maintenance", nil)
	}

	// d. Geofence check
	if !IsWithinGeofence(req.Lat, req.Lng, store) {
		return nil, errors.NewAPIError(errors.CodeStoreGeofenceMismatch, "You must be physically present inside the store to enter", nil)
	}

	// e. Auto-unbind existing active session at a DIFFERENT store if present
	existingSess, err := s.repo.GetActiveSessionByUser(ctx, userID)
	if err == nil && existingSess != nil {
		if existingSess.StoreID != store.ID {
			logger.Info("Auto-unbinding stale active session %s at store %s for user %s prior to binding to store %s",
				existingSess.ID, existingSess.StoreID, userID, store.ID)
			unbound, err := s.repo.UnbindUserActiveSession(ctx, userID)
			if err == nil && unbound != nil {
				_, _ = s.capacityMgr.DecrementCapacity(ctx, unbound.StoreID)
				s.publishSessionEnded(ctx, unbound.UserID, unbound.StoreID, unbound.ID, "previous_session_auto_cleanup")
			}
		} else {
			// User is re-binding to the SAME store (e.g. app restart) -> unbind existing and re-issue fresh
			unbound, err := s.repo.UnbindUserActiveSession(ctx, userID)
			if err == nil && unbound != nil {
				_, _ = s.capacityMgr.DecrementCapacity(ctx, unbound.StoreID)
			}
		}
	}

	// f. Capacity check & INCR with rollback safety
	ok, _, err := s.capacityMgr.TryIncrementCapacity(ctx, store.ID, store.CapacityMax)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeInternalError, "Failed to update store capacity", nil)
	}
	if !ok {
		return nil, errors.NewAPIError(errors.CodeStoreAtCapacity, "Store has reached maximum capacity, please try again shortly", nil)
	}

	// Guard for cleaning up INCR if DB insertion fails
	capacityIncremented := true
	defer func() {
		if !capacityIncremented {
			_, _ = s.capacityMgr.DecrementCapacity(ctx, store.ID)
		}
	}()

	// g. Insert store_sessions row
	sessID := uuid.New().String()
	now := time.Now()
	session := &StoreSession{
		ID:                   sessID,
		UserID:               userID,
		StoreID:              store.ID,
		DeviceID:             req.DeviceID,
		BoundAt:              now,
		CatalogVersionAtBind: store.CatalogVersion,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		capacityIncremented = false
		return nil, errors.NewAPIError(errors.CodeInternalError, "Failed to create store session", nil)
	}

	// h. Issue 4-hour Session JWT
	sessionTTL := 4 * time.Hour
	sessionToken, err := jwt.GenerateSessionToken(userID, store.ID, sessID, "CUSTOMER", s.jwtSecret, sessionTTL)
	if err != nil {
		capacityIncremented = false
		return nil, errors.NewAPIError(errors.CodeInternalError, "Failed to generate session token", nil)
	}

	expiresAtStr := now.Add(sessionTTL).Format(time.RFC3339)

	// i. Publish Kafka event `store.session_started`
	s.publishSessionStarted(ctx, userID, store.ID, sessID)

	logger.Info("User %s successfully bound to store %s (token: %s, session_id: %s)", userID, store.ID, tokenLogMask, sessID)

	return &StoreBindResponse{
		StoreID:          store.ID,
		StoreName:        store.Name,
		SessionToken:     sessionToken,
		SessionExpiresAt: expiresAtStr,
		CatalogVersion:   store.CatalogVersion,
		RFIDEnabled:      store.RFIDEnabled,
	}, nil
}

func (s *SessionManager) UnbindStore(ctx context.Context, userID, sessionID, reason string) error {
	var unbound *StoreSession
	var err error

	if sessionID != "" {
		unbound, err = s.repo.UnbindSession(ctx, sessionID)
	} else {
		unbound, err = s.repo.UnbindUserActiveSession(ctx, userID)
	}

	if err != nil || unbound == nil {
		// Idempotent success (already unbound)
		return nil
	}

	// Transitioned active session -> DECR capacity
	_, _ = s.capacityMgr.DecrementCapacity(ctx, unbound.StoreID)

	if reason == "" {
		reason = "customer_exit"
	}
	s.publishSessionEnded(ctx, unbound.UserID, unbound.StoreID, unbound.ID, reason)

	logger.Info("Unbound session %s for user %s at store %s (reason: %s)", unbound.ID, unbound.UserID, unbound.StoreID, reason)
	return nil
}

func (s *SessionManager) GetActiveSession(ctx context.Context, userID string) (*StoreSessionResponse, error) {
	sess, err := s.repo.GetActiveSessionByUser(ctx, userID)
	if err != nil || sess == nil {
		return nil, errors.NewAPIError(errors.CodeNoActiveSession, "No active store session found for user", nil)
	}

	store, err := s.repo.GetStoreByID(ctx, sess.StoreID)
	if err != nil || store == nil {
		return nil, errors.NewAPIError(errors.CodeStoreNotFound, "Associated store not found", nil)
	}

	sessionTTL := 4 * time.Hour
	sessionToken, err := jwt.GenerateSessionToken(userID, store.ID, sess.ID, "CUSTOMER", s.jwtSecret, sessionTTL)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeInternalError, "Failed to generate session token", nil)
	}

	expiresAtStr := sess.BoundAt.Add(sessionTTL).Format(time.RFC3339)

	return &StoreSessionResponse{
		StoreID:          store.ID,
		StoreName:        store.Name,
		SessionToken:     sessionToken,
		SessionExpiresAt: expiresAtStr,
		CatalogVersion:   store.CatalogVersion,
	}, nil
}

func (s *SessionManager) AutoExpireStaleSessionsJob(ctx context.Context) {
	staleDuration := 3 * time.Hour
	expired, err := s.repo.AutoExpireStaleSessions(ctx, staleDuration)
	if err != nil {
		logger.Error("AutoExpireStaleSessions error: %v", err)
		return
	}

	for _, sess := range expired {
		_, _ = s.capacityMgr.DecrementCapacity(ctx, sess.StoreID)
		s.publishSessionEnded(ctx, sess.UserID, sess.StoreID, sess.ID, "auto_expired")
		logger.Info("Auto-expired stale store session %s for user %s at store %s (bound_at: %v)",
			sess.ID, sess.UserID, sess.StoreID, sess.BoundAt)
	}
}

func (s *SessionManager) publishSessionStarted(ctx context.Context, userID, storeID, sessionID string) {
	if s.kafkaProducer == nil {
		return
	}
	payload := map[string]interface{}{
		"user_id":    userID,
		"store_id":   storeID,
		"session_id": sessionID,
		"ts":         time.Now().Unix(),
	}
	_ = s.kafkaProducer.PublishEvent(ctx, "store.session_started", sessionID, payload)
}

func (s *SessionManager) publishSessionEnded(ctx context.Context, userID, storeID, sessionID, reason string) {
	if s.kafkaProducer == nil {
		return
	}
	payload := map[string]interface{}{
		"user_id":    userID,
		"store_id":   storeID,
		"session_id": sessionID,
		"reason":     reason,
		"ts":         time.Now().Unix(),
	}
	_ = s.kafkaProducer.PublishEvent(ctx, "store.session_ended", sessionID, payload)
}
