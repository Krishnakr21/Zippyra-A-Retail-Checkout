package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type TestRedisClient struct {
	mu    sync.RWMutex
	store map[string]string
}

func NewTestRedisClient() *TestRedisClient {
	return &TestRedisClient{
		store: make(map[string]string),
	}
}

func (m *TestRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewBoolCmd(ctx)
	if _, exists := m.store[key]; exists {
		cmd.SetVal(false)
		return cmd
	}
	m.store[key] = fmt.Sprintf("%v", value)
	cmd.SetVal(true)
	return cmd
}

func (m *TestRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmd := redis.NewStringCmd(ctx)
	val, exists := m.store[key]
	if !exists {
		cmd.SetErr(redis.Nil)
		return cmd
	}
	cmd.SetVal(val)
	return cmd
}

func (m *TestRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewStatusCmd(ctx)
	m.store[key] = fmt.Sprintf("%v", value)
	cmd.SetVal("OK")
	return cmd
}

func (m *TestRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := redis.NewIntCmd(ctx)
	var count int64 = 0
	for _, k := range keys {
		if _, exists := m.store[k]; exists {
			delete(m.store, k)
			count++
		}
	}
	cmd.SetVal(count)
	return cmd
}

func (m *TestRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	cmd.SetVal("PONG")
	return cmd
}

func (m *TestRedisClient) Close() error {
	return nil
}
