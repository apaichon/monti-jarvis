#!/usr/bin/env bash
set -Eeuo pipefail

# Reproducible UAT for the Sprint 31 platform billing/usage surface. The
# fixture script owns all writes and the EXIT trap removes them again.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FIXTURE_SCRIPT="${FIXTURE_SCRIPT:-${ROOT_DIR}/scripts/sprint32-usage-fixtures.sh}"
API_URL="${API_URL:-http://localhost:8091}"
SCOPE="${FIXTURE_SCOPE:-uat-s31-demo}"
START_DATE="${FIXTURE_START_DATE:-2026-07-18}"
END_DATE="${FIXTURE_END_DATE:-2026-07-18}"
EVIDENCE_DIR="${UAT_EVIDENCE_DIR:-${ROOT_DIR}/var/uat/sprint32}"
EVIDENCE_FILE="${UAT_EVIDENCE_FILE:-${EVIDENCE_DIR}/${SCOPE}.json}"
MISMATCH_EVIDENCE_FILE="${UAT_MISMATCH_EVIDENCE_FILE:-${EVIDENCE_DIR}/${SCOPE}-mismatch.json}"

die() {
  echo "sprint32-uat: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

[[ -x "$FIXTURE_SCRIPT" ]] || die "fixture script is not executable: $FIXTURE_SCRIPT"
[[ -n "${PLATFORM_ADMIN_TOKEN:-}" ]] || die "PLATFORM_ADMIN_TOKEN is required"

require_cmd curl
require_cmd jq
require_cmd redis-cli

TENANT_ID="${SCOPE}-tenant"
SECONDARY_TENANT_ID="${SCOPE}-tenant-mismatch"
REDIS_URL="${REDIS_URL:-redis://localhost:6379/4}"
REDIS_PREFIX="${REDIS_PREFIX:-monti_jarvis:}"
START_MONTH="${START_DATE:0:4}${START_DATE:5:2}"
START_DAY="${START_DATE//-/}"
MONTH_KEY="${REDIS_PREFIX}quota:${TENANT_ID}:minutes:${START_MONTH}"
DAY_KEY="${REDIS_PREFIX}call_daily:${TENANT_ID}:${START_DAY}"
SECONDARY_MONTH_KEY="${REDIS_PREFIX}quota:${SECONDARY_TENANT_ID}:minutes:${START_MONTH}"
SECONDARY_DAY_KEY="${REDIS_PREFIX}call_daily:${SECONDARY_TENANT_ID}:${START_DAY}"

cleanup() {
  if ! "$FIXTURE_SCRIPT" reset >/dev/null; then
    echo "sprint32-uat: fixture cleanup failed for ${SCOPE}; run '${FIXTURE_SCRIPT} reset' manually" >&2
  fi
}
trap cleanup EXIT

"$FIXTURE_SCRIPT" load

mkdir -p "$EVIDENCE_DIR"
fetch_usage() {
  local tenant_id="$1"
  local evidence_file="$2"
  curl -fsS \
    -H "Authorization: Bearer ${PLATFORM_ADMIN_TOKEN}" \
    --get "${API_URL%/}/api/platform/billing/usage" \
    --data-urlencode "start_date=${START_DATE}" \
    --data-urlencode "end_date=${END_DATE}" \
    --data-urlencode "tenant_id=${tenant_id}" \
    --data-urlencode "limit=1" \
    --data-urlencode "offset=0" \
    >"$evidence_file"
}

fetch_usage "$TENANT_ID" "$EVIDENCE_FILE"
fetch_usage "$SECONDARY_TENANT_ID" "$MISMATCH_EVIDENCE_FILE"

jq -e \
  --arg tenant_id "$TENANT_ID" \
  --arg start_date "$START_DATE" \
  --arg end_date "$END_DATE" \
  '
    .range.start_date == $start_date and
    .range.end_date == $end_date and
    .billing.paid_orders == 1 and
    .billing.paid_amount_minor == 450000 and
    .billing.currency == "764" and
    .billing.status == "current" and
    .ai_cost.observed_events == 1 and
    .ai_cost.estimated_events == 1 and
    .ai_cost.unavailable_events == 1 and
    .ai_cost.observed_cost_microunits == 12500 and
    .ai_cost.estimated_cost_microunits == 900000 and
    .ai_cost.status == "warning" and
    .reconciliation.orders_entitlements == "ok" and
    any(.tenants[]?;
      .tenant_id == $tenant_id and
      .paid_orders == 1 and
      .paid_amount_minor == 450000 and
      .reporting_minutes == 2 and
      .ai_observed_events == 1 and
      .ai_estimated_events == 1 and
      .ai_unavailable_events == 1
    )
  ' "$EVIDENCE_FILE" >/dev/null || die "billing/usage response did not match the controlled fixture"

[[ "$(redis-cli -u "$REDIS_URL" GET "$MONTH_KEY")" == "40" ]] || die "monthly Redis quota fixture was not preserved"
[[ "$(redis-cli -u "$REDIS_URL" GET "$DAY_KEY")" == "12" ]] || die "daily Redis quota fixture was not preserved"
[[ "$(redis-cli -u "$REDIS_URL" GET "$SECONDARY_MONTH_KEY")" == "7" ]] || die "secondary monthly Redis quota fixture was not preserved"
[[ "$(redis-cli -u "$REDIS_URL" GET "$SECONDARY_DAY_KEY")" == "3" ]] || die "secondary daily Redis quota fixture was not preserved"

jq -e \
  --arg tenant_id "$SECONDARY_TENANT_ID" \
  '
    .billing.paid_orders == 1 and
    .billing.paid_amount_minor == 210000 and
    any(.tenants[]?;
      .tenant_id == $tenant_id and
      .paid_orders == 1 and
      .paid_amount_minor == 210000 and
      .package.status == "unassigned"
    )
  ' "$MISMATCH_EVIDENCE_FILE" >/dev/null || die "entitlement-mismatch fixture was not visible in the tenant aggregate"

echo "passed fixture_scope=${SCOPE} evidence=${EVIDENCE_FILE} mismatch_evidence=${MISMATCH_EVIDENCE_FILE}"
