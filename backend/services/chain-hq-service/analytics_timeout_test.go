package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zippyra/backend/shared/jwt"
)

func TestDashboard_AnalyticsTimeout_Returns200WithAnalyticsUnavailableFlag(t *testing.T) {
	// Set up a mock analytics server that times out (responds after a delay)
	analyticsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Immediately return error to simulate timeout/unavailability
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer analyticsServer.Close()

	handler := NewDashboardHandler()
	handler.analyticsServiceURL = analyticsServer.URL // Point to mock that returns 503

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

	// Must still return 200
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 even when analytics is unavailable, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	// Must include existing count fields
	if resp["total_stores"] == nil || resp["active_stores"] == nil {
		t.Fatalf("expected store count fields in dashboard response even with analytics unavailable, got %v", resp)
	}

	// Must include analytics_unavailable=true flag
	if analyticsUnavailable, ok := resp["analytics_unavailable"].(bool); !ok || !analyticsUnavailable {
		t.Fatalf("expected analytics_unavailable=true when analytics-service is down, got %v", resp["analytics_unavailable"])
	}

	// Must NOT include fabricated revenue numbers
	if resp["total_revenue_paise"] != nil || resp["total_orders"] != nil {
		t.Fatalf("must not include revenue data when analytics-service is unavailable, got %v", resp)
	}
}
