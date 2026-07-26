#!/usr/bin/env bash
set -e

echo "=== Starting Zippyra Monorepo Services ==="

echo "--> 1. Checking Local Infrastructure (Docker)..."
if docker info >/dev/null 2>&1; then
  echo "    Docker daemon is running. Launching containers..."
  docker-compose up -d || true
else
  echo "    ⚠️ Docker daemon not active. Microservices will run in mock/in-memory mode."
fi

echo "--> 2. Launching Backend Services..."
(cd backend && ./manage.sh start-all) &

echo "--> 3. Launching Web Apps..."
(cd web && pnpm dev) &

echo "=== Zippyra All Systems Running! ==="
