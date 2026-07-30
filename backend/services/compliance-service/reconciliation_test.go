package main

import (
	"context"
	"testing"
)

func TestReconciliationJob_RunsAndSavesReport(t *testing.T) {
	repo := NewMemoryRepository()
	job := NewReconciliationJob(repo, "")

	ctx := context.Background()

	report, err := job.RunReconciliationForDate(ctx, "2026-07-31")
	if err != nil || report == nil {
		t.Fatalf("RunReconciliationForDate failed: %v", err)
	}

	if report.ReportDate != "2026-07-31" {
		t.Fatalf("Expected report date 2026-07-31, got %s", report.ReportDate)
	}

	fetched, err := repo.GetSettlementReportByDate(ctx, "2026-07-31")
	if err != nil || fetched == nil {
		t.Fatalf("Failed to fetch report from repository: %v", err)
	}
}
