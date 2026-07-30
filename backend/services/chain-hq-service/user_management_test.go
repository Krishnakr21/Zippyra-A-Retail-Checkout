package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

func TestInviteUser_NonOwner_Returns403(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewUserManagementHandler(repo, nil)

	financeClaims := &jwt.Claims{
		UserID:   "user-finance-1",
		ChainID:  "chain-100",
		Role:     RoleFinance,
		UserType: "CHAIN_HQ",
	}

	body, _ := json.Marshal(InviteUserRequest{
		Phone: "+919876543211",
		Name:  "Ops User",
		Role:  RoleOperations,
	})

	req := httptest.NewRequest("POST", "/v1/chain-hq/users", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", financeClaims))
	w := httptest.NewRecorder()

	handler.HandleInviteUser(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestInviteUser_AttemptingOwnerRole_Returns400(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewUserManagementHandler(repo, nil)

	ownerClaims := &jwt.Claims{
		UserID:   "user-owner-1",
		ChainID:  "chain-100",
		Role:     RoleOwner,
		UserType: "CHAIN_HQ",
	}

	body, _ := json.Marshal(InviteUserRequest{
		Phone: "+919876543212",
		Name:  "Second Owner Attempt",
		Role:  RoleOwner,
	})

	req := httptest.NewRequest("POST", "/v1/chain-hq/users", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", ownerClaims))
	w := httptest.NewRecorder()

	handler.HandleInviteUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 when attempting to invite OWNER role, got %d", w.Code)
	}
}

func TestDeactivateUser_SelfDeactivation_Returns400CannotDeactivateSelf(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewUserManagementHandler(repo, nil)

	owner := &ChainHQUser{
		ID:       "owner-99",
		ChainID:  "chain-100",
		Phone:    "+919876543213",
		Name:     "Owner Self",
		Role:     RoleOwner,
		IsActive: true,
	}
	_ = repo.CreateUser(context.Background(), owner)

	ownerClaims := &jwt.Claims{
		UserID:   owner.ID,
		ChainID:  owner.ChainID,
		Role:     RoleOwner,
		UserType: "CHAIN_HQ",
	}

	req := httptest.NewRequest("DELETE", "/v1/chain-hq/users/"+owner.ID, nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", ownerClaims))
	req = mux.SetURLVars(req, map[string]string{"id": owner.ID})
	w := httptest.NewRecorder()

	handler.HandleDeactivateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for self deactivation, got %d", w.Code)
	}

	var resp errors.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error.Code != CodeCannotDeactivateSelf {
		t.Fatalf("expected error code CANNOT_DEACTIVATE_SELF, got %s", resp.Error.Code)
	}
}

func TestAdminProvisionOwner_SecondAttempt_Returns409ChainAlreadyHasOwner(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewUserManagementHandler(repo, nil)

	// Existing owner
	existingOwner := &ChainHQUser{
		ChainID:  "chain-200",
		Phone:    "+919876543214",
		Name:     "Existing Owner",
		Role:     RoleOwner,
		IsActive: true,
	}
	_ = repo.CreateUser(context.Background(), existingOwner)

	adminClaims := &jwt.Claims{
		AdminID:  "admin-1",
		UserType: "ADMIN",
		StepUpAt: time.Now().Unix(),
	}

	body, _ := json.Marshal(AdminProvisionOwnerRequest{
		ChainID: "chain-200",
		Phone:   "+919876543215",
		Name:    "Second Owner",
	})

	req := httptest.NewRequest("POST", "/v1/chain-hq/admin/owner", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", adminClaims))
	w := httptest.NewRecorder()

	handler.HandleAdminProvisionOwner(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}

	var resp errors.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error.Code != CodeChainAlreadyHasOwner {
		t.Fatalf("expected error code CHAIN_ALREADY_HAS_OWNER, got %s", resp.Error.Code)
	}
}
