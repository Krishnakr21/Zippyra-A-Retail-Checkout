package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type OfferReconciliationJob struct {
	compiler    *OfferCompiler
	repo        OfferRepository
	rdb         redis.Cmdable
	mu          sync.RWMutex
	lastRunTime time.Time
}

func NewOfferReconciliationJob(compiler *OfferCompiler, repo OfferRepository, rdb redis.Cmdable) *OfferReconciliationJob {
	return &OfferReconciliationJob{
		compiler:    compiler,
		repo:        repo,
		rdb:         rdb,
		lastRunTime: time.Now().UTC(),
	}
}

func (j *OfferReconciliationJob) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[INFO] [OfferReconciliationJob] Stopping reconciliation worker...")
			return
		case <-ticker.C:
			j.RunOnce(ctx)
		}
	}
}

func (j *OfferReconciliationJob) RunOnce(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	j.mu.Lock()
	j.lastRunTime = time.Now().UTC()
	j.mu.Unlock()

	log.Println("[INFO] [OfferReconciliationJob] Running hourly reconciliation sweep...")

	stores, err := j.repo.ListStoresForChain(ctx, "chain-default-001")
	if err != nil {
		log.Printf("[WARN] [OfferReconciliationJob] Failed to list stores for reconciliation: %v", err)
		return
	}

	for _, storeID := range stores {
		j.reconcileStore(ctx, storeID)
	}
}

func (j *OfferReconciliationJob) reconcileStore(ctx context.Context, storeID string) {
	chainID, err := j.repo.GetStoreChainID(ctx, storeID)
	if err != nil {
		return
	}

	offers, err := j.repo.ListActiveOffersForStore(ctx, chainID, storeID)
	if err != nil {
		return
	}

	var expectedRules []*OfferRule
	for _, o := range offers {
		expectedRules = append(expectedRules, mapOfferToOfferRule(o))
	}
	if expectedRules == nil {
		expectedRules = []*OfferRule{}
	}

	expectedHash := hashRules(expectedRules)

	redisKey := fmt.Sprintf("offer_rules:%s", storeID)
	cachedVal, err := j.rdb.Get(ctx, redisKey).Result()
	if err != nil || cachedVal == "" {
		// Key missing or error -> recompile
		log.Printf("[INFO] [OfferReconciliationJob] Missing or invalid Redis key %s. Triggering recompile...", redisKey)
		_ = j.compiler.CompileAndPublish(ctx, storeID)
		return
	}

	var cachedRules []*OfferRule
	_ = jsonUnmarshal([]byte(cachedVal), &cachedRules)
	cachedHash := hashRules(cachedRules)

	if expectedHash != cachedHash {
		log.Printf("[WARN] [OfferReconciliationJob] Hash mismatch for store %s (cached=%s, expected=%s). Recompiling...", storeID, cachedHash, expectedHash)
		_ = j.compiler.CompileAndPublish(ctx, storeID)
	}
}

func (j *OfferReconciliationJob) IsHealthy() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return time.Since(j.lastRunTime) <= 2*time.Hour
}

func hashRules(rules []*OfferRule) string {
	b, _ := jsonMarshal(rules)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

var jsonMarshal = func(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

var jsonUnmarshal = func(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
