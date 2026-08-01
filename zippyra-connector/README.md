# Zippyra Standalone ERP Connector (`zippyra-connector`)

`zippyra-connector` is a lightweight, zero-dependency Go binary designed for store IT on-premise installation. It acts as an integration bridge between cloud-hosted `integration-service` and local ERP software (**Tally ERP** or **Busy ERP**).

---

## Architecture Overview

```
 ┌───────────────────────────┐           ┌───────────────────────────────┐
 │ Zippyra Cloud             │           │ Local Store Windows PC        │
 │ (integration-service)     │           │ (No inbound cloud ports)      │
 │                           │           │                               │
 │ GET  /pull-queue  ◄───────┼───Poll────┼─── zippyra-connector          │
 │ POST /pull-queue/ack ─────┼───Ack─────┼───   ├─ tally_adapter         │
 │ POST /webhook ◄───────────┼───Push────┼───   ├─ busy_adapter         │
 └───────────────────────────┘           │   └─ local_status_server      │
                                         │       (127.0.0.1:8085/status) │
                                         └───────────────────────────────┘
```

---

## The `ErpAdapter` Interface Contract

All ERP integrations implement the unified `ErpAdapter` interface defined in `internal/erp_adapter/adapter.go`:

```go
type ErpAdapter interface {
    ApplyPriceUpdate(ctx context.Context, barcode string, pricePaise int64) error
    ApplyStockAdjustment(ctx context.Context, barcode string, qtyDelta int64, reason string) error
    ApplyGrn(ctx context.Context, items []GrnItem) error
    PollLocalChanges(ctx context.Context, since time.Time) ([]LocalChange, error)
    HealthCheck(ctx context.Context) error
}
```

Adding support for a new ERP target (e.g. SAP Business One or Marg ERP) requires implementing this single interface contract without modifying core sync loop scheduler logic.

---

## Resilient Execution & Failure Posture

1. **Panic Recovery**: The main loop wraps each tick cycle in a panic recovery block (`recover()`). The process never crashes on malformed data, network drops, or ERP timeouts.
2. **Batched Acknowledgments**: Successful sync jobs are batched into a single `AckQueue` API call per tick cycle.
3. **Failed Job Retention**: Jobs that fail local ERP application are omitted from the acknowledgment call, remaining in `DELIVERED` status on Chain HQ dashboards for human intervention.
4. **HMAC Webhook Signatures**: Local ERP edits polled by the connector are signed using HMAC-SHA256 (`X-Signature` header) before posting to `integration-service`.
5. **Local Status Server**: Exposes diagnostic metrics on `http://127.0.0.1:8085/status`.

---

## Building & Testing

```bash
# Run unit test suite
make test

# Cross-compile for Windows, Linux, and macOS
make all
```
