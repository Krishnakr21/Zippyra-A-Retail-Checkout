package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/jwt"
)

func TestDeviceSessionsManagement(t *testing.T) {
	repo := NewMemoryRepository()
	handler := NewAuthHandler(repo, nil, nil)
	jwtSecret := "zippyra-dev-jwt-secret-key-32bytes"
	router := mux.NewRouter()

	router.HandleFunc("/v1/auth/sessions", handler.HandleGetUserSessions).Methods(http.MethodGet)
	router.HandleFunc("/v1/auth/sessions", handler.HandleRevokeAllOtherSessions).Methods(http.MethodDelete)
	router.HandleFunc("/v1/auth/sessions/{id}", handler.HandleRevokeSessionByID).Methods(http.MethodDelete)

	user, _ := repo.CreateUserWithPhone(nil, "+919999988888")

	// Issue 3 sessions: session-1 (current), session-2, session-3
	token1, _, _ := issueSession(nil, repo, user, "device-1", "iPhone 14", jwtSecret)
	claims1, _ := jwt.ParseAndVerifyToken(token1, jwtSecret)
	currentSessionID := claims1.SessionID

	token2, _, _ := issueSession(nil, repo, user, "device-2", "Pixel 7", jwtSecret)
	claims2, _ := jwt.ParseAndVerifyToken(token2, jwtSecret)
	session2ID := claims2.SessionID

	_, _, _ = issueSession(nil, repo, user, "device-3", "iPad Air", jwtSecret)

	t.Run("GET /v1/auth/sessions lists active user sessions with is_current flag", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/auth/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+token1)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Sessions []SessionDTO `json:"sessions"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Sessions) != 3 {
			t.Fatalf("expected 3 sessions, got %d", len(resp.Sessions))
		}

		currentCount := 0
		for _, s := range resp.Sessions {
			if s.IsCurrent {
				currentCount++
				if s.ID != currentSessionID {
					t.Errorf("expected current session ID %s, got %s", currentSessionID, s.ID)
				}
			}
		}
		if currentCount != 1 {
			t.Errorf("expected exactly 1 current session, got %d", currentCount)
		}
	})

	t.Run("DELETE /v1/auth/sessions/{id} on current session returns 400 USE_LOGOUT_FOR_CURRENT_SESSION", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/sessions/"+currentSessionID, nil)
		req.Header.Set("Authorization", "Bearer "+token1)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "USE_LOGOUT_FOR_CURRENT_SESSION") {
			t.Errorf("expected error code USE_LOGOUT_FOR_CURRENT_SESSION, got body: %s", rr.Body.String())
		}
	})

	t.Run("DELETE /v1/auth/sessions/{id} revokes specific other session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/sessions/"+session2ID, nil)
		req.Header.Set("Authorization", "Bearer "+token1)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		sessions, _ := repo.GetUserSessions(nil, user.ID)
		for _, s := range sessions {
			if s.ID == session2ID {
				t.Errorf("session2 should have been removed from active sessions")
			}
		}
	})

	t.Run("DELETE /v1/auth/sessions revokes all other sessions except current", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+token1)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		sessions, _ := repo.GetUserSessions(nil, user.ID)
		if len(sessions) != 1 {
			t.Fatalf("expected only 1 session remaining (current), got %d", len(sessions))
		}
		if sessions[0].ID != currentSessionID {
			t.Errorf("expected remaining session to be currentSessionID %s, got %s", currentSessionID, sessions[0].ID)
		}
	})
}
