# Sentry Alert Rules & PagerDuty Integration Runbook

## Overview & Architecture Clarification

The Zippyra platform employs a bi-furcated observability and crash-reporting architecture:

1. **Mobile Applications (`mobile-customer-app` & `mobile-staff-app`)**:
   - Initialized with Sentry (`SENTRY_DSN_MOBILE`).
   - Captures unhandled Flutter exceptions, fatal crashes, UI thread hangs, and release health session statistics.
   - Stack traces are automatically tagged with a `feature` key (`payment`, `exit`, `auth`, `onboarding`, `cart`, `shift`, `inventory`, `general`) via `beforeSend` frame inspection.

2. **Backend Services (`backend-services`)**:
   - Backend services rely on OpenTelemetry, Prometheus metrics (`http_requests_total`, `kafka_consumergroup_lag`, `dlq_messages_total`), and Grafana Alertmanager.
   - Direct PagerDuty pages trigger on HTTP 5xx error rate spikes (>1% over 5m), DLQ message accumulation (>0), and Kafka consumer lag (>1000).
   - Mobile client crashes flow exclusively into Sentry to prevent metric pollution in Prometheus.

---

## Configured Sentry Alert Rules & PagerDuty Triggers

Sentry projects are linked to the platform's primary **PagerDuty Integration** via Terraform (`infra/observability/sentry_alerts.tf`).

### Alert Rule 1: High-Impact Crash Threshold (SEV2)
- **Condition**: `> 10 unique users affected by the same issue within 1 hour`.
- **Target**: PagerDuty SEV2 Incident.
- **Rationale**: Prevents notification fatigue for isolated single-user edge cases while immediately escalating systemic bugs affecting multiple users.

### Alert Rule 2: Critical Feature Crash (SEV1 Immediate)
- **Condition**: `New issue (first-seen) with tag feature IN [payment, exit, auth]`.
- **Target**: PagerDuty SEV1 Immediate Page.
- **Rationale**: Any unhandled exception in self-checkout payments, RFID exit validation, or authentication threatens core revenue and store security. Triggers on the **first occurrence** without waiting for user count thresholds.

### Alert Rule 3: Release Health Degradation
- **Condition**: `Crash-free session rate drops below 99.0% for a release version within 24 hours`.
- **Target**: PagerDuty Release Health Alert.
- **Rationale**: Catches regression releases quickly post-rollout across Customer or Staff App deployments.

---

## Flutter App Feature-Tagging Implementation

Both Flutter apps initialize Sentry in `main.dart` with a `beforeSend` callback inspecting stacktrace frames to append the `feature` tag:

```dart
options.beforeSend = (event, {hint}) {
  String feature = 'general';
  final exceptions = event.exceptions;
  if (exceptions != null && exceptions.isNotEmpty) {
    final frames = exceptions.first.stacktrace?.frames;
    if (frames != null) {
      for (final frame in frames) {
        final path = (frame.absPath ?? frame.package ?? '').toLowerCase();
        if (path.contains('payment') || path.contains('checkout') || path.contains('razorpay')) {
          feature = 'payment';
          break;
        } else if (path.contains('exit') || path.contains('gate') || path.contains('qr')) {
          feature = 'exit';
          break;
        } else if (path.contains('auth') || path.contains('login') || path.contains('otp')) {
          feature = 'auth';
          break;
        }
      }
    }
  }
  event.setTag('feature', feature);
  return event;
};
```

---

## Weekly Digest & Ongoing Visibility

- **Sentry Built-in Digest**: Automated weekly reports enabled under `Project Settings > Notifications > Weekly Digest` for `mobile-customer-app` and `mobile-staff-app`.
- **Slack Summary**: Top 5 crash issues by frequency posted every Monday at 09:00 IST to `#mobile-engineering-alerts`.
