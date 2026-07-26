# Pre-Pilot Go-Live Checklist

## 1. Security & Auth
- [ ] Ed25519 key pair rotated and stored in AWS Secrets Manager / Vault.
- [ ] TLS / SSL Pinning verified in `zippyra_core`.
- [ ] Play Integrity / App Attestation enabled for Android.
- [ ] DPDP compliance data deletion hooks active.

## 2. Payments & Checkout
- [ ] Razorpay / Cashfree HMAC webhook verification enabled.
- [ ] UPI pending checkout polling & 90s timeout tested.
- [ ] Exit validation Ed25519 single-use QR verified with MQTT turnstile gates.

## 3. Infra & Monitoring
- [ ] OpenTelemetry spans verified across HTTP, Postgres, Redis, and Kafka.
- [ ] ClickHouse shrinkage & analytics pipelines connected.
- [ ] Kafka DLQ consumer and alerting verified.
