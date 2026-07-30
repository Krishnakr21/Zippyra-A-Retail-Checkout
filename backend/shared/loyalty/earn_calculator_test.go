package loyalty

import (
	"testing"
)

func TestCalculateEarnPoints(t *testing.T) {
	// 1 point per ₹10 spent (1000 paise = ₹10)
	tests := []struct {
		name       string
		totalPaise int64
		multiplier float64
		expected   int64
	}{
		{"Zero total", 0, 1.0, 0},
		{"Below ₹10", 500, 1.0, 0},
		{"₹100 Bronze 1.0x", 10000, 1.0, 10},
		{"₹100 Silver 1.2x", 10000, 1.2, 12},
		{"₹100 Gold 1.5x", 10000, 1.5, 15},
		{"₹100 Platinum 2.0x", 10000, 2.0, 20},
		{"₹255.50 Bronze 1.0x", 25550, 1.0, 25},
		{"₹255.50 Gold 1.5x", 25550, 1.5, 37}, // floor(25 * 1.5) = floor(37.5) = 37
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateEarnPoints(tt.totalPaise, tt.multiplier)
			if got != tt.expected {
				t.Errorf("CalculateEarnPoints(%d, %f) = %d; want %d", tt.totalPaise, tt.multiplier, got, tt.expected)
			}
		})
	}
}

func TestCalculateTier(t *testing.T) {
	tests := []struct {
		name           string
		lifetimePoints int64
		expectedTier   string
	}{
		{"0 points", 0, "BRONZE"},
		{"4999 points", 4999, "BRONZE"},
		{"5000 points", 5000, "SILVER"},
		{"19999 points", 19999, "SILVER"},
		{"20000 points", 20000, "GOLD"},
		{"49999 points", 49999, "GOLD"},
		{"50000 points", 50000, "PLATINUM"},
		{"100000 points", 100000, "PLATINUM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTier(tt.lifetimePoints, DefaultTiers)
			if got.Tier != tt.expectedTier {
				t.Errorf("CalculateTier(%d) = %s; want %s", tt.lifetimePoints, got.Tier, tt.expectedTier)
			}
		})
	}
}
