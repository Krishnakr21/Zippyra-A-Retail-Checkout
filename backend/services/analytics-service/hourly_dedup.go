package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type HourlyDedupGuard interface {
	ShouldIncrementHourly(ctx context.Context, orderID string) (bool, error)
}

type RedisHourlyDedupGuard struct {
	client *redis.Client
}

func NewRedisHourlyDedupGuard(client *redis.Client) *RedisHourlyDedupGuard {
	return &RedisHourlyDedupGuard{client: client}
}

func (r *RedisHourlyDedupGuard) ShouldIncrementHourly(ctx context.Context, orderID string) (bool, error) {
	if r.client == nil {
		return true, nil
	}

	key := fmt.Sprintf("analytics_hourly_processed:%s", orderID)
	// SETNX returns true if the key was set (i.e. first time processing order)
	ok, err := r.client.SetNX(ctx, key, "1", 7*24*time.Hour).Result()
	if err != nil {
		return true, nil // Fallback to processing if Redis query fails
	}
	return ok, nil
}

type MemoryHourlyDedupGuard struct {
	mu        sync.RWMutex
	processed map[string]bool
}

func NewMemoryHourlyDedupGuard() *MemoryHourlyDedupGuard {
	return &MemoryHourlyDedupGuard{
		processed: make(map[string]bool),
	}
}

func (m *MemoryHourlyDedupGuard) ShouldIncrementHourly(ctx context.Context, orderID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.processed[orderID] {
		return false, nil // Redelivery detected -> skip second increment
	}
	m.processed[orderID] = true
	return true, nil
}
