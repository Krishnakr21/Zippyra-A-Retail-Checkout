package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/middleware"
	"github.com/zippyra/backend/shared/redis"
	"github.com/zippyra/backend/shared/validator"
)

type AuthHandler struct {
	repo           Repository
	otpManager     OTPManager
	googleVerifier GoogleTokenVerifier
	jwtSecret      string
	googleClientID string
	redisClient    *redis.Client
}

func NewAuthHandler(repo Repository, otpMgr OTPManager, verifier GoogleTokenVerifier, rdb ...*redis.Client) *AuthHandler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	clientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	if clientID == "" {
		clientID = "105329356913-onj9okrsj316t4fgrbqr1h5vfm7os14m.apps.googleusercontent.com"
	}
	var rc *redis.Client
	if len(rdb) > 0 {
		rc = rdb[0]
	}
	return &AuthHandler{
		repo:           repo,
		otpManager:     otpMgr,
		googleVerifier: verifier,
		jwtSecret:      secret,
		googleClientID: clientID,
		redisClient:    rc,
	}
}

// HandleJWKS retrieves JWKS public keys
// @Summary Get JWKS Public Keys
// @Description Retrieve JSON Web Key Set for JWT verification
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /.well-known/jwks.json [get]
func (h *AuthHandler) HandleJWKS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "octet",
				"use": "sig",
				"alg": "HS256",
				"kid": "zippyra-key-v1",
				"k":   "zippyra-dev-jwt-secret-key-32bytes",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_ = json.NewEncoder(w).Encode(jwks)
}

// HandleSendOTP sends a one-time password
// @Summary Send OTP
// @Description Send one-time password to user phone or email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body SendOTPRequest true "Send OTP payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.APIError
// @Failure 429 {object} errors.APIError
// @Router /v1/auth/otp/send [post]
func (h *AuthHandler) HandleSendOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req SendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	req.Channel = strings.ToLower(strings.TrimSpace(req.Channel))
	req.Identifier = strings.TrimSpace(req.Identifier)

	if req.Channel != "phone" && req.Channel != "email" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Channel must be 'phone' or 'email'", nil)
		return
	}

	if req.Channel == "phone" {
		if !validator.ValidatePhone(req.Identifier) {
			errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid phone number format. Must be +91 followed by 10 digits starting with 6-9", nil)
			return
		}
	} else {
		if !validator.ValidateEmail(req.Identifier) {
			errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid email address format", nil)
			return
		}
	}

	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = strings.Split(forwarded, ",")[0]
	}

	_, err := h.otpManager.SendOTP(r.Context(), req.Channel, req.Identifier, ip)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			status := http.StatusBadRequest
			if apiErr.Code == errors.CodeRateLimitExceeded {
				status = http.StatusTooManyRequests
			} else if apiErr.Code == errors.CodeOTPLocked {
				status = http.StatusForbidden
			}
			errors.WriteError(w, status, apiErr.Code, apiErr.Message, apiErr.Details)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to send OTP", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "OTP sent successfully",
	})
}

// HandleVerifyOTP verifies a one-time password and returns JWT session tokens
// @Summary Verify OTP
// @Description Verify OTP code and obtain Access and Refresh tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "Verify OTP payload"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} errors.APIError
// @Failure 403 {object} errors.APIError
// @Failure 500 {object} errors.APIError
// @Router /v1/auth/otp/verify [post]
func (h *AuthHandler) HandleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	req.Channel = strings.ToLower(strings.TrimSpace(req.Channel))
	req.Identifier = strings.TrimSpace(req.Identifier)
	req.OTP = strings.TrimSpace(req.OTP)

	if req.Channel != "phone" && req.Channel != "email" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Channel must be 'phone' or 'email'", nil)
		return
	}
	if req.OTP == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "OTP code is required", nil)
		return
	}

	// 1. Verify OTP
	if err := h.otpManager.VerifyOTP(r.Context(), req.Channel, req.Identifier, req.OTP); err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			status := http.StatusBadRequest
			if apiErr.Code == errors.CodeOTPLocked {
				status = http.StatusForbidden
			}
			errors.WriteError(w, status, apiErr.Code, apiErr.Message, apiErr.Details)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to verify OTP", nil)
		return
	}

	// 2. Find or Create User
	var user *User
	var err error
	isNewUser := false

	if req.Channel == "phone" {
		user, err = h.repo.GetUserByPhone(r.Context(), req.Identifier)
		if err != nil {
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Database error retrieving user", nil)
			return
		}
		if user == nil {
			user, err = h.repo.CreateUserWithPhone(r.Context(), req.Identifier)
			if err != nil {
				errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create user", nil)
				return
			}
			isNewUser = true
		} else {
			_ = h.repo.UpdateUserVerifiedAt(r.Context(), user.ID, "phone")
		}
	} else {
		user, err = h.repo.GetUserByEmail(r.Context(), req.Identifier)
		if err != nil {
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Database error retrieving user", nil)
			return
		}
		if user == nil {
			user, err = h.repo.CreateUserWithEmail(r.Context(), req.Identifier)
			if err != nil {
				errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create user", nil)
				return
			}
			isNewUser = true
		} else {
			_ = h.repo.UpdateUserVerifiedAt(r.Context(), user.ID, "email")
		}
	}

	// 3. Issue Session
	accessToken, refreshToken, err := issueSession(r.Context(), h.repo, user, req.DeviceID, req.DeviceLabel, h.jwtSecret)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to issue session tokens", nil)
		return
	}

	resp := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         ToUserDTO(user),
		IsNewUser:    isNewUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleGoogleOAuth authenticates via Google ID token
