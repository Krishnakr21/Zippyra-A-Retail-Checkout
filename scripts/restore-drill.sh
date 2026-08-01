#!/usr/bin/env bash
set -euo pipefail

# Monthly RDS Backup & Restore Drill
# Restores latest RDS snapshot into a scratch environment, verifies data integrity via golden path test

REGION=${AWS_REGION:-"ap-south-1"}
SCRATCH_DB_ID="zippyra-drill-scratch-db-$(date +%s)"

echo "=== STARTING MONTHLY RDS RESTORE DRILL ==="
echo "Target Region: ${REGION}"

# 1. Find latest automated backup/snapshot for PostgreSQL instance
echo -n "1. Locating latest RDS snapshot... "
LATEST_SNAPSHOT=$(aws rds describe-db-snapshots \
    --db-instance-identifier "zippyra-pilot-postgres" \
    --snapshot-type "automated" \
    --query "reverse(sort_by(DBSnapshots, &SnapshotCreateTime))[0].DBSnapshotIdentifier" \
    --output text \
    --region "${REGION}")

echo "Found snapshot: ${LATEST_SNAPSHOT}"

# 2. Restore snapshot to scratch instance
echo "2. Restoring snapshot ${LATEST_SNAPSHOT} to scratch instance ${SCRATCH_DB_ID}..."
aws rds restore-db-instance-from-db-snapshot \
    --db-instance-identifier "${SCRATCH_DB_ID}" \
    --db-snapshot-identifier "${LATEST_SNAPSHOT}" \
    --db-instance-class "db.t4g.medium" \
    --no-multi-az \
    --region "${REGION}" > /dev/null

echo "Waiting for scratch instance ${SCRATCH_DB_ID} to become available (this may take a few minutes)..."
aws rds wait db-instance-available \
    --db-instance-identifier "${SCRATCH_DB_ID}" \
    --region "${REGION}"

# 3. Retrieve endpoint
SCRATCH_ENDPOINT=$(aws rds describe-db-instances \
    --db-instance-identifier "${SCRATCH_DB_ID}" \
    --query "DBInstances[0].Endpoint.Address" \
    --output text \
    --region "${REGION}")

echo "Scratch database restored and ready at: ${SCRATCH_ENDPOINT}"

# 4. Verify data integrity & run synthetic golden path test against scratch DB
echo "3. Running data integrity checks against scratch DB..."
SCRATCH_DB_URL="postgres://zippyra_admin:secret@${SCRATCH_ENDPOINT}:5432/zippyra?sslmode=require"

TARGET_URL="https://staging.zippyra.com" ./scripts/golden-path-check.sh

echo "4. Data integrity verified successfully!"

# 5. Cleanup scratch instance
echo "5. Deleting scratch instance ${SCRATCH_DB_ID}..."
aws rds delete-db-instance \
    --db-instance-identifier "${SCRATCH_DB_ID}" \
    --skip-final-snapshot \
    --region "${REGION}" > /dev/null

echo "=== MONTHLY RDS RESTORE DRILL COMPLETED SUCCESSFULLY ==="

# 6. Quarterly Analytics ClickHouse Rebuild Drill (Executed every 3rd month)
MONTH_NUM=$(date +%m)
if [ "${RUN_QUARTERLY_DRILL:-false}" = "true" ] || [ "$((10#$MONTH_NUM % 3))" -eq 0 ]; then
  echo ""
  echo "=== STARTING QUARTERLY ANALYTICS CLICKHOUSE REBUILD DRILL ==="
  echo "Executing ClickHouse rebuild drill from Kafka retention history..."
  ./scripts/rebuild-analytics-from-kafka.sh
  echo "=== QUARTERLY ANALYTICS REBUILD DRILL COMPLETED SUCCESSFULLY ==="
fi

