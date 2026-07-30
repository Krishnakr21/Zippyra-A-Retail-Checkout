package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type DynamicMockCatalogEngine struct {
	livePrices map[string]int64
}

func (d *DynamicMockCatalogEngine) GetProductByBarcode(ctx context.Context, storeID, barcode string) (*ProductDTO, error) {
	price := int64(25000)
	if p, ok := d.livePrices[barcode]; ok {
		price = p
	}
	return &ProductDTO{
		ID:         "p1",
		StoreID:    storeID,
		Barcode:    barcode,
		Name:       "Price Test Coffee",
		PricePaise: price,
		HSNCode:    "0901",
	}, nil
}

func TestCheckoutInit_StalePriceSnapshot_Rejection(t *testing.T) {
	holdMgr := NewMemoryHoldManager()
	cartStore := NewRedisCartStore(nil)
	offerEngine := NewMemoryOfferEngine()
	checkoutRepo := NewMemoryCheckoutRepository()
	lockMgr := NewMemoryLockManager()
	mockCatalog := &DynamicMockCatalogEngine{livePrices: make(map[string]int64)}

	handler := NewCartHandler(cartStore, holdMgr, offerEngine, checkoutRepo, lockMgr, mockCatalog, nil, "test-secret")

	ctx := context.Background()
	storeID := "store-price-1"
	userID := "user-price-1"
	barcode := "8901030300011"

	_ = holdMgr.SetAvailableQty(ctx, storeID, barcode, 10)

	// Item scanned at ₹250.00 (25000 paise)
	cartItem := &CartItem{
		Barcode:            barcode,
		Name:               "Price Test Coffee",
		Qty:                1,
		PricePaiseSnapshot: 25000,
		LineTotalPaise:     25000,
		HSNCode:            "0901",
	}
	_ = cartStore.UpsertCartItem(ctx, storeID, userID, cartItem)

	// Admin updates live price to ₹300.00 (30000 paise) in catalog-service
	mockCatalog.livePrices[barcode] = 30000

	// User attempts checkout init
	req := httptest.NewRequest(http.MethodPost, "/v1/cart/checkout/init", nil)
	req.Header.Set("X-User-ID", userID)
	req.Header.Set("X-Store-ID", storeID)
	w := httptest.NewRecorder()

	handler.HandleCheckoutInit(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 Conflict for stale price snapshot checkout, got %d", w.Code)
	}

	// Assert lock was released so user is not stuck
	locked, _ := lockMgr.AcquireCheckoutLock(ctx, userID)
	if !locked {
		t.Errorf("Expected checkout lock to be released after price change rejection")
	}
}

func TestCheckoutSession_ConsumeClearsCartAndHolds(t *testing.T) {
	holdMgr := NewMemoryHoldManager()
	cartStore := NewRedisCartStore(nil)
	checkoutRepo := NewMemoryCheckoutRepository()
	lockMgr := NewMemoryLockManager()

	internalHandler := NewInternalCartHandler(checkoutRepo, cartStore, holdMgr, lockMgr, "test-secret")

	ctx := context.Background()
	storeID := "store-consume-1"
	userID := "user-consume-1"
	barcode := "8901030300011"

	// Seed item in cart & hold
	_ = holdMgr.SetAvailableQty(ctx, storeID, barcode, 10)
	_ = holdMgr.CheckStockAndReserveHold(ctx, storeID, userID, barcode, 2)

	cartItem := &CartItem{
		Barcode:            barcode,
		Name:               "Consume Coffee",
		Qty:                2,
		PricePaiseSnapshot: 20000,
		LineTotalPaise:     40000,
		HSNCode:            "0901",
	}
	_ = cartStore.UpsertCartItem(ctx, storeID, userID, cartItem)
	_, _ = lockMgr.AcquireCheckoutLock(ctx, userID)

	// Create pending checkout session
	session := &CheckoutSession{
		ID:            "sess-consume-100",
		UserID:        userID,
		StoreID:       storeID,
		Items:         []*CartItem{cartItem},
		SubtotalPaise: 40000,
		TotalPaise:    42000,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(10 * time.Minute),
	}
	_ = checkoutRepo.CreateCheckoutSession(ctx, session)

	// Call POST /v1/cart/internal/checkout-session/sess-consume-100/consume
	req := httptest.NewRequest(http.MethodPost, "/v1/cart/internal/checkout-session/sess-consume-100/consume", nil)
	req.Header.Set("X-Internal-API-Key", "zippyra-internal-secret-key-32bytes")
	w := httptest.NewRecorder()

	internalHandler.HandleConsumeCheckoutSession(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on checkout session consumption, got %d: %s", w.Code, w.Body.String())
	}

	// Assert cart is cleared
	items, _, _ := cartStore.GetCart(ctx, storeID, userID)
	if len(items) != 0 {
		t.Errorf("Expected cart to be cleared after consumption, got %d items", len(items))
	}

	// Assert lock is released
	locked, _ := lockMgr.AcquireCheckoutLock(ctx, userID)
	if !locked {
		t.Errorf("Expected checkout lock to be released after consumption")
	}
}

func TestBackgroundCleaner_ReleasesExpiredSessionLock(t *testing.T) {
	checkoutRepo := NewMemoryCheckoutRepository()
	lockMgr := NewMemoryLockManager()

	ctx := context.Background()
	userID := "user-expired-lock"
	storeID := "store-1"

	// Create already expired checkout session
	session := &CheckoutSession{
		ID:            "sess-stale-1",
		UserID:        userID,
		StoreID:       storeID,
		Items:         []*CartItem{},
		SubtotalPaise: 1000,
		TotalPaise:    1000,
		Status:        "PENDING",
		CreatedAt:     time.Now().Add(-15 * time.Minute),
		ExpiresAt:     time.Now().Add(-5 * time.Minute), // Expired 5 mins ago
	}
	_ = checkoutRepo.CreateCheckoutSession(ctx, session)
	_, _ = lockMgr.AcquireCheckoutLock(ctx, userID)

	// Run expire stale sessions logic
	expired, err := checkoutRepo.ExpireStaleSessions(ctx)
	if err != nil || len(expired) != 1 {
		t.Fatalf("Expected 1 expired session, got %d", len(expired))
	}

	for _, s := range expired {
		_ = lockMgr.ReleaseCheckoutLock(ctx, s.UserID)
	}

	// Assert lock is released so customer can scan/checkout again
	locked, _ := lockMgr.AcquireCheckoutLock(ctx, userID)
	if !locked {
		t.Errorf("Expected expired session lock to be released for retry")
	}
}
