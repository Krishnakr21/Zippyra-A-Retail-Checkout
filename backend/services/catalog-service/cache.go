package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheManager interface {
	GetSKU(ctx context.Context, storeID, barcode string) (*Product, error)
	SetSKU(ctx context.Context, storeID, barcode string, product *Product) error
	DeleteSKU(ctx context.Context, storeID, barcode string) error
	GetCategoryTree(ctx context.Context, chainID string) ([]*Category, error)
	SetCategoryTree(ctx context.Context, chainID string, categories []*Category) error
}

type RedisCacheManager struct {
	client *redis.Client
}

func NewRedisCacheManager(client *redis.Client) CacheManager {
	return &RedisCacheManager{client: client}
}

func (r *RedisCacheManager) GetSKU(ctx context.Context, storeID, barcode string) (*Product, error) {
	if r.client == nil {
		return nil, nil
	}

	key := fmt.Sprintf("sku:%s:%s", storeID, barcode)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var product Product
	if err := json.Unmarshal([]byte(val), &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *RedisCacheManager) SetSKU(ctx context.Context, storeID, barcode string, product *Product) error {
	if r.client == nil || product == nil {
		return nil
	}

	key := fmt.Sprintf("sku:%s:%s", storeID, barcode)
	val, err := json.Marshal(product)
	if err != nil {
		return err
	}

	// 6 Hours TTL write-through SKU cache
	return r.client.Set(ctx, key, val, 6*time.Hour).Err()
}

func (r *RedisCacheManager) DeleteSKU(ctx context.Context, storeID, barcode string) error {
	if r.client == nil {
		return nil
	}
	key := fmt.Sprintf("sku:%s:%s", storeID, barcode)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCacheManager) GetCategoryTree(ctx context.Context, chainID string) ([]*Category, error) {
	if r.client == nil {
		return nil, nil
	}

	key := fmt.Sprintf("categories:%s", chainID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var categories []*Category
	if err := json.Unmarshal([]byte(val), &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *RedisCacheManager) SetCategoryTree(ctx context.Context, chainID string, categories []*Category) error {
	if r.client == nil {
		return nil
	}

	key := fmt.Sprintf("categories:%s", chainID)
	val, err := json.Marshal(categories)
	if err != nil {
		return err
	}

	// 1 Hour TTL for category tree
	return r.client.Set(ctx, key, val, 1*time.Hour).Err()
}

// MemoryCacheManager for test and dev fallback
type MemoryCacheManager struct {
	mu         sync.Mutex
	skuMap     map[string]*Product
	categories map[string][]*Category
}

func NewMemoryCacheManager() CacheManager {
	return &MemoryCacheManager{
		skuMap:     make(map[string]*Product),
		categories: make(map[string][]*Category),
	}
}

func (m *MemoryCacheManager) GetSKU(ctx context.Context, storeID, barcode string) (*Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", storeID, barcode)
	return m.skuMap[key], nil
}

func (m *MemoryCacheManager) SetSKU(ctx context.Context, storeID, barcode string, product *Product) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", storeID, barcode)
	m.skuMap[key] = product
	return nil
}

func (m *MemoryCacheManager) DeleteSKU(ctx context.Context, storeID, barcode string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", storeID, barcode)
	delete(m.skuMap, key)
	return nil
}

func (m *MemoryCacheManager) GetCategoryTree(ctx context.Context, chainID string) ([]*Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.categories[chainID], nil
}

func (m *MemoryCacheManager) SetCategoryTree(ctx context.Context, chainID string, categories []*Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.categories[chainID] = categories
	return nil
}
