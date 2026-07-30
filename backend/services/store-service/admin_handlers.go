package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/validator"
)

// ── GSTIN / state helpers (kept here for self-manage validation) ─────────────

var stateToGSTINPrefix = map[string]string{
	"JAMMU AND KASHMIR": "01", "JK": "01",
	"HIMACHAL PRADESH":  "02", "HP": "02",
	"PUNJAB":            "03", "PB": "03",
	"CHANDIGARH":        "04", "CH": "04",
	"UTTARAKHAND":       "05", "UK": "05",
	"HARYANA":           "06", "HR": "06",
	"DELHI":             "07", "DL": "07",
	"RAJASTHAN":         "08", "RJ": "08",
	"UTTAR PRADESH":     "09", "UP": "09",
	"BIHAR":             "10", "BR": "10",
	"WEST BENGAL":       "19", "WB": "19",
	"GUJARAT":           "24", "GJ": "24",
	"MAHARASHTRA":       "27", "MH": "27",
	"KARNATAKA":         "29", "KA": "29",
	"TAMIL NADU":        "33", "TN": "33",
	"TELANGANA":         "36", "TG": "36", "TS": "36",
	"ANDHRA PRADESH":    "37", "AP": "37",
}

func validateGSTINAndState(gstin, state string) (string, string) {
	if gstin == "" {
		return "", ""
	}
	valid, err := validator.ValidateGSTIN(gstin)
	if !valid || err != nil {
		return errors.CodeGSTINChecksumInvalid, "GSTIN checksum is invalid"
	}
	if state != "" {
		gstinState := validator.GSTINStateCode(gstin)
		stUpper := strings.ToUpper(strings.TrimSpace(state))
		expectedPrefix := stateToGSTINPrefix[stUpper]
		if expectedPrefix == "" {
			expectedPrefix = stUpper
		}
		if gstinState != expectedPrefix {
			return errors.CodeGSTINStateMismatch, "GSTIN state code does not match store state"
		}
	}
	return "", ""
}

// ── Shared JWT extraction helpers ─────────────────────────────────────────────

func getClaimsFromRequest(r *http.Request, jwtSecret string) *jwt.Claims {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseAndVerifyToken(tokenStr, jwtSecret)
		if err == nil && claims != nil {
			return claims
		}
	}
	return &jwt.Claims{
		UserID:   r.Header.Get("X-User-ID"),
		UserType: r.Header.Get("X-User-Type"),
		Role:     r.Header.Get("X-User-Role"),
		StoreID:  r.Header.Get("X-Store-ID"),
	}
}

// ── InternalAdminWriteHandler — SYSTEM-JWT-only write paths ──────────────────
// These are ONLY called by admin-store-service; they are NOT externally reachable.

type InternalAdminWriteHandler struct {
	repo      Repository
	jwtSecret string
}

func NewInternalAdminWriteHandler(repo Repository, jwtSecret string) *InternalAdminWriteHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &InternalAdminWriteHandler{repo: repo, jwtSecret: jwtSecret}
}

func (h *InternalAdminWriteHandler) requireSystem(r *http.Request) bool {
	claims := getClaimsFromRequest(r, h.jwtSecret)
	return claims != nil && (claims.UserType == "SYSTEM" || r.Header.Get("X-Internal-Service") == "admin-store-service")
}

// POST /v1/store/internal/admin-write/stores
func (h *InternalAdminWriteHandler) HandleCreateStore(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}

	var req AdminStoreCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	store := &Store{
		ChainID:              req.ChainID,
		Name:                 req.Name,
		Address:              req.Address,
		City:                 req.City,
		State:                req.State,
		Pincode:              req.Pincode,
		GSTIN:                req.GSTIN,
		Lat:                  req.Lat,
		Lng:                  req.Lng,
		GeofenceRadiusMeters: req.GeofenceRadiusMeters,
		CapacityMax:          req.CapacityMax,
		OpeningTime:          req.OpeningTime,
		ClosingTime:          req.ClosingTime,
		Timezone:             req.Timezone,
		RFIDEnabled:          req.RFIDEnabled,
		Status:               "ACTIVE",
	}

	// Chain validation now lives in admin-store-service; skip it here.
	// Direct insert into stores table.
	if err := h.repo.CreateStoreInternal(r.Context(), store); err != nil {
		logger.Error("Internal CreateStore failed: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create store", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(store)
}

// PUT /v1/store/internal/admin-write/stores/{id}/geofence
func (h *InternalAdminWriteHandler) HandleUpdateGeofence(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req AdminGeofenceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}
	if err := h.repo.UpdateGeofence(r.Context(), storeID, req.Polygon, req.RadiusMeters); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update geofence", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// PUT /v1/store/internal/admin-write/stores/{id}/hours
func (h *InternalAdminWriteHandler) HandleUpdateHours(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req AdminHoursUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}
	if err := h.repo.UpdateHours(r.Context(), storeID, req.OpeningTime, req.ClosingTime, req.Timezone); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update hours", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// PUT /v1/store/internal/admin-write/stores/{id}/capacity
func (h *InternalAdminWriteHandler) HandleUpdateCapacity(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req AdminCapacityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}
	if err := h.repo.UpdateCapacity(r.Context(), storeID, req.CapacityMax); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update capacity", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// PUT /v1/store/internal/admin-write/stores/{id}/status
func (h *InternalAdminWriteHandler) HandleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req AdminStatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}
	if err := h.repo.UpdateStatus(r.Context(), storeID, req.Status); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update status", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// PUT /v1/store/internal/admin-write/stores/{id}/payment-setup
