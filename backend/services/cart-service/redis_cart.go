package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type CartStore interface {
	GetCart(ctx context.Context, storeID, userID string) ([]*CartItem, string, error)
	UpsertCartItem(ctx context.Context, storeID, userID string, item *CartItem) error
	RemoveCartItem(ctx context.Context, storeID, userID, barcode string) error
	ClearCart(ctx context.Context, storeID, userID string) error
	SetCoupon(ctx context.Context, storeID, userID, code string) error
	GetCoupon(ctx context.Context, storeID, userID string) (string, error)
	RemoveCoupon(ctx context.Context, storeID, userID string) error
}

type RedisCartStore struct {
	rdb         redis.Cmdable
	mu          sync.Mutex
	memCart     map[string]map[string]*CartItem
	memCoupons  map[string]string
}

func NewRedisCartStore(rdb redis.Cmdable) CartStore {
	return &RedisCartStore{
		rdb:        rdb,
		memCart:    make(map[string]map[string]*CartItem),
		memCoupons: make(map[string]string),
	}
}

func (r *RedisCartStore) GetCart(ctx context.Context, storeID, userID string) ([]*CartItem, string, error) {
	if r.rdb == nil {
		r.mu.Lock()
		defer r.mu.Unlock()

		cartKey := fmt.Sprintf("%s:%s", storeID, userID)
		couponCode := r.memCoupons[cartKey]

		var items []*CartItem
		if m, ok := r.memCart[cartKey]; ok {
			for _, item := range m {
				cp := *item
				cp.PricePaise = cp.PricePaiseSnapshot
				if cp.LineTotalPaise <= 0 {
					cp.LineTotalPaise = cp.PricePaise * int64(cp.Qty)
				}
				items = append(items, &cp)
			}
		}
		return items, couponCode, nil
	}

	cartKey := fmt.Sprintf("cart:%s:%s", storeID, userID)
	couponKey := fmt.Sprintf("cart_coupon:%s:%s", storeID, userID)

	pipe := r.rdb.Pipeline()
	hgetAllCmd := pipe.HGetAll(ctx, cartKey)
	getCouponCmd := pipe.Get(ctx, couponKey)
	_, _ = pipe.Exec(ctx)

	rawItems, err := hgetAllCmd.Result()
	if err != nil && err != redis.Nil {
		return nil, "", err
	}

	couponCode, _ := getCouponCmd.Result()

	var items []*CartItem
	for _, jsonStr := range rawItems {
		var item CartItem
		if err := json.Unmarshal([]byte(jsonStr), &item); err == nil {
			if item.LineTotalPaise <= 0 {
				item.LineTotalPaise = item.PricePaiseSnapshot * int64(item.Qty)
			}
			item.PricePaise = item.PricePaiseSnapshot
			items = append(items, &item)
		}
	}

	return items, couponCode, nil
}

func (r *RedisCartStore) UpsertCartItem(ctx context.Context, storeID, userID string, item *CartItem) error {
	if r.rdb == nil {
		r.mu.Lock()
		defer r.mu.Unlock()

		cartKey := fmt.Sprintf("%s:%s", storeID, userID)
		if _, ok := r.memCart[cartKey]; !ok {
			r.memCart[cartKey] = make(map[string]*CartItem)
		}
		cp := *item
		r.memCart[cartKey][item.Barcode] = &cp
		return nil
	}

	cartKey := fmt.Sprintf("cart:%s:%s", storeID, userID)

	itemJSON, err := json.Marshal(item)
	if err != nil {
		return err
	}

	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, cartKey, item.Barcode, string(itemJSON))
	pipe.Expire(ctx, cartKey, 2*3600*time.Second)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisCartStore) RemoveCartItem(ctx context.Context, storeID, userID, barcode string) error {
	if r.rdb == nil {
		r.mu.Lock()
		defer r.mu.Unlock()

		cartKey := fmt.Sprintf("%s:%s", storeID, userID)
		if m, ok := r.memCart[cartKey]; ok {
			delete(m, barcode)
		}
		return nil
	}

	cartKey := fmt.Sprintf("cart:%s:%s", storeID, userID)
	return r.rdb.HDel(ctx, cartKey, barcode).Err()
}

func (r *RedisCartStore) ClearCart(ctx context.Context, storeID, userID string) error {
	if r.rdb == nil {
		r.mu.Lock()
		defer r.mu.Unlock()

		cartKey := fmt.Sprintf("%s:%s", storeID, userID)
		delete(r.memCart, cartKey)
		delete(r.memCoupons, cartKey)
		return nil
	}

	cartKey := fmt.Sprintf("cart:%s:%s", storeID, userID)
	couponKey := fmt.Sprintf("cart_coupon:%s:%s", storeID, userID)

	pipe := r.rdb.Pipeline()
	pipe.Del(ctx, cartKey)
	pipe.Del(ctx, couponKey)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisCartStore) SetCoupon(ctx context.Context, storeID, userID, code string) error {
	if r.rdb == nil {
		r.mu.Lock()
		defer r.mu.Unlock()

		cartKey := fmt.Sprintf("%s:%s", storeID, userID)
		r.memCoupons[cartKey] = code
		return nil
	}

	couponKey := fmt.Sprintf("cart_coupon:%s:%s", storeID, userID)
	return r.rdb.Set(ctx, couponKey, code, 2*3600*time.Second).Err()
}

func (r *RedisCartStore) GetCoupon(ctx context.Context, storeID, userID string) (string, error) {
	if r.rdb == nil {
		r.mu.Lock()
		defer r.mu.Unlock()

		cartKey := fmt.Sprintf("%s:%s", storeID, userID)
		return r.memCoupons[cartKey], nil
	}

	couponKey := fmt.Sprintf("cart_coupon:%s:%s", storeID, userID)
	val, err := r.rdb.Get(ctx, couponKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisCartStore) RemoveCoupon(ctx context.Context, storeID, userID string) error {
	if r.rdb == nil {
		r.mu.Lock()
		defer r.mu.Unlock()

		cartKey := fmt.Sprintf("%s:%s", storeID, userID)
		delete(r.memCoupons, cartKey)
		return nil
	}

	couponKey := fmt.Sprintf("cart_coupon:%s:%s", storeID, userID)
	return r.rdb.Del(ctx, couponKey).Err()
}
