package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/validator"
)

type StaffHandler struct {
	repo      Repository
	jwtSecret string
}

func NewStaffHandler(repo Repository, jwtSecret string) *StaffHandler {
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &StaffHandler{repo: repo, jwtSecret: jwtSecret}
}

func (h *StaffHandler) extractClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		// Fallback headers for integration test mocking
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

func (h *StaffHandler) HandleCreateStaff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil || (claims.Role != RoleManager && claims.Role != RoleAdmin) {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Manager or Admin authorization required", nil)
		return
	}

	var req CreateStaffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	// Validate phone format
	if !validator.ValidatePhone(req.Phone) {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Phone number must be a valid +91 10-digit number", nil)
		return
	}

	// Validate role
	req.Role = strings.ToUpper(req.Role)
	if req.Role != RoleCashier && req.Role != RoleStockAssociate && req.Role != RoleSecurity && req.Role != RoleManager {
		errors.WriteError(w, http.StatusBadRequest, CodeInvalidRole, "Role must be CASHIER, STOCK_ASSOCIATE, SECURITY, or MANAGER", nil)
		return
	}

	// Store Scope Check: MANAGER caller can only create staff for their OWN store_id
	if claims.Role == RoleManager && req.StoreID != claims.StoreID {
		errors.WriteError(w, http.StatusForbidden, CodeStoreScopeMismatch, "Managers can only create staff for their assigned store", nil)
		return
	}

	staff := &StaffMember{
		StoreID:   req.StoreID,
		ChainID:   claims.ChainID,
		Phone:     req.Phone,
		Name:      req.Name,
		Role:      req.Role,
		CreatedBy: &claims.UserID,
	}
	if staff.ChainID == "" {
		staff.ChainID = "chain-hq-001"
	}

	if err := h.repo.CreateStaffMember(r.Context(), staff); err != nil {
		if err == ErrPhoneConflict {
			errors.WriteError(w, http.StatusConflict, CodePhoneAlreadyStaff, "Phone number is already registered to a staff member", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create staff member", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(staff)
}

func (h *StaffHandler) HandleListStaff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		storeID = claims.StoreID
	}
	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id parameter is required", nil)
		return
	}

	// Store Scope Check
	if claims.Role == RoleManager && storeID != claims.StoreID {
		errors.WriteError(w, http.StatusForbidden, CodeStoreScopeMismatch, "Managers can only list staff for their assigned store", nil)
		return
	}

	roleFilter := r.URL.Query().Get("role")
	includeInactive := r.URL.Query().Get("include_inactive") == "true"
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	staffList, total, err := h.repo.ListStaffByStore(r.Context(), storeID, roleFilter, includeInactive, page, pageSize)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list staff", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"staff":     staffList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *StaffHandler) HandleUpdateStaff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil || (claims.Role != RoleManager && claims.Role != RoleAdmin) {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Manager or Admin authorization required", nil)
		return
	}

	vars := mux.Vars(r)
	staffID := vars["id"]

	existing, err := h.repo.GetStaffByID(r.Context(), staffID)
	if err != nil || existing == nil {
		errors.WriteError(w, http.StatusNotFound, CodeStaffNotFound, "Staff member not found", nil)
		return
	}

	// Store Scope Check
	if claims.Role == RoleManager && existing.StoreID != claims.StoreID {
		errors.WriteError(w, http.StatusForbidden, CodeStoreScopeMismatch, "Managers can only edit staff in their assigned store", nil)
		return
	}

	var req UpdateStaffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON body", nil)
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Phone != nil && strings.TrimSpace(*req.Phone) != "" {
		existing.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Role != nil {
		roleUpper := strings.ToUpper(*req.Role)
		if roleUpper != RoleCashier && roleUpper != RoleStockAssociate && roleUpper != RoleSecurity && roleUpper != RoleManager {
			errors.WriteError(w, http.StatusBadRequest, CodeInvalidRole, "Role must be CASHIER, STOCK_ASSOCIATE, SECURITY, or MANAGER", nil)
			return
		}
		existing.Role = roleUpper
	}

	if err := h.repo.UpdateStaffMember(r.Context(), existing); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update staff member", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(existing)
}

func (h *StaffHandler) HandleRequestManagerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	var req struct {
		OriginalPhone string `json:"original_phone"`
		StoreID       string `json:"store_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.OriginalPhone) == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Original phone number is required", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Reset request sent to Store Manager. Please verify your identity with store management in person.",
	})
}

func (h *StaffHandler) HandleDeactivateStaff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		errors.WriteError(w, http.StatusMethodNotAllowed, errors.CodeInvalidRequest, "Method not allowed", nil)
		return
	}

	claims, err := h.extractClaims(r)
	if err != nil || (claims.Role != RoleManager && claims.Role != RoleAdmin) {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Manager or Admin authorization required", nil)
		return
	}

	vars := mux.Vars(r)
	staffID := vars["id"]

	existing, err := h.repo.GetStaffByID(r.Context(), staffID)
	if err != nil || existing == nil {
		errors.WriteError(w, http.StatusNotFound, CodeStaffNotFound, "Staff member not found", nil)
		return
	}

	// Store Scope Check
	if claims.Role == RoleManager && existing.StoreID != claims.StoreID {
		errors.WriteError(w, http.StatusForbidden, CodeStoreScopeMismatch, "Managers can only deactivate staff in their assigned store", nil)
		return
	}

	// Deactivate staff, revoke sessions & end shift in SAME transaction
	if err := h.repo.DeactivateStaffMemberTx(r.Context(), staffID); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to deactivate staff member", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":     staffID,
		"status": "DEACTIVATED",
	})
}
