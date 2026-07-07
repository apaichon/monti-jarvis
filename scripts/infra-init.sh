#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
ENV_FILE="$ROOT_DIR/infra/.env.dev"

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
MINIO_BUCKET=${MINIO_BUCKET:-monti-jarvis}

echo "==> Ensuring shared Postgres (poc-gml-postgres) is reachable..."
if docker ps --format '{{.Names}}' | grep -qx 'poc-gml-postgres'; then
  docker exec poc-gml-postgres pg_isready -U postgres -d postgres
else
  echo "note: poc-gml-postgres not running — start shared Postgres or point POSTGRES_URL elsewhere"
fi

echo "==> Ensuring Postgres database monti_jarvis exists..."
psql "$POSTGRES_ADMIN_URL" -v ON_ERROR_STOP=1 -tc "SELECT 1 FROM pg_database WHERE datname = 'monti_jarvis'" | grep -q 1 || \
  psql "$POSTGRES_ADMIN_URL" -v ON_ERROR_STOP=1 -c 'CREATE DATABASE monti_jarvis'

echo "==> Ensuring Postgres schema/tables exist..."
psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 <<SQL
CREATE SCHEMA IF NOT EXISTS "$POSTGRES_SCHEMA";
CREATE TABLE IF NOT EXISTS "$POSTGRES_SCHEMA".calls (
  id text PRIMARY KEY,
  agent_id text NOT NULL DEFAULT 'ava',
  title text NOT NULL DEFAULT 'Inbound call',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS "$POSTGRES_SCHEMA".messages (
  id bigserial PRIMARY KEY,
  call_id text NOT NULL REFERENCES "$POSTGRES_SCHEMA".calls(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('caller', 'agent')),
  content text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS "$POSTGRES_SCHEMA".call_sessions (
  id text PRIMARY KEY,
  tenant_id text NOT NULL DEFAULT 'demo',
  room_name text NOT NULL UNIQUE,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'ended')),
  started_at timestamptz NOT NULL DEFAULT now(),
  ended_at timestamptz,
  recording_key text
);
CREATE TABLE IF NOT EXISTS "$POSTGRES_SCHEMA".call_turns (
  id bigserial PRIMARY KEY,
  call_id text NOT NULL REFERENCES "$POSTGRES_SCHEMA".call_sessions(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('caller', 'agent', 'system')),
  content text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
SQL

if docker ps --format '{{.Names}}' | grep -qx 'poc-gml-minio'; then
  echo "==> Ensuring MinIO bucket $MINIO_BUCKET exists..."
  docker exec poc-gml-minio sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin >/dev/null && mc mb -p local/$MINIO_BUCKET >/dev/null || true"
fi

if docker ps --format '{{.Names}}' | grep -qx 'monti-nats'; then
  echo "NATS monti-nats is running on :4222"
else
  echo "note: run 'make infra-up' or 'docker compose -f infra/docker-compose.yml up -d' for NATS"
fi

if docker ps --format '{{.Names}}' | grep -qx 'monti-livekit'; then
  echo "LiveKit monti-livekit is running on :7880"
else
  echo "note: run 'make infra-up' or 'docker compose -f infra/docker-compose.yml up -d' for LiveKit"
fi

echo "infra ready: Postgres monti_jarvis schema $POSTGRES_SCHEMA, Redis DB 4, MinIO $MINIO_BUCKET, NATS :4222, LiveKit :7880"