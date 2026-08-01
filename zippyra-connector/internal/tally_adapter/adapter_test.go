package tally_adapter

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/logging"
)

func TestTallyAdapter_Operations(t *testing.T) {
	logger, _ := logging.NewLogger("")

	var lastXML string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		lastXML = string(body)

		if strings.Contains(lastXML, "List of Companies") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<ENVELOPE><RESPONSE>Tally Running</RESPONSE></ENVELOPE>`))
			return
		}

		if strings.Contains(lastXML, "List of Stock Items") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<ENVELOPE><BODY><DATA><TALLYMESSAGE><STOCKITEM NAME="8901001"><RATE>50.00/PCS</RATE></STOCKITEM></TALLYMESSAGE></DATA></BODY></ENVELOPE>`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<ENVELOPE><RESPONSE>CREATED</RESPONSE></ENVELOPE>`))
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL, logger)
	ctx := context.Background()

	// 1. HealthCheck
	if err := adapter.HealthCheck(ctx); err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	// 2. ApplyPriceUpdate
	if err := adapter.ApplyPriceUpdate(ctx, "8901001", 5000); err != nil {
		t.Fatalf("ApplyPriceUpdate failed: %v", err)
	}
	if !strings.Contains(lastXML, `STOCKITEM NAME="8901001"`) || !strings.Contains(lastXML, `50.00/PCS`) {
		t.Errorf("ApplyPriceUpdate XML incorrect: %s", lastXML)
	}

	// 3. ApplyStockAdjustment
	if err := adapter.ApplyStockAdjustment(ctx, "8901001", 10, "Inventory Count Audit"); err != nil {
		t.Fatalf("ApplyStockAdjustment failed: %v", err)
	}
	if !strings.Contains(lastXML, `VOUCHER VCHTYPE="Stock Journal"`) || !strings.Contains(lastXML, `10 PCS`) {
		t.Errorf("ApplyStockAdjustment XML incorrect: %s", lastXML)
	}

	// 4. ApplyGrn
	items := []erp_adapter.GrnItem{
		{Barcode: "8901001", Qty: 20, CostPaise: 4000},
	}
	if err := adapter.ApplyGrn(ctx, items); err != nil {
		t.Fatalf("ApplyGrn failed: %v", err)
	}
	if !strings.Contains(lastXML, `VOUCHER VCHTYPE="Purchase"`) || !strings.Contains(lastXML, `20 PCS`) {
		t.Errorf("ApplyGrn XML incorrect: %s", lastXML)
	}

	// 5. PollLocalChanges
	changes, err := adapter.PollLocalChanges(ctx, time.Now())
	if err != nil {
		t.Fatalf("PollLocalChanges failed: %v", err)
	}
	if len(changes) != 1 || changes[0].Barcode != "8901001" {
		t.Errorf("PollLocalChanges incorrect result: %v", changes)
	}
}
