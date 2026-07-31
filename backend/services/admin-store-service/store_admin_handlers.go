package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/logger"
	"github.com/zippyra/backend/shared/validator"
)

// stateToGSTINPrefix maps Indian state names / abbreviations to GSTIN state codes.
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

func validateGSTINAndState(gstin, state string) (errCode, errMsg string) {
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

// StoreClientIface is the interface that StoreServiceClient implements,
// exposed here so unit tests can inject a mock without a live store-service.
type StoreClientIface interface {
	CreateStore(req *CreateStoreRequest) (*StoreResponse, error)
	UpdateGeofence(storeID string, req *UpdateGeofenceRequest) error
	UpdateHours(storeID string, req *UpdateHoursRequest) error
	UpdateCapacity(storeID string, req *UpdateCapacityRequest) error
	UpdateStatus(storeID string, req *UpdateStatusRequest) error
	UpdatePaymentSetup(storeID string, req *UpdatePaymentSetupRequest) error
	RotateQRTokens(storeID string, req *RotateQRTokensRequest) (map[string]interface{}, error)
	GetQRTokens(storeID string) (map[string]interface{}, error)
	ListStores(query string) (map[string]interface{}, error)
	GetStoreByID(storeID string) (*StoreResponse, error)
}

// StoreAdminHandler handles all /v1/admin-store/stores/* endpoints.
// It validates chain ownership locally (chains table is here), then delegates
// every actual store DB write to store-service via StoreClientIface.
type StoreAdminHandler struct {
	repo        ChainRepository
	storeClient StoreClientIface
	jwtSecret   string
}

func NewStoreAdminHandler(repo ChainRepository, storeClient StoreClientIface, jwtSecret string) *StoreAdminHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &StoreAdminHandler{repo: repo, storeClient: storeClient, jwtSecret: jwtSecret}
}

func (h *StoreAdminHandler) getClaims(r *http.Request) *jwt.Claims {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
		if err == nil && claims != nil {
			return claims
		}
	}
	return &jwt.Claims{
		UserID:   r.Header.Get("X-User-ID"),
		UserType: r.Header.Get("X-User-Type"),
		Role:     r.Header.Get("X-User-Role"),
		ChainID:  r.Header.Get("X-Chain-ID"),
	}
}

func (h *StoreAdminHandler) requireAdmin(r *http.Request) *jwt.Claims {
	claims := h.getClaims(r)
	if claims == nil {
		return nil
	}
	if claims.UserType != "ADMIN" && claims.Role != "ADMIN" {
		return nil
	}
	return claims
}

// validateChainActive checks the chains table local to this service and returns
// the chain if ACTIVE, or writes an appropriate error to w and returns nil.
func (h *StoreAdminHandler) validateChainActive(w http.ResponseWriter, r *http.Request, chainID string) *Chain {
	chain, err := h.repo.GetChainByID(r.Context(), chainID)
	if err != nil || chain == nil {
		errors.WriteError(w, http.StatusBadRequest, CodeChainNotFound, "Chain not found", nil)
		return nil
	}
	if chain.Status != ChainStatusActive {
		errors.WriteError(w, http.StatusBadRequest, CodeChainSuspended, "Cannot create store under a suspended chain", nil)
		return nil
	}
	return chain
}

// POST /v1/admin-store/stores
func (h *StoreAdminHandler) HandleCreateStore(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAdmin(r)
	if claims == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}

	var req CreateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	// Resolve chain: fall back to claims.ChainID if not provided
	chainID := req.ChainID
	if chainID == "" {
		chainID = claims.ChainID
	}
	req.ChainID = chainID

	// Local chain status validation — chains table is owned here, no cross-service call needed
	if h.validateChainActive(w, r, chainID) == nil {
		return
	}

	// GSTIN checksum + state mismatch validation (same logic as original)
	if errCode, errMsg := validateGSTINAndState(req.GSTIN, req.State); errCode != "" {
		errors.WriteError(w, http.StatusBadRequest, errCode, errMsg, nil)
		return
	}

	// Delegate the actual DB write to store-service via SYSTEM JWT
	store, err := h.storeClient.CreateStore(&req)
	if err != nil {
		logger.Error("Failed to create store via store-service: %v", err)
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create store", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(store)
}

// PUT /v1/admin-store/stores/{id}/geofence
func (h *StoreAdminHandler) HandleUpdateGeofence(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req UpdateGeofenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if err := h.storeClient.UpdateGeofence(storeID, &req); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update geofence", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// PUT /v1/admin-store/stores/{id}/hours
func (h *StoreAdminHandler) HandleUpdateHours(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req UpdateHoursRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if err := h.storeClient.UpdateHours(storeID, &req); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update store hours", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// PUT /v1/admin-store/stores/{id}/capacity
func (h *StoreAdminHandler) HandleUpdateCapacity(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req UpdateCapacityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if err := h.storeClient.UpdateCapacity(storeID, &req); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update capacity", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// PUT /v1/admin-store/stores/{id}/status  (Step-Up required for INACTIVE/UNDER_MAINTENANCE)
func (h *StoreAdminHandler) HandleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAdmin(r)
	if claims == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	targetStatus := strings.ToUpper(req.Status)
	if targetStatus == StoreStatusInactive || targetStatus == StoreStatusUnderMaintenance {
		if claims.StepUpAt <= 0 || (time.Now().Unix()-claims.StepUpAt) > 600 {
			errors.WriteError(w, http.StatusForbidden, "STEP_UP_REQUIRED", "Fresh 2FA step-up required to deactivate store", nil)
			return
		}
	}
	req.Status = targetStatus

	if err := h.storeClient.UpdateStatus(storeID, &req); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update store status", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// PUT /v1/admin-store/stores/{id}/payment-setup
func (h *StoreAdminHandler) HandleUpdatePaymentSetup(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req UpdatePaymentSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	if err := h.storeClient.UpdatePaymentSetup(storeID, &req); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update payment setup", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UPDATED"})
}

// POST /v1/admin-store/stores/{id}/qr-tokens/rotate
func (h *StoreAdminHandler) HandleRotateQRTokens(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	var req RotateQRTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}
	if len(req.GateIDs) == 0 {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "gate_ids required", nil)
		return
	}

	result, err := h.storeClient.RotateQRTokens(storeID, &req)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to rotate QR tokens", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// GET /v1/admin-store/stores/{id}/qr-tokens
func (h *StoreAdminHandler) HandleGetQRTokens(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(r) == nil {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "ADMIN JWT required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	result, err := h.storeClient.GetQRTokens(storeID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to get QR tokens", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// GET /v1/admin-store/stores
func (h *StoreAdminHandler) HandleListStores(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || (claims.UserType != "ADMIN" && claims.UserType != "CHAIN_HQ" && claims.Role != "ADMIN") {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Authorized admin or chain HQ session required", nil)
		return
	}

	result, err := h.storeClient.ListStores(r.URL.RawQuery)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list stores", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// GET /v1/admin-store/stores/{id}
func (h *StoreAdminHandler) HandleGetStore(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || (claims.UserType != "ADMIN" && claims.UserType != "CHAIN_HQ" && claims.Role != "ADMIN") {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Authorized admin or chain HQ session required", nil)
		return
	}
	storeID := mux.Vars(r)["id"]

	store, err := h.storeClient.GetStoreByID(storeID)
	if err != nil || store == nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Store not found", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(store)
}
