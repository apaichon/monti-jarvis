#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
COMPOSE_FILE="$ROOT_DIR/infra/docker-compose.yml"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required for infra-up"
  exit 1
fi

echo "==> Starting Monti Jarvis compose (Postgres, Redis, MinIO, NATS, LiveKit, ClickHouse)..."
docker compose -f "$COMPOSE_FILE" up -d

echo "==> Waiting for core services..."
for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do
  PG_OK=0
  REDIS_OK=0
  NATS_OK=0
  LK_OK=0
  CH_OK=0
  docker ps --format '{{.Names}}' | grep -Eqx 'monti-postgres|poc-gml-postgres' && PG_OK=1 || true
  docker ps --format '{{.Names}}' | grep -Eqx 'monti-redis|poc-gml-redis' && REDIS_OK=1 || true
  docker ps --format '{{.Names}}' | grep -qx 'monti-nats' && NATS_OK=1 || true
  docker ps --format '{{.Names}}' | grep -qx 'monti-livekit' && LK_OK=1 || true
  docker ps --format '{{.Names}}' | grep -qx 'monti-clickhouse' && CH_OK=1 || true
  if [ "$PG_OK" = 1 ] && [ "$REDIS_OK" = 1 ] && [ "$NATS_OK" = 1 ] && [ "$LK_OK" = 1 ] && [ "$CH_OK" = 1 ]; then
    PG_READY=0
    for c in monti-postgres poc-gml-postgres; do
      if docker ps --format '{{.Names}}' | grep -qx "$c" \
        && docker exec "$c" pg_isready -U postgres -d postgres >/dev/null 2>&1; then
        PG_READY=1
        break
      fi
    done
    if [ "$PG_READY" = 1 ] && curl -fsS http://localhost:8123/ping >/dev/null 2>&1; then
      break
    fi
  fi
  sleep 1
done

"$ROOT_DIR/scripts/infra-init.sh"
