package busy_adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/logging"
)

type Adapter struct {
	endpoint   string
	httpClient *http.Client
	logger     *logging.Logger
}

func NewAdapter(endpoint string, logger *logging.Logger) *Adapter {
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080/api"
	}
	return &Adapter{
		endpoint: strings.TrimRight(endpoint, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (a *Adapter) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", a.endpoint+"/ping", nil)
	if err != nil {
		return err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Busy health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Busy health check returned status %d", resp.StatusCode)
	}

	return nil
}

func (a *Adapter) ApplyPriceUpdate(ctx context.Context, barcode string, pricePaise int64) error {
	payload := map[string]interface{}{
		"item_code":   barcode,
		"price_paise": pricePaise,
		"sales_price": float64(pricePaise) / 100.0,
	}

	if err := a.postJSON(ctx, "/item/update-price", payload); err != nil {
		return fmt.Errorf("failed to apply Busy price update: %w", err)
	}

	a.logger.Info("[BusyAdapter] Applied price update barcode=%s pricePaise=%d", barcode, pricePaise)
	return nil
}

func (a *Adapter) ApplyStockAdjustment(ctx context.Context, barcode string, qtyDelta int64, reason string) error {
	payload := map[string]interface{}{
		"item_code": barcode,
		"qty_delta": qtyDelta,
		"reason":    reason,
	}

	if err := a.postJSON(ctx, "/stock/adjust", payload); err != nil {
		return fmt.Errorf("failed to apply Busy stock adjustment: %w", err)
	}

	a.logger.Info("[BusyAdapter] Applied stock adjustment barcode=%s qtyDelta=%d", barcode, qtyDelta)
	return nil
}

func (a *Adapter) ApplyGrn(ctx context.Context, items []erp_adapter.GrnItem) error {
	payload := map[string]interface{}{
		"items": items,
	}

	if err := a.postJSON(ctx, "/grn/create", payload); err != nil {
		return fmt.Errorf("failed to apply Busy GRN: %w", err)
	}

	a.logger.Info("[BusyAdapter] Applied GRN with %d items", len(items))
	return nil
}

func (a *Adapter) PollLocalChanges(ctx context.Context, since time.Time) ([]erp_adapter.LocalChange, error) {
	reqURL := fmt.Sprintf("%s/items/modified?since=%s", a.endpoint, url.QueryEscape(since.Format(time.RFC3339)))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Busy poll changes failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Busy returned status %d: %s", resp.StatusCode, string(body))
	}

	var busyChanges []struct {
		ItemCode  string                 `json:"item_code"`
		EventType string                 `json:"event_type"`
		Payload   map[string]interface{} `json:"payload"`
		UpdatedAt string                 `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&busyChanges); err != nil {
		return []erp_adapter.LocalChange{}, nil
	}

	var result []erp_adapter.LocalChange
	for _, c := range busyChanges {
		t, err := time.Parse(time.RFC3339, c.UpdatedAt)
		if err != nil {
			t = time.Now()
		}
		evt := c.EventType
		if evt == "" {
			evt = "CATALOG_PRICE_CHANGED"
		}
		result = append(result, erp_adapter.LocalChange{
			EventType: evt,
			Barcode:   c.ItemCode,
			Payload:   c.Payload,
			Timestamp: t,
		})
	}

	return result, nil
}

func (a *Adapter) postJSON(ctx context.Context, path string, data interface{}) error {
	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.endpoint+path, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