func (h *InternalAdminWriteHandler) HandleUpdatePaymentSetup(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req UpdatePaymentSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}
	if err := h.repo.UpdatePaymentSetup(r.Context(), storeID, req.RazorpayAccountID, req.RazorpayKYCStatus, req.PaymentSetupNote); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update payment setup", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// POST /v1/store/internal/admin-write/stores/{id}/qr-tokens/rotate
func (h *InternalAdminWriteHandler) HandleRotateQRTokens(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req AdminRotateQRTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}
	if len(req.GateIDs) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "gate_ids required", nil)
		return
	}

	var generatedTokens []*StoreQRToken
	for _, gateID := range req.GateIDs {
		_ = h.repo.DeactivateQRTokens(r.Context(), storeID, gateID)

		rawBytes := make([]byte, 16)
		_, _ = rand.Read(rawBytes)
		rawToken := hex.EncodeToString(rawBytes)

		token := &StoreQRToken{
			ID:        uuid.New().String(),
			StoreID:   storeID,
			GateID:    gateID,
			Token:     rawToken,
			IsActive:  true,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := h.repo.CreateQRToken(r.Context(), token); err == nil {
			generatedTokens = append(generatedTokens, token)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"tokens": generatedTokens})
}

// GET /v1/store/internal/admin-write/stores/{id}/qr-tokens
func (h *InternalAdminWriteHandler) HandleGetQRTokens(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	tokens, err := h.repo.GetActiveQRTokens(r.Context(), storeID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to get QR tokens", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"tokens": tokens})
}

// GET /v1/store/internal/admin-write/stores
func (h *InternalAdminWriteHandler) HandleListStores(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}
	chainID := r.URL.Query().Get("chain_id")
	statusFilter := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	stores, total, err := h.repo.ListStores(r.Context(), chainID, statusFilter, page, pageSize)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list stores", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"stores":    stores,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GET /v1/store/internal/admin-write/stores/{id}
func (h *InternalAdminWriteHandler) HandleGetStore(w http.ResponseWriter, r *http.Request) {
	if !h.requireSystem(r) {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "SYSTEM JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]
	store, err := h.repo.GetStoreByID(r.Context(), storeID)
	if err != nil || store == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Store not found", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(store)
}

// ── SelfManageHandler — MANAGER-JWT-only narrow write paths ──────────────────
// Store managers updating their OWN store's hours or capacity from Retailer Dashboard.
// These do NOT go through admin-store-service (which is ADMIN-only).

type SelfManageHandler struct {
	repo      Repository
	jwtSecret string
}

func NewSelfManageHandler(repo Repository, jwtSecret string) *SelfManageHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &SelfManageHandler{repo: repo, jwtSecret: jwtSecret}
}

// requireManager validates that the caller has MANAGER role and that the
// store_id in the path matches the store_id scoped in their JWT.
func (h *SelfManageHandler) requireManagerForStore(r *http.Request, pathStoreID string) bool {
	claims := getClaimsFromRequest(r, h.jwtSecret)
	if claims == nil {
		return false
	}
	if claims.Role != "MANAGER" {
		return false
	}
	// Strict store-scope: manager may only update their own assigned store
	if claims.StoreID != "" && claims.StoreID != pathStoreID {
		return false
	}
	return true
}

// PUT /v1/store/self-manage/stores/{id}/hours
func (h *SelfManageHandler) HandleUpdateHours(w http.ResponseWriter, r *http.Request) {
	storeID := mux.Vars(r)["id"]
	if !h.requireManagerForStore(r, storeID) {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden,
			"MANAGER JWT scoped to this store required", nil)
		return
	}

	var req AdminHoursUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}
	if err := h.repo.UpdateHours(r.Context(), storeID, req.OpeningTime, req.ClosingTime, req.Timezone); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update store hours", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// PUT /v1/store/self-manage/stores/{id}/capacity
func (h *SelfManageHandler) HandleUpdateCapacity(w http.ResponseWriter, r *http.Request) {
	storeID := mux.Vars(r)["id"]
	if !h.requireManagerForStore(r, storeID) {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden,
			"MANAGER JWT scoped to this store required", nil)
		return
	}

	var req AdminCapacityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}
	if err := h.repo.UpdateCapacity(r.Context(), storeID, req.CapacityMax); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update capacity", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}
