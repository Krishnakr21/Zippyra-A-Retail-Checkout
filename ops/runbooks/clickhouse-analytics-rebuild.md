# ClickHouse Analytics Data Loss & Kafka Rebuild Runbook

## Overview & Trigger Conditions
This runbook applies when `analytics-service` experiences data loss, table corruption, or instance failure on ClickHouse Cloud. Because ClickHouse acts as a pure event-sourced read model, historical analytics data is repopulated by replaying Kafka message history.

---

## Recovery Window & Limitations
- **Recovery Window**: Bounded by MSK per-topic retention override (**30 Days** / `2,592,000,000 ms`).
- **Target Topics**: `order.completed`, `order.returned`, `payment.confirmed`, `inventory.stock_updated`, `analytics.event`.
- **Limitation**: Any data loss occurring before 30 days cannot be recovered via Kafka offset reset.

---

## Step-by-Step Recovery Procedure

### Step 1: Provision / Verify ClickHouse Instance
Ensure ClickHouse Cloud instance is active and accessible via HTTP endpoint (`CLICKHOUSE_URL`).

### Step 2: Execute Rebuild Script
Run the automated rebuild script from the repo root:
```bash
KAFKA_BROKERS="b-1.msk.ap-south-1.amazonaws.com:9092,b-2.msk.ap-south-1.amazonaws.com:9092" \
CLICKHOUSE_URL="https://clickhouse.zippyra.com:8443" \
ANALYTICS_CONSUMER_GROUP="analytics-service-group" \
./scripts/rebuild-analytics-from-kafka.sh
```

### Step 3: Monitor Ingestion Progress
Monitor ClickHouse table row counts as consumers process historical messages:
```sql
SELECT 'sales_events', count() FROM sales_events
UNION ALL
SELECT 'order_items_events', count() FROM order_items_events
UNION ALL
SELECT 'funnel_events', count() FROM funnel_events;
```

---

## Cost & Infrastructure Notes
- **ClickHouse Cloud Dev Tier**: No automated PITR backups.
- **Production Upgrade**: Minimum **$260/month** for automated 24-hour ClickHouse Cloud snapshots.
