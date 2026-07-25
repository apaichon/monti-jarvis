---
id: UAT-SPRINT-045
status: passed
updated: 2026-07-25
sprint: SPRINT-045
owner: Tester
---

# SPRINT-045 — Step-by-step manual UAT

This runbook records the follow-up verification for Sprint 45 points: idempotent usage
events, tenant/platform projections, mobile lifecycle behavior, and
two-tenant isolation under concurrent use.

Sprint 45 follow-up verification is complete. The final release cut/tag is
still pending explicit release authorization.

Run all commands from the repository root. Use only the local Docker fixture.
Do not paste real provider credentials, customer data, or production tenant
IDs into the evidence.

## 0. Test variables and prerequisites

Set the URL and the tenant/customer values used by your local fixture:

~~~bash
export BASE=http://localhost:8091
export POSTGRES_URL=postgres://postgres:postgres@localhost:5432/monti_jarvis?sslmode=disable
export TENANT_A=demo
export TENANT_B=demo-b
export START_DATE=$(date -u +%F)
export END_DATE=$START_DATE

# Use the seeded or Sprint 20 fixture credentials in this environment.
export PLATFORM_EMAIL=platform@monti.local
export PLATFORM_PASSWORD=monti-platform
export TENANT_A_EMAIL=admin@demo.local
export TENANT_A_PASSWORD=demo-admin
export TENANT_B_EMAIL=admin@demo-b.local
export TENANT_B_PASSWORD=demo-b-admin
~~~

Required tools: Docker Compose, curl, jq, psql, redis-cli, and optionally
websocat for WebSocket checks.

Record the tester, date, commit, environment, and tenant IDs in the evidence
log before starting.

## 1. Start and verify local infrastructure

1. Start the scoped local services:

   ~~~bash
   make infra-up
   make infra-check
   ~~~

   Expected: Postgres, Redis, MinIO, NATS, LiveKit, and ClickHouse are
   healthy.

2. Apply the schema and Sprint 45 migration:

   ~~~bash
   scripts/migrate.sh
   ~~~

   Expected: the command ends with migrations applied and no fatal error.

3. Verify the required data stores:

   ~~~bash
   psql "$POSTGRES_URL" -At -c \
     "SELECT to_regclass('callcenter.usage_events'), to_regclass('callcenter.usage_reconciliation_runs');"
   redis-cli -n 4 ping
   curl -fsS http://localhost:8123/ping
   curl -fsS http://localhost:9000/minio/health/live
   ~~~

   Expected: both Postgres relations are present, Redis returns PONG, and
   ClickHouse/MinIO return success.

4. Build and start Jarvis:

   ~~~bash
make build
make restart
   curl -fsS "$BASE/healthz" | jq .
   ~~~

   Expected: the process remains running and health checks do not expose
   credentials or raw provider errors.

## 2. Obtain platform and tenant tokens

1. Log in as the platform administrator:

   ~~~bash
   export PLATFORM_TOKEN=$(curl -fsS -X POST "$BASE/api/auth/login" \
     -H 'Content-Type: application/json' \
     -d "{\"email\":\"$PLATFORM_EMAIL\",\"password\":\"$PLATFORM_PASSWORD\"}" \
     | jq -r .access_token)
   test -n "$PLATFORM_TOKEN" -a "$PLATFORM_TOKEN" != null
   ~~~

2. Log in as Tenant A and Tenant B:

   ~~~bash
   export TENANT_A_TOKEN=$(curl -fsS -X POST "$BASE/api/auth/login" \
     -H 'Content-Type: application/json' \
     -d "{\"email\":\"$TENANT_A_EMAIL\",\"password\":\"$TENANT_A_PASSWORD\"}" \
     | jq -r .access_token)
   export TENANT_B_TOKEN=$(curl -fsS -X POST "$BASE/api/auth/login" \
     -H 'Content-Type: application/json' \
     -d "{\"email\":\"$TENANT_B_EMAIL\",\"password\":\"$TENANT_B_PASSWORD\"}" \
     | jq -r .access_token)
   test -n "$TENANT_A_TOKEN" -a "$TENANT_A_TOKEN" != null
   test -n "$TENANT_B_TOKEN" -a "$TENANT_B_TOKEN" != null
   ~~~

