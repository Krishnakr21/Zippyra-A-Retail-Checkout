package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type NotificationHandler struct {
	repo      NotificationRepository
	jwtSecret string
}

func NewNotificationHandler(repo NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		repo:      repo,
		jwtSecret: "dev-secret-key-change-in-prod",
	}
}

func (h *NotificationHandler) extractClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
		if err == nil {
			return claims, nil
		}
	}

	userID := r.Header.Get("X-User-ID")
	role := r.Header.Get("X-User-Role")
	if userID != "" {
		return &jwt.Claims{
			UserID: userID,
			Role:   role,
		}, nil
	}
	return nil, sharedErrors.NewAPIError(sharedErrors.CodeUnauthorized, "Unauthorized", nil)
}

// 1. POST /v1/notification/device-tokens
func (h *NotificationHandler) RegisterDeviceToken(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req RegisterDeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.FCMToken == "" || req.DeviceID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid payload", nil)
		return
	}

	userType := UserTypeCustomer
	if claims.Role == "STAFF" || claims.Role == "SECURITY" || claims.Role == "MANAGER" {
		userType = UserTypeStaff
	}

	token := &DeviceToken{
		UserID:   claims.UserID,
		UserType: userType,
		FCMToken: req.FCMToken,
		Platform: req.Platform,
		DeviceID: req.DeviceID,
	}

	if err := h.repo.UpsertDeviceToken(r.Context(), token); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to register token", nil)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "REGISTERED"})
}

// 2. DELETE /v1/notification/device-tokens/{device_id}
func (h *NotificationHandler) DeactivateDeviceToken(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["device_id"]

	if err := h.repo.DeactivateDeviceToken(r.Context(), claims.UserID, deviceID); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to deactivate token", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// 3. GET /v1/notification/preferences
func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	prefs, err := h.repo.ListPreferences(r.Context(), claims.UserID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list preferences", nil)
		return
	}

	type PreferenceResponse struct {
		UserID           string           `json:"user_id"`
		UserType         UserType         `json:"user_type"`
		NotificationType NotificationType `json:"notification_type"`
		Channel          Channel          `json:"channel"`
		IsMandatory      bool             `json:"is_mandatory"`
		UpdatedAt        interface{}      `json:"updated_at"`
	}

	var responseItems []*PreferenceResponse
	for _, p := range prefs {
		responseItems = append(responseItems, &PreferenceResponse{
			UserID:           p.UserID,
			UserType:         p.UserType,
			NotificationType: p.NotificationType,
			Channel:          p.Channel,
			IsMandatory:      MandatoryNotificationTypes[p.NotificationType],
			UpdatedAt:        p.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"preferences": responseItems})
}

// 4. PUT /v1/notification/preferences
func (h *NotificationHandler) UpdatePreference(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req UpdatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid payload", nil)
		return
	}

	// Reject attempts to set MANDATORY types to NONE
	if MandatoryNotificationTypes[req.NotificationType] && req.Channel == ChannelNone {
		sharedErrors.WriteError(w, http.StatusBadRequest, "CANNOT_DISABLE_MANDATORY_NOTIFICATION", "Mandatory transactional notifications cannot be set to NONE", nil)
		return
	}

	userType := UserTypeCustomer
	if claims.Role == "STAFF" || claims.Role == "SECURITY" || claims.Role == "MANAGER" {
		userType = UserTypeStaff
	}

	pref := &NotificationPreference{
		UserID:           claims.UserID,
		UserType:         userType,
		NotificationType: req.NotificationType,
		Channel:          req.Channel,
	}

	if err := h.repo.UpsertPreference(r.Context(), pref); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to update preference", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pref)
}

// 5. GET /v1/notification/inbox?page={n}
func (h *NotificationHandler) GetInbox(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	page := 1
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if parsed, err := strconv.Atoi(pStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	logs, err := h.repo.ListUserInbox(r.Context(), claims.UserID, page, 20)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to fetch inbox", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"notifications": logs, "page": page})
}

// 6. PUT /v1/notification/inbox/{id}/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.repo.MarkNotificationRead(r.Context(), claims.UserID, id); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to mark read", nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "MARKED_READ"})
}

// 7. GET /v1/notification/inbox/unread-count
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	count, err := h.repo.GetUnreadCount(r.Context(), claims.UserID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to fetch unread count", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"unread_count": count})
}
