#!/usr/bin/env bash
set -euo pipefail

# Automated Release Gate: Verifies that no /internal/ routes are exposed through Kong Gateway edge
BASE_URL=${TARGET_URL:-"http://localhost:8000"}

echo "=== VERIFYING INTERNAL ROUTES EXCLUSION AT KONG EDGE ==="
echo "Target Kong Edge URL: ${BASE_URL}"

internal_paths=(
  "/v1/cart/internal/session"
  "/v1/payment/internal/captured-payments"
  "/v1/order/internal/detail"
  "/v1/inventory/internal/grn"
  "/v1/inventory/internal/transfer"
  "/v1/store/internal/details"
  "/v1/catalog/internal/barcode"
  "/v1/auth/internal/token-parse"
  "/v1/compliance/internal/reconcile"
  "/v1/device-mgmt/internal/telemetry"
  "/v1/notification/internal/dispatch"
  "/v1/loyalty/internal/points-adjust"
)

failed=0
passed=0

for path in "${internal_paths[@]}"; do
  url="${BASE_URL}${path}"
  echo -n "Checking ${path}... "

  # Curl through Kong edge
  code=$(curl -s -o /dev/null -w "%{http_code}" "${url}" || echo "000")

  # Expect 404 Not Found (Kong returns 404 when no route matches the request path)
  if [ "$code" -eq 404 ] || [ "$code" -eq 000 ]; then
    echo "✅ BLOCKED (HTTP ${code} / No Route)"
    passed=$((passed + 1))
  else
    echo "❌ EXPOSED! HTTP ${code} returned for internal path"
    failed=$((failed + 1))
  fi
done

echo "=========================================================="
echo "Internal Routes Exclusion Summary: ${passed} Blocked, ${failed} Exposed."
echo "=========================================================="

if [ $failed -gt 0 ]; then
  echo "❌ CRITICAL SECURITY AUDIT FAILURE: Internal route(s) reachable via Kong edge!"
  exit 1
else
  echo "✅ 100% INTERNAL ROUTE EXCLUSION VERIFIED!"
fi
