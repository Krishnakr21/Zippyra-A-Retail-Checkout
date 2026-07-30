package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type MockCatalogEngine struct {
	products map[string]*ProductDTO
}

func (m *MockCatalogEngine) GetProductByBarcode(ctx context.Context, storeID, barcode string) (*ProductDTO, error) {
	if p, ok := m.products[barcode]; ok {
		return p, nil
	}
	return &ProductDTO{
		ID:         "p1",
		StoreID:    storeID,
		Barcode:    barcode,
		Name:       "Test Coffee",
		PricePaise: 25000,
		HSNCode:    "0901",
	}, nil
}

func TestScan_CumulativeHold_And_OutOfStock(t *testing.T) {
	holdMgr := NewMemoryHoldManager()
	cartStore := NewRedisCartStore(nil)
	offerEngine := NewMemoryOfferEngine()
	checkoutRepo := NewMemoryCheckoutRepository()
	lockMgr := NewMemoryLockManager()
	catalogEngine := &MockCatalogEngine{products: make(map[string]*ProductDTO)}

	handler := NewCartHandler(cartStore, holdMgr, offerEngine, checkoutRepo, lockMgr, catalogEngine, nil, "test-secret")

	ctx := context.Background()
	storeID := "store-1"
	userID := "user-1"
	barcode := "8901030300011" // Valid EAN-13

	// Set available_qty = 3 in stock
	_ = holdMgr.SetAvailableQty(ctx, storeID, barcode, 3)

	// 1. Scan 1st item (qty=2) -> Success
	body1 := `{"barcode":"8901030300011", "qty":2}`
	req1 := httptest.NewRequest(http.MethodPost, "/v1/cart/scan", strings.NewReader(body1))
	req1.Header.Set("X-User-ID", userID)
	req1.Header.Set("X-Store-ID", storeID)
	w1 := httptest.NewRecorder()

	handler.HandleScanItem(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on first scan, got %d: %s", w1.Code, w1.Body.String())
	}

	// Assert hold quantity is 2
	items, _, _ := cartStore.GetCart(ctx, storeID, userID)
	if len(items) != 1 || items[0].Qty != 2 {
		t.Errorf("Expected cart qty 2, got %v", items)
	}

	// 2. Scan 2nd item (qty=2) -> total requested = 2 + 2 = 4 > available (3) -> 409 OUT_OF_STOCK
	body2 := `{"barcode":"8901030300011", "qty":2}`
	req2 := httptest.NewRequest(http.MethodPost, "/v1/cart/scan", strings.NewReader(body2))
	req2.Header.Set("X-User-ID", userID)
	req2.Header.Set("X-Store-ID", storeID)
	w2 := httptest.NewRecorder()

	handler.HandleScanItem(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for out of stock scan, got %d", w2.Code)
	}

	// Assert cart quantity remains 2 (no partial increment occurred)
	itemsAfter, _, _ := cartStore.GetCart(ctx, storeID, userID)
	if len(itemsAfter) != 1 || itemsAfter[0].Qty != 2 {
		t.Errorf("Expected cart qty to remain 2 after failed scan, got %d", itemsAfter[0].Qty)
	}
}

func TestConcurrentScans_RaceConditionOverflowProtection(t *testing.T) {
	holdMgr := NewMemoryHoldManager()
	ctx := context.Background()
	storeID := "store-race-1"
	barcode := "8901030300011"

	// Stock available_qty = 10
	_ = holdMgr.SetAvailableQty(ctx, storeID, barcode, 10)

	// Launch 50 concurrent scans of qty=1 each across different users
	const concurrentUsers = 50
	var successfulScans int64
	var failedScans int64

	var wg sync.WaitGroup
	wg.Add(concurrentUsers)

	for i := 0; i < concurrentUsers; i++ {
		userID := fmt.Sprintf("user-%d", i)
		go func(uid string) {
			defer wg.Done()
			err := holdMgr.CheckStockAndReserveHold(ctx, storeID, uid, barcode, 1)
			if err == nil {
				atomic.AddInt64(&successfulScans, 1)
			} else {
				atomic.AddInt64(&failedScans, 1)
			}
		}(userID)
	}

	wg.Wait()

	// Assert NEVER more than 10 successful reservations
	if successfulScans != 10 {
		t.Errorf("Expected exactly 10 successful holds out of 50 concurrent requests, got %d (failed: %d)", successfulScans, failedScans)
	}

	if failedScans != 40 {
		t.Errorf("Expected 40 rejected requests, got %d", failedScans)
	}
}
