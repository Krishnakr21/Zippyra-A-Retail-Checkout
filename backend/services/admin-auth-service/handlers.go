package main

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/audit"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

var digitRegex = regexp.MustCompile(`^\d+$`)

func isDigitOnly(s string, length int) bool {
	s = strings.TrimSpace(s)
	if length > 0 && len(s) != length {
		return false
	}
	return digitRegex.MatchString(s)
}

type AdminAuthHandler struct {
	repo           Repository
	googleVal      GoogleTokenValidator
	auditPub       *audit.Publisher
	jwtSecret      string
	allowedDomain string
}

func NewAdminAuthHandler(repo Repository, googleVal GoogleTokenValidator, auditPub *audit.Publisher) *AdminAuthHandler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	domain := os.Getenv("ALLOWED_ADMIN_EMAIL_DOMAIN")
	if domain == "" {
		domain = "zippyra.com"
	}
	domain = strings.TrimPrefix(domain, "@")

	return &AdminAuthHandler{
		repo:           repo,
		googleVal:      googleVal,
		auditPub:       auditPub,
		jwtSecret:      secret,
		allowedDomain: domain,
	}
}

func (h *AdminAuthHandler) getClaims(r *http.Request) *jwt.Claims {
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

func (h *AdminAuthHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "id_token is required", nil)
		return
	}

	payload, err := h.googleVal.Validate(r.Context(), req.IDToken)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, err.Error(), nil)
		return
	}

	// Verify hosted domain
	if !strings.EqualFold(payload.HD, h.allowedDomain) {
		errors.WriteError(w, http.StatusForbidden, CodeDomainNotAllowed, "Email domain is not authorized for admin access", nil)
		return
	}

	admin, err := h.repo.GetAdminByEmail(r.Context(), payload.Email)
	if err != nil || !admin.IsActive {
		errors.WriteError(w, http.StatusForbidden, CodeAdminNotRegistered, "Account is not registered or is inactive", nil)
		return
	}

	// Link google_sub on first login
	if admin.GoogleSub == nil || *admin.GoogleSub == "" {
		sub := payload.Sub
		admin.GoogleSub = &sub
		_ = h.repo.UpdateAdmin(r.Context(), admin)
	}

	// Branch on 2FA state
	if admin.TOTPEnabledAt == nil {
		setupClaims := &jwt.Claims{
			AdminID:  admin.ID,
			Email:    admin.Email,
			Role:     admin.Role,
			UserType: "ADMIN_2FA_SETUP",
		}
		setupToken, err := jwt.GenerateToken(setupClaims, h.jwtSecret, 10*time.Minute)
		if err != nil {
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to generate setup token", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"setup_required": true,
			"setup_token":    setupToken,
		})
		return
	}

	verifyClaims := &jwt.Claims{
		AdminID:  admin.ID,
		Email:    admin.Email,
		Role:     admin.Role,
		UserType: "ADMIN_2FA_VERIFY",
	}
	verifyToken, err := jwt.GenerateToken(verifyClaims, h.jwtSecret, 5*time.Minute)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to generate verify token", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"totp_required": true,
		"verify_token":  verifyToken,
	})
}

func (h *AdminAuthHandler) HandleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN_2FA_SETUP" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "ADMIN_2FA_SETUP token required", nil)
		return
	}

	admin, err := h.repo.GetAdminByID(r.Context(), claims.AdminID)
	if err != nil || !admin.IsActive {
		errors.WriteError(w, http.StatusForbidden, CodeAdminNotRegistered, "Admin user not found", nil)
		return
	}

	secret, otpauthURI, err := GenerateTOTPKey(admin.Email)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to generate TOTP secret", nil)
		return
	}

	encrypted, err := EncryptTOTPSecret(secret)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to encrypt TOTP secret", nil)
		return
	}

	admin.TOTPSecretEncrypted = encrypted
	if err := h.repo.UpdateAdmin(r.Context(), admin); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to save TOTP secret", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"otpauth_uri":    otpauthURI,
		"secret_display": secret,
	})
}

