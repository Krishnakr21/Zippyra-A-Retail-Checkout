package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zippyra/backend/shared/jwt"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Failed to open sqlite memory db: %v", err)
	}
	db.SetMaxOpenConns(1)

	schemas := []string{
		`CREATE TABLE IF NOT EXISTS exit_attempts (
			id VARCHAR(36) PRIMARY KEY,
			order_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			store_id VARCHAR(36) NOT NULL,
			gate_id VARCHAR(50) NOT NULL,
			result VARCHAR(30) NOT NULL,
			is_alarm BOOLEAN NOT NULL DEFAULT 0,
			rfid_tag_ids TEXT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS staff_overrides (
			id VARCHAR(36) PRIMARY KEY,
			order_id VARCHAR(36) NULL,
			store_id VARCHAR(36) NOT NULL,
			gate_id VARCHAR(50) NOT NULL,
			staff_user_id VARCHAR(36) NOT NULL,
			reason TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, s := range schemas {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("Failed to execute schema setup: %v", err)
		}
	}

	return db
}

func setupTestEnvironment(t *testing.T) (*sql.DB, *TestRedisClient, *MockMQTTClient, *ExitHandler) {
	db := setupTestDB(t)
	redisClient := NewTestRedisClient()

	mqttClient := NewMockMQTTClient()
	repo := NewPostgresRepository(db)
	secret := "secret-32bytes-secret-32bytes-12"
	verifier := NewJWTVerifier(secret, nil)
	metrics := NewAlarmMetrics()

	handler := NewExitHandler(repo, redisClient, nil, mqttClient, verifier, metrics, secret)
	return db, redisClient, mqttClient, handler
}

func TestValidate_ValidToken_OpensGate(t *testing.T) {
	db, redisClient, mqttClient, handler := setupTestEnvironment(t)
	defer db.Close()
	defer redisClient.Close()

	ctx := context.Background()
	secret := "secret-32bytes-secret-32bytes-12"

	// Create valid exit token
	exitToken, err := jwt.GenerateExitToken("ord-valid-1", "user-1", "store-1", "", secret, 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate exit token: %v", err)
	}

	// Create device token for Store 1 Gate A
	deviceToken, err := jwt.GenerateDeviceToken("dev-1", "gate-A", "store-1", secret, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate device token: %v", err)
	}

	body, _ := json.Marshal(ValidateExitRequest{ExitToken: exitToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/exit/validate", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+deviceToken)

	rec := httptest.NewRecorder()
	handler.ValidateExitHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var resp ValidateExitResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Result != "OPEN" || resp.OrderID != "ord-valid-1" {
		t.Fatalf("Unexpected validate response: %+v", resp)
	}

	// Check Redis key exit_used:ord-valid-1 is set
	usedVal, err := redisClient.Get(ctx, "exit_used:ord-valid-1").Result()
	if err != nil || usedVal != "1" {
		t.Fatalf("Expected exit_used key to be set in Redis, got val=%s, err=%v", usedVal, err)
	}

	// Check MQTT publish
	published := mqttClient.GetPublished()
	if len(published) != 1 || published[0].Cmd.Cmd != "OPEN" {
		t.Fatalf("Expected 1 MQTT OPEN command, got %d commands", len(published))
	}
}

func TestValidate_WrongStore_AlarmAndDeny(t *testing.T) {
	db, redisClient, mqttClient, handler := setupTestEnvironment(t)
	defer db.Close()
	defer redisClient.Close()

	secret := "secret-32bytes-secret-32bytes-12"

	// Exit token for Store A
	exitToken, _ := jwt.GenerateExitToken("ord-wrong-store", "user-1", "store-A", "", secret, 10*time.Minute)

	// Device token for Store B Gate B
	deviceToken, _ := jwt.GenerateDeviceToken("dev-2", "gate-B", "store-B", secret, 24*time.Hour)

	body, _ := json.Marshal(ValidateExitRequest{ExitToken: exitToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/exit/validate", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+deviceToken)

	rec := httptest.NewRecorder()
	handler.ValidateExitHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var resp ValidateExitResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Result != "DENY" || resp.Reason != ResultWrongStore {
		t.Fatalf("Expected DENY with WRONG_STORE reason, got %+v", resp)
	}

	// Verify NO MQTT open command was sent
	published := mqttClient.GetPublished()
	if len(published) != 0 {
		t.Fatalf("Expected 0 MQTT commands sent on store mismatch, got %d", len(published))
	}
}

func TestValidate_ExpiredToken_DenyWithoutAlarm(t *testing.T) {
	db, redisClient, _, handler := setupTestEnvironment(t)
	defer db.Close()
	defer redisClient.Close()

	secret := "secret-32bytes-secret-32bytes-12"

	// Expired Exit Token (-1 minute)
	exitToken, _ := jwt.GenerateExitToken("ord-expired", "user-1", "store-1", "", secret, -1*time.Minute)
	deviceToken, _ := jwt.GenerateDeviceToken("dev-1", "gate-A", "store-1", secret, 24*time.Hour)

	body, _ := json.Marshal(ValidateExitRequest{ExitToken: exitToken})
	req := httptest.NewRequest(http.MethodPost, "/v1/exit/validate", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+deviceToken)

	rec := httptest.NewRecorder()
	handler.ValidateExitHandler(rec, req)

	var resp ValidateExitResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Result != "DENY" || resp.Reason != ResultQRExpired {
		t.Fatalf("Expected DENY with QR_EXPIRED reason, got %+v", resp)
	}

	// Verify in DB is_alarm = false
	attempt, err := handler.repo.GetLatestExitAttemptByOrderID(context.Background(), "ord-expired")
	if err != nil || attempt == nil {
		t.Fatalf("Failed to fetch logged exit attempt: %v", err)
	}

	if attempt.IsAlarm {
		t.Fatalf("Expected is_alarm=false for expired token, got is_alarm=true")
	}
}
