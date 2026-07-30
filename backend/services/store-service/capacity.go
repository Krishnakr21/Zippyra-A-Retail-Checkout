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

type CapacityManager interface {
	TryIncrementCapacity(ctx context.Context, storeID string, maxCapacity int) (bool, int, error)
	DecrementCapacity(ctx context.Context, storeID string) (int, error)
	GetLiveCapacity(ctx context.Context, storeID string) (int, error)
	CheckRateLimits(ctx context.Context, userID, clientIP string) error
}

type RedisCapacityManager struct {
	client *redis.Client
}

func NewRedisCapacityManager(client *redis.Client) CapacityManager {
	return &RedisCapacityManager{client: client}
}

func (r *RedisCapacityManager) TryIncrementCapacity(ctx context.Context, storeID string, maxCapacity int) (bool, int, error) {
	if r.client == nil {
		return true, 1, nil
	}

	key := fmt.Sprintf("capacity:%s", storeID)
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	if int(val) > maxCapacity {
		// Immediately rollback overflow increment
		_ = r.client.Decr(ctx, key).Val()
		return false, int(val) - 1, nil
	}

	return true, int(val), nil
}

func (r *RedisCapacityManager) DecrementCapacity(ctx context.Context, storeID string) (int, error) {
	if r.client == nil {
		return 0, nil
	}

	key := fmt.Sprintf("capacity:%s", storeID)
	val, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if val < 0 {
		_ = r.client.Set(ctx, key, 0, 0).Err()
		return 0, nil
	}

	return int(val), nil
}

func (r *RedisCapacityManager) GetLiveCapacity(ctx context.Context, storeID string) (int, error) {
	if r.client == nil {
		return 0, nil
	}

	key := fmt.Sprintf("capacity:%s", storeID)
	valStr, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	count, _ := strconv.Atoi(valStr)
	if count < 0 {
		count = 0
	}
	return count, nil
}

func (r *RedisCapacityManager) CheckRateLimits(ctx context.Context, userID, clientIP string) error {
	if r.client == nil {
		return nil
	}

	// 1. User rate limit: 20 per hour
	if userID != "" {
		userKey := fmt.Sprintf("qr_bind_ratelimit:%s", userID)
		userCount, err := r.client.Incr(ctx, userKey).Result()
		if err == nil {
			if userCount == 1 {
				_ = r.client.Expire(ctx, userKey, 1*time.Hour).Err()
			}
			if userCount > 20 {
				return errors.NewAPIError(errors.CodeRateLimitExceeded, "Store bind rate limit exceeded for user (max 20/hour)", nil)
			}
		}
	}

	// 2. IP rate limit: 60 per hour
	if clientIP != "" {
		ipKey := fmt.Sprintf("qr_bind_ratelimit:ip:%s", clientIP)
		ipCount, err := r.client.Incr(ctx, ipKey).Result()
		if err == nil {
			if ipCount == 1 {
				_ = r.client.Expire(ctx, ipKey, 1*time.Hour).Err()
			}
			if ipCount > 60 {
				return errors.NewAPIError(errors.CodeRateLimitExceeded, "Store bind rate limit exceeded for IP (max 60/hour)", nil)
			}
		}
	}

	return nil
}

type MemoryCapacityManager struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewMemoryCapacityManager() CapacityManager {
	return &MemoryCapacityManager{counts: make(map[string]int)}
}

func (m *MemoryCapacityManager) TryIncrementCapacity(ctx context.Context, storeID string, maxCapacity int) (bool, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.counts[storeID]
	if current >= maxCapacity {
		return false, current, nil
	}
	m.counts[storeID] = current + 1
	return true, current + 1, nil
}

func (m *MemoryCapacityManager) DecrementCapacity(ctx context.Context, storeID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.counts[storeID]
	if current <= 0 {
		m.counts[storeID] = 0
		return 0, nil
	}
	m.counts[storeID] = current - 1
	return current - 1, nil
}

func (m *MemoryCapacityManager) GetLiveCapacity(ctx context.Context, storeID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counts[storeID], nil
}

func (m *MemoryCapacityManager) CheckRateLimits(ctx context.Context, userID, clientIP string) error {
	return nil
}
