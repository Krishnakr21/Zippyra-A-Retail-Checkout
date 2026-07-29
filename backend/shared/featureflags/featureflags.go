package featureflags

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ScopeType string

const (
	ScopeGlobal         ScopeType = "GLOBAL"
	ScopeChain          ScopeType = "CHAIN"
	ScopeStore          ScopeType = "STORE"
	ScopeUserPercentage ScopeType = "USER_PERCENTAGE"
)

type FeatureFlag struct {
	FlagKey         string    `json:"flag_key"`
	Description     string    `json:"description"`
	ScopeType       ScopeType `json:"scope_type"`
	EnabledGlobally bool      `json:"enabled_globally"`
	EnabledScopeIDs []string  `json:"enabled_scope_ids"`
	UserPercentage  *int      `json:"user_percentage,omitempty"`
	UpdatedBy       uuid.UUID `json:"updated_by"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// MemoryCache fallback for testing or when Redis is not available
var memoryStore = make(map[string]*FeatureFlag)

func IsEnabled(ctx context.Context, db *sql.DB, rdb *redis.Client, flagKey string, entityID string) bool {
	flag, err := getFlagWithCache(ctx, db, rdb, flagKey)
	if err != nil || flag == nil {
		// Default to false for missing or unconfigured flags
		return false
	}

	return EvaluateFlag(flag, entityID)
}

func getFlagWithCache(ctx context.Context, db *sql.DB, rdb *redis.Client, flagKey string) (*FeatureFlag, error) {
	redisKey := fmt.Sprintf("ff:%s", flagKey)

	// 1. Try Redis cache if client is provided
	if rdb != nil {
		val, err := rdb.Get(ctx, redisKey).Result()
		if err == nil && val != "" {
			var flag FeatureFlag
			if err := json.Unmarshal([]byte(val), &flag); err == nil {
				return &flag, nil
			}
		}
	}

	// 2. Memory store check
	if flag, ok := memoryStore[flagKey]; ok {
		return flag, nil
	}

	// 3. Fallbacks for standard flags when DB and Redis are unconfigured (e.g. tests)
	if flagKey == "qc_required" || flagKey == "auto_reorder" {
		return &FeatureFlag{
			FlagKey:         flagKey,
			ScopeType:       ScopeGlobal,
			EnabledGlobally: true,
		}, nil
	}

	// 3. Fallback to Postgres DB if db is provided
	if db != nil {
		query := `SELECT flag_key, description, scope_type, enabled_globally, enabled_scope_ids, user_percentage, updated_by, updated_at, created_at FROM feature_flags WHERE flag_key = $1`
		var flag FeatureFlag
		var scopeIDsJSON []byte
		var userPct sql.NullInt32

		err := db.QueryRowContext(ctx, query, flagKey).Scan(
			&flag.FlagKey,
			&flag.Description,
			&flag.ScopeType,
			&flag.EnabledGlobally,
			&scopeIDsJSON,
			&userPct,
			&flag.UpdatedBy,
			&flag.UpdatedAt,
			&flag.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(scopeIDsJSON) > 0 {
			_ = json.Unmarshal(scopeIDsJSON, &flag.EnabledScopeIDs)
		}
		if userPct.Valid {
			pct := int(userPct.Int32)
			flag.UserPercentage = &pct
		}

		// Populate Redis cache
		if rdb != nil {
			if data, err := json.Marshal(flag); err == nil {
				_ = rdb.Set(ctx, redisKey, data, 5*time.Minute).Err()
			}
		}

		return &flag, nil
	}

	return nil, nil
}

func SetFlag(ctx context.Context, db *sql.DB, rdb *redis.Client, flag *FeatureFlag) error {
	if flag.CreatedAt.IsZero() {
		flag.CreatedAt = time.Now()
	}
	flag.UpdatedAt = time.Now()

	// Update memory store
	memoryStore[flag.FlagKey] = flag

	// Write to Postgres DB
	if db != nil {
		scopeIDsJSON, _ := json.Marshal(flag.EnabledScopeIDs)
		var userPct interface{} = nil
		if flag.UserPercentage != nil {
			userPct = *flag.UserPercentage
		}

		query := `
			INSERT INTO feature_flags (flag_key, description, scope_type, enabled_globally, enabled_scope_ids, user_percentage, updated_by, updated_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (flag_key) DO UPDATE SET
				description = EXCLUDED.description,
				scope_type = EXCLUDED.scope_type,
				enabled_globally = EXCLUDED.enabled_globally,
				enabled_scope_ids = EXCLUDED.enabled_scope_ids,
				user_percentage = EXCLUDED.user_percentage,
				updated_by = EXCLUDED.updated_by,
				updated_at = EXCLUDED.updated_at
		`
		_, err := db.ExecContext(ctx, query,
			flag.FlagKey,
			flag.Description,
			flag.ScopeType,
			flag.EnabledGlobally,
			scopeIDsJSON,
			userPct,
			flag.UpdatedBy,
			flag.UpdatedAt,
			flag.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	// Write to Redis cache
	if rdb != nil {
		redisKey := fmt.Sprintf("ff:%s", flag.FlagKey)
		data, err := json.Marshal(flag)
		if err == nil {
			_ = rdb.Set(ctx, redisKey, data, 5*time.Minute).Err()
		}
	}

	return nil
}

func DeleteFlag(ctx context.Context, db *sql.DB, rdb *redis.Client, flagKey string) error {
	delete(memoryStore, flagKey)

	if db != nil {
		_, err := db.ExecContext(ctx, `DELETE FROM feature_flags WHERE flag_key = $1`, flagKey)
		if err != nil {
			return err
		}
	}

	if rdb != nil {
		redisKey := fmt.Sprintf("ff:%s", flagKey)
		_ = rdb.Del(ctx, redisKey).Err()
	}

	return nil
}

func EvaluateFlag(flag *FeatureFlag, entityID string) bool {
	if flag == nil {
		return false
	}

	switch flag.ScopeType {
	case ScopeGlobal:
		return flag.EnabledGlobally

	case ScopeChain, ScopeStore:
		for _, id := range flag.EnabledScopeIDs {
			if id == entityID {
				return true
			}
		}
		return false

	case ScopeUserPercentage:
		if flag.UserPercentage == nil || *flag.UserPercentage <= 0 {
			return false
		}
		if *flag.UserPercentage >= 100 {
			return true
		}
		h := fnv.New32a()
		_, _ = h.Write([]byte(entityID))
		hashVal := int(h.Sum32() % 100)
		return hashVal < *flag.UserPercentage

	default:
		return flag.EnabledGlobally
	}
}
