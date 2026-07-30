package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/audit"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/validator"
)

type UserManagementHandler struct {
	repo     Repository
	auditPub *audit.Publisher
}

func NewUserManagementHandler(repo Repository, auditPub *audit.Publisher) *UserManagementHandler {
	return &UserManagementHandler{
		repo:     repo,
		auditPub: auditPub,
	}
}

func (h *UserManagementHandler) getClaims(r *http.Request) *jwt.Claims {
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

func (h *UserManagementHandler) HandleInviteUser(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "CHAIN_HQ" || claims.Role != RoleOwner {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "OWNER role required to invite chain users", nil)
		return
	}

	var req InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !validator.ValidatePhone(req.Phone) || req.Name == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Valid phone (+91...) and name are required", nil)
		return
	}

	req.Role = strings.ToUpper(strings.TrimSpace(req.Role))
	if req.Role != RoleFinance && req.Role != RoleOperations {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Role must be FINANCE or OPERATIONS (cannot invite another OWNER)", nil)
		return
	}

	// Always restrict to caller's chain_id
	chainID := claims.ChainID
	callerID := claims.UserID

	existing, _ := h.repo.GetUserByPhone(r.Context(), req.Phone)
	if existing != nil {
		errors.WriteError(w, http.StatusConflict, errors.CodeIdentifierTaken, "User with this phone number is already registered", nil)
		return
	}

	user := &ChainHQUser{
		ChainID:   chainID,
		Phone:     req.Phone,
		Name:      req.Name,
		Role:      req.Role,
		IsActive:  true,
		CreatedBy: &callerID,
	}

	if err := h.repo.CreateUser(r.Context(), user); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to create chain user", nil)
		return
	}

	if h.auditPub != nil {
		h.auditPub.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       callerID,
			ActionType:    "chain_hq.user_invited",
			TargetType:    "chain_hq_user",
			TargetID:      user.ID,
			SourceService: "chain-hq-service",
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       map[string]interface{}{"chain_id": chainID, "role": user.Role, "phone": user.Phone},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserManagementHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "CHAIN_HQ" || claims.ChainID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authorized CHAIN_HQ session required", nil)
		return
	}

	roleFilter := r.URL.Query().Get("role")
	users, err := h.repo.ListUsersByChainID(r.Context(), claims.ChainID, roleFilter)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list chain users", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"users": users})
}

func (h *UserManagementHandler) HandleDeactivateUser(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "CHAIN_HQ" || claims.Role != RoleOwner {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "OWNER role required to deactivate users", nil)
		return
	}

	vars := mux.Vars(r)
	targetID := vars["id"]

	if claims.UserID == targetID {
		errors.WriteError(w, http.StatusBadRequest, CodeCannotDeactivateSelf, "Chain Owners cannot deactivate their own account", nil)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), targetID)
	if err != nil || user.ChainID != claims.ChainID {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "User not found in your chain", nil)
		return
	}

	user.IsActive = false
	user.UpdatedAt = time.Now().UTC()
	_ = h.repo.UpdateUser(r.Context(), user)
	_ = h.repo.RevokeUserSessions(r.Context(), targetID)

	if h.auditPub != nil {
		h.auditPub.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       claims.UserID,
			ActionType:    "chain_hq.user_deactivated",
			TargetType:    "chain_hq_user",
			TargetID:      targetID,
			SourceService: "chain-hq-service",
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       map[string]interface{}{"chain_id": claims.ChainID},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deactivated", "id": targetID})
}

func (h *UserManagementHandler) HandleAdminProvisionOwner(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "ADMIN" {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Internal ADMIN JWT required to provision chain owner", nil)
		return
	}

	var req AdminProvisionOwnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ChainID == "" || !validator.ValidatePhone(req.Phone) || req.Name == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "chain_id, name, and valid phone required", nil)
		return
	}

	// Check if active owner already exists for this chain
	existingOwner, _ := h.repo.GetActiveOwnerByChainID(r.Context(), req.ChainID)
	if existingOwner != nil {
		errors.WriteError(w, http.StatusConflict, CodeChainAlreadyHasOwner, "Chain already has an active OWNER provisioned", nil)
		return
	}

	adminID := claims.AdminID
	owner := &ChainHQUser{
		ChainID:   req.ChainID,
		Phone:     req.Phone,
		Name:      req.Name,
		Role:      RoleOwner,
		IsActive:  true,
		CreatedBy: &adminID,
	}

	if err := h.repo.CreateUser(r.Context(), owner); err != nil {
		if err == ErrUserExists {
			errors.WriteError(w, http.StatusConflict, errors.CodeIdentifierTaken, "Phone number is already in use", nil)
			return
		}
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to provision owner", nil)
		return
	}

	if h.auditPub != nil {
		h.auditPub.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       adminID,
			ActionType:    "chain_hq.owner_provisioned",
			TargetType:    "chain_hq_user",
			TargetID:      owner.ID,
			SourceService: "chain-hq-service",
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       map[string]interface{}{"chain_id": req.ChainID, "phone": req.Phone},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(owner)
}
