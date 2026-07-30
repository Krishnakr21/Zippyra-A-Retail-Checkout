package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestMessageToClosedTicket_Returns409(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewTicketHandler(repo, nil, nil)
	r := mux.NewRouter()
	SetupRoutes(r, handler, nil)

	ticket := &SupportTicket{
		ID:            "t-closed-1",
		RequesterID:   "cust-close-1",
		RequesterType: "CUSTOMER",
		Subject:       "Test",
		Description:   "Test",
		Status:        StatusClosed,
		SLADueAt:      time.Now().Add(24 * time.Hour),
	}
	_ = repo.CreateTicket(context.Background(), ticket)

	payload := AddMessageRequest{Body: "Trying to post to closed ticket"}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/v1/support/tickets/t-closed-1/messages", bytes.NewBuffer(bodyBytes))
	req.Header.Set("X-User-ID", "cust-close-1")
	req.Header.Set("X-User-Role", "CUSTOMER")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected HTTP 409 Conflict when posting to closed ticket, got %d", rr.Code)
	}
}

func TestCustomerReply_WaitingOnCustomer_AutoTransitionsToAssigned(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewTicketHandler(repo, nil, nil)
	r := mux.NewRouter()
	SetupRoutes(r, handler, nil)

	ticket := &SupportTicket{
		ID:            "t-waiting-1",
		RequesterID:   "cust-wait-1",
		RequesterType: "CUSTOMER",
		Subject:       "Test",
		Description:   "Test",
		Status:        StatusWaitingOnCustomer,
		SLADueAt:      time.Now().Add(24 * time.Hour),
	}
	_ = repo.CreateTicket(context.Background(), ticket)

	payload := AddMessageRequest{Body: "Here is the requested screenshot"}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/v1/support/tickets/t-waiting-1/messages", bytes.NewBuffer(bodyBytes))
	req.Header.Set("X-User-ID", "cust-wait-1")
	req.Header.Set("X-User-Role", "CUSTOMER")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d", rr.Code)
	}

	updated, _ := repo.GetTicketByID(context.Background(), "t-waiting-1")
	if updated.Status != StatusAssigned {
		t.Errorf("Expected status to auto-transition to ASSIGNED, got %s", updated.Status)
	}
}

func TestReopenWindow_Within7DaysSucceeds_After7DaysExpired(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewTicketHandler(repo, nil, nil)
	r := mux.NewRouter()
	SetupRoutes(r, handler, nil)

	agentID := "agent-100"

	// Ticket 1: Resolved 3 days ago (within 7 day window)
	res3DaysAgo := time.Now().Add(-3 * 24 * time.Hour)
	t1 := &SupportTicket{
		ID:              "t-reopen-3d",
		RequesterID:     "cust-reopen-1",
		RequesterType:   "CUSTOMER",
		Subject:         "Test",
		Description:     "Test",
		Status:          StatusResolved,
		AssignedAgentID: &agentID,
		ResolvedAt:      &res3DaysAgo,
		SLADueAt:        time.Now().Add(24 * time.Hour),
	}
	_ = repo.CreateTicket(context.Background(), t1)

	req1 := httptest.NewRequest("POST", "/v1/support/tickets/t-reopen-3d/reopen", nil)
	req1.Header.Set("X-User-ID", "cust-reopen-1")
	req1.Header.Set("X-User-Role", "CUSTOMER")

	rr1 := httptest.NewRecorder()
	r.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("Expected HTTP 200 for reopen within 7 days, got %d", rr1.Code)
	}
	t1Updated, _ := repo.GetTicketByID(context.Background(), "t-reopen-3d")
	if t1Updated.Status != StatusOpen || t1Updated.AssignedAgentID != nil {
		t.Errorf("Expected status OPEN and assigned_agent_id cleared, got status %s, agent %v", t1Updated.Status, t1Updated.AssignedAgentID)
	}

	// Ticket 2: Resolved 10 days ago (outside 7 day window)
	res10DaysAgo := time.Now().Add(-10 * 24 * time.Hour)
	t2 := &SupportTicket{
		ID:            "t-reopen-10d",
		RequesterID:   "cust-reopen-1",
		RequesterType: "CUSTOMER",
		Subject:       "Test",
		Description:   "Test",
		Status:        StatusResolved,
		ResolvedAt:    &res10DaysAgo,
		SLADueAt:      time.Now().Add(24 * time.Hour),
	}
	_ = repo.CreateTicket(context.Background(), t2)

	req2 := httptest.NewRequest("POST", "/v1/support/tickets/t-reopen-10d/reopen", nil)
	req2.Header.Set("X-User-ID", "cust-reopen-1")
	req2.Header.Set("X-User-Role", "CUSTOMER")

	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected HTTP 400 REOPEN_WINDOW_EXPIRED after 7 days, got %d", rr2.Code)
	}
}

func TestResolveStatus_WithoutResolutionNote_Returns400(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewTicketHandler(repo, nil, nil)
	r := mux.NewRouter()
	SetupRoutes(r, handler, nil)

	ticket := &SupportTicket{
		ID:            "t-resolve-1",
		RequesterID:   "cust-res-1",
		RequesterType: "CUSTOMER",
		Subject:       "Test",
		Description:   "Test",
		Status:        StatusAssigned,
		SLADueAt:      time.Now().Add(24 * time.Hour),
	}
	_ = repo.CreateTicket(context.Background(), ticket)

	// Attempt to set status to RESOLVED without resolution_note
	payload := UpdateStatusRequest{
		Status: StatusResolved,
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/v1/support/tickets/t-resolve-1/status", bytes.NewBuffer(bodyBytes))
	req.Header.Set("X-User-ID", "agent-admin-1")
	req.Header.Set("X-User-Role", "ADMIN")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected HTTP 400 RESOLUTION_NOTE_REQUIRED, got %d", rr.Code)
	}
}
