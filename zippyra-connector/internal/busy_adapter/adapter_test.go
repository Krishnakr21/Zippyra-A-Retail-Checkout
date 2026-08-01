package busy_adapter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/logging"
)

func TestBusyAdapter_Operations(t *testing.T) {
	logger, _ := logging.NewLogger("")

	var lastBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" && r.URL.Path == "/items/modified" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{
					"item_code": "8902001",
					"event_type": "CATALOG_PRICE_CHANGED",
					"payload": {"price_paise": 6000},
					"updated_at": "2026-08-01T12:00:00Z"
				}
			]`))
			return
		}

		bodyBytes, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(bodyBytes, &lastBody)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL, logger)
	ctx := context.Background()

	// 1. HealthCheck
	if err := adapter.HealthCheck(ctx); err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	// 2. ApplyPriceUpdate
	if err := adapter.ApplyPriceUpdate(ctx, "8902001", 6000); err != nil {
		t.Fatalf("ApplyPriceUpdate failed: %v", err)
	}
	if lastBody["item_code"] != "8902001" || lastBody["price_paise"] != float64(6000) {
		t.Errorf("ApplyPriceUpdate body incorrect: %v", lastBody)
	}

	// 3. ApplyStockAdjustment
	if err := adapter.ApplyStockAdjustment(ctx, "8902001", -5, "Damage Spoilage"); err != nil {
		t.Fatalf("ApplyStockAdjustment failed: %v", err)
	}
	if lastBody["item_code"] != "8902001" || lastBody["qty_delta"] != float64(-5) {
		t.Errorf("ApplyStockAdjustment body incorrect: %v", lastBody)
	}

	// 4. ApplyGrn
	items := []erp_adapter.GrnItem{
		{Barcode: "8902001", Qty: 50, CostPaise: 4500},
	}
	if err := adapter.ApplyGrn(ctx, items); err != nil {
		t.Fatalf("ApplyGrn failed: %v", err)
	}

	// 5. PollLocalChanges
	changes, err := adapter.PollLocalChanges(ctx, time.Now())
	if err != nil {
		t.Fatalf("PollLocalChanges failed: %v", err)
	}
	if len(changes) != 1 || changes[0].Barcode != "8902001" {
		t.Errorf("PollLocalChanges incorrect result: %v", changes)
	}
}
