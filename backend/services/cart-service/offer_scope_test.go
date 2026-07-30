package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOfferScopeEnforcement(t *testing.T) {
	repo := NewMemoryOfferRepository()
	compiler := NewOfferCompiler(repo, nil)
	adminHandler := NewOfferAdminHandler(repo, compiler, nil)

	router := SetupRoutes(nil, nil, adminHandler, nil, nil)

	t.Run("MANAGER attempting chain-wide offer creation gets 403 CHAIN_WIDE_REQUIRES_OWNER", func(t *testing.T) {
		body := CreateOfferRequest{
			ChainID:           "chain-001",
			StoreID:           nil, // chain-wide
			Type:              "PERCENT_OFF",
			AppliesTo:         "ALL",
			RuleConfig:        map[string]interface{}{"percent": 15.0},
			MinCartValuePaise: 0,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/cart/admin/offers", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Role", "MANAGER")
		req.Header.Set("X-Store-ID", "store-100")
		req.Header.Set("X-Chain-ID", "chain-001")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %d", rr.Code)
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte("CHAIN_WIDE_REQUIRES_OWNER")) {
			t.Errorf("expected error code CHAIN_WIDE_REQUIRES_OWNER, got body: %s", rr.Body.String())
		}
	})

	t.Run("CHAIN_HQ FINANCE role attempting chain-wide create gets 403 CHAIN_WIDE_REQUIRES_OWNER", func(t *testing.T) {
		body := CreateOfferRequest{
			ChainID:           "chain-001",
			StoreID:           nil, // chain-wide
			Type:              "FLAT_OFF",
			AppliesTo:         "ALL",
			RuleConfig:        map[string]interface{}{"flat_paise": 500.0},
			MinCartValuePaise: 0,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/v1/cart/admin/offers", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Role", "FINANCE")
		req.Header.Set("X-Chain-ID", "chain-001")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for FINANCE role creating chain-wide offer, got %d", rr.Code)
		}
	})

	t.Run("MANAGER attempting to edit chain-wide offer gets 403 CANNOT_EDIT_CHAIN_WIDE_OFFER", func(t *testing.T) {
		// 1. Create a chain-wide offer as OWNER
		chainWideOffer := &Offer{
			ID:                "offer-cw-123",
			ChainID:           "chain-001",
			StoreID:           nil,
			Type:              "PERCENT_OFF",
			AppliesTo:         "ALL",
			RuleConfig:        map[string]interface{}{"percent": 10.0},
			MinCartValuePaise: 0,
			Priority:          10,
			IsActive:          true,
			CreatedBy:         "owner-1",
		}
		_ = repo.CreateOffer(nil, chainWideOffer)

		// 2. MANAGER tries to PUT update it
		updateBody := UpdateOfferRequest{
			Type:              "PERCENT_OFF",
			AppliesTo:         "ALL",
			RuleConfig:        map[string]interface{}{"percent": 20.0},
			MinCartValuePaise: 0,
			Priority:          10,
		}
		bodyBytes, _ := json.Marshal(updateBody)

		req := httptest.NewRequest(http.MethodPut, "/v1/cart/admin/offers/offer-cw-123", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Role", "MANAGER")
		req.Header.Set("X-Store-ID", "store-100")
		req.Header.Set("X-Chain-ID", "chain-001")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden when MANAGER edits chain-wide offer, got %d", rr.Code)
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte("CANNOT_EDIT_CHAIN_WIDE_OFFER")) {
			t.Errorf("expected CANNOT_EDIT_CHAIN_WIDE_OFFER error code, got %s", rr.Body.String())
		}
	})
}
