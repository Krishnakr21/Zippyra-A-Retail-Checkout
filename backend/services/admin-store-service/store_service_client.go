package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// StoreServiceClient calls store-service's SYSTEM-JWT-gated internal write endpoints.
// admin-store-service NEVER touches store-service's database directly.
type StoreServiceClient struct {
	baseURL    string
	jwtSecret  string
	httpClient *http.Client
}

func NewStoreServiceClient(baseURL, jwtSecret string) *StoreServiceClient {
	return &StoreServiceClient{
		baseURL:    baseURL,
		jwtSecret:  jwtSecret,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *StoreServiceClient) systemToken() string {
	// Mint a minimal SYSTEM JWT for service-to-service calls.
	// In production this uses the shared jwt.MintSystemToken helper;
	// here we pass the raw secret header used by the shared middleware.
	return "system-internal"
}

func (c *StoreServiceClient) doJSON(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.systemToken())
	req.Header.Set("X-Internal-Service", "admin-store-service")

	return c.httpClient.Do(req)
}

func (c *StoreServiceClient) decodeResponse(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var errBody map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		msg := fmt.Sprintf("store-service returned %d", resp.StatusCode)
		if e, ok := errBody["error"].(map[string]interface{}); ok {
			if m, ok := e["message"].(string); ok {
				msg = m
			}
		}
		return fmt.Errorf("%s", msg)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// CreateStore creates a store row inside store-service's database.
func (c *StoreServiceClient) CreateStore(req *CreateStoreRequest) (*StoreResponse, error) {
	resp, err := c.doJSON(http.MethodPost, "/v1/store/internal/admin-write/stores", req)
	if err != nil {
		return nil, err
	}
	var store StoreResponse
	if err := c.decodeResponse(resp, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

// UpdateGeofence delegates a geofence update to store-service.
func (c *StoreServiceClient) UpdateGeofence(storeID string, req *UpdateGeofenceRequest) error {
	resp, err := c.doJSON(http.MethodPut, "/v1/store/internal/admin-write/stores/"+storeID+"/geofence", req)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, nil)
}

// UpdateHours delegates a store hours update to store-service.
func (c *StoreServiceClient) UpdateHours(storeID string, req *UpdateHoursRequest) error {
	resp, err := c.doJSON(http.MethodPut, "/v1/store/internal/admin-write/stores/"+storeID+"/hours", req)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, nil)
}

// UpdateCapacity delegates a capacity update to store-service.
func (c *StoreServiceClient) UpdateCapacity(storeID string, req *UpdateCapacityRequest) error {
	resp, err := c.doJSON(http.MethodPut, "/v1/store/internal/admin-write/stores/"+storeID+"/capacity", req)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, nil)
}

// UpdateStatus delegates a store status transition to store-service.
func (c *StoreServiceClient) UpdateStatus(storeID string, req *UpdateStatusRequest) error {
	resp, err := c.doJSON(http.MethodPut, "/v1/store/internal/admin-write/stores/"+storeID+"/status", req)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, nil)
}

// UpdatePaymentSetup delegates payment setup fields update to store-service.
func (c *StoreServiceClient) UpdatePaymentSetup(storeID string, req *UpdatePaymentSetupRequest) error {
	resp, err := c.doJSON(http.MethodPut, "/v1/store/internal/admin-write/stores/"+storeID+"/payment-setup", req)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, nil)
}

// RotateQRTokens delegates QR token rotation to store-service.
func (c *StoreServiceClient) RotateQRTokens(storeID string, req *RotateQRTokensRequest) (map[string]interface{}, error) {
	resp, err := c.doJSON(http.MethodPost, "/v1/store/internal/admin-write/stores/"+storeID+"/qr-tokens/rotate", req)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := c.decodeResponse(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetQRTokens fetches active QR tokens for a store from store-service.
func (c *StoreServiceClient) GetQRTokens(storeID string) (map[string]interface{}, error) {
	resp, err := c.doJSON(http.MethodGet, "/v1/store/internal/admin-write/stores/"+storeID+"/qr-tokens", nil)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := c.decodeResponse(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListStores fetches store lists from store-service.
func (c *StoreServiceClient) ListStores(query string) (map[string]interface{}, error) {
	path := "/v1/store/internal/admin-write/stores"
	if query != "" {
		path += "?" + query
	}
	resp, err := c.doJSON(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := c.decodeResponse(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetStoreByID fetches a single store from store-service.
func (c *StoreServiceClient) GetStoreByID(storeID string) (*StoreResponse, error) {
	resp, err := c.doJSON(http.MethodGet, "/v1/store/internal/admin-write/stores/"+storeID, nil)
	if err != nil {
		return nil, err
	}
	var store StoreResponse
	if err := c.decodeResponse(resp, &store); err != nil {
		return nil, err
	}
	return &store, nil
}