func (h *AdminAuthHandler) HandleTOTPConfirm(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN_2FA_SETUP" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "ADMIN_2FA_SETUP token required", nil)
		return
	}

	var req TOTPConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !isDigitOnly(req.Code, 6) {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "6-digit code is required", nil)
		return
	}

	admin, err := h.repo.GetAdminByID(r.Context(), claims.AdminID)
	if err != nil || !admin.IsActive {
		errors.WriteError(w, http.StatusForbidden, CodeAdminNotRegistered, "Admin user not found", nil)
		return
	}

	if len(admin.TOTPSecretEncrypted) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "TOTP setup not initiated", nil)
		return
	}

	secret, err := DecryptTOTPSecret(admin.TOTPSecretEncrypted)
	if err != nil || !ValidateTOTPCode(req.Code, secret) {
		errors.WriteError(w, http.StatusBadRequest, CodeTOTPInvalid, "Invalid TOTP code", nil)
		return
	}

	now := time.Now().UTC()
	admin.TOTPEnabledAt = &now
	admin.TOTPFailedAttempts = 0
	admin.TOTPLockedUntil = nil
	_ = h.repo.UpdateAdmin(r.Context(), admin)

	sessionID := uuid.New().String()
	sess := &AdminSession{
		ID:        sessionID,
		AdminID:   admin.ID,
		CreatedAt: now,
	}
	_ = h.repo.CreateSession(r.Context(), sess)

	fullClaims := &jwt.Claims{
		AdminID:   admin.ID,
		Email:     admin.Email,
		Role:      admin.Role,
		UserType:  "ADMIN",
		SessionID: sessionID,
		StepUpAt:  now.Unix(),
	}
	accessToken, _ := jwt.GenerateToken(fullClaims, h.jwtSecret, 1*time.Hour)
	rawRefresh, hashRefresh, _ := jwt.GenerateRefreshToken()
	_ = hashRefresh

	if h.auditPub != nil {
		h.auditPub.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       admin.ID,
			ActorName:     admin.Name,
			ActionType:    "admin.totp_enabled",
			TargetType:    "admin_user",
			TargetID:      admin.ID,
			SourceService: "admin-auth-service",
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       map[string]interface{}{"email": admin.Email, "role": admin.Role},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": rawRefresh,
		"admin": map[string]interface{}{
			"id":         admin.ID,
			"email":      admin.Email,
			"name":       admin.Name,
			"role":       admin.Role,
			"step_up_at": now.Unix(),
		},
	})
}

func (h *AdminAuthHandler) HandleTOTPVerify(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN_2FA_VERIFY" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "ADMIN_2FA_VERIFY token required", nil)
		return
	}

	var req TOTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !isDigitOnly(req.Code, 6) {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "6-digit code is required", nil)
		return
	}

	admin, err := h.repo.GetAdminByID(r.Context(), claims.AdminID)
	if err != nil || !admin.IsActive {
		errors.WriteError(w, http.StatusForbidden, CodeAdminNotRegistered, "Admin user not found", nil)
		return
	}

	now := time.Now().UTC()
	if admin.TOTPLockedUntil != nil && admin.TOTPLockedUntil.After(now) {
		w.Header().Set("Retry-After", "900")
		errors.WriteError(w, http.StatusTooManyRequests, CodeTOTPLocked, "Account temporarily locked due to failed 2FA attempts", nil)
		return
	}

	secret, err := DecryptTOTPSecret(admin.TOTPSecretEncrypted)
	if err != nil || !ValidateTOTPCode(req.Code, secret) {
		admin.TOTPFailedAttempts++
		if admin.TOTPFailedAttempts >= 5 {
			lockedUntil := now.Add(15 * time.Minute)
			admin.TOTPLockedUntil = &lockedUntil
			admin.TOTPFailedAttempts = 0
			_ = h.repo.UpdateAdmin(r.Context(), admin)
			w.Header().Set("Retry-After", "900")
			errors.WriteError(w, http.StatusTooManyRequests, CodeTOTPLocked, "Account locked for 15 minutes due to 5 consecutive 2FA failures", nil)
			return
		}
		_ = h.repo.UpdateAdmin(r.Context(), admin)
		errors.WriteError(w, http.StatusBadRequest, CodeTOTPInvalid, "Invalid TOTP code", nil)
		return
	}

	admin.TOTPFailedAttempts = 0
	admin.TOTPLockedUntil = nil
	_ = h.repo.UpdateAdmin(r.Context(), admin)

	sessionID := uuid.New().String()
	sess := &AdminSession{
		ID:        sessionID,
		AdminID:   admin.ID,
		CreatedAt: now,
	}
	_ = h.repo.CreateSession(r.Context(), sess)

	fullClaims := &jwt.Claims{
		AdminID:   admin.ID,
		Email:     admin.Email,
		Role:      admin.Role,
		UserType:  "ADMIN",
		SessionID: sessionID,
		StepUpAt:  now.Unix(),
	}
	accessToken, _ := jwt.GenerateToken(fullClaims, h.jwtSecret, 1*time.Hour)
	rawRefresh, _, _ := jwt.GenerateRefreshToken()

	if h.auditPub != nil {
		h.auditPub.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       admin.ID,
			ActorName:     admin.Name,
			ActionType:    "admin.login_succeeded",
			TargetType:    "admin_user",
			TargetID:      admin.ID,
			SourceService: "admin-auth-service",
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       map[string]interface{}{"email": admin.Email, "role": admin.Role},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": rawRefresh,
		"admin": map[string]interface{}{
			"id":         admin.ID,
			"email":      admin.Email,
			"name":       admin.Name,
			"role":       admin.Role,
			"step_up_at": now.Unix(),
		},
	})
}

