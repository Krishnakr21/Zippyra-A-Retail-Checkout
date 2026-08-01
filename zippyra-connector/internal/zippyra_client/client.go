package zippyra_client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/logging"
)

type SyncJob struct {
	JobID           string                 `json:"job_id"`
	ConnectionID    string                 `json:"connection_id"`
	SourceEventType string                 `json:"source_event_type"` // CATALOG_PRICE_CHANGED, STOCK_ADJUSTED, GRN_CREATED
	Payload         map[string]interface{} `json:"payload"`
	CreatedAt       string                 `json:"created_at"`
}

type Client struct {
	baseURL       string
	connectionID  string
	agentAPIKey   string
	webhookSecret string
	httpClient    *http.Client
	logger        *logging.Logger
	maxRetries    int
}

func NewClient(baseURL, connectionID, agentAPIKey, webhookSecret string, logger *logging.Logger) *Client {
	return &Client{
		baseURL:       strings.TrimRight(baseURL, "/"),
		connectionID:  connectionID,
		agentAPIKey:   agentAPIKey,
		webhookSecret: webhookSecret,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger:     logger,
		maxRetries: 3,
	}
}

func (c *Client) PullQueue(ctx context.Context) ([]SyncJob, error) {
	url := fmt.Sprintf("%s/v1/integration/connections/%s/pull-queue", c.baseURL, c.connectionID)

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create pull request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.agentAPIKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logger.Warn("[ZippyraClient] Pull attempt %d failed: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
			c.logger.Warn("[ZippyraClient] Pull attempt %d returned %v", attempt, lastErr)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		var jobs []SyncJob
		if err := json.Unmarshal(bodyBytes, &jobs); err != nil {
			// Check if response is wrapped in {"jobs": [...]}
			var wrapped struct {
				Jobs []SyncJob `json:"jobs"`
			}
			if errWrap := json.Unmarshal(bodyBytes, &wrapped); errWrap == nil {
				return wrapped.Jobs, nil
			}
			return nil, fmt.Errorf("failed to decode pull queue response: %w", err)
		}

		return jobs, nil
	}

	return nil, fmt.Errorf("pull queue failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *Client) AckQueue(ctx context.Context, jobIDs []string) error {
	if len(jobIDs) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s/v1/integration/connections/%s/pull-queue/ack", c.baseURL, c.connectionID)

	payload := map[string]interface{}{
		"job_ids": jobIDs,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal ack payload: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create ack request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.agentAPIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logger.Warn("[ZippyraClient] Ack attempt %d failed: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
			c.logger.Warn("[ZippyraClient] Ack attempt %d returned %v", attempt, lastErr)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		c.logger.Info("[ZippyraClient] Successfully acked %d jobs", len(jobIDs))
		return nil
	}

	return fmt.Errorf("ack queue failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *Client) SendWebhook(ctx context.Context, change erp_adapter.LocalChange) error {
	url := fmt.Sprintf("%s/v1/integration/connections/%s/webhook", c.baseURL, c.connectionID)

	bodyBytes, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook change payload: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write(bodyBytes)
	signature := hex.EncodeToString(mac.Sum(nil))

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.agentAPIKey)
		req.Header.Set("X-Signature", signature)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logger.Warn("[ZippyraClient] Webhook attempt %d failed: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
			c.logger.Warn("[ZippyraClient] Webhook attempt %d returned %v", attempt, lastErr)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		c.logger.Info("[ZippyraClient] Webhook successfully sent for local change event=%s barcode=%s", change.EventType, change.Barcode)
		return nil
	}

	return fmt.Errorf("send webhook failed after %d retries: %w", c.maxRetries, lastErr)
}