// @Summary Google OAuth Authentication
// @Description Authenticate user via Google ID Token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body GoogleOAuthRequest true "Google OAuth payload"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} errors.APIError
// @Failure 401 {object} errors.APIError
// @Router /v1/auth/oauth/google [post]
func (h *AuthHandler) HandleGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req GoogleOAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if req.IDToken == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "id_token is required", nil)
		return
	}

	// 1. Verify Google ID token
	payload, err := h.googleVerifier.VerifyIDToken(r.Context(), req.IDToken, h.googleClientID)
	if err != nil {
		logger.Error("Google ID token verification failed: %v", err)
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeGoogleTokenInvalid, "Invalid or expired Google ID token", nil)
		return
	}

	logger.Info("Google OAuth verified sub: %s, email: %s", payload.Sub, logger.MaskEmail(payload.Email))

	// 2. Account Resolution / Linking
	var user *User
	isNewUser := false

	// Check if google_sub exists
	user, err = h.repo.GetUserByGoogleSub(r.Context(), payload.Sub)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Database query error", nil)
		return
	}

	if user != nil {
		// Log in existing Google user
		_ = h.repo.UpdateAuthProviderLast(r.Context(), user.ID, "google")
	} else if payload.Email != "" {
		// Check if user previously signed up via Email OTP
		existingEmailUser, err := h.repo.GetUserByEmail(r.Context(), payload.Email)
		if err != nil {
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Database query error", nil)
			return
		}

		if existingEmailUser != nil {
			// LINK google_sub to existing user account
			if err := h.repo.LinkGoogleSubToUser(r.Context(), existingEmailUser.ID, payload.Sub); err != nil {
				errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to link Google account to user", nil)
				return
			}
			user = existingEmailUser
			user.GoogleSub = &payload.Sub
			prov := "google"
			user.AuthProviderLast = &prov
			isNewUser = false
		} else {
			// Create new user with google_sub + email
			user, err = h.repo.CreateUserWithGoogle(r.Context(), payload.Sub, payload.Email, payload.EmailVerified)
			if err != nil {
				errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create user for Google login", nil)
				return
			}
			isNewUser = true
		}
	} else {
		// Create new user with google_sub
		user, err = h.repo.CreateUserWithGoogle(r.Context(), payload.Sub, "", payload.EmailVerified)
		if err != nil {
			errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create user for Google login", nil)
			return
		}
		isNewUser = true
	}

	// 3. Issue Session
	accessToken, refreshToken, err := issueSession(r.Context(), h.repo, user, req.DeviceID, req.DeviceLabel, h.jwtSecret)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to issue session tokens", nil)
		return
	}

	resp := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         ToUserDTO(user),
		IsNewUser:    isNewUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// POST /v1/auth/refresh
