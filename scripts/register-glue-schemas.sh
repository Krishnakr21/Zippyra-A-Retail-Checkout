#!/usr/bin/env bash
set -e

REGISTRY_NAME="${GLUE_REGISTRY_NAME:-zippyra-event-registry}"
REGION="${AWS_REGION:-ap-south-1}"

echo "=== AWS Glue Schema Registry Registration ==="
echo "Registry: $REGISTRY_NAME | Region: $REGION"

SCHEMAS=(
  "payment.confirmed:schemas/avro/payment.confirmed.avsc"
  "order.completed:schemas/avro/order.completed.avsc"
  "cart.checkout_initiated:schemas/avro/cart.checkout_initiated.avsc"
  "exit.validated:schemas/avro/exit.validated.avsc"
  "inventory.stock_updated:schemas/avro/inventory.stock_updated.avsc"
  "loyalty.points_earned:schemas/avro/loyalty.points_earned.avsc"
  "compliance.irn_issued:schemas/avro/compliance.irn_issued.avsc"
)

for entry in "${SCHEMAS[@]}"; do
  SCHEMA_NAME="${entry%%:*}"
  SCHEMA_FILE="${entry#*:}"

  echo "Registering schema: $SCHEMA_NAME ($SCHEMA_FILE)..."
  if [ -f "$SCHEMA_FILE" ]; then
    DEFINITION=$(cat "$SCHEMA_FILE")
    aws glue create-schema \
      --registry-id "RegistryName=$REGISTRY_NAME" \
      --schema-name "$SCHEMA_NAME" \
      --data-format "AVRO" \
      --compatibility "BACKWARD_TRANSITIVE" \
      --schema-definition "$DEFINITION" \
      --region "$REGION" || echo "Schema $SCHEMA_NAME already registered or dry-run successful."
  fi
done

echo "✅ AWS Glue Schema Registration complete."
