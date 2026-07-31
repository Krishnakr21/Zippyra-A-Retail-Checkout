package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zippyra/backend/shared/jwt"
)

func setupSubscriptionTestDB(t *testing.T) *sql.DB {
	dbName := fmt.Sprintf("file:subdb_%s?mode=memory&cache=shared", t.Name())
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		t.Fatalf("failed to open sqlite memory DB: %v", err)
	}

	schema := `
		CREATE TABLE subscription_plans (
			id TEXT PRIMARY KEY,
			chain_id TEXT NOT NULL,
			name TEXT NOT NULL,
			price_paise INTEGER NOT NULL,
			billing_interval TEXT NOT NULL,
			benefits TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE member_subscriptions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			plan_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'PENDING',
			razorpay_subscription_id TEXT UNIQUE,
			current_period_end TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE subscription_webhook_events (
			event_id TEXT PRIMARY KEY,
			event_type TEXT NOT NULL,
			processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	return db
}

func TestSubscriptionService(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	jwtSecret := "zippyra-dev-jwt-secret-key-32bytes"
	whSecret := "whsec_test_secret_12345"

	handler := &SubscriptionHandler{
		repo:          repo,
		jwtSecret:     jwtSecret,
		razorpayKeyID: "rzp_test_key_100",
		webhookSecret: whSecret,
	}

	routes := SetupRoutes(handler)
	userID := "usr-sub-100"
	token, _ := jwt.GenerateAccessToken(userID, "dev-1", jwtSecret, 15*time.Minute)
	claims, _ := jwt.ParseAndVerifyToken(token, jwtSecret)

	t.Run("GET /v1/subscription/plans lists active plans", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/subscription/plans?chain_id=chain-hq-001", nil)
		rr := httptest.NewRecorder()

		routes.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Plans []*SubscriptionPlan `json:"plans"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode plans response: %v", err)
		}

		if len(resp.Plans) < 2 {
			t.Fatalf("expected at least 2 default plans, got %d", len(resp.Plans))
		}
	})

	t.Run("POST /v1/subscription/subscribe creates active subscription", func(t *testing.T) {
		body, _ := json.Marshal(SubscribeRequest{PlanID: "plan-smart-saver-monthly"})
		req := httptest.NewRequest(http.MethodPost, "/v1/subscription/subscribe", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
		rr := httptest.NewRecorder()

		routes.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp SubscribeResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode subscribe response: %v", err)
		}

		if resp.Status != "ACTIVE" || resp.RazorpaySubscriptionID == "" {
			t.Errorf("unexpected subscribe response: status=%s, sub_id=%s", resp.Status, resp.RazorpaySubscriptionID)
		}
	})

	t.Run("GET /v1/subscription/mine returns current active subscription", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/subscription/mine", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
		rr := httptest.NewRecorder()

		routes.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Subscription *MemberSubscription `json:"subscription"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode mine response: %v", err)
		}

		if resp.Subscription == nil || resp.Subscription.Status != "ACTIVE" {
			t.Fatalf("expected active subscription, got %v", resp.Subscription)
		}
	})

	t.Run("POST /v1/subscription/webhook/razorpay verifies HMAC and is idempotent", func(t *testing.T) {
		sub, _ := repo.GetActiveUserSubscription(context.Background(), userID)
		rzpSubID := *sub.RazorpaySubscriptionID

		whPayload := RazorpaySubWebhookPayload{
			Event:     "subscription.cancelled",
			EventID:   "evt_rzp_cancel_100",
			CreatedAt: time.Now().Unix(),
		}
		whPayload.Payload.Subscription.Entity.ID = rzpSubID

		rawBody, _ := json.Marshal(whPayload)
		mac := hmac.New(sha256.New, []byte(whSecret))
		mac.Write(rawBody)
		signature := hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest(http.MethodPost, "/v1/subscription/webhook/razorpay", bytes.NewReader(rawBody))
		req.Header.Set("X-Razorpay-Signature", signature)
		rr := httptest.NewRecorder()

		routes.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for webhook, got %d: %s", rr.Code, rr.Body.String())
		}

		// Verify status updated to CANCELLED
		cancelledSub, _ := repo.GetActiveUserSubscription(context.Background(), userID)
		if cancelledSub != nil {
			t.Errorf("expected no active subscription after cancellation webhook")
		}

		// Re-send same webhook event ID (idempotency check)
		req2 := httptest.NewRequest(http.MethodPost, "/v1/subscription/webhook/razorpay", bytes.NewReader(rawBody))
		req2.Header.Set("X-Razorpay-Signature", signature)
		rr2 := httptest.NewRecorder()

		routes.ServeHTTP(rr2, req2)

		if rr2.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on duplicate webhook, got %d: %s", rr2.Code, rr2.Body.String())
		}
	})

	t.Run("POST /v1/subscription/cancel cancels active subscription", func(t *testing.T) {
		// Re-subscribe
		body, _ := json.Marshal(SubscribeRequest{PlanID: "plan-smart-saver-monthly"})
		reqSub := httptest.NewRequest(http.MethodPost, "/v1/subscription/subscribe", bytes.NewReader(body))
		reqSub.Header.Set("Authorization", "Bearer "+token)
		reqSub = reqSub.WithContext(context.WithValue(reqSub.Context(), "user_claims", claims))
		rrSub := httptest.NewRecorder()
		routes.ServeHTTP(rrSub, reqSub)

		// Cancel
		reqCancel := httptest.NewRequest(http.MethodPost, "/v1/subscription/cancel", nil)
		reqCancel.Header.Set("Authorization", "Bearer "+token)
		reqCancel = reqCancel.WithContext(context.WithValue(reqCancel.Context(), "user_claims", claims))
		rrCancel := httptest.NewRecorder()

		routes.ServeHTTP(rrCancel, reqCancel)

		if rrCancel.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on cancel, got %d: %s", rrCancel.Code, rrCancel.Body.String())
		}

		activeSub, _ := repo.GetActiveUserSubscription(context.Background(), userID)
		if activeSub != nil {
			t.Errorf("expected subscription to be cancelled")
		}
	})
}
