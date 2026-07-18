#!/usr/bin/env bash
set -Eeuo pipefail

# Controlled Sprint 32 billing-usage fixtures. The scope guard is deliberate:
# this script must never target a shared/demo tenant by accident.

ACTION="${1:-load}"
SCOPE="${FIXTURE_SCOPE:-uat-s31-demo}"
POSTGRES_SCHEMA="${POSTGRES_SCHEMA:-callcenter}"
REDIS_URL="${REDIS_URL:-redis://localhost:6379/4}"
REDIS_PREFIX="${REDIS_PREFIX:-monti_jarvis:}"
CLICKHOUSE_URL="${CLICKHOUSE_URL:-http://localhost:8123}"
CLICKHOUSE_DB="${CLICKHOUSE_DB:-monti_jarvis}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-monti}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-monti}"
START_DATE="${FIXTURE_START_DATE:-2026-07-18}"
END_DATE="${FIXTURE_END_DATE:-2026-07-18}"

die() {
  echo "sprint32-fixtures: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

[[ "$SCOPE" =~ ^uat-s31-[a-z0-9-]+$ ]] || die "FIXTURE_SCOPE must match uat-s31-<safe-suffix>"
[[ "$POSTGRES_SCHEMA" =~ ^[a-z_][a-z0-9_]*$ ]] || die "POSTGRES_SCHEMA must be a simple identifier"
[[ "$CLICKHOUSE_DB" =~ ^[a-z_][a-z0-9_]*$ ]] || die "CLICKHOUSE_DB must be a simple identifier"
[[ "$START_DATE" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || die "FIXTURE_START_DATE must be YYYY-MM-DD"
[[ "$END_DATE" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || die "FIXTURE_END_DATE must be YYYY-MM-DD"

TENANT_ID="${SCOPE}-tenant"
SECONDARY_TENANT_ID="${SCOPE}-tenant-mismatch"
START_MONTH="${START_DATE:0:4}${START_DATE:5:2}"
START_DAY="${START_DATE//-/}"
MONTH_KEY="${REDIS_PREFIX}quota:${TENANT_ID}:minutes:${START_MONTH}"
DAY_KEY="${REDIS_PREFIX}call_daily:${TENANT_ID}:${START_DAY}"
SECONDARY_MONTH_KEY="${REDIS_PREFIX}quota:${SECONDARY_TENANT_ID}:minutes:${START_MONTH}"
SECONDARY_DAY_KEY="${REDIS_PREFIX}call_daily:${SECONDARY_TENANT_ID}:${START_DAY}"

require_cmd psql
require_cmd redis-cli
require_cmd curl
require_cmd jq

[[ -n "${POSTGRES_URL:-}" ]] || die "POSTGRES_URL is required"

ch_query() {
  local query="$1"
  local body="${2:-}"
  local encoded
  encoded="$(jq -nr --arg q "$query" '$q|@uri')"
  curl -fsS -u "${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}" \
    --data-binary "$body" "${CLICKHOUSE_URL%/}/?query=${encoded}&database=${CLICKHOUSE_DB}" >/dev/null
}

load_postgres() {
  local has_payment_orders
  has_payment_orders="$(psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -qAtc "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = '${POSTGRES_SCHEMA}' AND table_name = 'payment_orders')")"
  [[ "$has_payment_orders" == "t" ]] || die "payment_orders is not initialized in schema ${POSTGRES_SCHEMA}"

  psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 \
    -v fixture_scope="$SCOPE" -v tenant_id="$TENANT_ID" \
    -v schema="$POSTGRES_SCHEMA" -v start_date="$START_DATE" -v end_date="$END_DATE" \
    <<'SQL'
BEGIN;
INSERT INTO :"schema".tenants (id, slug, name, status, created_by, updated_by)
VALUES (:'tenant_id', :'fixture_scope', 'Sprint 32 UAT tenant', 'active', 'sprint32-fixture', 'sprint32-fixture')
ON CONFLICT (id) DO UPDATE SET status = 'active', updated_at = now(), updated_by = 'sprint32-fixture';

INSERT INTO :"schema".tenants (id, slug, name, status, created_by, updated_by)
VALUES (:'fixture_scope' || '-tenant-mismatch', :'fixture_scope' || '-mismatch', 'Sprint 32 UAT mismatch tenant', 'active', 'sprint32-fixture', 'sprint32-fixture')
ON CONFLICT (id) DO UPDATE SET status = 'active', updated_at = now(), updated_by = 'sprint32-fixture';

INSERT INTO :"schema".tenant_entitlements
  (id, tenant_id, package_id, rules_schema_id, rules_snapshot, status, valid_from, created_by, updated_by)
SELECT :'fixture_scope' || '-entitlement', :'tenant_id', package_id, rules_schema_id, rules, 'active', now(), 'sprint32-fixture', 'sprint32-fixture'
FROM :"schema".package_limits WHERE package_id = 'pkg-starter'
ON CONFLICT (id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id, status = 'active', valid_until = NULL, updated_at = now(), updated_by = 'sprint32-fixture';

INSERT INTO :"schema".payment_orders
  (id, tenant_id, package_id, order_no, amount_cents, currency, status, provider, payment_method, transaction_id, paid_at, created_at, updated_at, created_by, updated_by)
VALUES
  (:'fixture_scope' || '-order-paid', :'tenant_id', 'pkg-starter', :'fixture_scope' || '-paid', 450000, '764', 'paid', 'fixture', 'test', :'fixture_scope' || '-tx-paid', (:'start_date' || ' 00:15:00+00')::timestamptz, (:'start_date' || ' 00:15:00+00')::timestamptz, now(), 'sprint32-fixture', 'sprint32-fixture'),
  (:'fixture_scope' || '-order-unpaid', :'tenant_id', 'pkg-starter', :'fixture_scope' || '-unpaid', 990000, '764', 'pending', 'fixture', 'test', '', NULL, (:'end_date' || ' 23:15:00+00')::timestamptz, now(), 'sprint32-fixture', 'sprint32-fixture'),
  (:'fixture_scope' || '-order-mismatch', :'fixture_scope' || '-tenant-mismatch', 'pkg-starter', :'fixture_scope' || '-mismatch', 210000, '764', 'paid', 'fixture', 'test', :'fixture_scope' || '-tx-mismatch', (:'start_date' || ' 01:15:00+00')::timestamptz, (:'start_date' || ' 01:15:00+00')::timestamptz, now(), 'sprint32-fixture', 'sprint32-fixture')
ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, amount_cents = EXCLUDED.amount_cents, paid_at = EXCLUDED.paid_at, updated_at = now(), updated_by = 'sprint32-fixture';
COMMIT;
SQL
}

load_redis() {
  redis-cli -u "$REDIS_URL" SET "$MONTH_KEY" 40 >/dev/null
  redis-cli -u "$REDIS_URL" SET "$DAY_KEY" 12 >/dev/null
  redis-cli -u "$REDIS_URL" SET "$SECONDARY_MONTH_KEY" 7 >/dev/null
  redis-cli -u "$REDIS_URL" SET "$SECONDARY_DAY_KEY" 3 >/dev/null
}

load_clickhouse() {
  local call_rows ai_rows
  call_rows=$(cat <<JSON
{"fact_id":"${SCOPE}-call-1","tenant_id":"${TENANT_ID}","call_id":"${SCOPE}-call-1","conversation_record_id":"${SCOPE}-conversation-1","avatar_id":"fixture-avatar","channel":"chat","source":"chat","status":"archived","started_at":"${START_DATE} 00:20:00","ended_at":"${START_DATE} 00:22:00","usage_date":"${START_DATE}","duration_seconds":120,"source_updated_at":"${START_DATE} 00:22:00","created_at":"${START_DATE} 00:22:00","updated_at":"${START_DATE} 00:22:00","created_by":"sprint32-fixture","updated_by":"sprint32-fixture"}
{"fact_id":"${SCOPE}-call-mismatch","tenant_id":"${SECONDARY_TENANT_ID}","call_id":"${SCOPE}-call-mismatch","conversation_record_id":"${SCOPE}-conversation-mismatch","avatar_id":"fixture-avatar","channel":"voice","source":"voice","status":"archived","started_at":"${START_DATE} 01:20:00","ended_at":"${START_DATE} 01:21:00","usage_date":"${START_DATE}","duration_seconds":60,"source_updated_at":"${START_DATE} 01:21:00","created_at":"${START_DATE} 01:21:00","updated_at":"${START_DATE} 01:21:00","created_by":"sprint32-fixture","updated_by":"sprint32-fixture"}
JSON
)
  ai_rows=$(cat <<JSON
{"fact_id":"${SCOPE}-ai-observed","tenant_id":"${TENANT_ID}","call_id":"${SCOPE}-call-1","conversation_record_id":"${SCOPE}-conversation-1","provider":"gemini","model":"fixture-text","modality":"text","measurement_state":"observed","input_units":1000,"output_units":400,"audio_seconds":0,"rate_version":"fixture-r1","cost_microunits":12500,"currency":"USD","usage_date":"${START_DATE}","source_updated_at":"${START_DATE} 00:22:00","updated_at":"${START_DATE} 00:22:00"}
{"fact_id":"${SCOPE}-ai-observed","tenant_id":"${TENANT_ID}","call_id":"${SCOPE}-call-1","conversation_record_id":"${SCOPE}-conversation-1","provider":"gemini","model":"fixture-text","modality":"text","measurement_state":"observed","input_units":1000,"output_units":400,"audio_seconds":0,"rate_version":"fixture-r1","cost_microunits":12500,"currency":"USD","usage_date":"${START_DATE}","source_updated_at":"${START_DATE} 00:22:01","updated_at":"${START_DATE} 00:22:01"}
{"fact_id":"${SCOPE}-ai-estimated","tenant_id":"${TENANT_ID}","call_id":"${SCOPE}-call-1","conversation_record_id":"${SCOPE}-conversation-1","provider":"gemini","model":"fixture-live","modality":"voice","measurement_state":"estimated","input_units":0,"output_units":0,"audio_seconds":90,"rate_version":"fixture-r1","cost_microunits":900000,"currency":"USD","usage_date":"${START_DATE}","source_updated_at":"${START_DATE} 00:23:00","updated_at":"${START_DATE} 00:23:00"}
{"fact_id":"${SCOPE}-ai-unavailable","tenant_id":"${TENANT_ID}","call_id":"${SCOPE}-call-1","conversation_record_id":"${SCOPE}-conversation-1","provider":"gemini","model":"fixture-unknown","modality":"text","measurement_state":"unavailable","input_units":0,"output_units":0,"audio_seconds":0,"rate_version":"unconfigured","cost_microunits":0,"currency":"USD","usage_date":"${START_DATE}","source_updated_at":"${START_DATE} 00:24:00","updated_at":"${START_DATE} 00:24:00"}
JSON
)
  ch_query "INSERT INTO ${CLICKHOUSE_DB}.call_center_usage_facts (fact_id, tenant_id, call_id, conversation_record_id, avatar_id, channel, source, status, started_at, ended_at, usage_date, duration_seconds, source_updated_at, created_at, updated_at, created_by, updated_by) FORMAT JSONEachRow" "$call_rows"
  ch_query "INSERT INTO ${CLICKHOUSE_DB}.ai_cost_usage_facts (fact_id, tenant_id, call_id, conversation_record_id, provider, model, modality, measurement_state, input_units, output_units, audio_seconds, rate_version, cost_microunits, currency, usage_date, source_updated_at, updated_at) FORMAT JSONEachRow" "$ai_rows"
}

reset_postgres() {
  psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -v tenant_id="$TENANT_ID" -v secondary_tenant_id="$SECONDARY_TENANT_ID" -v schema="$POSTGRES_SCHEMA" \
    <<'SQL'
DELETE FROM :"schema".payment_orders WHERE tenant_id IN (:'tenant_id', :'secondary_tenant_id');
DELETE FROM :"schema".tenant_entitlements WHERE tenant_id = :'tenant_id';
DELETE FROM :"schema".tenants WHERE id = :'tenant_id';
DELETE FROM :"schema".tenants WHERE id = :'secondary_tenant_id';
SQL
}

reset_redis() {
  redis-cli -u "$REDIS_URL" DEL "$MONTH_KEY" "$DAY_KEY" "$SECONDARY_MONTH_KEY" "$SECONDARY_DAY_KEY" >/dev/null
}

reset_clickhouse() {
  ch_query "ALTER TABLE ${CLICKHOUSE_DB}.call_center_usage_facts DELETE WHERE tenant_id IN ('${TENANT_ID}', '${SECONDARY_TENANT_ID}')"
  ch_query "ALTER TABLE ${CLICKHOUSE_DB}.ai_cost_usage_facts DELETE WHERE tenant_id IN ('${TENANT_ID}', '${SECONDARY_TENANT_ID}')"
}

case "$ACTION" in
  load)
    load_postgres
    load_redis
    load_clickhouse
    echo "loaded fixture_scope=${SCOPE} tenant_id=${TENANT_ID} dates=${START_DATE}:${END_DATE}"
    ;;
  reset)
    reset_clickhouse
    reset_redis
    reset_postgres
    echo "reset fixture_scope=${SCOPE} tenant_id=${TENANT_ID}"
    ;;
  *)
    die "usage: $0 [load|reset]"
    ;;
esac
