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

func TestDeactivateStaff_RevokesSessionAndEndsShiftInSameCall(t *testing.T) {
	repo := NewMemoryRepository()
	staffH := NewStaffHandler(repo, "dev-secret")

	// 1. Create active staff
	staff := &StaffMember{
		ID:       "staff-deact-1",
		StoreID:  "store-1",
		ChainID:  "chain-1",
		Phone:    "+919876511111",
		Name:     "Test Staff",
		Role:     "CASHIER",
		IsActive: true,
	}
	_ = repo.CreateStaffMember(context.Background(), staff)

	// 2. Create active session
	sess := &StaffSession{
		ID:         "sess-active-1",
		StaffID:    staff.ID,
		AuthMethod: AuthMethodOTP,
	}
	_ = repo.CreateSession(context.Background(), sess)

	// 3. Create active shift
	shift := &StaffShift{
		ID:      "shift-active-1",
		StaffID: staff.ID,
		StoreID: "store-1",
	}
	_ = repo.StartShiftTx(context.Background(), shift)

	// 4. Perform DELETE /v1/retailer-auth/staff/{id} as MANAGER of store-1
	req := httptest.NewRequest("DELETE", "/v1/retailer-auth/staff/staff-deact-1", nil)
	req.Header.Set("X-User-ID", "manager-1")
	req.Header.Set("X-Store-ID", "store-1")
	req.Header.Set("X-User-Role", "MANAGER")
	req = mux.SetURLVars(req, map[string]string{"id": "staff-deact-1"})
	w := httptest.NewRecorder()

	staffH.HandleDeactivateStaff(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on staff deactivation, got %d", w.Code)
	}

	// 5. Verify staff deactivated
	deactStaff, _ := repo.GetStaffByID(context.Background(), staff.ID)
	if deactStaff.IsActive {
		t.Fatalf("expected staff to be deactivated")
	}

	// 6. Verify session revoked
	revokedSess, _ := repo.GetSessionByID(context.Background(), sess.ID)
	if revokedSess.RevokedAt == nil {
		t.Fatalf("expected session to be revoked on staff deactivation")
	}

	// 7. Verify shift ended
	activeShift, _ := repo.GetActiveShift(context.Background(), staff.ID)
	if activeShift != nil {
		t.Fatalf("expected shift to be ended on staff deactivation")
	}
}

func TestStoreScopeMismatch_ManagerStoreA_AttemptingActionsOnStoreB_Returns403(t *testing.T) {
	repo := NewMemoryRepository()
	staffH := NewStaffHandler(repo, "dev-secret")

	// Store B staff
	staffB := &StaffMember{
		ID:       "staff-b-1",
		StoreID:  "store-B",
		ChainID:  "chain-1",
		Phone:    "+919876522222",
		Name:     "Store B Staff",
		Role:     "CASHIER",
		IsActive: true,
	}
	_ = repo.CreateStaffMember(context.Background(), staffB)

	// Manager of Store A attempting to create staff for Store B
	bodyCreate, _ := json.Marshal(map[string]string{
		"store_id": "store-B",
		"phone":    "+919876533333",
		"name":     "New Staff",
		"role":     "CASHIER",
	})
	reqCreate := httptest.NewRequest("POST", "/v1/retailer-auth/staff", bytes.NewReader(bodyCreate))
	reqCreate.Header.Set("X-User-ID", "manager-A")
	reqCreate.Header.Set("X-Store-ID", "store-A")
	reqCreate.Header.Set("X-User-Role", "MANAGER")
	wCreate := httptest.NewRecorder()
	staffH.HandleCreateStaff(wCreate, reqCreate)

	if wCreate.Code != http.StatusForbidden {
		t.Fatalf("Create staff: expected 403 Forbidden for Store Scope Mismatch, got %d", wCreate.Code)
	}

	// Manager of Store A attempting to list staff for Store B
	reqList := httptest.NewRequest("GET", "/v1/retailer-auth/staff?store_id=store-B", nil)
	reqList.Header.Set("X-User-ID", "manager-A")
	reqList.Header.Set("X-Store-ID", "store-A")
	reqList.Header.Set("X-User-Role", "MANAGER")
	wList := httptest.NewRecorder()
	staffH.HandleListStaff(wList, reqList)

	if wList.Code != http.StatusForbidden {
		t.Fatalf("List staff: expected 403 Forbidden for Store Scope Mismatch, got %d", wList.Code)
	}

	// Manager of Store A attempting to edit staff for Store B
	bodyEdit, _ := json.Marshal(map[string]string{"name": "Hacked Name"})
	reqEdit := httptest.NewRequest("PUT", "/v1/retailer-auth/staff/staff-b-1", bytes.NewReader(bodyEdit))
	reqEdit.Header.Set("X-User-ID", "manager-A")
	reqEdit.Header.Set("X-Store-ID", "store-A")
	reqEdit.Header.Set("X-User-Role", "MANAGER")
	reqEdit = mux.SetURLVars(reqEdit, map[string]string{"id": "staff-b-1"})
	wEdit := httptest.NewRecorder()
	staffH.HandleUpdateStaff(wEdit, reqEdit)

	if wEdit.Code != http.StatusForbidden {
		t.Fatalf("Edit staff: expected 403 Forbidden for Store Scope Mismatch, got %d", wEdit.Code)
	}

	// Manager of Store A attempting to deactivate staff for Store B
	reqDeact := httptest.NewRequest("DELETE", "/v1/retailer-auth/staff/staff-b-1", nil)
	reqDeact.Header.Set("X-User-ID", "manager-A")
	reqDeact.Header.Set("X-Store-ID", "store-A")
	reqDeact.Header.Set("X-User-Role", "MANAGER")
	reqDeact = mux.SetURLVars(reqDeact, map[string]string{"id": "staff-b-1"})
	wDeact := httptest.NewRecorder()
	staffH.HandleDeactivateStaff(wDeact, reqDeact)

	if wDeact.Code != http.StatusForbidden {
		t.Fatalf("Deactivate staff: expected 403 Forbidden for Store Scope Mismatch, got %d", wDeact.Code)
	}
}