func (h *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if req.RefreshToken == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "refresh_token is required", nil)
		return
	}

	hash := jwt.HashRefreshToken(req.RefreshToken)
	session, err := h.repo.GetSessionByRefreshHash(r.Context(), hash)
	if err != nil || session == nil || session.RevokedAt != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid or revoked refresh token", nil)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), session.UserID)
	if err != nil || user == nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "User associated with refresh token not found", nil)
		return
	}

	// Update last used at timestamp
	_ = h.repo.UpdateSessionLastUsed(r.Context(), session.ID)

	// Revoke old session
	_ = h.repo.RevokeSession(r.Context(), hash)

	deviceLabel := ""
	if session.DeviceLabel != nil {
		deviceLabel = *session.DeviceLabel
	}

	// Issue new session
	accessToken, refreshToken, err := issueSession(r.Context(), h.repo, user, req.DeviceID, deviceLabel, h.jwtSecret)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to refresh session", nil)
		return
	}

	resp := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         ToUserDTO(user),
		IsNewUser:    false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// POST /v1/auth/logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req LogoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.RefreshToken != "" {
		hash := jwt.HashRefreshToken(req.RefreshToken)
		_ = h.repo.RevokeSession(r.Context(), hash)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// GET /v1/auth/me
func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Bearer token required in Authorization header", nil)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Invalid or expired access token", nil)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "User not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"user": ToUserDTO(user),
	})
}

type AdminUserItemDTO struct {
	ID               string `json:"id"`
	PhoneMasked      string `json:"phone_masked"`
	EmailMasked      string `json:"email_masked"`
	Phone            string `json:"phone,omitempty"`
	Email            string `json:"email,omitempty"`
	EmailVerified    bool   `json:"email_verified"`
	PhoneVerified    bool   `json:"phone_verified"`
	AuthProviderLast string `json:"auth_provider_last"`
}

func (h *AuthHandler) HandleAdminListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	phoneFilter := r.URL.Query().Get("phone")
	emailFilter := r.URL.Query().Get("email")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	users, total, err := h.repo.ListUsersAdmin(r.Context(), phoneFilter, emailFilter, page, pageSize)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list users", nil)
		return
	}

	var dtos []AdminUserItemDTO
	for _, u := range users {
		item := AdminUserItemDTO{
			ID:            u.ID,
			PhoneMasked:   maskPhone(u.Phone),
			EmailMasked:   maskEmail(u.Email),
			EmailVerified: u.EmailVerifiedAt != nil,
			PhoneVerified: u.PhoneVerifiedAt != nil,
		}
		if u.AuthProviderLast != nil {
			item.AuthProviderLast = *u.AuthProviderLast
		}
		dtos = append(dtos, item)
	}

	if dtos == nil {
		dtos = []AdminUserItemDTO{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"users":     dtos,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AuthHandler) HandleAdminGetUserDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "User ID missing", nil)
		return
	}
	userID := parts[4]

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "User not found", nil)
		return
	}

	phoneStr := ""
	if user.Phone != nil {
		phoneStr = *user.Phone
	}
	emailStr := ""
	if user.Email != nil {
		emailStr = *user.Email
	}

	item := AdminUserItemDTO{
		ID:            user.ID,
		PhoneMasked:   maskPhone(user.Phone),
		EmailMasked:   maskEmail(user.Email),
		Phone:         phoneStr,
		Email:         emailStr,
		EmailVerified: user.EmailVerifiedAt != nil,
		PhoneVerified: user.PhoneVerifiedAt != nil,
	}
	if user.AuthProviderLast != nil {
		item.AuthProviderLast = *user.AuthProviderLast
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

// PUT /v1/auth/me/name (CUSTOMER JWT)
func (h *AuthHandler) HandleUpdateName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, ok := r.Context().Value(middleware.UserClaimsKey).(*jwt.Claims)
	if !ok || claims == nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authentication required", nil)
		return
	}

	var req UpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	trimmedName := strings.TrimSpace(req.Name)
	if trimmedName == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Name cannot be empty", nil)
		return
	}
	if len(trimmedName) > 100 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Name exceeds maximum length of 100 characters", nil)
		return
	}

	user, err := h.repo.UpdateUserName(r.Context(), claims.UserID, trimmedName)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update profile name", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ToUserDTO(user))
}

func maskPhone(phone *string) string {
	if phone == nil || *phone == "" {
		return ""
	}
	p := *phone
	if len(p) <= 4 {
		return "****"
	}
	return p[:3] + "XXXXXX" + p[len(p)-4:]
}

func maskEmail(email *string) string {
	if email == nil || *email == "" {
		return ""
	}
	parts := strings.Split(*email, "@")
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***@***.***"
	}
	uname := parts[0]
	maskedName := string(uname[0]) + "***"
	return maskedName + "@" + parts[1]
}