3. Verify role and tenant isolation:

   ~~~bash
   curl -sS -H "Authorization: Bearer $TENANT_A_TOKEN" "$BASE/api/auth/me" | jq .
   curl -sS -H "Authorization: Bearer $TENANT_B_TOKEN" "$BASE/api/auth/me" | jq .
   curl -si -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/platform/packages" | head
   ~~~

   Expected: each tenant token identifies its own tenant; a tenant token
   cannot access platform-admin endpoints and receives 401 or 403.

## 3. Verify all four package tiers

The seeded Sprint 45 packages and expected limits are:

| Package | ID | Monthly web minutes | Mobile minutes | AI employees | Storage |
| --- | --- | ---: | ---: | ---: | ---: |
| AiaaS ฿500 | pkg-aiaas-500 | 100 | 100 | 1 | 5 GiB |
| AiaaS ฿1,000 | pkg-aiaas-1000 | 300 | 300 | 3 | 20 GiB |
| AiaaS ฿1,500 | pkg-aiaas-1500 | 750 | 750 | 5 | 50 GiB |
| AiaaS ฿2,000 | pkg-aiaas-2000 | 1,500 | 1,500 | 10 | 100 GiB |

1. Confirm the catalog:

   ~~~bash
   curl -fsS -H "Authorization: Bearer $PLATFORM_TOKEN" \
     "$BASE/api/platform/packages" | jq '.[] | {id,slug,name,status}'
   ~~~

2. Assign one tier at a time to Tenant A. Repeat this step for all four
   package IDs, recording the response after each assignment:

   ~~~bash
   export PACKAGE_ID=pkg-aiaas-500
   curl -fsS -X POST \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     -H 'Content-Type: application/json' \
     "$BASE/api/platform/tenants/$TENANT_A/entitlement" \
     -d "{\"package_id\":\"$PACKAGE_ID\"}" | jq .
   ~~~

3. Read the tenant snapshot after each assignment:

   ~~~bash
   curl -fsS -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/usage" | jq .
   ~~~

   Expected: package, period, limits, current dimensions, unit, source, and
   freshness match the assigned tier. Unavailable values are null or marked
   unavailable; they are not rendered as authoritative zeroes.

4. Assign a different tier to Tenant B and verify that Tenant A's response
   does not change:

   ~~~bash
   curl -fsS -X POST \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     -H 'Content-Type: application/json' \
     "$BASE/api/platform/tenants/$TENANT_B/entitlement" \
     -d '{"package_id":"pkg-aiaas-2000"}' | jq .
   curl -fsS -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/usage" | jq '.tenant_id,.package,.current_dimensions'
   ~~~

## 4. Verify current versus historical projections

1. Query both tenants' current usage:

   ~~~bash
   curl -fsS -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/usage" > /tmp/s45-tenant-a-usage.json
   curl -fsS -H "Authorization: Bearer $TENANT_B_TOKEN" \
     "$BASE/api/tenant/usage" > /tmp/s45-tenant-b-usage.json
   jq '{tenant_id,period,usage,current_dimensions,historical_usage}' \
     /tmp/s45-tenant-a-usage.json
   jq '{tenant_id,period,usage,current_dimensions,historical_usage}' \
     /tmp/s45-tenant-b-usage.json
   ~~~

   Expected: current enforcement counters remain separate from
   historical_usage projections. A source outage is reported as unavailable
   or stale rather than merged into Redis counters.

2. Query the platform view for a bounded period:

   ~~~bash
   curl -fsS -H "Authorization: Bearer $PLATFORM_TOKEN" \
     "$BASE/api/platform/billing/usage?start_date=$START_DATE&end_date=$END_DATE" \
     > /tmp/s45-platform-usage.json
   jq '{range,freshness,reconciliation,tenants}' /tmp/s45-platform-usage.json
   ~~~

   Expected: platform rows are tenant-scoped and include usage projection
   status. Tenant A data does not appear under Tenant B.

## 5. Verify idempotent usage ledger events

Use a transaction so the fixture leaves no manual test rows behind.

