package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuditConsumer_DuplicateEvent_InsertsExactlyOnce(t *testing.T) {
	repo := NewMemoryRepository()
	consumer := NewAuditConsumer(repo)
	ctx := context.Background()

	event := AdminAction{
		ActorID:       "admin-001",
		ActorName:     "Jane Doe",
		ActionType:    "store.created",
		TargetType:    "store",
		TargetID:      "store-123",
		Payload:       map[string]interface{}{"name": "Test Store"},
		SourceService: "store-service",
		RequestID:     "req-abc-123",
	}
	eventBytes, _ := json.Marshal(event)

	// Process the same event twice
	err1 := consumer.ProcessMessage(ctx, []byte("store-123"), eventBytes)
	if err1 != nil {
		t.Fatalf("first ProcessMessage failed: %v", err1)
	}

	err2 := consumer.ProcessMessage(ctx, []byte("store-123"), eventBytes)
	if err2 != nil {
		t.Fatalf("second ProcessMessage should succeed (idempotent skip): %v", err2)
	}

	// Verify exactly one row exists
	actions, total, err := repo.ListActions(ctx, AuditFilter{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatalf("ListActions failed: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected exactly 1 action after duplicate delivery, got %d", total)
	}
	if actions[0].ActionType != "store.created" {
		t.Fatalf("expected action_type store.created, got %s", actions[0].ActionType)
	}
}

func TestAuditHandler_ListActions_WithFilters(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAuditHandler(repo, nil, "zippyra-dev-jwt-secret-key-32bytes", nil)
	ctx := context.Background()

	// Seed actions from different services
	_ = repo.CreateAction(ctx, &AdminAction{
		ActorID: "admin-001", ActionType: "store.created", TargetType: "store",
		TargetID: "s1", SourceService: "store-service", RequestID: "req-1",
	})
	_ = repo.CreateAction(ctx, &AdminAction{
		ActorID: "admin-001", ActionType: "staff.created", TargetType: "staff",
		TargetID: "st1", SourceService: "retailer-auth-service", RequestID: "req-2",
	})
	_ = repo.CreateAction(ctx, &AdminAction{
		ActorID: "admin-002", ActionType: "product.created", TargetType: "product",
		TargetID: "p1", SourceService: "catalog-service", RequestID: "req-3",
	})

	// Unfiltered — all 3
	req := httptest.NewRequest("GET", "/v1/audit/actions?page=1&page_size=100", nil)
	w := httptest.NewRecorder()
	handler.HandleListActions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["total"].(float64) != 3 {
		t.Fatalf("expected total 3, got %v", resp["total"])
	}

	// Filter by actor_id
	req2 := httptest.NewRequest("GET", "/v1/audit/actions?actor_id=admin-001&page=1&page_size=100", nil)
	w2 := httptest.NewRecorder()
	handler.HandleListActions(w2, req2)

	var resp2 map[string]interface{}
	_ = json.Unmarshal(w2.Body.Bytes(), &resp2)
	if resp2["total"].(float64) != 2 {
		t.Fatalf("expected 2 actions for admin-001, got %v", resp2["total"])
	}

	// Filter by action_type
	req3 := httptest.NewRequest("GET", "/v1/audit/actions?action_type=staff.created&page=1&page_size=100", nil)
	w3 := httptest.NewRecorder()
	handler.HandleListActions(w3, req3)

	var resp3 map[string]interface{}
	_ = json.Unmarshal(w3.Body.Bytes(), &resp3)
	if resp3["total"].(float64) != 1 {
		t.Fatalf("expected 1 action for staff.created, got %v", resp3["total"])
	}
}
