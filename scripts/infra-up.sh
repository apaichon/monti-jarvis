#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
COMPOSE_FILE="$ROOT_DIR/infra/docker-compose.yml"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required for infra-up"
  exit 1
fi

echo "==> Starting Monti Jarvis compose (NATS, LiveKit)..."
docker compose -f "$COMPOSE_FILE" up -d

echo "==> Waiting for NATS and LiveKit..."
for i in 1 2 3 4 5 6 7 8 9 10; do
  NATS_OK=0
  LK_OK=0
  docker ps --format '{{.Names}}' | grep -qx 'monti-nats' && NATS_OK=1 || true
  docker ps --format '{{.Names}}' | grep -qx 'monti-livekit' && LK_OK=1 || true
  if [ "$NATS_OK" = 1 ] && [ "$LK_OK" = 1 ]; then
    break
  fi
  sleep 1
done

"$ROOT_DIR/scripts/infra-init.sh"