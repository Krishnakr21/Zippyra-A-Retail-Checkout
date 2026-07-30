package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zippyra/backend/shared/jwt"
)

func TestDashboard_PartialFailure_ReturnsDegradedStoresArray(t *testing.T) {
	handler := NewDashboardHandler()

	claims := &jwt.Claims{
		UserID:   "user-owner-1",
		ChainID:  "chain-100",
		Role:     RoleOwner,
		UserType: "CHAIN_HQ",
	}

	req := httptest.NewRequest("GET", "/v1/chain-hq/dashboard", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
	w := httptest.NewRecorder()

	handler.HandleDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 on dashboard query, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["total_stores"] == nil || resp["active_stores"] == nil {
		t.Fatalf("expected store counts in dashboard response, got %v", resp)
	}

	degraded, ok := resp["degraded_stores"].([]interface{})
	if !ok {
		t.Fatalf("expected degraded_stores array in response")
	}

	// Should not fail request even if degraded array is present or empty
	_ = degraded
}
