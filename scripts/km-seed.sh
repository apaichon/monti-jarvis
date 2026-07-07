#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
ENV_FILE="$ROOT_DIR/infra/.env.dev"
PORT=${PORT:-8091}

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

PORT=${PORT:-8091}
BASE_URL="http://localhost:${PORT}"
AUTH_DISABLED=${AUTH_DISABLED:-true}

auth_header() {
  if [ "$AUTH_DISABLED" = "true" ] || [ "$AUTH_DISABLED" = "1" ]; then
    return 0
  fi
  TOKEN=$(curl -fsS -X POST "$BASE_URL/api/auth/login" \
    -H 'content-type: application/json' \
    -d '{"email":"platform@monti.local","password":"monti-platform"}' \
    | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")
  printf 'Authorization: Bearer %s' "$TOKEN"
}

HDR=$(auth_header || true)
if [ -n "$HDR" ]; then
  curl -fsS -X POST "$BASE_URL/api/km/seed" -H "$HDR"
else
  curl -fsS -X POST "$BASE_URL/api/km/seed"
fi