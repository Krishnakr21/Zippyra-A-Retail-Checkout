# Runbook: Payment Gateway Down (SEV1)

## Symptoms
- Alert: `ServiceHighErrorRate` on `payment-service` (>1% HTTP 5xx errors)
- PagerDuty P1 triggered
- Razorpay API failing or unreachable

## Immediate Actions

1. **Verify Gateway Status**:
   - Check Razorpay status page: `https://status.razorpay.com`
   - Check payment-service metrics in Grafana (`Payment Gateway Health` dashboard)

2. **Enable Cash Fallback Mode**:
   - If Razorpay is down, enable store cash/UPI POS fallback mode in store-service:
   ```bash
   kubectl set env deployment/store-service -n zippyra-prod ALLOW_CASH_ONLY_FALLBACK=true
   ```

3. **Check Outbox Relay Queue**:
   - Ensure `payment.confirmed` outbox relay worker is not deadlocked:
   ```bash
   kubectl logs -n zippyra-prod -l app=payment-service --tail=100 | grep "outbox"
   ```

4. **Rollback if Caused by Deploy**:
   - If error rate spiked post-deploy, trigger rollback:
   ```bash
   kubectl rollout undo deployment/payment-service -n zippyra-prod
   ```
