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

type TransferClient interface {
	CreateTransfer(ctx context.Context, req CreateTransferRequest, claims *jwt.SessionClaims) (*TransferOrder, error)
	ListTransfers(ctx context.Context, chainID, status string) ([]*TransferOrder, error)
	GetTransfer(ctx context.Context, id string) (*TransferOrder, error)
	ApproveTransfer(ctx context.Context, id string) (map[string]interface{}, error)
	RejectTransfer(ctx context.Context, id string, reason string) (map[string]interface{}, error)
	ShipTransfer(ctx context.Context, id string, req ShipTransferRequest) (map[string]interface{}, error)
	ReceiveTransfer(ctx context.Context, id string, req ReceiveTransferRequest) (map[string]interface{}, error)
}

type httpTransferClient struct {
	baseURL    string
	httpClient *http.Client
	jwtSecret  string
}

func NewTransferClient(baseURL string, jwtSecret string) TransferClient {
	if baseURL == "" {
		baseURL = "http://localhost:8090"
	}
	if jwtSecret == "" {
		jwtSecret = "zippyra-dev-jwt-secret-key-32bytes"
	}
	return &httpTransferClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		jwtSecret: jwtSecret,
	}
}

func (c *httpTransferClient) generateSystemToken() string {
	claims := &jwt.Claims{
		UserID:   "system-transfer-client",
		UserType: "SYSTEM",
		Role:     "SYSTEM",
	}
	token, _ := jwt.GenerateToken(claims, c.jwtSecret, 1*time.Hour)
	return token
}

func (c *httpTransferClient) CreateTransfer(ctx context.Context, req CreateTransferRequest, claims *jwt.SessionClaims) (*TransferOrder, error) {
	url := fmt.Sprintf("%s/v1/transfer/internal/transfers", c.baseURL)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.generateSystemToken())
	if claims != nil {
		httpReq.Header.Set("X-User-ID", claims.UserID)
		httpReq.Header.Set("X-User-Type", claims.UserType)
		httpReq.Header.Set("X-User-Role", claims.Role)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var res TransferOrder
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *httpTransferClient) ListTransfers(ctx context.Context, chainID, status string) ([]*TransferOrder, error) {
	url := fmt.Sprintf("%s/v1/transfer/internal/transfers?chain_id=%s&status=%s", c.baseURL, chainID, status)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.generateSystemToken())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var res struct {
		Transfers []*TransferOrder `json:"transfers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.Transfers, nil
}

func (c *httpTransferClient) GetTransfer(ctx context.Context, id string) (*TransferOrder, error) {
	url := fmt.Sprintf("%s/v1/transfer/internal/transfers/%s", c.baseURL, id)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.generateSystemToken())

	resp, err := c.httpClient.Do(httpReq)
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

	var res TransferOrder
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *httpTransferClient) ApproveTransfer(ctx context.Context, id string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/transfer/internal/transfers/%s/approve", c.baseURL, id)

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.generateSystemToken())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var res map[string]interface{}
	_ = json.Unmarshal(respBody, &res)
	return res, nil
}

func (c *httpTransferClient) RejectTransfer(ctx context.Context, id string, reason string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/transfer/internal/transfers/%s/reject", c.baseURL, id)

	payload := map[string]string{"reason": reason}
	bodyBytes, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.generateSystemToken())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var res map[string]interface{}
	_ = json.Unmarshal(respBody, &res)
	return res, nil
}

func (c *httpTransferClient) ShipTransfer(ctx context.Context, id string, req ShipTransferRequest) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/transfer/internal/transfers/%s/ship", c.baseURL, id)

	bodyBytes, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.generateSystemToken())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var res map[string]interface{}
	_ = json.Unmarshal(respBody, &res)
	return res, nil
}

func (c *httpTransferClient) ReceiveTransfer(ctx context.Context, id string, req ReceiveTransferRequest) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/transfer/internal/transfers/%s/receive", c.baseURL, id)

	bodyBytes, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.generateSystemToken())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var res map[string]interface{}
	_ = json.Unmarshal(respBody, &res)
	return res, nil
}
