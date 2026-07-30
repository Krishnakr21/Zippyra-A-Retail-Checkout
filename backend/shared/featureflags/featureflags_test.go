package featureflags

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFeatureFlag_GlobalScope(t *testing.T) {
	ctx := context.Background()
	flagKey := "test.global_feature"

	flag := &FeatureFlag{
		FlagKey:         flagKey,
		Description:     "Test Global Flag",
		ScopeType:       ScopeGlobal,
		EnabledGlobally: true,
		UpdatedBy:       uuid.New(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := SetFlag(ctx, nil, nil, flag)
	if err != nil {
		t.Fatalf("Failed to set flag: %v", err)
	}

	if !IsEnabled(ctx, nil, nil, flagKey, "any-store-id") {
		t.Errorf("Expected flag to be enabled globally")
	}

	// Disable flag
	flag.EnabledGlobally = false
	_ = SetFlag(ctx, nil, nil, flag)

	if IsEnabled(ctx, nil, nil, flagKey, "any-store-id") {
		t.Errorf("Expected flag to be disabled globally")
	}
}

func TestFeatureFlag_StoreScope(t *testing.T) {
	ctx := context.Background()
	flagKey := "test.store_feature"

	flag := &FeatureFlag{
		FlagKey:         flagKey,
		Description:     "Store Specific Feature",
		ScopeType:       ScopeStore,
		EnabledScopeIDs: []string{"store-mumbai-01", "store-delhi-02"},
		UpdatedBy:       uuid.New(),
	}

	_ = SetFlag(ctx, nil, nil, flag)

	if !IsEnabled(ctx, nil, nil, flagKey, "store-mumbai-01") {
		t.Errorf("Expected flag to be enabled for store-mumbai-01")
	}
	if IsEnabled(ctx, nil, nil, flagKey, "store-bangalore-03") {
		t.Errorf("Expected flag to be disabled for store-bangalore-03")
	}
}

func TestFeatureFlag_UserPercentageScope(t *testing.T) {
	ctx := context.Background()
	flagKey := "test.gradual_rollout"
	pct := 50

	flag := &FeatureFlag{
		FlagKey:        flagKey,
		Description:    "Gradual Rollout Feature",
		ScopeType:      ScopeUserPercentage,
		UserPercentage: &pct,
		UpdatedBy:      uuid.New(),
	}

	_ = SetFlag(ctx, nil, nil, flag)

	// User percentage deterministic evaluation
	user1Enabled := IsEnabled(ctx, nil, nil, flagKey, "user-uuid-100")
	user2Enabled := IsEnabled(ctx, nil, nil, flagKey, "user-uuid-200")

	// Verify consistency across repeated calls
	if IsEnabled(ctx, nil, nil, flagKey, "user-uuid-100") != user1Enabled {
		t.Errorf("Expected deterministic flag evaluation for user-uuid-100")
	}
	if IsEnabled(ctx, nil, nil, flagKey, "user-uuid-200") != user2Enabled {
		t.Errorf("Expected deterministic flag evaluation for user-uuid-200")
	}
}
