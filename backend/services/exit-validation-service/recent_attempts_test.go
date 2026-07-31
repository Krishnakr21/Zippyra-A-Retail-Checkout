package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetRecentExitAttempts_ReturnsAttemptsWithFullDetail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	handler := NewExitHandler(repo, nil, nil, nil, nil, nil, "secret")

	ctx := context.Background()

	// Insert test exit attempt
	attempt := &ExitAttempt{
		OrderID:    "ord-recent-1",
		UserID:     "user-100",
		StoreID:    "store-recent-1",
		GateID:     "gate-A",
		Result:     "WRONG_STORE",
		IsAlarm:    true,
		RFIDTagIDs: json.RawMessage(`["rfid-1"]`),
	}
	_ = repo.CreateExitAttempt(ctx, attempt)

	req := httptest.NewRequest("GET", "/v1/exit/recent-attempts?store_id=store-recent-1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.HandleGetRecentExitAttempts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	attemptsList, ok := resp["attempts"].([]interface{})
	if !ok || len(attemptsList) == 0 {
		t.Fatalf("expected non-empty attempts list, got %v", resp)
	}

	firstAttempt := attemptsList[0].(map[string]interface{})
	if firstAttempt["order_id"] != "ord-recent-1" || firstAttempt["result"] != "WRONG_STORE" || firstAttempt["is_alarm"] != true {
		t.Fatalf("expected full exit attempt detail with is_alarm=true, got %v", firstAttempt)
	}
}
