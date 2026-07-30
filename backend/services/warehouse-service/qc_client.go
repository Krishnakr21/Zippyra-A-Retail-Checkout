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
	"github.com/zippyra/backend/shared/logger"
)

type QCLineItemCreatePayload struct {
	GRNLineItemID string  `json:"grn_line_item_id"`
	Barcode       string  `json:"barcode"`
	QtyReceived   int     `json:"qty_received"`
	QCStatus      string  `json:"qc_status,omitempty"`
	QCNote        *string `json:"qc_note,omitempty"`
}

type QCLineItemUpdatePayload struct {
	GRNLineItemID string  `json:"grn_line_item_id"`
	QCStatus      string  `json:"qc_status"`
	QCNote        *string `json:"qc_note,omitempty"`
}

type QCLineItemResponse struct {
	GRNLineItemID string  `json:"grn_line_item_id"`
	Barcode       string  `json:"barcode"`
	QtyReceived   int     `json:"qty_received"`
	QCStatus      string  `json:"qc_status"`
	QCNote        *string `json:"qc_note,omitempty"`
}

type QCReviewResponse struct {
	ID            string               `json:"id"`
	GRNID         string               `json:"grn_id"`
	StoreID       string               `json:"store_id"`
	LineItems     []QCLineItemResponse `json:"line_items"`
	OverallStatus string               `json:"overall_status"`
	ReviewedBy    *string              `json:"reviewed_by,omitempty"`
	CompletedAt   *time.Time           `json:"completed_at,omitempty"`
}

type QCClient interface {
	CreateReview(ctx context.Context, grnID, storeID string, items []QCLineItemCreatePayload) (*QCReviewResponse, error)
	GetReview(ctx context.Context, grnID string) (*QCReviewResponse, error)
	UpdateReview(ctx context.Context, grnID string, updates []QCLineItemUpdatePayload) (*QCReviewResponse, error)
	IsReviewComplete(ctx context.Context, grnID string) (bool, error)
}

type httpQCClient struct {
	baseURL    string
	httpClient *http.Client
	jwtSecret  string
}

func NewQCClient(baseURL string, jwtSecret string) QCClient {
	if baseURL == "" {
		baseURL = "http://localhost:8089"
	}
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &httpQCClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		jwtSecret: jwtSecret,
	}
}

func (c *httpQCClient) generateSystemToken() string {
	claims := &jwt.Claims{
		UserID:   "system-qc-client",
		UserType: "SYSTEM",
		Role:     "SYSTEM",
	}
	token, _ := jwt.GenerateToken(claims, c.jwtSecret, 1*time.Hour)
	return token
}

func (c *httpQCClient) CreateReview(ctx context.Context, grnID, storeID string, items []QCLineItemCreatePayload) (*QCReviewResponse, error) {
	url := fmt.Sprintf("%s/v1/qc/internal/reviews", c.baseURL)

	payload := map[string]interface{}{
		"grn_id":     grnID,
		"store_id":   storeID,
		"line_items": items,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.generateSystemToken())
	req.Header.Set("X-User-Type", "SYSTEM")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var res QCReviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *httpQCClient) GetReview(ctx context.Context, grnID string) (*QCReviewResponse, error) {
	url := fmt.Sprintf("%s/v1/qc/internal/reviews/%s", c.baseURL, grnID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.generateSystemToken())
	req.Header.Set("X-User-Type", "SYSTEM")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var res QCReviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *httpQCClient) UpdateReview(ctx context.Context, grnID string, updates []QCLineItemUpdatePayload) (*QCReviewResponse, error) {
	url := fmt.Sprintf("%s/v1/qc/internal/reviews/%s", c.baseURL, grnID)

	payload := map[string]interface{}{
		"line_item_updates": updates,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.generateSystemToken())
	req.Header.Set("X-User-Type", "SYSTEM")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var res QCReviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *httpQCClient) IsReviewComplete(ctx context.Context, grnID string) (bool, error) {
	url := fmt.Sprintf("%s/v1/qc/internal/reviews/%s/is-complete", c.baseURL, grnID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+c.generateSystemToken())
	req.Header.Set("X-User-Type", "SYSTEM")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("[QCClient] IsReviewComplete returned status %d for grn %s", resp.StatusCode, grnID)
		return false, nil
	}

	var res struct {
		IsComplete bool `json:"is_complete"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, err
	}
	return res.IsComplete, nil
}
