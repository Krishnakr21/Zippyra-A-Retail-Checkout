package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestShiftStart_DuplicateActiveShift_Returns409ShiftAlreadyActive(t *testing.T) {
	repo := NewMemoryRepository()
	shiftSvc := NewShiftService(repo)
	shiftH := NewShiftHandler(shiftSvc, "dev-secret")

	staffID := "staff-shift-1"
	storeID := "store-1"

	// First start -> 201 Created
	req1 := httptest.NewRequest("POST", "/v1/retailer-auth/shift/start", nil)
	req1.Header.Set("X-User-ID", staffID)
	req1.Header.Set("X-Store-ID", storeID)
	req1.Header.Set("X-User-Role", "CASHIER")
	w1 := httptest.NewRecorder()
	shiftH.HandleStartShift(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first shift start: expected 201 Created, got %d", w1.Code)
	}

	// Second start while active -> 409 Conflict
	req2 := httptest.NewRequest("POST", "/v1/retailer-auth/shift/start", nil)
	req2.Header.Set("X-User-ID", staffID)
	req2.Header.Set("X-Store-ID", storeID)
	req2.Header.Set("X-User-Role", "CASHIER")
	w2 := httptest.NewRecorder()
	shiftH.HandleStartShift(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("second shift start: expected 409 Conflict, got %d", w2.Code)
	}
}

func TestShiftStart_RaceCondition_OnlyOneShiftCreated(t *testing.T) {
	repo := NewMemoryRepository()
	shiftSvc := NewShiftService(repo)
	shiftH := NewShiftHandler(shiftSvc, "dev-secret")

	staffID := "staff-race-1"
	storeID := "store-1"

	var wg sync.WaitGroup
	results := make(chan int, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("POST", "/v1/retailer-auth/shift/start", nil)
			req.Header.Set("X-User-ID", staffID)
			req.Header.Set("X-Store-ID", storeID)
			req.Header.Set("X-User-Role", "CASHIER")
			w := httptest.NewRecorder()
			shiftH.HandleStartShift(w, req)
			results <- w.Code
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	conflictCount := 0
	for code := range results {
		if code == http.StatusCreated {
			successCount++
		} else if code == http.StatusConflict {
			conflictCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 shift start to succeed, got %d", successCount)
	}
	if conflictCount != 9 {
		t.Fatalf("expected 9 shift starts to get 409 Conflict, got %d", conflictCount)
	}
}
