package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zippyra/backend/shared/errors"
)

type HoldManager interface {
	CheckStockAndReserveHold(ctx context.Context, storeID, userID, barcode string, deltaQty int) error
	ReleaseHold(ctx context.Context, storeID, userID, barcode string, qtyToRelease int) error
	ReleaseAllUserHolds(ctx context.Context, storeID, userID string, items []*CartItem) error
	GetAvailableQty(ctx context.Context, storeID, barcode string) (int, error)
	SetAvailableQty(ctx context.Context, storeID, barcode string, qty int) error
}

type RedisHoldManager struct {
	rdb              redis.Cmdable
	reserveLuaSha    string
	reserveLuaScript *redis.Script
}

var reserveLua = redis.NewScript(`
	local availKey = KEYS[1]
	local globalHoldKey = KEYS[2]
	local ownerHoldKey = KEYS[3]

	local delta = tonumber(ARGV[1])

	local availVal = redis.call('GET', availKey)
	if not availVal then
		return -1 -- Stock key missing
	end
	local avail = tonumber(availVal)

	local currentHold = tonumber(redis.call('GET', globalHoldKey) or '0')

	if (currentHold + delta) > avail then
		return 0 -- OUT_OF_STOCK
	end

	redis.call('INCRBY', globalHoldKey, delta)
	redis.call('INCRBY', ownerHoldKey, delta)
	redis.call('EXPIRE', ownerHoldKey, 1800)
	return 1
`)

func NewRedisHoldManager(rdb redis.Cmdable) HoldManager {
	return &RedisHoldManager{
		rdb:              rdb,
		reserveLuaScript: reserveLua,
	}
}

func (h *RedisHoldManager) CheckStockAndReserveHold(ctx context.Context, storeID, userID, barcode string, deltaQty int) error {
	if deltaQty <= 0 {
		return nil
	}

	availKey := fmt.Sprintf("available_qty:%s:%s", storeID, barcode)
	globalHoldKey := fmt.Sprintf("hold:%s:%s", storeID, barcode)
	ownerHoldKey := fmt.Sprintf("hold_owner:%s:%s:%s", storeID, userID, barcode)

	keys := []string{availKey, globalHoldKey, ownerHoldKey}
	res, err := h.reserveLuaScript.Run(ctx, h.rdb, keys, deltaQty).Result()
	if err != nil {
		return errors.NewAPIError(errors.CodeInternalError, "Failed to execute hold reservation script", nil)
	}

	resInt, _ := res.(int64)
	if resInt == 0 || resInt == -1 {
		return errors.NewAPIError(errors.CodeOutOfStock, fmt.Sprintf("Item %s is out of stock", barcode), nil)
	}

	return nil
}

func (h *RedisHoldManager) ReleaseHold(ctx context.Context, storeID, userID, barcode string, qtyToRelease int) error {
	if qtyToRelease <= 0 {
		return nil
	}

	globalHoldKey := fmt.Sprintf("hold:%s:%s", storeID, barcode)
	ownerHoldKey := fmt.Sprintf("hold_owner:%s:%s:%s", storeID, userID, barcode)

	// Get current user's held portion
	ownerValStr, err := h.rdb.Get(ctx, ownerHoldKey).Result()
	if err == redis.Nil || ownerValStr == "" {
		return nil
	}

	ownerVal, _ := strconv.Atoi(ownerValStr)
	if qtyToRelease > ownerVal {
		qtyToRelease = ownerVal
	}

	pipe := h.rdb.Pipeline()
	pipe.DecrBy(ctx, globalHoldKey, int64(qtyToRelease))
	pipe.DecrBy(ctx, ownerHoldKey, int64(qtyToRelease))
	pipe.Expire(ctx, ownerHoldKey, 1800*time.Second)
	_, _ = pipe.Exec(ctx)

	return nil
}

func (h *RedisHoldManager) ReleaseAllUserHolds(ctx context.Context, storeID, userID string, items []*CartItem) error {
	for _, item := range items {
		_ = h.ReleaseHold(ctx, storeID, userID, item.Barcode, item.Qty)
	}
	return nil
}

func (h *RedisHoldManager) GetAvailableQty(ctx context.Context, storeID, barcode string) (int, error) {
	availKey := fmt.Sprintf("available_qty:%s:%s", storeID, barcode)
	val, err := h.rdb.Get(ctx, availKey).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func (h *RedisHoldManager) SetAvailableQty(ctx context.Context, storeID, barcode string, qty int) error {
	availKey := fmt.Sprintf("available_qty:%s:%s", storeID, barcode)
	return h.rdb.Set(ctx, availKey, qty, 0).Err()
}

// MemoryHoldManager for tests
type MemoryHoldManager struct {
	mu           sync.Mutex
	availableQty map[string]int
	globalHold   map[string]int
	ownerHold    map[string]int
}

func NewMemoryHoldManager() HoldManager {
	return &MemoryHoldManager{
		availableQty: make(map[string]int),
		globalHold:   make(map[string]int),
		ownerHold:    make(map[string]int),
	}
}

func (m *MemoryHoldManager) CheckStockAndReserveHold(ctx context.Context, storeID, userID, barcode string, deltaQty int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", storeID, barcode)
	avail := m.availableQty[key]
	currentHold := m.globalHold[key]

	if (currentHold + deltaQty) > avail {
		return errors.NewAPIError(errors.CodeOutOfStock, fmt.Sprintf("Item %s is out of stock", barcode), nil)
	}

	m.globalHold[key] += deltaQty
	ownerKey := fmt.Sprintf("%s:%s:%s", storeID, userID, barcode)
	m.ownerHold[ownerKey] += deltaQty
	return nil
}

func (m *MemoryHoldManager) ReleaseHold(ctx context.Context, storeID, userID, barcode string, qtyToRelease int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ownerKey := fmt.Sprintf("%s:%s:%s", storeID, userID, barcode)
	ownerVal := m.ownerHold[ownerKey]
	if ownerVal <= 0 {
		return nil
	}

	if qtyToRelease > ownerVal {
		qtyToRelease = ownerVal
	}

	key := fmt.Sprintf("%s:%s", storeID, barcode)
	m.globalHold[key] -= qtyToRelease
	if m.globalHold[key] < 0 {
		m.globalHold[key] = 0
	}

	m.ownerHold[ownerKey] -= qtyToRelease
	if m.ownerHold[ownerKey] <= 0 {
		delete(m.ownerHold, ownerKey)
	}
	return nil
}

func (m *MemoryHoldManager) ReleaseAllUserHolds(ctx context.Context, storeID, userID string, items []*CartItem) error {
	for _, item := range items {
		_ = m.ReleaseHold(ctx, storeID, userID, item.Barcode, item.Qty)
	}
	return nil
}

func (m *MemoryHoldManager) GetAvailableQty(ctx context.Context, storeID, barcode string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", storeID, barcode)
	return m.availableQty[key], nil
}

func (m *MemoryHoldManager) SetAvailableQty(ctx context.Context, storeID, barcode string, qty int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", storeID, barcode)
	m.availableQty[key] = qty
	return nil
}
