#!/usr/bin/env bash
# Bring up a monti-jarvis server for E2E on E2E_PORT (default 8099), building
# the web apps + Go binary first. If one is already listening, reuse it.
#
# Test posture: AUTH_DISABLED=false (so platform guards are meaningful), the
# registration rate-limit stub disabled (so repeated signups don't 429), and a
# known JWT secret. Postgres/Redis/etc. come from infra/.env.dev, which the
# server loads itself (godotenv does not override these explicit env vars).
set -euo pipefail

PORT="${E2E_PORT:-8099}"
BINARY="monti-jarvis"
E2E_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT="$(cd "$E2E_DIR/.." && pwd)"

if curl -fsS "http://localhost:${PORT}/healthz" >/dev/null 2>&1; then
  echo "e2e: reusing server already listening on :${PORT}"
  exit 0
fi

cd "$ROOT"

if [ "${E2E_SKIP_BUILD:-0}" != "1" ]; then
  echo "e2e: building web apps + server (E2E_SKIP_BUILD=1 to skip)…"
  make build
fi

export PORT="${PORT}"
export APP_PUBLIC_URL="http://localhost:${PORT}"
export AUTH_DISABLED="${AUTH_DISABLED:-false}"
export TENANT_REGISTER_ENABLED=true
export TENANT_REGISTER_RATE_LIMIT="${TENANT_REGISTER_RATE_LIMIT:-100000}"
export JWT_SECRET="${JWT_SECRET:-e2e-jwt-secret-please-change-32-bytes-minimum-0}"

echo "e2e: starting ${BINARY} on :${PORT} (AUTH_DISABLED=${AUTH_DISABLED})…"
if [ -x "./${BINARY}" ]; then
  exec "./${BINARY}"
fi
exec go run ./cmd/server
