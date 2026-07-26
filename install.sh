#!/usr/bin/env bash
set -e

export PATH="$PATH:$HOME/.pub-cache/bin"

echo "=== Bootstrapping Zippyra Toolchains ==="

echo "--> 1. Installing Backend Go Dependencies..."
if [ -d "backend" ]; then
  (cd backend && go mod download)
fi

echo "--> 2. Installing Web PNPM Dependencies..."
if [ -d "web" ]; then
  (cd web && pnpm install)
fi

echo "--> 3. Bootstrapping Mobile Flutter Workspace..."
if [ -d "mobile" ]; then
  (cd mobile/packages/zippyra_core && flutter pub get)
  (cd mobile/customer_app && flutter pub get)
  (cd mobile/staff_app && flutter pub get)
fi

echo "=== Zippyra Bootstrap Complete! ==="
