# Post-Mortem: {Incident Title}

| Field | Value |
|---|---|
| **Date of Incident** | YYYY-MM-DD |
| **Severity** | SEV{1-4} |
| **Duration** | HH:MM (from first detection to full recovery confirmation) |
| **Author(s)** | {Names of post-mortem authors — typically the incident commander + primary responder} |
| **Status** | Draft / Reviewed / Action Items Tracked |
| **Incident ID** | {PagerDuty incident ID or internal tracking ID} |
| **Statuspage Incident** | {Link to status.zippyra.com incident if one was posted} |

---

## Summary

<!-- 2-3 sentences, plain language, suitable for a non-engineer to understand
     what happened and its customer impact. This section is shared directly
     with retail partners and leadership — write accordingly. -->

{What happened, which user-facing capability was affected, and for how long.}

---

## Timeline

<!-- Reconstruct this from the BUILT observability stack — use OpenTelemetry
     correlation IDs, Grafana dashboard timestamps, Prometheus alert firing
     times, and PagerDuty incident logs rather than relying on memory.

     All times in IST (UTC+05:30) unless otherwise noted. -->

| Time (IST) | Event |
|---|---|
| HH:MM | **DETECTION**: {How the incident was first detected — alert name from `infra/observability/alert_rules.yml`, customer report, manual observation, etc.} |
| HH:MM | **ACKNOWLEDGED**: {On-call engineer acknowledged PagerDuty page} |
| HH:MM | **INVESTIGATION**: {Key diagnostic steps — which Grafana dashboards were checked, which logs were queried, what hypotheses were tested and ruled out} |
| HH:MM | **ROOT CAUSE IDENTIFIED**: {When the actual cause was pinpointed} |
| HH:MM | **FIX DEPLOYED**: {What was deployed — commit SHA, config change, infrastructure action} |
| HH:MM | **RECOVERY CONFIRMED**: {When metrics returned to normal — reference specific SLO recording rules from `alert_rules.yml` (e.g., `job:payment_success_rate:ratio_rate5m` back above threshold)} |
| HH:MM | **STATUSPAGE RESOLVED**: {When status.zippyra.com incident was marked Resolved} |

---

## Impact

<!-- Quantify using data from the built observability and analytics stacks.
     Pull SLO/error-budget consumption from Prometheus recording rules. -->

- **Customers affected**: {Count or estimate}
- **Orders affected**: {Count — query from order-service or analytics-service}
- **Stores affected**: {Count and store IDs if localized}
- **Revenue impact**: {₹ amount if payment-related, or "No direct revenue impact"}
- **SLO budget consumed**: {e.g., "Payment success rate SLO dropped to 98.2% for 12 minutes, consuming approximately 35% of monthly error budget" — pull from `job:payment_success_rate:ratio_rate5m`}
- **Failed exit validations**: {Count if exit-gate related — pull from `exit_validation_alarms_total`}

---

## Root Cause

<!-- Identify the SYSTEMIC cause, not "engineer X made a mistake."

     Every service in this platform has extensive idempotency guards, timeout
     configurations, circuit-breaker protections, dead-letter queues, and
     retry policies built in specifically because we design for failure.
     A post-mortem should ask "why didn't our existing safeguards catch this"
     rather than "who broke it."

     Be specific and technical — name the exact component, configuration,
     or interaction that failed. -->

{Description of the systemic root cause.}

---

## Contributing Factors

<!-- Secondary conditions that made the incident worse, longer, or harder to
     detect. These are often the most actionable findings. -->

- {e.g., "The relevant Grafana dashboard did not have a panel for this specific metric, delaying diagnosis by ~8 minutes."}
- {e.g., "The alert threshold for `KafkaConsumerLagHighP1` was set at 1000, but the degradation manifested at lag ~600 which did not trigger the alert."}
- {e.g., "The runbook for this scenario (`ops/runbooks/...`) did not cover this specific failure mode."}

---

## What Went Well

<!-- Explicitly required section. Identify what worked correctly during the
     incident — safeguards that DID prevent worse harm, fast detection,
     effective communication, etc. This balances the retrospective and
     reinforces which existing protections are worth keeping and extending. -->

- {e.g., "The DLQ consumer correctly captured the 47 failed messages, preventing data loss. All were reprocessed successfully after the fix."}
- {e.g., "PagerDuty alert fired within 2 minutes of the SLO breach, well within the 15-minute SEV1 response target."}
- {e.g., "The circuit breaker on payment-service correctly fell back to the secondary gateway (Cashfree) after 3 consecutive Razorpay failures, limiting the blast radius."}

---

## Action Items

<!-- Each item must be a concrete, trackable change — a specific code fix,
     a new alert rule, a runbook update, a configuration change.
     NOT vague items like "be more careful" or "improve monitoring."
     Link to the issue tracker — don't duplicate tracking here. -->

| # | Action | Owner | Due Date | Tracker Link | Status |
|---|---|---|---|---|---|
| 1 | {e.g., "Add Grafana panel for `kafka_consumergroup_lag` filtered to payment consumer groups"} | {Name} | YYYY-MM-DD | {Jira/Linear link} | Open |
| 2 | {e.g., "Lower `KafkaConsumerLagHighP1` alert threshold from 1000 to 500 in `alert_rules.yml`"} | {Name} | YYYY-MM-DD | {Jira/Linear link} | Open |
| 3 | {e.g., "Add runbook `ops/runbooks/razorpay-timeout-spike.md` covering this failure mode"} | {Name} | YYYY-MM-DD | {Jira/Linear link} | Open |
| 4 | {e.g., "Update Statuspage automation to auto-set Payment Processing to Degraded when `job:payment_success_rate:ratio_rate5m` drops below 0.995"} | {Name} | YYYY-MM-DD | {Jira/Linear link} | Open |

---

## Blameless Language Guide

> This section is a standing reminder kept at the bottom of every post-mortem.
> It is not incident-specific — do not remove it.

**Blameless post-mortems are a core operating principle of Zippyra's engineering culture.** The platform's architecture — idempotent message processing, circuit breakers, dead-letter queues, multi-gateway payment failover, RFID/barcode dual-validation — is designed around the assumption that failures will happen. Post-mortems exist to improve the *system*, not to assign blame to individuals.

### Language principles:

- ✅ Write: *"The deploy pipeline allowed an untested config change to reach production"*
- ❌ Not: *"X pushed a bad config"*

- ✅ Write: *"The alert threshold was too permissive to detect this degradation pattern"*
- ❌ Not: *"The on-call engineer should have noticed sooner"*

- ✅ Write: *"The runbook did not cover this failure mode, leading to a 15-minute diagnostic delay"*
- ❌ Not: *"The engineer didn't know what to do"*

**Focus discussion on**: System gaps, process improvements, missing safeguards, inadequate observability, and architectural changes that would prevent recurrence — not individual performance.

**Reference**: This principle is stated in the platform's original operational culture notes and is operationalized through this template.
