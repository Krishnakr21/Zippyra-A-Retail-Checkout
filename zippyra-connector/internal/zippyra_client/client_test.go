package zippyra_client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/logging"
)

func TestZippyraClient_PullAckAndWebhook(t *testing.T) {
	connID := "conn-test-123"
	apiKey := "agent-api-key-abc"
	webhookSecret := "webhook-secret-xyz"

	logger, _ := logging.NewLogger("")

	var receivedHMAC string
	var ackedJobs []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Bearer Auth
		if r.Header.Get("Authorization") != "Bearer "+apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/v1/integration/connections/" + connID + "/pull-queue":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"job_id": "job-1",
					"connection_id": "conn-test-123",
					"source_event_type": "CATALOG_PRICE_CHANGED",
					"payload": {"barcode": "8901001", "price_paise": 5000}
				}
			]`))

		case "/v1/integration/connections/" + connID + "/pull-queue/ack":
			body, _ := io.ReadAll(r.Body)
			ackedJobs = append(ackedJobs, string(body))
			w.WriteHeader(http.StatusOK)

		case "/v1/integration/connections/" + connID + "/webhook":
			receivedHMAC = r.Header.Get("X-Signature")
			w.WriteHeader(http.StatusOK)

		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := NewClient(ts.URL, connID, apiKey, webhookSecret, logger)
	ctx := context.Background()

	// 1. Pull Queue Test
	jobs, err := client.PullQueue(ctx)
	if err != nil {
		t.Fatalf("PullQueue failed: %v", err)
	}
	if len(jobs) != 1 || jobs[0].JobID != "job-1" {
		t.Errorf("Expected 1 job 'job-1', got: %v", jobs)
	}

	// 2. Ack Queue Test
	err = client.AckQueue(ctx, []string{"job-1"})
	if err != nil {
		t.Fatalf("AckQueue failed: %v", err)
	}
	if len(ackedJobs) == 0 {
		t.Errorf("Expected ack call to server")
	}

	// 3. Webhook HMAC Test
	change := erp_adapter.LocalChange{
		EventType: "CATALOG_PRICE_CHANGED",
		Barcode:   "8901001",
		Payload:   map[string]interface{}{"price_paise": 5500},
		Timestamp: time.Now(),
	}
	err = client.SendWebhook(ctx, change)
	if err != nil {
		t.Fatalf("SendWebhook failed: %v", err)
	}

	if receivedHMAC == "" {
		t.Errorf("Expected X-Signature header on webhook call")
	}

	// Verify HMAC calculation manually
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	// Marshaled change
	bodyBytes, _ := json.Marshal(change)
	mac.Write(bodyBytes)
	expectedHMAC := hex.EncodeToString(mac.Sum(nil))

	if receivedHMAC != expectedHMAC {
		t.Errorf("HMAC mismatch! Expected %s, got %s", expectedHMAC, receivedHMAC)
	}
}
