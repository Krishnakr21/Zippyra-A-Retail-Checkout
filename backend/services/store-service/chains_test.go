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

// ── Internal Admin-Write endpoint tests ──────────────────────────────────────

func TestInternalCreateStore_SystemJWT_Success(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewInternalAdminWriteHandler(repo, "dev-secret")

	body, _ := json.Marshal(AdminStoreCreateRequest{
		ChainID:     "chain-x",
		Name:        "Internal Store",
		Address:     "100 Main St",
		City:        "Bengaluru",
		State:       "Karnataka",
		Pincode:     "560001",
		CapacityMax: 50,
		OpeningTime: "08:00:00",
		ClosingTime: "22:00:00",
		Timezone:    "Asia/Kolkata",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/store/internal/admin-write/stores", bytes.NewReader(body))
	req.Header.Set("X-Internal-Service", "admin-store-service")
	w := httptest.NewRecorder()
	handler.HandleCreateStore(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}
	var store Store
	_ = json.Unmarshal(w.Body.Bytes(), &store)
	if store.ID == "" {
		t.Fatal("expected non-empty store ID")
	}
	if store.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE status, got %s", store.Status)
	}
}

func TestInternalCreateStore_NonSystem_Returns401(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewInternalAdminWriteHandler(repo, "dev-secret")

	body, _ := json.Marshal(AdminStoreCreateRequest{Name: "Unauthorized"})
	req := httptest.NewRequest(http.MethodPost, "/v1/store/internal/admin-write/stores", bytes.NewReader(body))
	req.Header.Set("X-User-Type", "STAFF") // not SYSTEM
	w := httptest.NewRecorder()
	handler.HandleCreateStore(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for non-SYSTEM caller, got %d", w.Code)
	}
}

func TestInternalUpdateCapacity_SystemJWT_Success(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewInternalAdminWriteHandler(repo, "dev-secret")

	// Pre-create a store
	_ = repo.CreateStoreInternal(context.Background(), &Store{ID: "store-cap-1", ChainID: "chain-1", Name: "Test", Status: "ACTIVE"})

	body, _ := json.Marshal(AdminCapacityUpdateRequest{CapacityMax: 200})
	req := httptest.NewRequest(http.MethodPut, "/v1/store/internal/admin-write/stores/store-cap-1/capacity", bytes.NewReader(body))
	req.Header.Set("X-Internal-Service", "admin-store-service")
	req = mux.SetURLVars(req, map[string]string{"id": "store-cap-1"})
	w := httptest.NewRecorder()
	handler.HandleUpdateCapacity(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	updated, _ := repo.GetStoreByID(context.Background(), "store-cap-1")
	if updated.CapacityMax != 200 {
		t.Fatalf("expected capacity 200, got %d", updated.CapacityMax)
	}
}

// ── Self-Manage endpoint tests ───────────────────────────────────────────────

func TestSelfManage_ManagerCanUpdateOwnStoreHours(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewSelfManageHandler(repo, "dev-secret")

	_ = repo.CreateStoreInternal(context.Background(), &Store{
		ID: "store-mgr-1", ChainID: "chain-1", Name: "My Store", Status: "ACTIVE",
		OpeningTime: "08:00:00", ClosingTime: "20:00:00", Timezone: "Asia/Kolkata",
	})

	body, _ := json.Marshal(AdminHoursUpdateRequest{
		OpeningTime: "09:00:00",
		ClosingTime: "21:00:00",
		Timezone:    "Asia/Kolkata",
	})
	req := httptest.NewRequest(http.MethodPut, "/v1/store/self-manage/stores/store-mgr-1/hours", bytes.NewReader(body))
	req.Header.Set("X-User-Role", "MANAGER")
	req.Header.Set("X-Store-ID", "store-mgr-1") // scoped to their own store
	req = mux.SetURLVars(req, map[string]string{"id": "store-mgr-1"})
	w := httptest.NewRecorder()
	handler.HandleUpdateHours(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	updated, _ := repo.GetStoreByID(context.Background(), "store-mgr-1")
	if updated.OpeningTime != "09:00:00" {
		t.Fatalf("expected opening_time 09:00:00, got %s", updated.OpeningTime)
	}
}

func TestSelfManage_ManagerCannotUpdateOtherStoreHours(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewSelfManageHandler(repo, "dev-secret")

	_ = repo.CreateStoreInternal(context.Background(), &Store{
		ID: "store-other-1", ChainID: "chain-1", Name: "Other Store", Status: "ACTIVE",
	})

	body, _ := json.Marshal(AdminHoursUpdateRequest{OpeningTime: "10:00:00", ClosingTime: "18:00:00", Timezone: "Asia/Kolkata"})
	req := httptest.NewRequest(http.MethodPut, "/v1/store/self-manage/stores/store-other-1/hours", bytes.NewReader(body))
	req.Header.Set("X-User-Role", "MANAGER")
	req.Header.Set("X-Store-ID", "store-mgr-1") // scoped to a DIFFERENT store
	req = mux.SetURLVars(req, map[string]string{"id": "store-other-1"})
	w := httptest.NewRecorder()
	handler.HandleUpdateHours(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-store access, got %d", w.Code)
	}
}

func TestSelfManage_StaffRole_Returns403(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewSelfManageHandler(repo, "dev-secret")

	_ = repo.CreateStoreInternal(context.Background(), &Store{
		ID: "store-staff-1", ChainID: "chain-1", Name: "Staff Store", Status: "ACTIVE",
	})

	body, _ := json.Marshal(AdminCapacityUpdateRequest{CapacityMax: 999})
	req := httptest.NewRequest(http.MethodPut, "/v1/store/self-manage/stores/store-staff-1/capacity", bytes.NewReader(body))
	req.Header.Set("X-User-Role", "STAFF") // STAFF, not MANAGER
	req.Header.Set("X-Store-ID", "store-staff-1")
	req = mux.SetURLVars(req, map[string]string{"id": "store-staff-1"})
	w := httptest.NewRecorder()
	handler.HandleUpdateCapacity(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for STAFF role, got %d", w.Code)
	}
}

func TestSelfManage_ManagerCanUpdateOwnStoreCapacity(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewSelfManageHandler(repo, "dev-secret")

	_ = repo.CreateStoreInternal(context.Background(), &Store{
		ID: "store-cap-mgr", ChainID: "chain-1", Name: "Cap Store", Status: "ACTIVE", CapacityMax: 50,
	})

	body, _ := json.Marshal(AdminCapacityUpdateRequest{CapacityMax: 150})
	req := httptest.NewRequest(http.MethodPut, "/v1/store/self-manage/stores/store-cap-mgr/capacity", bytes.NewReader(body))
	req.Header.Set("X-User-Role", "MANAGER")
	req.Header.Set("X-Store-ID", "store-cap-mgr")
	req = mux.SetURLVars(req, map[string]string{"id": "store-cap-mgr"})
	w := httptest.NewRecorder()
	handler.HandleUpdateCapacity(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	updated, _ := repo.GetStoreByID(context.Background(), "store-cap-mgr")
	if updated.CapacityMax != 150 {
		t.Fatalf("expected capacity 150, got %d", updated.CapacityMax)
	}
}
