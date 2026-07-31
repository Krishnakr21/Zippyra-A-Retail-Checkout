# Integration Service (`integration-service`)

`integration-service` manages ERP integrations with **SAP** (cloud-reachable direct integration) and **Tally / Busy** (on-premise agent-polled integration).

---

## Deployment Models & System Boundary

### 1. SAP Integration (`DIRECT` Mode)
- **Deployment**: SAP Cloud Connector / OData endpoints are cloud-reachable via public URL or direct VPC link.
- **Outbound**: `integration-service` makes direct HTTP POST requests with 15-second timeout. Failed attempts are retried with exponential backoff up to a 10-attempt ceiling by background worker.
- **Inbound**: SAP pushes webhooks directly to `/v1/integration/connections/{id}/webhook`.

### 2. Tally & Busy Integrations (`AGENT_POLLED` Mode) — **CONNECTOR AGENT BOUNDARY**
- **Architecture**: In real Indian retail deployments, Tally and Busy run as Windows desktop applications inside the store's local network (LAN) with **no public IP or inbound routing from cloud**.
- **System Boundary**:
  - **`integration-service` (Cloud)**:
    - Generates and stores SHA-256 hashes of agent API keys (`agent_api_key`).
    - Enqueues outbound ERP sync jobs (`erp_sync_jobs`) into a pull queue.
    - Exposes `/v1/integration/connections/{id}/pull-queue` and `/v1/integration/connections/{id}/pull-queue/ack` for unattended agents.
    - Receives inbound local edits pushed from agents to `/v1/integration/connections/{id}/webhook`.
  - **On-Premise Connector Agent (Local Binary - Separate Deliverable)**:
    - Separately distributed lightweight binary installed on the store machine running Tally/Busy.
    - Authenticates to `integration-service` using `Authorization: Bearer <agent_api_key>`.
    - Periodically polls `/pull-queue` (e.g. every 60s), pulls pending jobs, and applies them locally via Tally's XML-over-HTTP endpoint (`http://localhost:9000`) or Busy's API.
    - Sends local changes (e.g. manual price edits in Tally) back to `integration-service`'s webhook endpoint.

---

## Webhook Security & Idempotency
- **HMAC Signature Verification**: All `/webhook` calls must present `X-Signature` computed using HMAC-SHA256 of the raw body and the connection's secret. Invalid signatures return `400 INVALID_SIGNATURE`.
- **Idempotency**: Requests specify `X-Event-Id` (or fall back to `sha256(raw_body)`). `erp_webhook_events` enforces a `UNIQUE (connection_id, event_id)` constraint. Duplicate events return `200 DUPLICATE_IGNORED` immediately.
