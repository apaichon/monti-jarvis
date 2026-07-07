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