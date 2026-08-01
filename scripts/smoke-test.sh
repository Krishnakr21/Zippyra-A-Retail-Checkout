#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${STAGING_ALB_URL:-"https://api.staging.zippyra.com"}

echo "Running Smoke Tests against ${BASE_URL}..."

services=(
  "auth-service:/v1/auth/healthz/ready"
  "admin-auth-service:/v1/admin-auth/healthz/ready"
  "retailer-auth-service:/v1/retailer-auth/healthz/ready"
  "chain-hq-auth-service:/v1/chain-hq-auth/healthz/ready"
  "store-service:/v1/store/healthz/ready"
  "catalog-service:/v1/catalog/healthz/ready"
  "inventory-service:/v1/inventory/healthz/ready"
  "cart-service:/v1/cart/healthz/ready"
  "payment-service:/v1/payment/healthz/ready"
  "order-service:/v1/order/healthz/ready"
  "exit-validation-service:/v1/exit/healthz/ready"
  "loyalty-service:/v1/loyalty/healthz/ready"
  "notification-service:/v1/notification/healthz/ready"
  "analytics-service:/v1/analytics/healthz/ready"
  "compliance-service:/v1/compliance/healthz/ready"
  "device-mgmt-service:/v1/device/healthz/ready"
  "warehouse-service:/v1/warehouse/healthz/ready"
  "chain-hq-service:/v1/chain-hq/healthz/ready"
  "integration-service:/v1/integration/healthz/ready"
  "customer-support-service:/v1/support/healthz/ready"
  "staffing-service:/v1/staffing/healthz/ready"
  "audit-service:/v1/audit/healthz/ready"
)

failed=0
for entry in "${services[@]}"; do
  svc="${entry%%:*}"
  path="${entry#*:}"
  url="${BASE_URL}${path}"

  echo -n "Checking ${svc} at ${path}... "
  code=$(curl -s -o /dev/null -w "%{http_code}" "${url}" || echo "000")

  if [ "$code" -eq 200 ]; then
    echo "✅ PASS (200 OK)"
  else
    echo "❌ FAIL (HTTP ${code})"
    failed=$((failed + 1))
  fi
done

if [ $failed -gt 0 ]; then
  echo "❌ Smoke test failed: ${failed} service(s) unready."
  exit 1
else
  echo "✅ All 22 microservices passed smoke tests!"
fi
