package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/zippyra/backend/shared/jwt"
)

func TestReplay_ConcurrentValidate_SingleOpen(t *testing.T) {
	db, redisClient, mqttClient, handler := setupTestEnvironment(t)
	defer db.Close()
	defer redisClient.Close()

	secret := "secret-32bytes-secret-32bytes-12"
	exitToken, _ := jwt.GenerateExitToken("ord-replay-100", "user-1", "store-1", "", secret, 10*time.Minute)
	deviceToken, _ := jwt.GenerateDeviceToken("dev-1", "gate-A", "store-1", secret, 24*time.Hour)

	const concurrency = 20
	var wg sync.WaitGroup
	wg.Add(concurrency)

	results := make([]ValidateExitResponse, concurrency)
	resultsMu := sync.Mutex{}

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()

			body, _ := json.Marshal(ValidateExitRequest{ExitToken: exitToken})
			req := httptest.NewRequest(http.MethodPost, "/v1/exit/validate", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+deviceToken)

			rec := httptest.NewRecorder()
			handler.ValidateExitHandler(rec, req)

			var resp ValidateExitResponse
			_ = json.Unmarshal(rec.Body.Bytes(), &resp)

			resultsMu.Lock()
			results[idx] = resp
			resultsMu.Unlock()
		}(i)
	}

	wg.Wait()

	openCount := 0
	replayCount := 0

	for _, res := range results {
		if res.Result == "OPEN" {
			openCount++
		} else if res.Result == "DENY" && res.Reason == ResultQRAlreadyUsed {
			replayCount++
		}
	}

	if openCount != 1 {
		t.Fatalf("Expected exactly 1 OPEN response among 20 concurrent attempts, got %d", openCount)
	}

	if replayCount != concurrency-1 {
		t.Fatalf("Expected %d QR_ALREADY_USED rejections, got %d", concurrency-1, replayCount)
	}

	// Verify MQTT published exactly 1 OPEN command
	published := mqttClient.GetPublished()
	if len(published) != 1 {
		t.Fatalf("Expected exactly 1 MQTT OPEN command sent, got %d", len(published))
	}
}
