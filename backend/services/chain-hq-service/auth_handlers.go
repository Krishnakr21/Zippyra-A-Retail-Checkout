package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/validator"
)

type AuthHandler struct {
	repo      Repository
	jwtSecret string
}

func NewAuthHandler(repo Repository) *AuthHandler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &AuthHandler{
		repo:      repo,
		jwtSecret: secret,
	}
}

func (h *AuthHandler) getClaims(r *http.Request) *jwt.Claims {
	if val := r.Context().Value("user_claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	if val := r.Context().Value("claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	return nil
}

func (h *AuthHandler) HandleSendOTP(w http.ResponseWriter, r *http.Request) {
	var req SendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !validator.ValidatePhone(req.Phone) {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Valid E.164 phone number required (+91...)", nil)
		return
	}

	user, err := h.repo.GetUserByPhone(r.Context(), req.Phone)
	if err != nil || !user.IsActive {
		// Deliberately uninformative 403 response matching retailer-auth-service
		errors.WriteError(w, http.StatusForbidden, CodeChainHQUserNotRegistered, "Phone number is not registered for Chain HQ access", nil)
		return
	}

	// Log mock OTP for dev
	logger.Info("[Chain HQ Auth] OTP sent to %s (Mock OTP: 123456)", logger.MaskPhone(user.Phone))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "otp_sent",
		"message": "OTP sent successfully",
	})
}

func (h *AuthHandler) HandleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !validator.ValidatePhone(req.Phone) || req.OTP == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Phone and OTP are required", nil)
		return
	}

	user, err := h.repo.GetUserByPhone(r.Context(), req.Phone)
	if err != nil || !user.IsActive {
		errors.WriteError(w, http.StatusForbidden, CodeChainHQUserNotRegistered, "Phone number is not registered for Chain HQ access", nil)
		return
	}

	// Verify OTP (mock accepts 123456)
	if strings.TrimSpace(req.OTP) != "123456" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeOTPInvalid, "Invalid OTP code", nil)
		return
	}

	sessionID := uuid.New().String()
	sess := &ChainHQSession{
		ID:            sessionID,
		ChainHQUserID: user.ID,
		DeviceID:      req.DeviceID,
		CreatedAt:     time.Now().UTC(),
	}
	_ = h.repo.CreateSession(r.Context(), sess)

	claims := &jwt.Claims{
		UserID:    user.ID,
		ChainID:   user.ChainID,
		Role:      user.Role,
		UserType:  "CHAIN_HQ",
		SessionID: sessionID,
	}

	// 8 hour access TTL, 30 days refresh TTL
	accessToken, _ := jwt.GenerateToken(claims, h.jwtSecret, 8*time.Hour)
	rawRefresh, _, _ := jwt.GenerateRefreshToken()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": rawRefresh,
		"user": map[string]interface{}{
			"id":       user.ID,
			"chain_id": user.ChainID,
			"phone":    logger.MaskPhone(user.Phone),
			"name":     user.Name,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "CHAIN_HQ" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Valid CHAIN_HQ session required", nil)
		return
	}

	newAccessToken, _ := jwt.GenerateToken(claims, h.jwtSecret, 8*time.Hour)
	rawRefresh, _, _ := jwt.GenerateRefreshToken()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  newAccessToken,
		"refresh_token": rawRefresh,
	})
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims != nil && claims.UserID != "" {
		_ = h.repo.RevokeUserSessions(r.Context(), claims.UserID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}

func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "User not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        user.ID,
		"chain_id":  user.ChainID,
		"phone":     logger.MaskPhone(user.Phone),
		"name":      user.Name,
		"role":      user.Role,
		"is_active": user.IsActive,
	})
}
