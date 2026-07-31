package main

import (
	"testing"
)

func TestPeakHoursStaffingRecommendation_Formula_ComputesCorrectStaff(t *testing.T) {
	// Formula: recommended_staff := ceil(avg_transactions_per_week / throughput_per_hour)
	// Test case: 80 avg transactions/week over 4-week lookback with throughput=20 -> 80/20 = 4 staff members
	avgTx := 80.0
	throughput := 20
	recommended := CalculateRecommendedStaff(avgTx, throughput)

	if recommended != 4 {
		t.Fatalf("expected recommended_staff=4 for avgTx=80 and throughput=20, got %d", recommended)
	}

	// Test case 2: 25 avg transactions/week with throughput=20 -> ceil(25/20) = ceil(1.25) = 2 staff members
	rec2 := CalculateRecommendedStaff(25.0, 20)
	if rec2 != 2 {
		t.Fatalf("expected recommended_staff=2 for avgTx=25 and throughput=20, got %d", rec2)
	}
}
