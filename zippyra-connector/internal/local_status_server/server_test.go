package local_status_server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/logging"
)

type FakeErpAdapter struct {
	healthErr error
}

func (f *FakeErpAdapter) ApplyPriceUpdate(ctx context.Context, barcode string, pricePaise int64) error {
	return nil
}
func (f *FakeErpAdapter) ApplyStockAdjustment(ctx context.Context, barcode string, qtyDelta int64, reason string) error {
	return nil
}
func (f *FakeErpAdapter) ApplyGrn(ctx context.Context, items []erp_adapter.GrnItem) error {
	return nil
}
func (f *FakeErpAdapter) PollLocalChanges(ctx context.Context, since time.Time) ([]erp_adapter.LocalChange, error) {
	return nil, nil
}
func (f *FakeErpAdapter) HealthCheck(ctx context.Context) error {
	return f.healthErr
}

func TestLocalStatusServer(t *testing.T) {
	logger, _ := logging.NewLogger("")
	fakeAdapter := &FakeErpAdapter{}

	server, metrics := NewServer(8999, "TALLY", fakeAdapter, logger)
	if err := server.Start(); err != nil {
		t.Fatalf("Start server failed: %v", err)
	}
	defer func() { _ = server.Stop(context.Background()) }()

	metrics.UpdatePoll(3, errors.New("ERP connection timeout"))

	resp, err := http.Get("http://127.0.0.1:8999/status")
	if err != nil {
		t.Fatalf("GET /status failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("Decode JSON failed: %v", err)
	}

	if data["status"] != "DEGRADED" {
		t.Errorf("Expected status DEGRADED, got %v", data["status"])
	}
	if data["last_error"] != "ERP connection timeout" {
		t.Errorf("Expected last_error 'ERP connection timeout', got %v", data["last_error"])
	}
	if data["pending_jobs_count"] != float64(3) {
		t.Errorf("Expected pending_jobs_count 3, got %v", data["pending_jobs_count"])
	}
}
