#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
ENV_FILE="$ROOT_DIR/infra/.env.dev"
COMPOSE_FILE="$ROOT_DIR/infra/docker-compose.yml"

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
elif [ -f "$ROOT_DIR/infra/.env.example" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ROOT_DIR/infra/.env.example"
  set +a
fi

POSTGRES_ADMIN_URL=${POSTGRES_ADMIN_URL:-postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable}
POSTGRES_URL=${POSTGRES_URL:-postgres://postgres:postgres@localhost:5432/monti_jarvis?sslmode=disable}
POSTGRES_SCHEMA=${POSTGRES_SCHEMA:-callcenter}
REDIS_URL=${REDIS_URL:-redis://localhost:6379/4}
MINIO_BUCKET=${MINIO_BUCKET:-monti-jarvis}

CLICKHOUSE_URL=${CLICKHOUSE_URL:-http://localhost:8123}
CLICKHOUSE_DB=${CLICKHOUSE_DB:-monti_jarvis}
CLICKHOUSE_USER=${CLICKHOUSE_USER:-monti}
CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD:-monti}
echo "==> Dropping ClickHouse database $CLICKHOUSE_DB..."
if curl -fsS "$CLICKHOUSE_URL/ping" >/dev/null 2>&1; then
  curl -fsS "$CLICKHOUSE_URL/?user=$CLICKHOUSE_USER&password=$CLICKHOUSE_PASSWORD" \
    --data "DROP DATABASE IF EXISTS $CLICKHOUSE_DB" >/dev/null 2>&1 || echo "warn: could not drop ClickHouse database"
else
  echo "note: ClickHouse not reachable — skipping database drop"
fi

echo "==> Stopping Monti Jarvis compose services (NATS, LiveKit, ClickHouse)..."
if command -v docker >/dev/null 2>&1; then
  docker compose -f "$COMPOSE_FILE" down --remove-orphans 2>/dev/null || true
else
  echo "docker not found — skipping compose down"
fi

echo "==> Dropping Postgres database monti_jarvis..."
if command -v psql >/dev/null 2>&1; then
  psql "$POSTGRES_ADMIN_URL" -v ON_ERROR_STOP=1 <<'SQL' || echo "warn: could not drop monti_jarvis (is Postgres up?)"
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'monti_jarvis' AND pid <> pg_backend_pid();
DROP DATABASE IF EXISTS monti_jarvis;
SQL
else
  echo "psql not found — skipping database drop"
fi

echo "==> Flushing Redis DB for Monti..."
if command -v redis-cli >/dev/null 2>&1; then
  redis-cli -u "$REDIS_URL" FLUSHDB >/dev/null 2>&1 || echo "warn: could not flush Redis (is Redis up?)"
else
  REDIS_CONTAINER=""
  for candidate in monti-redis poc-gml-redis; do
    if docker ps --format '{{.Names}}' | grep -qx "$candidate"; then
      REDIS_CONTAINER=$candidate
      break
    fi
  done
  if [ -n "$REDIS_CONTAINER" ]; then
    DB_INDEX=$(printf '%s' "$REDIS_URL" | sed -n 's|.*/\([0-9][0-9]*\)$|\1|p')
    DB_INDEX=${DB_INDEX:-4}
    docker exec "$REDIS_CONTAINER" redis-cli -n "$DB_INDEX" FLUSHDB >/dev/null 2>&1 || echo "warn: could not flush Redis via docker"
  else
    echo "redis-cli not found — skipping Redis flush"
  fi
fi

echo "==> Removing MinIO bucket $MINIO_BUCKET..."
if docker ps --format '{{.Names}}' | grep -qx 'poc-gml-minio'; then
  docker exec poc-gml-minio sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin >/dev/null 2>&1 && mc rb --force local/$MINIO_BUCKET >/dev/null 2>&1 || true"
elif docker ps --format '{{.Names}}' | grep -qx 'monti-minio'; then
  docker exec monti-minio sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin >/dev/null 2>&1 && mc rb --force local/$MINIO_BUCKET >/dev/null 2>&1 || true"
else
  echo "minio container not found — skipping bucket removal"
fi

echo "infra destroyed: ClickHouse DB dropped, compose down, monti_jarvis DB dropped, Redis flushed, MinIO bucket removed"