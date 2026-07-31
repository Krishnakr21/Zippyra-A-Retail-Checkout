package main

import (
	"context"
	"testing"
	"time"
)

func TestSLAWarningJob_WarnsOnlyOnceAcrossMultipleTicks(t *testing.T) {
	repo := NewMemoryRepository()
	job := NewSLAWarningJob(repo, nil)

	// Create an URGENT ticket created 3.5 hours ago (SLA = 4h, 80% = 3.2h) -> crossed 80% window
	ticket := &SupportTicket{
		ID:            "ticket-sla-1",
		RequesterID:   "cust-sla-1",
		RequesterType: "CUSTOMER",
		Category:      CategoryExitGateIssue,
		Subject:       "Stuck at gate",
		Description:   "Help",
		Priority:      PriorityUrgent,
		Status:        StatusOpen,
		CreatedAt:     time.Now().Add(-210 * time.Minute), // 3.5h ago
		SLADueAt:      time.Now().Add(30 * time.Minute),   // 4h window
		IsSLAWarned:   false,
	}
	_ = repo.CreateTicket(context.Background(), ticket)

	// Tick 1: should detect ticket and mark is_sla_warned = true
	job.RunOnce(context.Background())

	t1, _ := repo.GetTicketByID(context.Background(), "ticket-sla-1")
	if !t1.IsSLAWarned {
		t.Errorf("Expected ticket to be marked as SLA warned on Tick 1")
	}

	// Tick 2: ListTicketsNearSLA should return 0 tickets because is_sla_warned = true
	nearSLATickets, _ := repo.ListTicketsNearSLA(context.Background())
	if len(nearSLATickets) != 0 {
		t.Errorf("Expected 0 tickets near SLA on Tick 2 after being warned, got %d", len(nearSLATickets))
	}
}
