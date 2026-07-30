package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestStaffPhoneUpdate_PermissionsAndStoreScope(t *testing.T) {
	repo := NewMemoryRepository()
	staffH := NewStaffHandler(repo, "zippyra-dev-jwt-secret-key-32bytes")

	staff := &StaffMember{
		ID:      "stf-101",
		StoreID: "store-001",
		Phone:   "+919876543210",
		Name:    "Ramesh Cashier",
		Role:    RoleCashier,
	}
	_ = repo.CreateStaffMember(context.Background(), staff)

	router := mux.NewRouter()
	router.HandleFunc("/v1/retailer-auth/staff/{id}", staffH.HandleUpdateStaff).Methods("PUT")

	// Test Case 1: Cashier role attempting to update staff -> 403 Forbidden
	phoneUpdate := "+919999900000"
	reqBody, _ := json.Marshal(map[string]string{"phone": phoneUpdate})

	req1 := httptest.NewRequest("PUT", "/v1/retailer-auth/staff/"+staff.ID, bytes.NewBuffer(reqBody))
	req1.Header.Set("X-User-Role", RoleCashier)
	req1.Header.Set("X-Store-ID", "store-001")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for Cashier role, got %d", w1.Code)
	}

	// Test Case 2: Manager of DIFFERENT store -> 403 Forbidden
	req2 := httptest.NewRequest("PUT", "/v1/retailer-auth/staff/"+staff.ID, bytes.NewBuffer(reqBody))
	req2.Header.Set("X-User-Role", RoleManager)
	req2.Header.Set("X-Store-ID", "store-002") // Mismatched store
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for store scope mismatch, got %d", w2.Code)
	}

	// Test Case 3: Manager of SAME store -> 200 OK
	req3 := httptest.NewRequest("PUT", "/v1/retailer-auth/staff/"+staff.ID, bytes.NewBuffer(reqBody))
	req3.Header.Set("X-User-Role", RoleManager)
	req3.Header.Set("X-Store-ID", "store-001") // Matching store
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for same store Manager, got %d: %s", w3.Code, w3.Body.String())
	}

	// Verify phone number updated in repository
	updated, _ := repo.GetStaffByID(context.Background(), staff.ID)
	if updated.Phone != phoneUpdate {
		t.Fatalf("expected staff phone to be updated to %s, got %s", phoneUpdate, updated.Phone)
	}
}
