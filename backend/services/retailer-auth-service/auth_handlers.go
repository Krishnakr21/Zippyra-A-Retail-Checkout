package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type AuthHandler struct {
	repo       Repository
	otpSvc     *OTPService
	pinSvc     *PINService
	jwtSecret  string
}

func NewAuthHandler(repo Repository, otpSvc *OTPService, pinSvc *PINService, jwtSecret string) *AuthHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &AuthHandler{
		repo:      repo,
		otpSvc:    otpSvc,
		pinSvc:    pinSvc,
		jwtSecret: jwtSecret,
	}
}

func (h *AuthHandler) HandleSendOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req SendOtpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}

	if err := h.otpSvc.SendOTP(r.Context(), req.Phone, clientIP); err != nil {
		if err.Error() == CodeStaffNotRegistered {
			// Byte-identical 403 response for unregistered and deactivated phones
			errors.WriteError(w, http.StatusForbidden, CodeStaffNotRegistered, "Staff member not registered or inactive", nil)
			return
		}
		if err.Error() == "RATE_LIMIT_EXCEEDED" {
			errors.WriteError(w, http.StatusTooManyRequests, errors.CodeRateLimitExceeded, "OTP send rate limit exceeded. Please try again later.", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to send OTP", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "OTP_SENT",
		"message": "OTP sent successfully via SMS",
	})
}

func (h *AuthHandler) HandleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req VerifyOtpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	staff, err := h.otpSvc.VerifyOTP(r.Context(), req.Phone, req.OTP)
	if err != nil {
		if err.Error() == CodeStaffNotRegistered {
			errors.WriteError(w, http.StatusForbidden, CodeStaffNotRegistered, "Staff member not registered or inactive", nil)
			return
		}
		if err.Error() == "TOO_MANY_ATTEMPTS" {
			errors.WriteError(w, http.StatusTooManyRequests, errors.CodeRateLimitExceeded, "Too many failed OTP attempts", nil)
			return
		}
		errors.WriteError(w, http.StatusBadRequest, errors.CodeOTPInvalid, "Invalid OTP code", nil)
		return
	}

	// Create staff_sessions entry (auth_method='OTP')
	sess := &StaffSession{
		StaffID:    staff.ID,
		DeviceID:   req.DeviceID,
		AuthMethod: AuthMethodOTP,
	}
	if err := h.repo.CreateSession(r.Context(), sess); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create session", nil)
		return
	}

	// Issue JWT
	accessToken, refreshToken, err := h.issueTokens(staff, sess.ID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to issue token", nil)
		return
	}

	resp := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Staff: StaffProfile{
			ID:        staff.ID,
			Name:      staff.Name,
			Role:      staff.Role,
			StoreID:   staff.StoreID,
			StoreName: "Downtown Store",
			HasPinSet: staff.PinHash != nil,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) HandleSetPin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authorization required", nil)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil || claims.SessionID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid authorization token", nil)
		return
	}

	sess, err := h.repo.GetSessionByID(r.Context(), claims.SessionID)
	if err != nil || sess == nil || sess.RevokedAt != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Session revoked or invalid", nil)
		return
	}

	var req SetPinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if err := h.pinSvc.SetPin(r.Context(), claims.UserID, req.PIN, sess); err != nil {
		if err.Error() == CodeStepUpRequired {
			errors.WriteError(w, http.StatusForbidden, CodeStepUpRequired, "PIN setup requires a fresh OTP authentication session within the last 10 minutes", nil)
			return
		}
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "PIN_SET",
		"message": "PIN updated successfully",
	})
}

func (h *AuthHandler) HandlePinLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req PinLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	staff, err := h.pinSvc.VerifyPinLogin(r.Context(), req.Phone, req.PIN)
	if err != nil {
		if err.Error() == CodePinNotSet {
			errors.WriteError(w, http.StatusBadRequest, CodePinNotSet, "PIN not set for staff member. Please log in with OTP.", nil)
			return
		}
		if err.Error() == CodePinLocked {
			errors.WriteError(w, http.StatusTooManyRequests, CodePinLocked, "PIN locked due to 5 failed attempts. Retry after 15 minutes.", map[string]interface{}{"retry_after_seconds": 900})
			return
		}
		if err.Error() == CodePinInvalid {
			errors.WriteError(w, http.StatusBadRequest, CodePinInvalid, "Invalid PIN", nil)
			return
		}
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		return
	}

	// Create staff_sessions entry (auth_method='PIN')
	sess := &StaffSession{
		StaffID:    staff.ID,
		DeviceID:   req.DeviceID,
		AuthMethod: AuthMethodPIN,
	}
	if err := h.repo.CreateSession(r.Context(), sess); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create session", nil)
		return
	}

	accessToken, refreshToken, err := h.issueTokens(staff, sess.ID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to issue token", nil)
		return
	}

	resp := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Staff: StaffProfile{
			ID:        staff.ID,
			Name:      staff.Name,
			Role:      staff.Role,
			StoreID:   staff.StoreID,
			StoreName: "Downtown Store",
			HasPinSet: true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	claims, err := jwt.ParseAndVerifyToken(req.RefreshToken, h.jwtSecret)
	if err != nil || claims == nil || claims.SessionID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid refresh token", nil)
		return
	}

	sess, err := h.repo.GetSessionByID(r.Context(), claims.SessionID)
	if err != nil || sess == nil || sess.RevokedAt != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Session revoked", nil)
		return
	}

	if req.DeviceID != "" && sess.DeviceID != "" && req.DeviceID != sess.DeviceID {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Device ID mismatch", nil)
		return
	}

	staff, err := h.repo.GetStaffByID(r.Context(), claims.UserID)
	if err != nil || staff == nil || !staff.IsActive {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Staff inactive or not found", nil)
		return
	}

	accessToken, newRefreshToken, err := h.issueTokens(staff, sess.ID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to issue token", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
		if err == nil && claims != nil && claims.SessionID != "" {
			_ = h.repo.RevokeSession(r.Context(), claims.SessionID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "LOGGED_OUT"})
}

func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authorization required", nil)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid authorization token", nil)
		return
	}

	staff, err := h.repo.GetStaffByID(r.Context(), claims.UserID)
	if err != nil || staff == nil || !staff.IsActive {
		errors.WriteError(w, http.StatusNotFound, CodeStaffNotFound, "Staff member not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(staff)
}

func (h *AuthHandler) issueTokens(staff *StaffMember, sessionID string) (string, string, error) {
	claims := &jwt.Claims{
		UserID:    staff.ID,
		StoreID:   staff.StoreID,
		ChainID:   staff.ChainID,
		Role:      staff.Role,
		UserType:  "STAFF",
		SessionID: sessionID,
	}

	accessToken, err := jwt.GenerateToken(claims, h.jwtSecret, 8*time.Hour)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.GenerateToken(claims, h.jwtSecret, 14*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
