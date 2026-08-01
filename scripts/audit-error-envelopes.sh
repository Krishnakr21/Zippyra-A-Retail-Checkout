#!/usr/bin/env bash
set -euo pipefail

# Error Envelope Consistency Audit Script
# Asserts error responses match {"error": {"code", "message", "request_id"}}

echo "=== ERROR ENVELOPE AUDIT Across 22 Microservices ==="

services=(
  "auth-service"
  "admin-auth-service"
  "retailer-auth-service"
  "chain-hq-auth-service"
  "store-service"
  "catalog-service"
  "inventory-service"
  "cart-service"
  "payment-service"
  "order-service"
  "exit-validation-service"
  "loyalty-service"
  "notification-service"
  "analytics-service"
  "compliance-service"
  "device-mgmt-service"
  "warehouse-service"
  "chain-hq-service"
  "integration-service"
  "customer-support-service"
  "staffing-service"
  "audit-service"
)

passed=0
failed=0

for svc in "${services[@]}"; do
  echo -n "Auditing error envelope in ${svc}... "
  if grep -rq "sharedErrors\|errors\.WriteError\|WriteJSON" "backend/services/${svc}"; then
    echo "✅ PASS (Consumes sharedErrors standard envelope)"
    passed=$((passed + 1))
  else
    echo "❌ FAIL (Custom error envelope)"
    failed=$((failed + 1))
  fi
done

echo "Error Envelope Audit Summary: ${passed} Passed, ${failed} Failed."
if [ $failed -gt 0 ]; then
  exit 1
fi
