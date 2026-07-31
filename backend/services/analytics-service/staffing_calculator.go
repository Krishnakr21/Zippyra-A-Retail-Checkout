package main

import (
	"math"
	"os"
	"strconv"
)

func GetDefaultThroughputPerHour() int {
	valStr := os.Getenv("ANALYTICS_STAFF_THROUGHPUT_PER_HOUR")
	if valStr != "" {
		if val, err := strconv.Atoi(valStr); err == nil && val > 0 {
			return val
		}
	}
	return 20 // Default assumed throughput per staff member per hour
}

// CalculateRecommendedStaff computes recommended staff for a peak hour cell:
// recommended_staff := ceil(avg_transactions_per_week / throughput_per_hour)
func CalculateRecommendedStaff(avgTransactionsPerWeek float64, throughputPerHour int) int {
	if throughputPerHour <= 0 {
		throughputPerHour = 20
	}
	if avgTransactionsPerWeek <= 0 {
		return 1 // Minimum 1 staff member on duty
	}

	staff := math.Ceil(avgTransactionsPerWeek / float64(throughputPerHour))
	if staff < 1 {
		return 1
	}
	return int(staff)
}
