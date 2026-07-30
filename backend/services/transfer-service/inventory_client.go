package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zippyra/backend/shared/jwt"
)

type TransferItemPayload struct {
	Barcode string `json:"barcode"`
	Qty     int    `json:"qty"`
}

type InventoryClient interface {
	ApplyTransferOut(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error
	ApplyTransferIn(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error
}

type HTTPInventoryClient struct {
	baseURL    string
	httpClient *http.Client
	jwtSecret  string
}

func NewHTTPInventoryClient(baseURL string, jwtSecret string) InventoryClient {
	if baseURL == "" {
		baseURL = "http://localhost:8088"
	}
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &HTTPInventoryClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		jwtSecret: jwtSecret,
	}
}

func (c *HTTPInventoryClient) generateSystemToken() string {
	claims := &jwt.Claims{
		UserID:   "system-transfer-client",
		UserType: "SYSTEM",
		Role:     "SYSTEM",
	}
	token, _ := jwt.GenerateToken(claims, c.jwtSecret, 1*time.Hour)
	return token
}

func (c *HTTPInventoryClient) ApplyTransferOut(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error {
	url := fmt.Sprintf("%s/v1/inventory/internal/apply-transfer-out", c.baseURL)

	payload := map[string]interface{}{
		"store_id":    storeID,
		"transfer_id": transferID,
		"items":       items,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.generateSystemToken())
	req.Header.Set("X-User-Type", "SYSTEM")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *HTTPInventoryClient) ApplyTransferIn(ctx context.Context, storeID, transferID string, items []TransferItemPayload) error {
	url := fmt.Sprintf("%s/v1/inventory/internal/apply-transfer-in", c.baseURL)

	payload := map[string]interface{}{
		"store_id":    storeID,
		"transfer_id": transferID,
		"items":       items,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.generateSystemToken())
	req.Header.Set("X-User-Type", "SYSTEM")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
