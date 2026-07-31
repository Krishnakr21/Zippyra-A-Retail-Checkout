package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zippyra/backend/shared/jwt"
)

func setupFeedbackTestDB(t *testing.T) *sql.DB {
	dbName := fmt.Sprintf("file:fbdb_%s?mode=memory&cache=shared", t.Name())
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		t.Fatalf("failed to open sqlite memory DB: %v", err)
	}

	schema := `
		CREATE TABLE feedback_submissions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			user_type TEXT NOT NULL,
			source_app TEXT NOT NULL,
			nps_score INTEGER,
			comment TEXT,
			context TEXT NOT NULL DEFAULT 'general',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	return db
}

func TestFeedbackService(t *testing.T) {
	db := setupFeedbackTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	handler := &TicketHandler{repo: repo}

	router := mux.NewRouter()
	router.HandleFunc("/v1/support/feedback", handler.SubmitFeedback).Methods(http.MethodPost)
	router.HandleFunc("/v1/support/feedback", handler.ListFeedback).Methods(http.MethodGet)

	jwtSecret := "zippyra-dev-jwt-secret-key-32bytes"
	userID := "usr-feedback-001"
	token, _ := jwt.GenerateAccessToken(userID, "dev-1", jwtSecret, 15*time.Minute)
	claims, _ := jwt.ParseAndVerifyToken(token, jwtSecret)

	t.Run("POST /v1/support/feedback succeeds with optional fields", func(t *testing.T) {
		score := 9
		comment := "Super smooth exit gate experience!"
		reqBody, _ := json.Marshal(CreateFeedbackRequest{
			NPSScore:  &score,
			Comment:   &comment,
			SourceApp: "CUSTOMER_APP",
			Context:   "post_checkout",
		})

		req := httptest.NewRequest(http.MethodPost, "/v1/support/feedback", bytes.NewReader(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp["id"] == "" {
			t.Errorf("expected non-empty feedback ID")
		}
	})

	t.Run("POST /v1/support/feedback succeeds with no optional fields", func(t *testing.T) {
		reqBody, _ := json.Marshal(CreateFeedbackRequest{
			SourceApp: "RETAILER_DASHBOARD",
		})

		req := httptest.NewRequest(http.MethodPost, "/v1/support/feedback", bytes.NewReader(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("GET /v1/support/feedback filters by source_app and min_score", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/support/feedback?source_app=CUSTOMER_APP&min_score=8", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Feedbacks []*FeedbackSubmission `json:"feedbacks"`
			Total     int                  `json:"total"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode list response: %v", err)
		}

		if resp.Total < 1 {
			t.Errorf("expected at least 1 feedback, got %d", resp.Total)
		}
	})
}
