package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type ShiftHandler struct {
	shiftSvc  *ShiftService
	jwtSecret string
}

func NewShiftHandler(shiftSvc *ShiftService, jwtSecret string) *ShiftHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &ShiftHandler{shiftSvc: shiftSvc, jwtSecret: jwtSecret}
}

func (h *ShiftHandler) extractClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		if role := r.Header.Get("X-User-Role"); role != "" {
			return &jwt.Claims{
				UserID:  r.Header.Get("X-User-ID"),
				StoreID: r.Header.Get("X-Store-ID"),
				ChainID: r.Header.Get("X-Chain-ID"),
				Role:    role,
			}, nil
		}
		return nil, errors.NewAPIError(errors.CodeUnauthorized, "Authorization header required", nil)
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
	if err != nil || claims == nil {
		sessClaims, errSess := jwt.ParseAndVerifySessionToken(tokenStr, h.jwtSecret)
		if errSess == nil && sessClaims != nil {
			return &jwt.Claims{
				UserID:  sessClaims.UserID,
				StoreID: sessClaims.StoreID,
				Role:    sessClaims.Role,
			}, nil
		}
		return nil, errors.NewAPIError(errors.CodeUnauthorized, "Invalid authorization token", nil)
	}
	return claims, nil
}

func (h *ShiftHandler) HandleStartShift(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	shift, err := h.shiftSvc.StartShift(r.Context(), claims.UserID, claims.StoreID)
	if err != nil {
		if strings.HasPrefix(err.Error(), CodeShiftAlreadyActive) {
			errors.WriteError(w, http.StatusConflict, CodeShiftAlreadyActive, err.Error(), nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to start shift", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shift)
}

func (h *ShiftHandler) HandleEndShift(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	if err := h.shiftSvc.EndShift(r.Context(), claims.UserID); err != nil {
		if err.Error() == CodeNoActiveShift {
			errors.WriteError(w, http.StatusNotFound, CodeNoActiveShift, "No active shift found to end", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to end shift", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "SHIFT_ENDED",
		"message": "Shift ended successfully",
	})
}

func (h *ShiftHandler) HandleGetCurrentShift(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil || claims.UserID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	shift, err := h.shiftSvc.GetCurrentShift(r.Context(), claims.UserID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to query active shift", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if shift == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"active": false})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"active": true,
		"shift":  shift,
	})
}

func (h *ShiftHandler) HandleGetShiftHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil || (claims.Role != RoleManager && claims.Role != RoleAdmin) {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Manager or Admin authorization required", nil)
		return
	}

	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		storeID = claims.StoreID
	}

	// Store Scope Check
	if claims.Role == RoleManager && storeID != claims.StoreID {
		errors.WriteError(w, http.StatusForbidden, CodeStoreScopeMismatch, "Managers can only view shift history for their assigned store", nil)
		return
	}

	var dateFrom, dateTo *time.Time
	if dFrom := r.URL.Query().Get("date_from"); dFrom != "" {
		if t, err := time.Parse(time.RFC3339, dFrom); err == nil {
			dateFrom = &t
		}
	}
	if dTo := r.URL.Query().Get("date_to"); dTo != "" {
		if t, err := time.Parse(time.RFC3339, dTo); err == nil {
			dateTo = &t
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	shifts, total, err := h.shiftSvc.GetShiftHistory(r.Context(), storeID, dateFrom, dateTo, page, pageSize)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to query shift history", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"shifts":    shifts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
