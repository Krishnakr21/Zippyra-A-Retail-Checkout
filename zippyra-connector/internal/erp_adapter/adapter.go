package erp_adapter

import (
	"context"
	"time"
)

type GrnItem struct {
	Barcode   string `json:"barcode"`
	Qty       int64  `json:"qty"`
	CostPaise int64  `json:"cost_paise"`
}

type LocalChange struct {
	EventType string                 `json:"event_type"` // e.g. CATALOG_PRICE_CHANGED, STOCK_ADJUSTED
	Barcode   string                 `json:"barcode"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

type ErpAdapter interface {
	ApplyPriceUpdate(ctx context.Context, barcode string, pricePaise int64) error
	ApplyStockAdjustment(ctx context.Context, barcode string, qtyDelta int64, reason string) error
	ApplyGrn(ctx context.Context, items []GrnItem) error
	PollLocalChanges(ctx context.Context, since time.Time) ([]LocalChange, error)
	HealthCheck(ctx context.Context) error // used by local_status_server
}
