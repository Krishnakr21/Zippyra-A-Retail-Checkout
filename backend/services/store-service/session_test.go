package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSession_AutoUnbindPreviousStore(t *testing.T) {
	repo := NewMemoryRepository()
	capacityMgr := NewMemoryCapacityManager()
	sessionMgr := NewSessionManager(repo, capacityMgr, "test-secret-key-32bytes-long-string", nil)
	ctx := context.Background()

	// 1. Create Store A and Store B
	storeA := &Store{
		ID:                   "store-A",
		ChainID:              "chain-1",
		Name:                 "Store A",
		Lat:                  12.9716,
		Lng:                  77.5946,
		GeofenceRadiusMeters: 500,
		CapacityMax:          20,
		Status:               "ACTIVE",
	}
	storeB := &Store{
		ID:                   "store-B",
		ChainID:              "chain-1",
		Name:                 "Store B",
		Lat:                  12.9716,
		Lng:                  77.5946,
		GeofenceRadiusMeters: 500,
		CapacityMax:          20,
		Status:               "ACTIVE",
	}

	_ = repo.CreateStore(ctx, storeA)
	_ = repo.CreateStore(ctx, storeB)

	// Create valid active QR tokens for both stores
	tokenA := &StoreQRToken{ID: uuid.New().String(), StoreID: storeA.ID, GateID: "GATE1", Token: "TOKEN_A_123456", IsActive: true, ExpiresAt: time.Now().Add(1 * time.Hour)}
	tokenB := &StoreQRToken{ID: uuid.New().String(), StoreID: storeB.ID, GateID: "GATE1", Token: "TOKEN_B_123456", IsActive: true, ExpiresAt: time.Now().Add(1 * time.Hour)}
	_ = repo.CreateQRToken(ctx, tokenA)
	_ = repo.CreateQRToken(ctx, tokenB)

	userID := "user-123"

	// 2. Bind User to Store A
	respA, err := sessionMgr.BindStore(ctx, userID, "127.0.0.1", &StoreBindRequest{
		QRToken:  tokenA.Token,
		Lat:      12.9716,
		Lng:      77.5946,
		DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("Failed to bind to Store A: %v", err)
	}
	if respA.StoreID != storeA.ID {
		t.Errorf("Expected store_id %s, got %s", storeA.ID, respA.StoreID)
	}

	capA, _ := capacityMgr.GetLiveCapacity(ctx, storeA.ID)
	if capA != 1 {
		t.Errorf("Expected Store A capacity 1, got %d", capA)
	}

	// 3. Bind User to Store B -> Should auto-unbind Store A
	respB, err := sessionMgr.BindStore(ctx, userID, "127.0.0.1", &StoreBindRequest{
		QRToken:  tokenB.Token,
		Lat:      12.9716,
		Lng:      77.5946,
		DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("Failed to bind to Store B: %v", err)
	}
	if respB.StoreID != storeB.ID {
		t.Errorf("Expected store_id %s, got %s", storeB.ID, respB.StoreID)
	}

	// Verify Store A capacity decremented back to 0
	capAAfter, _ := capacityMgr.GetLiveCapacity(ctx, storeA.ID)
	if capAAfter != 0 {
		t.Errorf("Expected Store A capacity to be 0 after auto-unbind, got %d", capAAfter)
	}

	// Verify Store B capacity incremented to 1
	capBAfter, _ := capacityMgr.GetLiveCapacity(ctx, storeB.ID)
	if capBAfter != 1 {
		t.Errorf("Expected Store B capacity to be 1, got %d", capBAfter)
	}

	// Verify user active session is at Store B
	activeSess, err := repo.GetActiveSessionByUser(ctx, userID)
	if err != nil || activeSess == nil {
		t.Fatalf("Expected active session for user, got nil")
	}
	if activeSess.StoreID != storeB.ID {
		t.Errorf("Expected active session store_id to be %s, got %s", storeB.ID, activeSess.StoreID)
	}
}

func TestSession_AutoExpireJob(t *testing.T) {
	repo := NewMemoryRepository()
	capacityMgr := NewMemoryCapacityManager()
	sessionMgr := NewSessionManager(repo, capacityMgr, "test-secret", nil)
	ctx := context.Background()

	store := &Store{ID: "store-expire-test", CapacityMax: 50}
	_ = repo.CreateStore(ctx, store)

	// Create a stale session bound 4 hours ago
	staleSess := &StoreSession{
		ID:                   "stale-sess-1",
		UserID:               "stale-user",
		StoreID:              store.ID,
		BoundAt:              time.Now().Add(-4 * time.Hour),
		CatalogVersionAtBind: 1,
	}
	_ = repo.CreateSession(ctx, staleSess)
	_, _, _ = capacityMgr.TryIncrementCapacity(ctx, store.ID, store.CapacityMax)

	capBefore, _ := capacityMgr.GetLiveCapacity(ctx, store.ID)
	if capBefore != 1 {
		t.Fatalf("Expected capacity 1 before expire job, got %d", capBefore)
	}

	// Run auto-expire job
	sessionMgr.AutoExpireStaleSessionsJob(ctx)

	// Verify session is unbound
	activeSess, _ := repo.GetActiveSessionByUser(ctx, "stale-user")
	if activeSess != nil {
		t.Errorf("Expected stale session to be unbound, but found active session")
	}

	capAfter, _ := capacityMgr.GetLiveCapacity(ctx, store.ID)
	if capAfter != 0 {
		t.Errorf("Expected capacity 0 after auto-expire, got %d", capAfter)
	}
}

func TestAdmin_InternalWriteSystemJWTGated(t *testing.T) {
	repo := NewMemoryRepository()
	internalHandler := NewInternalAdminWriteHandler(repo, "test-secret")

	ctx := context.Background()

	// Store belongs to Chain A
	storeInChainA := &Store{
		ID:      "store-chain-A",
		ChainID: "chain-A",
		Name:    "Chain A Store",
		Status:  "ACTIVE",
	}
	_ = repo.CreateStoreInternal(ctx, storeInChainA)

	// Caller without SYSTEM JWT attempts internal write
	reqBody := `{"opening_time":"09:00:00","closing_time":"21:00:00","timezone":"Asia/Kolkata"}`
	req := httptest.NewRequest(http.MethodPut, "/v1/store/internal/admin-write/stores/store-chain-A/hours", strings.NewReader(reqBody))
	req.Header.Set("X-User-Type", "STAFF")

	w := httptest.NewRecorder()
	internalHandler.HandleUpdateHours(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for non-SYSTEM call, got %d", w.Code)
	}
}
