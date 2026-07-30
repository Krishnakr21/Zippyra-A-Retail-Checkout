package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type OfferEngine interface {
	EvaluateOffers(ctx context.Context, storeID string, items []*CartItem, subtotalPaise int64) (int64, []string, error)
	SetStoreOfferRules(ctx context.Context, storeID string, rules []*OfferRule) error
}

type RedisOfferEngine struct {
	rdb redis.Cmdable
}

func NewRedisOfferEngine(rdb redis.Cmdable) OfferEngine {
	return &RedisOfferEngine{rdb: rdb}
}

func (e *RedisOfferEngine) EvaluateOffers(ctx context.Context, storeID string, items []*CartItem, subtotalPaise int64) (int64, []string, error) {
	if len(items) == 0 || subtotalPaise <= 0 {
		return 0, nil, nil
	}

	key := fmt.Sprintf("offer_rules:%s", storeID)
	val, err := e.rdb.Get(ctx, key).Result()
	if err == redis.Nil || strings.TrimSpace(val) == "" || strings.TrimSpace(val) == "[]" {
		return 0, []string{}, nil
	} else if err != nil {
		// On Redis connection issue, treat missing rules as no discount (never panic)
		return 0, []string{}, nil
	}

	var rules []*OfferRule
	if err := json.Unmarshal([]byte(val), &rules); err != nil || len(rules) == 0 {
		return 0, []string{}, nil
	}

	now := time.Now()

	// First-match-wins evaluation strategy
	for _, rule := range rules {
		if rule == nil {
			continue
		}

		if rule.ActiveFrom != nil && now.Before(*rule.ActiveFrom) {
			continue
		}
		if rule.ActiveUntil != nil && now.After(*rule.ActiveUntil) {
			continue
		}
		if subtotalPaise < rule.MinCartValuePaise {
			continue
		}

		discount, name, matched := applySingleRule(rule, items, subtotalPaise)
		if matched && discount > 0 {
			if rule.MaxDiscountPaise != nil && discount > *rule.MaxDiscountPaise {
				discount = *rule.MaxDiscountPaise
			}
			if discount > subtotalPaise {
				discount = subtotalPaise
			}
			return discount, []string{name}, nil
		}
	}

	return 0, []string{}, nil
}

func (e *RedisOfferEngine) SetStoreOfferRules(ctx context.Context, storeID string, rules []*OfferRule) error {
	key := fmt.Sprintf("offer_rules:%s", storeID)
	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	return e.rdb.Set(ctx, key, string(rulesJSON), 0).Err()
}

func applySingleRule(rule *OfferRule, items []*CartItem, subtotalPaise int64) (int64, string, bool) {
	switch strings.ToUpper(rule.Type) {
	case "PERCENT_OFF":
		discount := int64(math.Round(float64(subtotalPaise) * rule.Value / 100.0))
		return discount, fmt.Sprintf("%.0f%% Off Cart", rule.Value), true

	case "FLAT_OFF":
		discount := int64(rule.Value)
		return discount, fmt.Sprintf("₹%.2f Flat Off", rule.Value/100.0), true

	case "BOGO":
		// Buy 1 Get 1 Free on qualifying target barcodes
		var discount int64
		matched := false
		targetMap := make(map[string]bool)
		for _, b := range rule.TargetIDs {
			targetMap[b] = true
		}

		for _, item := range items {
			if rule.AppliesTo == "ALL" || targetMap[item.Barcode] {
				freeQty := item.Qty / 2
				if freeQty > 0 {
					discount += int64(freeQty) * item.PricePaiseSnapshot
					matched = true
				}
			}
		}
		if matched {
			return discount, "Buy 1 Get 1 Free", true
		}

	case "CATEGORY_PERCENT_OFF":
		var categorySubtotal int64
		matched := false
		catMap := make(map[string]bool)
		for _, c := range rule.TargetIDs {
			catMap[c] = true
		}

		for _, item := range items {
			if item.CategoryID != "" && catMap[item.CategoryID] {
				categorySubtotal += item.PricePaiseSnapshot * int64(item.Qty)
				matched = true
			}
		}

		if matched && categorySubtotal > 0 {
			discount := int64(math.Round(float64(categorySubtotal) * rule.Value / 100.0))
			return discount, fmt.Sprintf("%.0f%% Off Category", rule.Value), true
		}
	}

	return 0, "", false
}

// MemoryOfferEngine for unit tests
type MemoryOfferEngine struct {
	mu    sync.Mutex
	rules map[string][]*OfferRule
}

func NewMemoryOfferEngine() OfferEngine {
	return &MemoryOfferEngine{rules: make(map[string][]*OfferRule)}
}

func (m *MemoryOfferEngine) EvaluateOffers(ctx context.Context, storeID string, items []*CartItem, subtotalPaise int64) (int64, []string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rules, ok := m.rules[storeID]
	if !ok || len(rules) == 0 {
		return 0, []string{}, nil
	}

	now := time.Now()
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if rule.ActiveFrom != nil && now.Before(*rule.ActiveFrom) {
			continue
		}
		if rule.ActiveUntil != nil && now.After(*rule.ActiveUntil) {
			continue
		}
		if subtotalPaise < rule.MinCartValuePaise {
			continue
		}

		discount, name, matched := applySingleRule(rule, items, subtotalPaise)
		if matched && discount > 0 {
			if rule.MaxDiscountPaise != nil && discount > *rule.MaxDiscountPaise {
				discount = *rule.MaxDiscountPaise
			}
			if discount > subtotalPaise {
				discount = subtotalPaise
			}
			return discount, []string{name}, nil
		}
	}

	return 0, []string{}, nil
}

func (m *MemoryOfferEngine) SetStoreOfferRules(ctx context.Context, storeID string, rules []*OfferRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules[storeID] = rules
	return nil
}
