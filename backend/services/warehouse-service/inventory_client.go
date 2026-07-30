package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/logger"
)

type InventoryClient interface {
	ApplyGRN(ctx context.Context, storeID, grnID string, items []GRNItemPayload) (*ApplyGRNClientResponse, error)
	ApplyTransferOut(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error
	ApplyTransferIn(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error
}

type GRNItemPayload struct {
	Barcode       string `json:"barcode"`
	QtyReceived   int    `json:"qty_received"`
	UnitCostPaise int64  `json:"unit_cost_paise"`
}

type TransferItemPayload struct {
	Barcode string `json:"barcode"`
	Qty     int    `json:"qty"`
}

type ApplyGRNClientResponse struct {
	Applied        bool `json:"applied"`
	ItemsRequested int  `json:"items_requested"`
	ItemsApplied   int  `json:"items_applied"`
}

type HTTPInventoryClient struct {
	baseURL   string
	jwtSecret string
	client    *http.Client
}

func NewHTTPInventoryClient(baseURL, jwtSecret string) *HTTPInventoryClient {
	return &HTTPInventoryClient{
		baseURL:   baseURL,
		jwtSecret: jwtSecret,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *HTTPInventoryClient) generateSystemToken() string {
	token, _ := jwt.GenerateSessionToken("system-warehouse-service", "SYSTEM", "", "", c.jwtSecret, 1*time.Hour)
	return token
}

func (c *HTTPInventoryClient) ApplyGRN(ctx context.Context, storeID, grnID string, items []GRNItemPayload) (*ApplyGRNClientResponse, error) {
	url := c.baseURL + "/v1/inventory/internal/apply-grn"
	payload := map[string]interface{}{
		"store_id": storeID,
		"grn_id":   grnID,
		"items":    items,
	}

	var resp ApplyGRNClientResponse
	err := c.doRequestWithRetry(ctx, http.MethodPost, url, payload, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *HTTPInventoryClient) ApplyTransferOut(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error {
	url := c.baseURL + "/v1/inventory/internal/apply-transfer-out"
	payload := map[string]interface{}{
		"store_id":    storeID,
		"transfer_id": transferID,
		"items":       items,
	}
	return c.doRequestWithRetry(ctx, http.MethodPost, url, payload, nil)
}

func (c *HTTPInventoryClient) ApplyTransferIn(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error {
	url := c.baseURL + "/v1/inventory/internal/apply-transfer-in"
	payload := map[string]interface{}{
		"store_id":    storeID,
		"transfer_id": transferID,
		"items":       items,
	}
	return c.doRequestWithRetry(ctx, http.MethodPost, url, payload, nil)
}

func (c *HTTPInventoryClient) doRequestWithRetry(ctx context.Context, method, url string, body interface{}, out interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	maxAttempts := 2
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create http request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.generateSystemToken())

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("network error calling inventory-service: %w", err)
			logger.Warn("[INVENTORY CLIENT] Attempt %d failed: %v", attempt, lastErr)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out != nil {
				if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
					return fmt.Errorf("failed to decode inventory response: %w", err)
				}
			}
			return nil
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			// Do NOT retry 4xx errors! Surface 4xx immediately.
			var errResp struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&errResp)
			return fmt.Errorf("HTTP_%d: %s - %s", resp.StatusCode, errResp.Code, errResp.Message)
		}

		// 5xx status -> eligible for retry
		lastErr = fmt.Errorf("HTTP_%d: server error from inventory-service", resp.StatusCode)
		logger.Warn("[INVENTORY CLIENT] Attempt %d returned 5xx status %d", attempt, resp.StatusCode)
		time.Sleep(100 * time.Millisecond)
	}

	return lastErr
}
