# Zippyra Platform Operations & Incident Runbook

## Overview

This runbook outlines operational procedures, incident response severity levels, on-call escalation policies, and disaster recovery procedures for the Zippyra platform deployed on AWS `ap-south-1` (Mumbai).

---

## Severity Matrix & Escalation Policy

| Severity | Target Response | Definition & Impact | PagerDuty Notification |
|----------|-----------------|---------------------|------------------------|
| **SEV1** | < 15 mins | Core checkout, payment, or exit gate completely broken | P1 Immediate Page (On-call + Primary Lead) |
| **SEV2** | < 30 mins | High consumer lag, degraded latency, non-critical service down | P2 Page (On-call Engineer) |
| **SEV3** | < 2 hours | Minor feature degradation, single store issue | Slack #ops-alerts |
| **SEV4** | < 24 hours | Non-impacting bug or minor alert threshold breach | Jira Ticket |

**Post-Mortem Requirement**: SEV1 and SEV2 incidents require a blameless post-mortem within 48 hours of resolution. Copy `ops/post-mortem-template.md` → `ops/post-mortems/{date}-{short-title}.md`. See [Post-Mortem Process](post-mortems/README.md) for full workflow.

---

## Disaster Recovery (RTO / RPO & Backup Strategy)

### 1. Core Transactional Databases (RDS PostgreSQL)
- **Target RTO (Recovery Time Objective)**: 2 Hours
- **Target RPO (Recovery Point Objective)**: 1 Hour
- **Backup Window**: Daily 01:00-02:00 IST (19:30-20:30 UTC)
- **Retention**: 7 Days Automated Backups + Cross-Region Replication to `ap-south-2` (Hyderabad)
- **Monthly Restoration Drill**: Executed via `./scripts/restore-drill.sh`

### 2. Analytical Read-Model Store (ClickHouse - `analytics-service`)
- **Primary DR Mechanism**: **Kafka Event-Source Rebuild** (`scripts/rebuild-analytics-from-kafka.sh`).
  - Analytics data (`sales_events`, `order_items_events`, `funnel_events`, `transaction_hourly`) is a pure event-sourced read model.
  - On ClickHouse instance failure or data loss, consumer group offsets are reset to earliest available point in Kafka retention, and data is repopulated.
- **Kafka Topic Retention Override**:
  - MSK default topic retention is 7 days (168 hours).
  - **Per-Topic Override**: Retention extended to **30 days** (`retention.ms=2592000000`) in `infra/modules/msk/main.tf` for analytics source topics:
    - `order.completed`, `order.returned`, `payment.confirmed`, `inventory.stock_updated`, `analytics.event`.
- **Known Bounded Recovery Limitation**:
  - Rebuild capability is strictly bounded by the **30-day Kafka retention window**. Any data loss discovered >30 days post-incident cannot be recovered via Kafka replay.
- **ClickHouse Cloud Backup Tier & Cost**:
  - Dev/Pilot Tier: Manual ClickHouse dump/restore.
  - Production Tier Upgrade: Requires ClickHouse Cloud Production Instance ($0.36/GB/mo storage + compute, approx **+$260/month**) to enable automated 24-hour snapshot backups and point-in-time recovery.
- **Quarterly Drill**: Integrated into `./scripts/restore-drill.sh` (executed every 3rd month).

### 3. Device Telemetry Store (TimescaleDB - `device-mgmt-service`)
- **Data Criticality**: **Lower Priority**. Raw device heartbeats (`device_heartbeats`) have a 30-day hypertable retention policy.
- **Backup Strategy**:
  - Timescale Cloud Basic Tier daily automated snapshot.
  - No PITR or cross-region replication needed due to low financial/compliance risk.
- **Recovery Procedure**: Restore latest daily snapshot in Timescale Cloud console.

---

## Public Status Page (status.zippyra.com) Incident Response & Communication

Zippyra maintains a managed public status page at [status.zippyra.com](https://status.zippyra.com) to provide real-time self-service operational visibility to retail partners and prevent support call spikes during incidents.

### User-Facing Components Monitored

1. **Payment Processing**: Payment checkout authorization, gateway routing, and refund dispatching.
2. **Mobile App API**: Customer App API surface, catalog search, cart operations, and authentication.
3. **Retailer Dashboard**: Web dashboard for store managers, inventory controls, and shift management.
4. **Exit Gate / Anti-Theft System**: Physical store exit pass generation, gate barcode/RFID verification.
5. **Customer Support**: Ticket management and in-app grievance support tools.

### Automated vs Manual Incident Management

- **Automated Webhooks**: Prometheus Alertmanager automatically posts component status updates to `status.zippyra.com` via incoming webhooks when SEV1/SEV2 alerts fire (`PaymentSuccessRateSLOBreach`, `ExitGateAlarmSpike`, etc.).
- **Manual Incident Posting Protocol (On-Call Engineer SLA < 15m)**:
  1. Login to Statuspage Admin at `admin.statuspage.io` using Zippyra SSO credentials.
  2. Click **Create Incident**.
  3. **Title**: Concise description (e.g., *"Degraded Payment Authorization via Razorpay UPI"*).
  4. **Incident Status**: Select `Investigating` | `Identified` | `Monitoring` | `Resolved`.
  5. **Affected Components**: Select component and update status (`Degraded Performance` | `Partial Outage` | `Major Outage`).
  6. **Customer Impact Message**:
     > *"We are investigating reports of payment authorization latency affecting UPI checkouts. Scan-and-pay transactions may experience delays. Our engineering team is actively working with our gateway partner to resolve the issue. Next update in 20 minutes."*
  7. **Resolution**: Update incident to `Resolved` and restore component status to `Operational` once verified. Maintain 90-day incident audit history.

---

## Scenario Runbooks

- [Payment Gateway Down Runbook](runbooks/payment-gateway-down.md)
- [Exit Gate Alarm Spike Runbook](runbooks/exit-gate-alarm-spike.md)
- [Kafka Consumer Lag Spike Runbook](runbooks/kafka-consumer-lag.md)
- [ClickHouse Analytics Rebuild Runbook](runbooks/clickhouse-analytics-rebuild.md)


