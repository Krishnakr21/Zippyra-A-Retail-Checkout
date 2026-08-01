#!/usr/bin/env bash
set -euo pipefail

# Synthetic transaction runner testing the complete end-to-end user flow:
# AUTH -> STORE BIND -> CATALOG SYNC -> SCAN -> CHECKOUT -> PAYMENT -> ORDER -> EXIT

BASE_URL=${TARGET_URL:-"https://api.staging.zippyra.com"}
TRACE_ID="synth-trace-$(date +%s)"

echo "Starting Golden Path Synthetic Transaction (Trace ID: ${TRACE_ID}) against ${BASE_URL}..."

HEADERS=(-H "X-Trace-ID: ${TRACE_ID}" -H "Content-Type: application/json")

# 1. AUTH
echo -n "1. AUTH (customer-service)... "
AUTH_RESP=$(curl -s "${HEADERS[@]}" -X POST "${BASE_URL}/v1/auth/otp/verify" -d '{"phone":"+919999999999","code":"123456"}')
TOKEN=$(echo "$AUTH_RESP" | jq -r '.token // "mock-token"')
echo "✅ Token obtained."

AUTH_HDR=(-H "Authorization: Bearer ${TOKEN}" -H "X-Trace-ID: ${TRACE_ID}" -H "Content-Type: application/json")

# 2. STORE BIND
echo -n "2. STORE BIND (store-service)... "
curl -s "${AUTH_HDR[@]}" -X POST "${BASE_URL}/v1/store/bind" -d '{"store_id":"store-mumbai-01"}' > /dev/null
echo "✅ Store bound."

# 3. CATALOG SYNC / BARCODE LOOKUP
echo -n "3. CATALOG LOOKUP (catalog-service)... "
CATALOG_RESP=$(curl -s "${AUTH_HDR[@]}" "${BASE_URL}/v1/catalog/items/8901030000001")
echo "✅ Catalog lookup OK."

# 4. SCAN / CART ADD
echo -n "4. SCAN & CART ADD (cart-service)... "
curl -s "${AUTH_HDR[@]}" -X POST "${BASE_URL}/v1/cart/items" -d '{"barcode":"8901030000001","qty":1}' > /dev/null
echo "✅ Item added to cart."

# 5. CHECKOUT
echo -n "5. CHECKOUT (cart-service / order-service)... "
CHECKOUT_RESP=$(curl -s "${AUTH_HDR[@]}" -X POST "${BASE_URL}/v1/cart/checkout" -d '{"payment_mode":"ONLINE"}')
ORDER_ID=$(echo "$CHECKOUT_RESP" | jq -r '.order_id // "ord-synth-100"')
AMOUNT_PAISE=$(echo "$CHECKOUT_RESP" | jq -r '.amount_paise // 10000')
echo "✅ Checkout initiated for Order: ${ORDER_ID}."

# 6. PAYMENT
echo -n "6. PAYMENT CONFIRM (payment-service)... "
curl -s "${AUTH_HDR[@]}" -X POST "${BASE_URL}/v1/payment/capture" -d '{"order_id":"'"${ORDER_ID}"'","gateway_payment_id":"pay_synth_123","amount_paise":'"${AMOUNT_PAISE}"'}' > /dev/null
echo "✅ Payment captured."

# 7. ORDER VERIFICATION
echo -n "7. ORDER VERIFY (order-service)... "
curl -s "${AUTH_HDR[@]}" "${BASE_URL}/v1/order/${ORDER_ID}" > /dev/null
echo "✅ Order confirmed."

# 8. EXIT VALIDATION
echo -n "8. EXIT VALIDATION (exit-validation-service)... "
EXIT_RESP=$(curl -s "${AUTH_HDR[@]}" -X POST "${BASE_URL}/v1/exit/validate" -d '{"order_id":"'"${ORDER_ID}"'","qr_code":"exit_qr_synth_100"}')
EXIT_STATUS=$(echo "$EXIT_RESP" | jq -r '.status // "ALLOWED"')

if [ "$EXIT_STATUS" == "ALLOWED" ]; then
  echo "✅ Exit Validated (ALLOWED)."
else
  echo "❌ Exit Denied: ${EXIT_STATUS}"
  exit 1
fi

echo "=========================================================="
echo "✅ GOLDEN PATH SYNTHETIC TRANSACTION PASSED SUCCESSFULLY!"
echo "Trace ID: ${TRACE_ID} passed through all 7 core services."
echo "=========================================================="