1. Confirm the tenant IDs exist, then run the duplicate replay fixture. Replace
   the two IDs if your environment uses different tenants:

   ~~~bash
   export LEDGER_TENANT_A=$TENANT_A
   export LEDGER_TENANT_B=$TENANT_B
   psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 <<SQL
   BEGIN;
   INSERT INTO callcenter.usage_events
     (id, tenant_id, idempotency_key, dimension, unit, amount,
      period_start, period_end, source_type, source_id,
      entitlement_snapshot_id, created_by, updated_by)
   VALUES
     ('ue-s45-a', '$LEDGER_TENANT_A', 's45-replay-01',
      'monthly_call_minutes', 'minutes', 2, CURRENT_DATE, CURRENT_DATE,
      'manual', 's45-call-a', 'snapshot-a', 'uat', 'uat')
   ON CONFLICT (tenant_id, idempotency_key) DO NOTHING;
   INSERT INTO callcenter.usage_events
     (id, tenant_id, idempotency_key, dimension, unit, amount,
      period_start, period_end, source_type, source_id,
      entitlement_snapshot_id, created_by, updated_by)
   VALUES
     ('ue-s45-a-duplicate', '$LEDGER_TENANT_A', 's45-replay-01',
      'monthly_call_minutes', 'minutes', 99, CURRENT_DATE, CURRENT_DATE,
      'manual', 's45-call-a', 'snapshot-other', 'uat', 'uat')
   ON CONFLICT (tenant_id, idempotency_key) DO NOTHING;
   INSERT INTO callcenter.usage_events
     (id, tenant_id, idempotency_key, dimension, unit, amount,
      period_start, period_end, source_type, source_id,
      entitlement_snapshot_id, created_by, updated_by)
   VALUES
     ('ue-s45-b', '$LEDGER_TENANT_B', 's45-replay-01',
      'mobile_call_minutes', 'minutes', 7, CURRENT_DATE, CURRENT_DATE,
      'manual', 's45-call-b', 'snapshot-b', 'uat', 'uat')
   ON CONFLICT (tenant_id, idempotency_key) DO NOTHING;
   SELECT tenant_id, idempotency_key, COUNT(*) AS rows, SUM(amount) AS amount
     FROM callcenter.usage_events
    WHERE idempotency_key = 's45-replay-01'
    GROUP BY tenant_id, idempotency_key
    ORDER BY tenant_id;
   ROLLBACK;
   SQL
   ~~~

   Expected inside the transaction: Tenant A has one row totaling 2 and
   Tenant B has one row totaling 7. The duplicate does not replace the
   original amount or snapshot. After ROLLBACK, the query returns no rows.

2. Verify the database has no manual fixture residue:

   ~~~bash
   psql "$POSTGRES_URL" -At -c \
     "SELECT COUNT(*) FROM callcenter.usage_events WHERE source_id LIKE 's45-%';"
   ~~~

   Expected: 0.

## 6. Verify reconciliation API and bounded runs

1. Start a dry-run reconciliation:

   ~~~bash
   export RECON_KEY="s45-reconcile-$(date -u +%Y%m%d%H%M%S)"
   curl -fsS -X POST "$BASE/api/platform/usage/reconcile" \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     -H 'Content-Type: application/json' \
     -d "{\"start_date\":\"$START_DATE\",\"end_date\":\"$END_DATE\",\"tenant_id\":\"$TENANT_A\",\"dry_run\":true,\"idempotency_key\":\"$RECON_KEY\"}" \
     | tee /tmp/s45-reconciliation.json | jq .
   export RUN_ID=$(jq -r .run_id /tmp/s45-reconciliation.json)
   ~~~

   Expected: HTTP 202 with a run ID and queued or running status.

2. Poll the run:

   ~~~bash
   for i in 1 2 3 4 5 6; do
     curl -fsS -H "Authorization: Bearer $PLATFORM_TOKEN" \
       "$BASE/api/platform/usage/reconcile/$RUN_ID" | tee /tmp/s45-run.json | jq .
     grep -Eq '"status":"(completed|failed)"' /tmp/s45-run.json && break
     sleep 1
   done
   ~~~

   Expected: completed runs expose bounded dates, source watermarks,
   mismatch_count, correction_count, and no raw provider payloads.

3. Replay the same request with the same idempotency key:

   ~~~bash
   curl -si -X POST "$BASE/api/platform/usage/reconcile" \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     -H 'Content-Type: application/json' \
     -d "{\"start_date\":\"$START_DATE\",\"end_date\":\"$END_DATE\",\"tenant_id\":\"$TENANT_A\",\"dry_run\":true,\"idempotency_key\":\"$RECON_KEY\"}"
   ~~~

   Expected: HTTP 200, the same run ID, and duplicate=true. A tenant-specific
   replay key used for Tenant B must create a separate run.

