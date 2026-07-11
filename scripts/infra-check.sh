#!/usr/bin/env sh
set -eu

echo "Checking shared local infra..."

PG_CONTAINER=""
for candidate in monti-postgres poc-gml-postgres; do
  if docker ps --format '{{.Names}}' | grep -qx "$candidate"; then
    PG_CONTAINER=$candidate
    break
  fi
done
if [ -n "$PG_CONTAINER" ]; then
  docker exec "$PG_CONTAINER" pg_isready -U postgres -d postgres
else
  echo "postgres container not found (expected monti-postgres or poc-gml-postgres)"
fi

REDIS_CONTAINER=""
for candidate in monti-redis poc-gml-redis; do
  if docker ps --format '{{.Names}}' | grep -qx "$candidate"; then
    REDIS_CONTAINER=$candidate
    break
  fi
done
if [ -n "$REDIS_CONTAINER" ]; then
  docker exec "$REDIS_CONTAINER" redis-cli ping
else
  echo "redis container not found (expected monti-redis or poc-gml-redis)"
fi

MINIO_CONTAINER=""
for candidate in monti-minio poc-gml-minio; do
  if docker ps --format '{{.Names}}' | grep -qx "$candidate"; then
    MINIO_CONTAINER=$candidate
    break
  fi
done
if [ -n "$MINIO_CONTAINER" ]; then
  if docker exec "$MINIO_CONTAINER" sh -c 'command -v mc >/dev/null 2>&1'; then
    docker exec "$MINIO_CONTAINER" mc ready local
  elif curl -fsS http://localhost:9000/minio/health/live >/dev/null 2>&1; then
    echo "minio ok (localhost:9000)"
  else
    echo "minio container $MINIO_CONTAINER found but health check failed"
  fi
else
  echo "minio container not found (expected monti-minio or poc-gml-minio)"
fi

if docker ps --format '{{.Names}}' | grep -Eiq 'monti-nats|^nats$'; then
  echo "nats container found"
else
  echo "nats container not found (optional: monti-nats on :4222)"
fi

if docker ps --format '{{.Names}}' | grep -qx 'monti-livekit'; then
  echo "livekit container monti-livekit found"
else
  echo "livekit container monti-livekit not found (optional: :7880)"
fi

if curl -fsS http://localhost:8123/ping >/dev/null 2>&1; then
  echo "clickhouse ok (localhost:8123)"
elif docker ps --format '{{.Names}}' | grep -qx 'monti-clickhouse'; then
  docker exec monti-clickhouse wget -qO- http://localhost:8123/ping >/dev/null && echo "clickhouse container ok (host port not published)"
elif docker ps --format '{{.Names}}' | grep -qx 'poc-gml-clickhouse'; then
  echo "clickhouse poc-gml-clickhouse running but localhost:8123 unreachable — run 'make infra-up' for monti-clickhouse"
else
  echo "clickhouse not found (run 'make infra-up' for monti-clickhouse on :8123)"
fi
