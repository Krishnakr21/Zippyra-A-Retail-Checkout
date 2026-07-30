package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type AdminManagementHandler struct {
	repo          Repository
	allowedDomain string
}

func NewAdminManagementHandler(repo Repository) *AdminManagementHandler {
	domain := strings.TrimPrefix(strings.TrimSpace("zippyra.com"), "@")
	return &AdminManagementHandler{
		repo:          repo,
		allowedDomain: domain,
	}
}

func (h *AdminManagementHandler) getClaims(r *http.Request) *jwt.Claims {
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

func (h *AdminManagementHandler) HandleCreateAdmin(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN" || claims.Role != RoleSuperAdmin {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "SUPER_ADMIN role required", nil)
		return
	}

	var req CreateAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Name == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "email and name are required", nil)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	parts := strings.Split(req.Email, "@")
	if len(parts) != 2 || !strings.EqualFold(parts[1], h.allowedDomain) {
		errors.WriteError(w, http.StatusBadRequest, CodeDomainNotAllowed, "Email domain is not permitted for admin accounts", nil)
		return
	}

	if req.Role == "" {
		req.Role = RolePlatformAdmin
	}
	if req.Role != RoleSuperAdmin && req.Role != RolePlatformAdmin && req.Role != RoleSupport {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid admin role", nil)
		return
	}

	// Check if already exists
	existing, _ := h.repo.GetAdminByEmail(r.Context(), req.Email)
	if existing != nil {
		errors.WriteError(w, http.StatusConflict, CodeAdminAlreadyExists, "Admin account with this email already exists", nil)
		return
	}

	createdBy := claims.AdminID
	newAdmin := &AdminUser{
		Email:     req.Email,
		Name:      req.Name,
		Role:      req.Role,
		IsActive:  true,
		CreatedBy: &createdBy,
	}

	if err := h.repo.CreateAdmin(r.Context(), newAdmin); err != nil {
		if err == ErrAdminAlreadyExists {
			errors.WriteError(w, http.StatusConflict, CodeAdminAlreadyExists, "Admin user already exists", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create admin user", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newAdmin)
}

func (h *AdminManagementHandler) HandleListAdmins(w http.ResponseWriter, r *http.Request) {
	roleFilter := r.URL.Query().Get("role")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	admins, total, err := h.repo.ListAdmins(r.Context(), roleFilter, page, pageSize)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to fetch admin users", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": admins,
		"total": total,
	})
}

func (h *AdminManagementHandler) HandleUpdateAdminRole(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN" || claims.Role != RoleSuperAdmin {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "SUPER_ADMIN role required", nil)
		return
	}

	vars := mux.Vars(r)
	targetID := vars["id"]
	if targetID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Admin ID required", nil)
		return
	}

	var req UpdateAdminRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Role == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "role is required", nil)
		return
	}

	if req.Role != RoleSuperAdmin && req.Role != RolePlatformAdmin && req.Role != RoleSupport {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid role", nil)
		return
	}

	target, err := h.repo.GetAdminByID(r.Context(), targetID)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Admin user not found", nil)
		return
	}

	target.Role = req.Role
	target.UpdatedAt = time.Now().UTC()
	if err := h.repo.UpdateAdmin(r.Context(), target); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to update role", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(target)
}

func (h *AdminManagementHandler) HandleDeleteAdmin(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN" || claims.Role != RoleSuperAdmin {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "SUPER_ADMIN role required", nil)
		return
	}

	vars := mux.Vars(r)
	targetID := vars["id"]
	if targetID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Admin ID required", nil)
		return
	}

	// Prevent self deactivation
	if claims.AdminID == targetID {
		errors.WriteError(w, http.StatusBadRequest, CodeCannotDeactivateSelf, "Super Admins cannot deactivate their own account", nil)
		return
	}

	target, err := h.repo.GetAdminByID(r.Context(), targetID)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Admin user not found", nil)
		return
	}

	target.IsActive = false
	target.UpdatedAt = time.Now().UTC()
	_ = h.repo.UpdateAdmin(r.Context(), target)

	// Immediately revoke active sessions
	_ = h.repo.RevokeAdminSessions(r.Context(), targetID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deactivated",
		"id":     targetID,
	})
}
