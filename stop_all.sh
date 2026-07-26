#!/usr/bin/env bash

echo "=== Stopping Zippyra Monorepo Services ==="

echo "--> Stopping Web and Backend processes..."
pkill -f "pnpm dev" || true
pkill -f "go run" || true
(cd backend && ./manage.sh stop-all) || true

echo "--> Stopping Docker Containers..."
docker-compose down

echo "=== Zippyra Stopped Successfully! ==="
