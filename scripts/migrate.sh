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

POSTGRES_URL=${POSTGRES_URL:-postgres://postgres:postgres@localhost:5432/monti_jarvis?sslmode=disable}
POSTGRES_SCHEMA=${POSTGRES_SCHEMA:-callcenter}
CLICKHOUSE_URL=${CLICKHOUSE_URL:-http://localhost:8123}
CLICKHOUSE_DB=${CLICKHOUSE_DB:-monti_jarvis}
CLICKHOUSE_USER=${CLICKHOUSE_USER:-monti}
CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD:-monti}

echo "==> Applying Postgres migrations (schema=$POSTGRES_SCHEMA)..."
psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -v POSTGRES_SCHEMA="$POSTGRES_SCHEMA" \
  -f "$ROOT_DIR/scripts/migrations/001_audit_columns_postgres.sql"

psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -v POSTGRES_SCHEMA="$POSTGRES_SCHEMA" \
  -f "$ROOT_DIR/scripts/migrations/028_tenant_ai_extensibility.sql"

psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -v POSTGRES_SCHEMA="$POSTGRES_SCHEMA" \
  -f "$ROOT_DIR/scripts/migrations/029_usage_ledger.sql"

psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -v POSTGRES_SCHEMA="$POSTGRES_SCHEMA" \
  -f "$ROOT_DIR/scripts/migrations/030_referrals.sql"

psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -v POSTGRES_SCHEMA="$POSTGRES_SCHEMA" \
  -f "$ROOT_DIR/scripts/migrations/031_referral_bonus_quota.sql"
if [ -n "${POSTGRES_KM_READONLY_ROLE:-}" ] && [ -n "${POSTGRES_TICKET_WRITE_ROLE:-}" ]; then
  echo "==> Applying Sprint 41 capability grants..."
  psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 \
    -v POSTGRES_SCHEMA="$POSTGRES_SCHEMA" \
    -v KM_READ_ROLE="$POSTGRES_KM_READONLY_ROLE" \
    -v TICKET_WRITE_ROLE="$POSTGRES_TICKET_WRITE_ROLE" \
    -f "$ROOT_DIR/scripts/migrations/032_security_capability_grants.sql"
else
  echo "note: Sprint 41 capability grants skipped — set POSTGRES_KM_READONLY_ROLE and POSTGRES_TICKET_WRITE_ROLE"
fi
if curl -fsS "$CLICKHOUSE_URL/ping" >/dev/null 2>&1; then
  echo "==> Applying ClickHouse audit columns (db=$CLICKHOUSE_DB)..."
  CH_HAS_TABLES=0
  if curl -fsS "$CLICKHOUSE_URL/?database=$CLICKHOUSE_DB&user=$CLICKHOUSE_USER&password=$CLICKHOUSE_PASSWORD" \
    --data "EXISTS TABLE km_embeddings" 2>/dev/null | grep -q 1; then
    CH_HAS_TABLES=1
  fi
  if [ "$CH_HAS_TABLES" = 1 ]; then
    while IFS= read -r stmt || [ -n "$stmt" ]; do
      case "$stmt" in
        ''|'--'*) continue ;;
      esac
      curl -fsS "$CLICKHOUSE_URL/?database=$CLICKHOUSE_DB&user=$CLICKHOUSE_USER&password=$CLICKHOUSE_PASSWORD" \
        --data-binary "$stmt" >/dev/null || {
          echo "clickhouse migration failed: $stmt" >&2
          exit 1
        }
    done < "$ROOT_DIR/scripts/migrations/002_audit_columns_clickhouse.sql"
  else
    echo "note: ClickHouse tables not present — audit columns applied on CREATE via infra-init"
  fi
else
  echo "note: ClickHouse not reachable at $CLICKHOUSE_URL — skipped 002_audit_columns_clickhouse.sql"
fi

echo "migrations applied"
