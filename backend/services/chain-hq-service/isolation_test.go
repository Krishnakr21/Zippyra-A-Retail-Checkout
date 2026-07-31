package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zippyra/backend/shared/jwt"
)

func TestIsolation_UserCannotQueryOtherChainUsers(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewUserManagementHandler(repo, nil)

	// Seed users in two different chains
	_ = repo.CreateUser(context.Background(), &ChainHQUser{
		ChainID: "chain-A", Phone: "+919000000001", Name: "User A", Role: RoleOwner, IsActive: true,
	})
	_ = repo.CreateUser(context.Background(), &ChainHQUser{
		ChainID: "chain-B", Phone: "+919000000002", Name: "User B", Role: RoleOwner, IsActive: true,
	})

	claimsA := &jwt.Claims{
		UserID:   "user-a",
		ChainID:  "chain-A",
		Role:     RoleOwner,
		UserType: "CHAIN_HQ",
	}

	req := httptest.NewRequest("GET", "/v1/chain-hq/users", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", claimsA))
	w := httptest.NewRecorder()

	handler.HandleListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	users := resp["users"].([]interface{})
	if len(users) != 1 {
		t.Fatalf("expected exactly 1 user for chain-A, got %d", len(users))
	}

	u := users[0].(map[string]interface{})
	if u["chain_id"] != "chain-A" {
		t.Fatalf("expected chain_id chain-A, got %v", u["chain_id"])
	}
}
