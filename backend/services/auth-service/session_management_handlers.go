package main

import (
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

// GET /v1/auth/sessions
func (h *AuthHandler) HandleGetUserSessions(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	sessions, err := h.repo.GetUserSessions(r.Context(), claims.UserID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to retrieve user sessions", nil)
		return
	}

	dtos := make([]SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		label := s.DeviceID
		if s.DeviceLabel != nil && *s.DeviceLabel != "" {
			label = *s.DeviceLabel
		}

		dtos = append(dtos, SessionDTO{
			ID:          s.ID,
			DeviceID:    s.DeviceID,
			DeviceLabel: label,
			CreatedAt:   s.CreatedAt,
			LastUsedAt:  s.LastUsedAt,
			IsCurrent:   s.ID == claims.SessionID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": dtos,
	})
}

// DELETE /v1/auth/sessions/{id}
func (h *AuthHandler) HandleRevokeSessionByID(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	sessionID := vars["id"]
	if sessionID == "" {
		// Fallback URL parsing if mux vars not set
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		sessionID = parts[len(parts)-1]
	}

	if sessionID == claims.SessionID {
		errors.WriteError(w, http.StatusBadRequest, "USE_LOGOUT_FOR_CURRENT_SESSION", "Cannot revoke current session via this path. Use logout instead.", nil)
		return
	}

	if err := h.repo.RevokeSessionByID(r.Context(), sessionID, claims.UserID); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to revoke session", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Session revoked successfully",
	})
}

// DELETE /v1/auth/sessions
func (h *AuthHandler) HandleRevokeAllOtherSessions(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractAuthClaims(r)
	if err != nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	if err := h.repo.RevokeAllOtherSessions(r.Context(), claims.SessionID, claims.UserID); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to revoke other sessions", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "All other sessions revoked successfully",
	})
}

func (h *AuthHandler) extractAuthClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, stdErrors.New("missing authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		return nil, stdErrors.New("invalid token")
	}
	return claims, nil
}
