package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type OfferCompiler struct {
	repo OfferRepository
	rdb  redis.Cmdable
}

func NewOfferCompiler(repo OfferRepository, rdb redis.Cmdable) *OfferCompiler {
	return &OfferCompiler{
		repo: repo,
		rdb:  rdb,
	}
}

func (c *OfferCompiler) CompileAndPublish(ctx context.Context, storeID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	chainID, err := c.repo.GetStoreChainID(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to resolve chain_id for store %s: %w", storeID, err)
	}

	offers, err := c.repo.ListActiveOffersForStore(ctx, chainID, storeID)
	if err != nil {
		return fmt.Errorf("failed to query active offers for store %s: %w", storeID, err)
	}

	// Sort by priority ASC, then created_at ASC
	sort.SliceStable(offers, func(i, j int) bool {
		if offers[i].Priority != offers[j].Priority {
			return offers[i].Priority < offers[j].Priority
		}
		return offers[i].CreatedAt.Before(offers[j].CreatedAt)
	})

	var compiledRules []*OfferRule
	for _, o := range offers {
		rule := mapOfferToOfferRule(o)
		compiledRules = append(compiledRules, rule)
	}

	if compiledRules == nil {
		compiledRules = []*OfferRule{}
	}

	rulesJSON, err := json.Marshal(compiledRules)
	if err != nil {
		return fmt.Errorf("failed to marshal compiled rules: %w", err)
	}

	// 1. SET Redis key offer_rules:{storeID} (durable, no TTL)
	redisKey := fmt.Sprintf("offer_rules:%s", storeID)
	if c.rdb != nil {
		if err := c.rdb.Set(ctx, redisKey, string(rulesJSON), 0).Err(); err != nil {
			log.Printf("[WARN] [OfferCompiler] Failed to set Redis key %s: %v", redisKey, err)
		}
	}

	// 2. INSERT offer_rules_audit
	if err := c.repo.LogOfferRulesAudit(ctx, storeID, rulesJSON); err != nil {
		log.Printf("[WARN] [OfferCompiler] Failed to insert offer_rules_audit row for store %s: %v", storeID, err)
	}

	log.Printf("[INFO] [OfferCompiler] Successfully compiled and published %d offer rules for store %s", len(compiledRules), storeID)
	return nil
}

func mapOfferToOfferRule(o *Offer) *OfferRule {
	var val float64
	switch strings.ToUpper(o.Type) {
	case "PERCENT_OFF", "CATEGORY_PERCENT_OFF":
		if p, ok := o.RuleConfig["percent"]; ok {
			switch pv := p.(type) {
			case float64:
				val = pv
			case int:
				val = float64(pv)
			case int64:
				val = float64(pv)
			}
		}
	case "FLAT_OFF":
		if f, ok := o.RuleConfig["flat_paise"]; ok {
			switch fv := f.(type) {
			case float64:
				val = fv
			case int:
				val = float64(fv)
			case int64:
				val = float64(fv)
			}
		}
	case "BOGO":
		val = 1.0
	}

	var activeFrom *time.Time
	if !o.ActiveFrom.IsZero() {
		af := o.ActiveFrom
		activeFrom = &af
	}

	return &OfferRule{
		ID:                 o.ID,
		Type:               o.Type,
		Value:              val,
		AppliesTo:          o.AppliesTo,
		TargetIDs:          o.TargetIDs,
		MinCartValuePaise: o.MinCartValuePaise,
		MaxDiscountPaise:  o.MaxDiscountPaise,
		ActiveFrom:         activeFrom,
		ActiveUntil:        o.ActiveUntil,
	}
}
