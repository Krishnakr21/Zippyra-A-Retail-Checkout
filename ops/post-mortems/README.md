# Post-Mortem Records

This directory contains completed post-mortem documents for resolved SEV1 and SEV2 incidents on the Zippyra platform.

---

## Workflow

### When to write a post-mortem

A post-mortem is **mandatory** for every **SEV1** and **SEV2** incident, and **recommended** for any SEV3 incident where the root cause revealed a systemic gap worth documenting.

### Timeline

| Step | Deadline | Owner |
|---|---|---|
| **Incident resolved** | T+0 | On-call engineer / Incident commander |
| **Post-mortem draft completed** | T+48 hours | Incident commander + primary responder |
| **Review meeting** | T+5 business days | Full engineering team |
| **Action items tracked in issue tracker** | Same day as review meeting | Incident commander |
| **Post-mortem status updated to "Action Items Tracked"** | Same day as review meeting | Incident commander |

### How to create a post-mortem

1. **Copy the template**:
   ```bash
   cp ops/post-mortem-template.md ops/post-mortems/YYYY-MM-DD-short-title.md
   ```
   Example: `ops/post-mortems/2026-08-15-razorpay-upi-timeout-spike.md`

2. **Fill out every section**. Use the platform's built observability stack to reconstruct the timeline accurately:
   - **Prometheus / Grafana**: Pull alert firing times, SLO recording rule values (`job:payment_success_rate:ratio_rate5m`, `job:cart_scan_latency_p99:ms`, etc.), and metric graphs.
   - **OpenTelemetry traces**: Use correlation IDs from `infra/observability/otel-collector.yaml` to trace request flows across services.
   - **PagerDuty**: Pull acknowledgment and escalation timestamps.
   - **Kafka / DLQ metrics**: Check `kafka_consumergroup_lag` and `dlq.*` topic offsets from `alert_rules.yml` alert definitions.
   - **Statuspage**: Reference the public incident timeline from `status.zippyra.com` if one was posted.

3. **Share with the full engineering team** for review. The review meeting should be:
   - Blameless (see the language guide at the bottom of the template).
   - Focused on identifying systemic improvements, not assigning fault.
   - Time-boxed to 30-60 minutes.

4. **Track action items in the issue tracker** (Jira, Linear, or whatever the team uses). The post-mortem document should **link to** issue tracker items, not duplicate them. Each action item row in the post-mortem table should have a direct link to its tracker issue.

5. **Update the post-mortem status** to "Action Items Tracked" once all items are filed.

### Naming Convention

```
ops/post-mortems/YYYY-MM-DD-short-descriptive-title.md
```

Examples:
- `2026-07-22-payment-gateway-timeout-cascade.md`
- `2026-08-03-exit-gate-rfid-reader-firmware-crash.md`
- `2026-08-10-kafka-consumer-lag-analytics-pipeline.md`

Use lowercase, hyphens for spaces, and keep the title to 3-6 words that identify the incident at a glance.

### Retention

Post-mortems are retained indefinitely in this repository as institutional knowledge. They serve as:
- **Onboarding material** for new engineers understanding past failure modes.
- **Evidence of continuous improvement** for compliance and audit purposes (referenced in `ops/COMPLIANCE_AUDIT.md`).
- **Pattern detection** — recurring themes across post-mortems indicate systemic investment areas.

---

## Index

<!-- Add completed post-mortems here as they are created. -->

| Date | Severity | Title | Duration | Primary Impact |
|---|---|---|---|---|
| *No incidents yet* | — | — | — | — |
