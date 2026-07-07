#!/usr/bin/env sh
set -eu

echo "Checking shared local infra..."

if docker ps --format '{{.Names}}' | grep -qx 'poc-gml-postgres'; then
  docker exec poc-gml-postgres pg_isready -U postgres -d postgres
else
  echo "postgres container poc-gml-postgres not found"
fi

if docker ps --format '{{.Names}}' | grep -qx 'poc-gml-redis'; then
  docker exec poc-gml-redis redis-cli ping
else
  echo "redis container poc-gml-redis not found"
fi

if docker ps --format '{{.Names}}' | grep -qx 'poc-gml-minio'; then
  docker exec poc-gml-minio mc ready local
else
  echo "minio container poc-gml-minio not found"
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