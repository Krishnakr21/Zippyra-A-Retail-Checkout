package sync_loop

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/local_status_server"
	"github.com/zippyra/zippyra-connector/internal/logging"
	"github.com/zippyra/zippyra-connector/internal/zippyra_client"
)

type MockErpAdapter struct {
	failBarcode string
	panicOnCall bool
}

func (m *MockErpAdapter) ApplyPriceUpdate(ctx context.Context, barcode string, pricePaise int64) error {
	if m.panicOnCall {
		panic("simulated adapter panic!")
	}
	if barcode == m.failBarcode {
		return errors.New("ERP failed to update price")
	}
	return nil
}

func (m *MockErpAdapter) ApplyStockAdjustment(ctx context.Context, barcode string, qtyDelta int64, reason string) error {
	if m.panicOnCall {
		panic("simulated adapter panic!")
	}
	if barcode == m.failBarcode {
		return errors.New("ERP failed stock adjustment")
	}
	return nil
}

func (m *MockErpAdapter) ApplyGrn(ctx context.Context, items []erp_adapter.GrnItem) error {
	return nil
}

func (m *MockErpAdapter) PollLocalChanges(ctx context.Context, since time.Time) ([]erp_adapter.LocalChange, error) {
	return []erp_adapter.LocalChange{
		{
			EventType: "CATALOG_PRICE_CHANGED",
			Barcode:   "8909999",
			Payload:   map[string]interface{}{"price_paise": 7500},
			Timestamp: time.Now(),
		},
	}, nil
}

func (m *MockErpAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

func TestSyncLoop_FailedJobsOmittedFromAckBatch(t *testing.T) {
	logger, _ := logging.NewLogger("")

	var ackBody string
	var webhookHMAC string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/integration/connections/conn-1/pull-queue":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"job_id": "job-success-1", "connection_id": "conn-1", "source_event_type": "CATALOG_PRICE_CHANGED", "payload": {"barcode": "8900001", "price_paise": 4000}},
				{"job_id": "job-fail-2", "connection_id": "conn-1", "source_event_type": "STOCK_ADJUSTED", "payload": {"barcode": "FAIL_BARCODE", "qty_delta": -5}},
				{"job_id": "job-success-3", "connection_id": "conn-1", "source_event_type": "CATALOG_PRICE_CHANGED", "payload": {"barcode": "8900003", "price_paise": 4500}}
			]`))

		case "/v1/integration/connections/conn-1/pull-queue/ack":
			bodyBytes, _ := io.ReadAll(r.Body)
			ackBody = string(bodyBytes)
			w.WriteHeader(http.StatusOK)

		case "/v1/integration/connections/conn-1/webhook":
			webhookHMAC = r.Header.Get("X-Signature")
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	client := zippyra_client.NewClient(ts.URL, "conn-1", "api-key-1", "webhook-sec-1", logger)
	mockAdapter := &MockErpAdapter{failBarcode: "FAIL_BARCODE"}
	server, metrics := local_status_server.NewServer(8998, "TALLY", mockAdapter, logger)
	_ = server

	loop := NewSyncLoop(client, mockAdapter, metrics, 60, logger)

	ctx := context.Background()
	loop.RunTick(ctx)

	// Verify failed job job-fail-2 was OMITTED from batched ack
	if strings.Contains(ackBody, "job-fail-2") {
		t.Errorf("Ack body contained failed job-fail-2: %s", ackBody)
	}

	// Verify successful jobs ARE in the batched ack call
	if !strings.Contains(ackBody, "job-success-1") || !strings.Contains(ackBody, "job-success-3") {
		t.Errorf("Ack body missing successful jobs: %s", ackBody)
	}

	// Verify outbound webhook HMAC was attached
	if webhookHMAC == "" {
		t.Errorf("Expected HMAC signature on outbound webhook push")
	}

	if metrics.PendingJobsCount != 1 {
		t.Errorf("Expected 1 pending un-acked failed job in metrics, got %d", metrics.PendingJobsCount)
	}
}

func TestSyncLoop_PanicRecovery(t *testing.T) {
	logger, _ := logging.NewLogger("")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"job_id": "job-1", "source_event_type": "CATALOG_PRICE_CHANGED", "payload": {"barcode": "8900001", "price_paise": 4000}}]`))
	}))
	defer ts.Close()

	client := zippyra_client.NewClient(ts.URL, "conn-1", "api-key-1", "webhook-sec-1", logger)
	panickingAdapter := &MockErpAdapter{panicOnCall: true}
	server, metrics := local_status_server.NewServer(8997, "TALLY", panickingAdapter, logger)
	_ = server

	loop := NewSyncLoop(client, panickingAdapter, metrics, 60, logger)

	ctx := context.Background()

	// Should recover from panic without crashing the process
	loop.RunTick(ctx)

	if metrics.LastError == "" || !strings.Contains(metrics.LastError, "PANIC RECOVERED") {
		t.Errorf("Expected PANIC RECOVERED error recorded in metrics, got: %s", metrics.LastError)
	}
}