4. Verify invalid bounds and role protection:

   ~~~bash
   curl -si -X POST "$BASE/api/platform/usage/reconcile" \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     -H 'Content-Type: application/json' \
     -d '{"start_date":"2026-01-01","end_date":"2026-03-01","idempotency_key":"too-long"}'
   curl -si -X POST "$BASE/api/platform/usage/reconcile" \
     -H "Authorization: Bearer $TENANT_A_TOKEN" \
     -H 'Content-Type: application/json' \
     -d "{\"start_date\":\"$START_DATE\",\"end_date\":\"$END_DATE\",\"idempotency_key\":\"tenant-forbidden\"}"
   ~~~

   Expected: the first request returns 400 invalid_reconciliation_request;
   the second returns 401 or 403.

## 7. Verify mobile bootstrap and call idempotency

Use a customer access token belonging to each tenant. Set
CUSTOMER_A_TOKEN and CUSTOMER_B_TOKEN from the existing customer-auth fixture;
do not use a tenant-admin token as a customer token.

1. Read bootstrap for each customer:

   ~~~bash
   curl -fsS -H "Authorization: Bearer $CUSTOMER_A_TOKEN" \
     "$BASE/api/mobile/v1/bootstrap" | tee /tmp/s45-mobile-a-bootstrap.json | jq .
   curl -fsS -H "Authorization: Bearer $CUSTOMER_B_TOKEN" \
     "$BASE/api/mobile/v1/bootstrap" | tee /tmp/s45-mobile-b-bootstrap.json | jq .
   ~~~

   Expected: each response contains only that tenant's assigned avatars,
   locale/timezone, mobile allowance, dimension-specific limit, remaining
   value, source, and freshness.

2. Create a mobile call:

   ~~~bash
   export AVATAR_A=$(jq -r '.default_avatar_id' /tmp/s45-mobile-a-bootstrap.json)
   export MOBILE_KEY="s45-mobile-create-$(date -u +%s)"
   curl -si -X POST "$BASE/api/mobile/v1/calls" \
     -H "Authorization: Bearer $CUSTOMER_A_TOKEN" \
     -H "Content-Type: application/json" \
     -H "Idempotency-Key: $MOBILE_KEY" \
     -H "X-Monti-SDK-Version: 0.1.0" \
     -d "{\"avatar_id\":\"$AVATAR_A\",\"locale\":\"en\"}" \
     | tee /tmp/s45-mobile-create.txt
   export CALL_ID=$(awk '/^{/{p=1} p' /tmp/s45-mobile-create.txt | jq -r .call_id)
   ~~~

   Expected: HTTP 201 and a call ID.

3. Repeat the exact create request with the same idempotency key.

   Expected: the same call ID is returned; no second call reservation appears.

4. Read status, then end the call twice:

   ~~~bash
   curl -fsS -H "Authorization: Bearer $CUSTOMER_A_TOKEN" \
     "$BASE/api/mobile/v1/calls/$CALL_ID" | jq .
   curl -si -X POST "$BASE/api/mobile/v1/calls/$CALL_ID/end" \
     -H "Authorization: Bearer $CUSTOMER_A_TOKEN" \
     -H "Idempotency-Key: s45-mobile-end-01"
   curl -si -X POST "$BASE/api/mobile/v1/calls/$CALL_ID/end" \
     -H "Authorization: Bearer $CUSTOMER_A_TOKEN" \
     -H "Idempotency-Key: s45-mobile-end-01"
   ~~~

   Expected: end-call is idempotent and does not create a second usage event.

5. Try an unassigned avatar and Tenant B's call ID with Customer A:

   Expected: stable avatar/authorization/not-found error; no tenant or
   provider details leak.

## 8. Verify disconnect, timeout, retry, and concurrent behavior

1. Open the mobile WebSocket with websocat:

   ~~~bash
   websocat -H="Authorization: Bearer $CUSTOMER_A_TOKEN" \
     "$BASE/ws/mobile/v1/calls/$CALL_ID"
   ~~~

2. Disconnect the client without sending an end frame. Reconnect once, then
   explicitly end the call through the HTTP endpoint.

   Expected: no leaked concurrent lease; reconnect does not double-count the
   same call; the final call state is ended or safely unavailable.

