package loyalty

type TierConfig struct {
	Tier              string  `json:"tier"`
	MinLifetimePoints int64   `json:"min_lifetime_points"`
	EarnMultiplier    float64 `json:"earn_multiplier"`
	DisplayName       string  `json:"display_name"`
	DisplayOrder      int     `json:"display_order"`
}

var DefaultTiers = []TierConfig{
	{Tier: "BRONZE", MinLifetimePoints: 0, EarnMultiplier: 1.0, DisplayName: "Bronze Tier", DisplayOrder: 1},
	{Tier: "SILVER", MinLifetimePoints: 5000, EarnMultiplier: 1.2, DisplayName: "Silver Tier", DisplayOrder: 2},
	{Tier: "GOLD", MinLifetimePoints: 20000, EarnMultiplier: 1.5, DisplayName: "Gold Tier", DisplayOrder: 3},
	{Tier: "PLATINUM", MinLifetimePoints: 50000, EarnMultiplier: 2.0, DisplayName: "Platinum Tier", DisplayOrder: 4},
}

// CalculateTier returns the highest tier where MinLifetimePoints <= lifetimePoints.
// Tier status, once earned, does NOT downgrade automatically from returns/reversals.
func CalculateTier(lifetimePoints int64, tiers []TierConfig) TierConfig {
	if len(tiers) == 0 {
		tiers = DefaultTiers
	}

	best := tiers[0]
	for _, t := range tiers {
		if lifetimePoints >= t.MinLifetimePoints && t.MinLifetimePoints >= best.MinLifetimePoints {
			best = t
		}
	}
	return best
}

// CalculateNextTier returns the next higher tier configuration, or nil if already at max tier.
func CalculateNextTier(lifetimePoints int64, tiers []TierConfig) *TierConfig {
	if len(tiers) == 0 {
		tiers = DefaultTiers
	}

	current := CalculateTier(lifetimePoints, tiers)
	var next *TierConfig
	for i, t := range tiers {
		if t.Tier == current.Tier && i+1 < len(tiers) {
			next = &tiers[i+1]
			break
		}
	}
	return next
}
