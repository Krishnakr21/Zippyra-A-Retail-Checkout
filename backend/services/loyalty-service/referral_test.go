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
	"github.com/zippyra/backend/shared/jwt"
)

func TestReferralProgram(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	customerHandler := NewCustomerHandler(repo)
	jwtSecret := "zippyra-dev-jwt-secret-key-32bytes"
	ctx := context.Background()

	referrerUser, _ := repo.EnsureAccountExists(ctx, "usr-referrer-1")
	referredUser, _ := repo.EnsureAccountExists(ctx, "usr-referred-2")

	router := mux.NewRouter()
	router.HandleFunc("/v1/loyalty/referral-code", customerHandler.HandleGetReferralCode).Methods(http.MethodGet)
	router.HandleFunc("/v1/loyalty/referral/apply", customerHandler.HandleApplyReferral).Methods(http.MethodPost)

	t.Run("GET /v1/loyalty/referral-code returns code and share text", func(t *testing.T) {
		token, _ := jwt.GenerateAccessToken(referrerUser.UserID, "dev-1", jwtSecret, 15*time.Minute)
		claims, _ := jwt.ParseAndVerifyToken(token, jwtSecret)

		req := httptest.NewRequest(http.MethodGet, "/v1/loyalty/referral-code", nil)
		req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp ReferralCodeResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.ReferralCode != referrerUser.ReferralCode {
			t.Errorf("expected referral code %s, got %s", referrerUser.ReferralCode, resp.ReferralCode)
		}
		if resp.ReferrerRewardPoints != 100 || resp.ReferredRewardPoints != 50 {
			t.Errorf("unexpected reward points: referrer=%d, referred=%d", resp.ReferrerRewardPoints, resp.ReferredRewardPoints)
		}
	})

	t.Run("POST /v1/loyalty/referral/apply rejects self referral", func(t *testing.T) {
		token, _ := jwt.GenerateAccessToken(referrerUser.UserID, "dev-1", jwtSecret, 15*time.Minute)
		claims, _ := jwt.ParseAndVerifyToken(token, jwtSecret)

		body, _ := json.Marshal(ApplyReferralRequest{ReferralCode: referrerUser.ReferralCode})
		req := httptest.NewRequest(http.MethodPost, "/v1/loyalty/referral/apply", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("POST /v1/loyalty/referral/apply succeeds for valid code", func(t *testing.T) {
		token, _ := jwt.GenerateAccessToken(referredUser.UserID, "dev-2", jwtSecret, 15*time.Minute)
		claims, _ := jwt.ParseAndVerifyToken(token, jwtSecret)

		body, _ := json.Marshal(ApplyReferralRequest{ReferralCode: referrerUser.ReferralCode})
		req := httptest.NewRequest(http.MethodPost, "/v1/loyalty/referral/apply", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("POST /v1/loyalty/referral/apply rejects duplicate referral application", func(t *testing.T) {
		token, _ := jwt.GenerateAccessToken(referredUser.UserID, "dev-2", jwtSecret, 15*time.Minute)
		claims, _ := jwt.ParseAndVerifyToken(token, jwtSecret)

		body, _ := json.Marshal(ApplyReferralRequest{ReferralCode: referrerUser.ReferralCode})
		req := httptest.NewRequest(http.MethodPost, "/v1/loyalty/referral/apply", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), "user_claims", claims))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 Conflict, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Order completed triggers referral reward for both parties", func(t *testing.T) {
		// Process referred user's first order
		refUserID, refPts, refedPts, rewarded, err := repo.ProcessFirstOrderReferralReward(ctx, referredUser.UserID, "ord-first-001")
		if err != nil {
			t.Fatalf("ProcessFirstOrderReferralReward failed: %v", err)
		}

		if !rewarded || refUserID != referrerUser.UserID || refPts != 100 || refedPts != 50 {
			t.Fatalf("expected rewarded=true, refUserID=%s, refPts=100, refedPts=50; got rewarded=%v, refUserID=%s, refPts=%d, refedPts=%d",
				referrerUser.UserID, rewarded, refUserID, refPts, refedPts)
		}

		// Verify referrer points balance
		refAcc, _ := repo.GetAccountByUserID(ctx, referrerUser.UserID)
		if refAcc.PointsBalance != 100 {
			t.Errorf("expected referrer balance 100, got %d", refAcc.PointsBalance)
		}

		// Verify referred user points balance
		refedAcc, _ := repo.GetAccountByUserID(ctx, referredUser.UserID)
		if refedAcc.PointsBalance != 50 {
			t.Errorf("expected referred user balance 50, got %d", refedAcc.PointsBalance)
		}

		// Idempotency / duplicate order check
		_, _, _, secondReward, _ := repo.ProcessFirstOrderReferralReward(ctx, referredUser.UserID, "ord-second-002")
		if secondReward {
			t.Errorf("expected second order NOT to trigger referral reward")
		}
	})

	t.Run("Expired pending referral does not reward late first order", func(t *testing.T) {
		lateUser, err := repo.EnsureAccountExists(ctx, "usr-late-4")
		if err != nil || lateUser == nil {
			t.Fatalf("EnsureAccountExists failed for lateUser: %v", err)
		}

		err = repo.ApplyReferral(ctx, lateUser.UserID, referrerUser.ReferralCode)
		if err != nil {
			t.Fatalf("ApplyReferral failed: %v", err)
		}

		// Force expire the event
		db.Exec("UPDATE referral_events SET status = 'EXPIRED' WHERE referred_user_id = ?", lateUser.UserID)

		_, _, _, rewarded, _ := repo.ProcessFirstOrderReferralReward(ctx, lateUser.UserID, "ord-late-99")
		if rewarded {
			t.Errorf("expected expired referral to NOT be rewarded")
		}
	})
}
