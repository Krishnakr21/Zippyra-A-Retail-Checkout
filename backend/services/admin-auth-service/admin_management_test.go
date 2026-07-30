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
	"github.com/zippyra/backend/shared/jwt"
)

func TestCreateAdmin_BySuperAdmin_Success(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAdminManagementHandler(repo)

	superAdminClaims := &jwt.Claims{
		AdminID:  "super-admin-1",
		Email:    "super@zippyra.com",
		Role:     RoleSuperAdmin,
		UserType: "ADMIN",
		StepUpAt: time.Now().Unix(),
	}

	body, _ := json.Marshal(CreateAdminRequest{
		Email: "operator@zippyra.com",
		Name:  "Ops Team Member",
		Role:  RolePlatformAdmin,
	})

	req := httptest.NewRequest("POST", "/v1/admin-auth/admin/users", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", superAdminClaims))
	w := httptest.NewRecorder()

	handler.HandleCreateAdmin(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	created, err := repo.GetAdminByEmail(context.Background(), "operator@zippyra.com")
	if err != nil || created.Name != "Ops Team Member" {
		t.Fatalf("admin user not found in repo: %v", err)
	}
}

func TestDeleteAdmin_SelfDeactivation_Returns400CannotDeactivateSelf(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAdminManagementHandler(repo)

	superAdmin := &AdminUser{
		ID:       "super-admin-999",
		Email:    "super999@zippyra.com",
		Name:     "Super Admin",
		Role:     RoleSuperAdmin,
		IsActive: true,
	}
	_ = repo.CreateAdmin(context.Background(), superAdmin)

	superAdminClaims := &jwt.Claims{
		AdminID:  superAdmin.ID,
		Email:    superAdmin.Email,
		Role:     RoleSuperAdmin,
		UserType: "ADMIN",
		StepUpAt: time.Now().Unix(),
	}

	req := httptest.NewRequest("DELETE", "/v1/admin-auth/admin/users/"+superAdmin.ID, nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", superAdminClaims))
	req = mux.SetURLVars(req, map[string]string{"id": superAdmin.ID})
	w := httptest.NewRecorder()

	handler.HandleDeleteAdmin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for self-deactivation, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errMap, ok := resp["error"].(map[string]interface{})
	if !ok || errMap["code"] != CodeCannotDeactivateSelf {
		t.Fatalf("expected error code CANNOT_DEACTIVATE_SELF, got %v", resp)
	}
}