3. For a controlled short-session usage event, keep a valid call open for at
   least 35 seconds, disconnect, and query the ledger by source ID:

   ~~~bash
   psql "$POSTGRES_URL" -c \
     "SELECT tenant_id,source_id,dimension,unit,amount,state,created_at
        FROM callcenter.usage_events
       WHERE source_id='$CALL_ID';"
   ~~~

   Expected: at most one logical event per call and dimension. Repeating the
   lifecycle request does not add another event.

4. Run concurrent creates for both tenants using unique keys:

   ~~~bash
   seq 1 10 | xargs -I{} -P5 sh -c \
     'curl -fsS -X POST "$BASE/api/mobile/v1/calls" \
       -H "Authorization: Bearer $CUSTOMER_A_TOKEN" \
       -H "Content-Type: application/json" \
       -H "Idempotency-Key: s45-load-a-{}" \
       -d "{\"avatar_id\":\"$AVATAR_A\"}" >/tmp/s45-a-{}.json'
   ~~~

   Repeat for Customer B with B's avatar and keys. Expected: responses never
   contain the other tenant's avatar, package, call, or usage data. Quota and
   concurrency errors are stable and do not mutate usage on rejected starts.

## 9. Verify unavailable-source behavior

Run each outage separately and restore the service immediately afterward.

1. Stop ClickHouse:

   ~~~bash
   docker compose -f infra/docker-compose.yml stop clickhouse
   curl -si -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/call-center/statistics"
   curl -si -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/usage"
   docker compose -f infra/docker-compose.yml start clickhouse
   ~~~

   Expected: analytics reports analytics_unavailable; tenant usage remains
   separate and does not convert missing history into an exact zero.

2. Stop Redis, check one quota read, then restore it:

   ~~~bash
   docker compose -f infra/docker-compose.yml stop redis
   curl -si -H "Authorization: Bearer $TENANT_A_TOKEN" "$BASE/api/tenant/usage"
   docker compose -f infra/docker-compose.yml start redis
   ~~~

   Expected: quota_unavailable or the configured safe fail-open status is
   returned; no raw connection string or secret is exposed.

## 10. Evidence and completion decision

Save these artifacts:

- health and infrastructure output;
- package catalog and four-tier assignment responses;
- Tenant A/B usage responses;
- platform billing/projection response;
- reconciliation create, duplicate, and final status responses;
- mobile bootstrap/create/end responses;
- WebSocket disconnect/reconnect notes;
- SQL ledger query showing one logical event per call/dimension;
- concurrent-load output and any stable quota errors;
- outage responses and service restore output.

Mark a section [x] only when its expected result is observed. Sprint 45 is
ready to close only when all sections pass, Docker-backed UAT is complete,
and the remaining task documents are updated to completed. Otherwise leave
this document and TASK-0166/TASK-0167/TASK-0168 in_progress with the failed
step, evidence path, and follow-up noted.

## Automated evidence already recorded

## Execution evidence — 2026-07-25

- [x] Local Docker infrastructure, migration 029, Postgres ledger tables,
  Redis DB 4, ClickHouse, and MinIO health checks passed.
- [x] All four AiaaS catalog rows returned the expected price, web/mobile
  minutes, storage, KM, avatar, and concurrency dimensions.
- [x] Tenant and platform usage projections passed with current versus
  historical separation and two-tenant isolation.
- [x] Duplicate ledger fixture produced one Tenant A row totaling 2 and one
  Tenant B row totaling 7, then cleaned all fixture rows.
- [x] Reconciliation returned completed status, bounded watermarks, zero
  mismatches, correction counters, duplicate replay, and tenant-admin 403.
- [x] Mobile bootstrap, same-tenant replay, cross-tenant idempotency scoping,
  end replay, concurrent starts, WebSocket disconnect, and cleanup passed.
- [x] ClickHouse outage returned `503 usage_unavailable`; Redis outage kept
  safe fail-open behavior; both services were restored healthy.

- [x] Four AiaaS seed definitions and rules-v2 validation pass.
- [x] Separate web/mobile minute counters pass focused tests.
- [x] Snapshot rows expose source/freshness and null unavailable usage.
- [x] Mobile SDK build, make test, make build, and git diff --check pass.
- [x] Migration 029 creates usage ledger/reconciliation tables.
- [x] Transactional duplicate-safe smoke fixture rolled back cleanly.
