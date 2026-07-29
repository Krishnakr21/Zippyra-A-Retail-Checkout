package loyalty

import (
	"math"
)

// CalculateEarnPoints calculates points earned based on order's total_paise (post-discount)
// Rule: 1 point per ₹10 spent (1000 paise = ₹10), multiplied by tier earn_multiplier.
func CalculateEarnPoints(totalPaise int64, earnMultiplier float64) int64 {
	if totalPaise <= 0 {
		return 0
	}
	if earnMultiplier <= 0 {
		earnMultiplier = 1.0
	}

	basePoints := float64(totalPaise / 1000)
	earned := math.Floor(basePoints * earnMultiplier)
	return int64(earned)
}
