package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCapacity_ConcurrentOverflowProtection(t *testing.T) {
	capacityMgr := NewMemoryCapacityManager()
	ctx := context.Background()

	storeID := "store-cap-test"
	maxCapacity := 10

	var successCount int32
	var failCount int32

	var wg sync.WaitGroup
	concurrentRequests := 50

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, _, err := capacityMgr.TryIncrementCapacity(ctx, storeID, maxCapacity)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if ok {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}()
	}

	wg.Wait()

	if successCount != int32(maxCapacity) {
		t.Errorf("Expected exactly %d successful increments, got %d", maxCapacity, successCount)
	}

	if failCount != int32(concurrentRequests-maxCapacity) {
		t.Errorf("Expected %d failed increments due to capacity limit, got %d", concurrentRequests-maxCapacity, failCount)
	}

	finalCap, _ := capacityMgr.GetLiveCapacity(ctx, storeID)
	if finalCap > maxCapacity {
		t.Errorf("Live capacity counter (%d) exceeded maxCapacity (%d)", finalCap, maxCapacity)
	}
}
