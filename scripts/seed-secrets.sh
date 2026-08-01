#!/usr/bin/env bash
set -euo pipefail

# Scripts to seed AWS Secrets Manager from .env.example files across backend services
ENV=${1:-pilot}
REGION=${AWS_REGION:-ap-south-1}
PREFIX="zippyra/${ENV}"

echo "Seeding AWS Secrets Manager under prefix: ${PREFIX} in region ${REGION}..."

# Shared secret seeding
echo "Seeding shared secret ${PREFIX}/shared..."
aws secretsmanager create-secret --name "${PREFIX}/shared" \
    --description "Shared secrets for Zippyra" \
    --secret-string '{"JWT_SECRET":"dev-jwt-secret-key-32chars-min!!","DB_MAX_CONNS":"10"}' \
    --region "${REGION}" 2>/dev/null || \
aws secretsmanager put-secret-value --secret-id "${PREFIX}/shared" \
    --secret-string '{"JWT_SECRET":"dev-jwt-secret-key-32chars-min!!","DB_MAX_CONNS":"10"}' \
    --region "${REGION}"

# Iterate services
for service_dir in backend/services/*; do
  if [ -d "$service_dir" ]; then
    svc=$(basename "$service_dir")
    secret_name="${PREFIX}/${svc}"
    
    echo "Processing ${svc}..."
    
    # Read .env.example if present, convert KEY=VAL to JSON
    if [ -f "${service_dir}/.env.example" ]; then
      json_payload=$(jq -n -R '
        [inputs | select(length > 0 and (startswith("#") | not)) | capture("(?<k>^[^=]+)=(?<v>.*)")] |
        from_entries
      ' < "${service_dir}/.env.example")
    else
      json_payload='{"SERVICE_NAME":"'"${svc}"'"}'
    fi
    
    aws secretsmanager create-secret --name "${secret_name}" \
        --description "Secrets for ${svc}" \
        --secret-string "${json_payload}" \
        --region "${REGION}" 2>/dev/null || \
    aws secretsmanager put-secret-value --secret-id "${secret_name}" \
        --secret-string "${json_payload}" \
        --region "${REGION}"
        
    echo "✅ Seeded ${secret_name}"
  fi
done

echo "Done seeding AWS Secrets Manager."