func (h *AdminAuthHandler) HandleStepUp(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "ADMIN JWT required for step-up", nil)
		return
	}

	var req StepUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !isDigitOnly(req.Code, 6) {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "6-digit code is required", nil)
		return
	}

	admin, err := h.repo.GetAdminByID(r.Context(), claims.AdminID)
	if err != nil || !admin.IsActive {
		errors.WriteError(w, http.StatusForbidden, CodeAdminNotRegistered, "Admin user not found", nil)
		return
	}

	now := time.Now().UTC()
	if admin.TOTPLockedUntil != nil && admin.TOTPLockedUntil.After(now) {
		w.Header().Set("Retry-After", "900")
		errors.WriteError(w, http.StatusTooManyRequests, CodeTOTPLocked, "Account locked due to 2FA failures", nil)
		return
	}

	secret, err := DecryptTOTPSecret(admin.TOTPSecretEncrypted)
	if err != nil || !ValidateTOTPCode(req.Code, secret) {
		admin.TOTPFailedAttempts++
		if admin.TOTPFailedAttempts >= 5 {
			lockedUntil := now.Add(15 * time.Minute)
			admin.TOTPLockedUntil = &lockedUntil
			admin.TOTPFailedAttempts = 0
			_ = h.repo.UpdateAdmin(r.Context(), admin)
			w.Header().Set("Retry-After", "900")
			errors.WriteError(w, http.StatusTooManyRequests, CodeTOTPLocked, "Account locked for 15 minutes due to 5 consecutive 2FA failures", nil)
			return
		}
		_ = h.repo.UpdateAdmin(r.Context(), admin)
		errors.WriteError(w, http.StatusBadRequest, CodeTOTPInvalid, "Invalid TOTP code", nil)
		return
	}

	admin.TOTPFailedAttempts = 0
	admin.TOTPLockedUntil = nil
	_ = h.repo.UpdateAdmin(r.Context(), admin)

	freshClaims := &jwt.Claims{
		AdminID:   admin.ID,
		Email:     admin.Email,
		Role:      admin.Role,
		UserType:  "ADMIN",
		SessionID: claims.SessionID,
		StepUpAt:  now.Unix(),
	}
	accessToken, _ := jwt.GenerateToken(freshClaims, h.jwtSecret, 1*time.Hour)

	if h.auditPub != nil {
		h.auditPub.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       admin.ID,
			ActorName:     admin.Name,
			ActionType:    "admin.step_up_used",
			TargetType:    "admin_user",
			TargetID:      admin.ID,
			SourceService: "admin-auth-service",
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       map[string]interface{}{"step_up_at": now.Unix()},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": accessToken,
		"step_up_at":   now.Unix(),
	})
}

func (h *AdminAuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "refresh_token is required", nil)
		return
	}

	claims := h.getClaims(r)
	if claims == nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Valid token required for refresh context", nil)
		return
	}

	// Preserves OLD step_up_at from current claims
	freshClaims := &jwt.Claims{
		AdminID:   claims.AdminID,
		Email:     claims.Email,
		Role:      claims.Role,
		UserType:  "ADMIN",
		SessionID: claims.SessionID,
		StepUpAt:  claims.StepUpAt,
	}
	newAccessToken, _ := jwt.GenerateToken(freshClaims, h.jwtSecret, 1*time.Hour)
	newRefresh, _, _ := jwt.GenerateRefreshToken()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  newAccessToken,
		"refresh_token": newRefresh,
		"step_up_at":    claims.StepUpAt,
	})
}

func (h *AdminAuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims != nil && claims.AdminID != "" {
		_ = h.repo.RevokeAdminSessions(r.Context(), claims.AdminID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AdminAuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.AdminID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	admin, err := h.repo.GetAdminByID(r.Context(), claims.AdminID)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Admin not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         admin.ID,
		"email":      admin.Email,
		"name":       admin.Name,
		"role":       admin.Role,
		"is_active":  admin.IsActive,
		"step_up_at": claims.StepUpAt,
	})
}
