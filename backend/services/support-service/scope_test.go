package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestManagerScopeEnforcement_QueryingDifferentStoreReturns403(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewTicketHandler(repo, nil, nil)
	r := mux.NewRouter()
	SetupRoutes(r, handler, nil)

	storeA := "store-mumbai-01"
	storeB := "store-delhi-02"

	ticketA := &SupportTicket{
		ID:            "t-store-a",
		RequesterID:   "cust-a",
		RequesterType: "CUSTOMER",
		StoreID:       &storeA,
		Subject:       "Store A Issue",
		Description:   "Issue",
	}
	_ = repo.CreateTicket(context.Background(), ticketA)

	// Manager for Store A queries Store B
	req := httptest.NewRequest("GET", "/v1/support/tickets?store_id="+storeB, nil)
	req.Header.Set("X-User-ID", "mgr-store-a")
	req.Header.Set("X-User-Role", "MANAGER")
	req.Header.Set("X-Store-ID", storeA)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected HTTP 403 when Manager queries store_id other than their assigned store, got %d", rr.Code)
	}
}

func TestAdminScope_QueryingAnyStoreAllowed(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewTicketHandler(repo, nil, nil)
	r := mux.NewRouter()
	SetupRoutes(r, handler, nil)

	storeB := "store-delhi-02"

	ticketB := &SupportTicket{
		ID:            "t-store-b",
		RequesterID:   "cust-b",
		RequesterType: "CUSTOMER",
		StoreID:       &storeB,
		Subject:       "Store B Issue",
		Description:   "Issue",
	}
	_ = repo.CreateTicket(context.Background(), ticketB)

	// Admin queries Store B
	req := httptest.NewRequest("GET", "/v1/support/tickets?store_id="+storeB, nil)
	req.Header.Set("X-User-ID", "admin-global-1")
	req.Header.Set("X-User-Role", "ADMIN")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected HTTP 200 for Admin querying any store_id, got %d", rr.Code)
	}
}
