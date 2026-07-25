---
id: UAT-SPRINT-046-TASK-0169
status: deferred
updated: 2026-07-25
sprint: SPRINT-046
task: TASK-0169
owner: tester
---

# Sprint 46 referral foundation — manual UAT

> **Deferred at sprint close:** test later. This checklist remains the required
> manual verification before the referral foundation is treated as fully UAT
> accepted.

This runbook verifies the Sprint 46 referral attribution and qualification
foundation. It covers referral code creation, transactional signup attribution,
qualification gates, idempotency, tenant isolation, and the append-only status
event trail.

Bonus quota grants, expiry, consumption, reversal, and affiliate reporting are
not implemented in this slice and must not be marked as passed by this UAT.

Run from the repository root against a local fixture only. Do not use real
tenant IDs, customer data, or production payment credentials.

## 0. Test variables and prerequisites

Required: Docker Compose, Go, `curl`, `jq`, `psql`, and `make`.

Use authenticated mode so the tenant and platform authorization guards are
exercised:

```bash
set -a
source infra/.env.dev
set +a

export BASE=http://localhost:8091
export POSTGRES_URL="${POSTGRES_URL:-postgres://postgres:postgres@localhost:5432/monti_jarvis?sslmode=disable}"
export TENANT_A=demo
export PLATFORM_EMAIL=platform@monti.local
export PLATFORM_PASSWORD=monti-platform
export TENANT_A_EMAIL=admin@demo.local
export TENANT_A_PASSWORD=demo-admin
export TEMP_SLUG="referral-uat-$(date -u +%Y%m%d%H%M%S)"
export TEMP_EMAIL="${TEMP_SLUG}@example.test"
export TEMP_PASSWORD='Referral-UAT-123!'
```

Before starting, record the tester, date, commit under test, and environment.
Confirm `AUTH_DISABLED=false` and a valid local `JWT_SECRET` are set in
`infra/.env.dev`.

## 1. Start the local fixture and apply the migration

1. Start dependencies and initialize the local data stores:

   ```bash
   make infra-up
   make infra-check
   scripts/migrate.sh
   ```

   Expected: infrastructure checks pass and the migration ends with
   `migrations applied`.

2. Build and start Jarvis:

   ```bash
   make build
   make restart
   curl -fsS "$BASE/healthz" | jq .
   ```

   Expected: the server is healthy and remains running.

3. Verify the referral tables exist:

   ```bash
   psql "$POSTGRES_URL" -At -c \
     "SELECT to_regclass('callcenter.tenant_referral_codes'),
             to_regclass('callcenter.tenant_referrals'),
             to_regclass('callcenter.tenant_referral_events');"
   ```

   Expected: all three relations are returned.

## 2. Obtain admin tokens and verify roles

```bash
export PLATFORM_TOKEN=$(curl -fsS -X POST "$BASE/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$PLATFORM_EMAIL\",\"password\":\"$PLATFORM_PASSWORD\"}" \
  | jq -r .access_token)

export TENANT_A_TOKEN=$(curl -fsS -X POST "$BASE/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$TENANT_A_EMAIL\",\"password\":\"$TENANT_A_PASSWORD\"}" \
  | jq -r .access_token)

test -n "$PLATFORM_TOKEN" -a "$PLATFORM_TOKEN" != null
test -n "$TENANT_A_TOKEN" -a "$TENANT_A_TOKEN" != null

curl -fsS -H "Authorization: Bearer $PLATFORM_TOKEN" \
  "$BASE/api/auth/me" | jq '{role,tenant_id}'
curl -fsS -H "Authorization: Bearer $TENANT_A_TOKEN" \
  "$BASE/api/auth/me" | jq '{role,tenant_id}'
```

Expected: the first response is `platform_admin`; the second is
`tenant_admin` for `demo`.

## 3. Create and verify Tenant A's referral code

1. Request the tenant code:

   ```bash
   curl -fsS -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/referral" | tee /tmp/s46-referral-code.json | jq .
   export REFERRAL_CODE=$(jq -r .code /tmp/s46-referral-code.json)
   export REFERRAL_CODE_ID=$(jq -r .id /tmp/s46-referral-code.json)
   test -n "$REFERRAL_CODE" -a "$REFERRAL_CODE" != null
   ```

   Expected: HTTP 200, `tenant_id: demo`, `status: active`, and a non-empty
   stable code.

2. Request it again:

   ```bash
   curl -fsS -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/referral" | jq .
   ```

   Expected: the code and code ID are unchanged; no second code row is
   created.

3. Confirm tenant scoping. Without a Bearer token, the endpoint must reject
   the request:

   ```bash
   curl -si "$BASE/api/tenant/referral" | head -n 1
   ```

   Expected: HTTP 401 or 403.

## 4. Attribute a new tenant during signup

1. Register a temporary tenant using Tenant A's code:

   ```bash
   curl -sS -i -X POST "$BASE/api/public/tenant/register" \
     -H 'Content-Type: application/json' \
     -d "{\"company_name\":\"Referral UAT\",\"slug\":\"$TEMP_SLUG\",\"admin_email\":\"$TEMP_EMAIL\",\"admin_password\":\"$TEMP_PASSWORD\",\"admin_display_name\":\"Referral Tester\",\"referral_code\":\"$REFERRAL_CODE\"}" \
     | tee /tmp/s46-registration-response.txt
   ```

   Expected: HTTP 201. Save the response body and set the temporary tenant ID:

   ```bash
   export TEMP_TENANT_ID=$(sed -n '/^{/,$p' /tmp/s46-registration-response.txt | jq -r .tenant_id)
   test "$TEMP_TENANT_ID" = "$TEMP_SLUG"
   ```

2. Check Tenant A's referral list:

   ```bash
   curl -fsS -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/referrals" | tee /tmp/s46-referrals.json | jq .
   export REFERRAL_ID=$(jq -r '.referrals[0].id' /tmp/s46-referrals.json)
   test -n "$REFERRAL_ID" -a "$REFERRAL_ID" != null
   ```

   Expected: one row for `TEMP_TENANT_ID` with:

   - `referrer_tenant_id: demo`
   - `referred_tenant_id: TEMP_TENANT_ID`
   - `status: attributed`
   - `source: tenant_registration`

3. Verify the database record and initial immutable event:

   ```bash
   psql "$POSTGRES_URL" -c \
     "SELECT id, referrer_tenant_id, referred_tenant_id, code, status, source
        FROM callcenter.tenant_referrals WHERE id='$REFERRAL_ID';"
   psql "$POSTGRES_URL" -c \
     "SELECT from_status, to_status, reason, source, created_by
        FROM callcenter.tenant_referral_events
       WHERE referral_id='$REFERRAL_ID' ORDER BY event_at;"
   ```

   Expected: one `'' → attributed` event exists and no quota or entitlement
   row was created by referral attribution.

## 5. Verify invalid and duplicate signup behavior

1. Try an unknown referral code:

   ```bash
   export BAD_SLUG="${TEMP_SLUG}-bad"
   curl -si -X POST "$BASE/api/public/tenant/register" \
     -H 'Content-Type: application/json' \
     -d "{\"company_name\":\"Invalid Referral\",\"slug\":\"$BAD_SLUG\",\"admin_email\":\"$BAD_SLUG@example.test\",\"admin_password\":\"$TEMP_PASSWORD\",\"admin_display_name\":\"Tester\",\"referral_code\":\"not-a-real-code\"}" \
     | head -n 1
   ```

   Expected: HTTP 400. Confirm the invalid tenant was not inserted:

   ```bash
   psql "$POSTGRES_URL" -At -c \
     "SELECT COUNT(*) FROM callcenter.tenants WHERE id='$BAD_SLUG';"
   ```

   Expected: `0`.

2. Retry the original registration with the same slug:

   ```bash
   curl -si -X POST "$BASE/api/public/tenant/register" \
     -H 'Content-Type: application/json' \
     -d "{\"company_name\":\"Referral UAT Retry\",\"slug\":\"$TEMP_SLUG\",\"admin_email\":\"retry-$TEMP_EMAIL\",\"admin_password\":\"$TEMP_PASSWORD\",\"admin_display_name\":\"Retry\",\"referral_code\":\"$REFERRAL_CODE\"}" \
     | head -n 1
   ```

   Expected: HTTP 409 for the existing slug. Confirm the original referral
   count is still one and its code/referrer have not changed:

   ```bash
   psql "$POSTGRES_URL" -At -c \
     "SELECT COUNT(*) FROM callcenter.tenant_referrals WHERE referred_tenant_id='$TEMP_TENANT_ID';"
   ```

   Expected: `1`.

3. Optional disabled-code check. Disable the fixture code, attempt a new
   registration, and restore it immediately:

   ```bash
   psql "$POSTGRES_URL" -c \
     "UPDATE callcenter.tenant_referral_codes SET status='disabled' WHERE id='$REFERRAL_CODE_ID';"
   curl -si -X POST "$BASE/api/public/tenant/register" \
     -H 'Content-Type: application/json' \
     -d "{\"company_name\":\"Disabled Referral\",\"slug\":\"${TEMP_SLUG}-disabled\",\"admin_email\":\"disabled-$TEMP_EMAIL\",\"admin_password\":\"$TEMP_PASSWORD\",\"admin_display_name\":\"Tester\",\"referral_code\":\"$REFERRAL_CODE\"}" \
     | head -n 1
   psql "$POSTGRES_URL" -c \
     "UPDATE callcenter.tenant_referral_codes SET status='active' WHERE id='$REFERRAL_CODE_ID';"
   ```

   Expected: the disabled-code registration returns HTTP 400 and does not
   leave a tenant or referral row.

## 6. Verify qualification gates

1. Before activation, ask the platform to qualify the referral:

   ```bash
   curl -sS -i -X POST \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     "$BASE/api/platform/referrals/$REFERRAL_ID/qualify"
   ```

   Expected: HTTP 409 with `error: referral_not_qualified`,
   `status: pending`, and `qualification_reason: tenant_not_active`.

2. Submit a KYC fixture for the temporary tenant. This database shortcut is
   only for local UAT because the temporary signup has no mailbox verification
   step in this runbook:

   ```bash
   psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 \
     -v tenant_id="$TEMP_TENANT_ID" <<'SQL'
   INSERT INTO callcenter.tenant_kyc_profiles
     (tenant_id, status, created_by, updated_by)
   VALUES (:'tenant_id', 'submitted', 'manual-uat', 'manual-uat')
   ON CONFLICT (tenant_id) DO UPDATE SET
     status='submitted', rejection_reason='', reviewed_at=NULL,
     reviewed_by='', updated_by='manual-uat', updated_at=now();
   UPDATE callcenter.tenant_registrations
      SET status='submitted', updated_by='manual-uat', updated_at=now()
    WHERE tenant_id=:'tenant_id';
   SQL
   ```

3. Approve KYC as platform admin:

   ```bash
   curl -fsS -X POST \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     "$BASE/api/platform/tenants/$TEMP_TENANT_ID/kyc/approve" | jq .
   ```

   Expected: tenant status and KYC status are both `active`/`approved`.

4. Qualify again before any paid order:

   ```bash
   curl -sS -i -X POST \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     "$BASE/api/platform/referrals/$REFERRAL_ID/qualify"
   ```

   Expected: HTTP 409 with `qualification_reason: paid_order_required` and
   the referral remains `pending`.

5. Insert one paid, non-voided local fixture order:

   ```bash
   export PACKAGE_ID=$(psql "$POSTGRES_URL" -At -c \
     "SELECT id FROM callcenter.packages WHERE status='active' ORDER BY id LIMIT 1")
   export ORDER_ID="manual-referral-order-$(date -u +%s)"
   export ORDER_NO="MANUAL-REFERRAL-$(date -u +%s)"

   psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 \
     -v tenant_id="$TEMP_TENANT_ID" -v package_id="$PACKAGE_ID" \
     -v order_id="$ORDER_ID" -v order_no="$ORDER_NO" <<'SQL'
   INSERT INTO callcenter.payment_orders
     (id, tenant_id, package_id, order_no, amount_cents, currency, status,
      provider, payment_method, transaction_id, paid_at, created_by, updated_by)
   VALUES (:'order_id', :'tenant_id', :'package_id', :'order_no', 1000, '764',
           'paid', 'manual-uat', 'credit_card', 'manual-uat-txn', now(),
           'manual-uat', 'manual-uat');
   SQL
   ```

   Expected: the insert succeeds and the order is `paid`. Do not create a
   `payment_documents` row with `status='voided'` for this qualification case.

6. Qualify the referral:

   ```bash
   curl -fsS -X POST \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     "$BASE/api/platform/referrals/$REFERRAL_ID/qualify" | tee /tmp/s46-qualified.json | jq .
   ```

   Expected: HTTP 200, `status: qualified`, empty qualification reason, and a
   populated `qualified_at`.

7. Repeat qualification:

   ```bash
   curl -fsS -X POST \
     -H "Authorization: Bearer $PLATFORM_TOKEN" \
     "$BASE/api/platform/referrals/$REFERRAL_ID/qualify" | jq .
   psql "$POSTGRES_URL" -c \
     "SELECT from_status, to_status, reason, source
        FROM callcenter.tenant_referral_events
       WHERE referral_id='$REFERRAL_ID' ORDER BY event_at;"
   ```

   Expected: HTTP 200 with the same qualification result. No second
   qualification event is appended.

## 7. Verify authorization, isolation, and no quota mutation

1. A tenant token cannot qualify a referral:

   ```bash
   curl -si -X POST \
     -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/platform/referrals/$REFERRAL_ID/qualify" | head -n 1
   ```

   Expected: HTTP 401 or 403.

2. Confirm Tenant A sees only referrals it referred, and no endpoint exposes
   another tenant's private billing details:

   ```bash
   curl -fsS -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/referrals" | jq .
   psql "$POSTGRES_URL" -c \
     "SELECT referrer_tenant_id, referred_tenant_id, status
        FROM callcenter.tenant_referrals WHERE id='$REFERRAL_ID';"
   ```

   Expected: the tenant list contains only Tenant A's referral summary; the
   platform-only qualification endpoint is not tenant-accessible.

3. Verify qualification did not create or mutate base quota:

   ```bash
   psql "$POSTGRES_URL" -c \
     "SELECT id, tenant_id, package_id, status
        FROM callcenter.tenant_entitlements WHERE tenant_id='$TEMP_TENANT_ID';"
   psql "$POSTGRES_URL" -c \
     "SELECT COUNT(*) AS usage_events
        FROM callcenter.usage_events WHERE tenant_id='$TEMP_TENANT_ID';"
   ```

   Expected: referral qualification itself creates no entitlement and no usage
   event. Any pre-existing fixture entitlement/usage must be documented
   separately.

## 8. Cleanup and evidence

1. Save the evidence files and database output:

   ```bash
   curl -fsS -H "Authorization: Bearer $TENANT_A_TOKEN" \
     "$BASE/api/tenant/referrals" > /tmp/s46-referrals-final.json
   psql "$POSTGRES_URL" -c \
     "SELECT * FROM callcenter.tenant_referral_events WHERE referral_id='$REFERRAL_ID' ORDER BY event_at;" \
     > /tmp/s46-referral-events.txt
   ```

2. Remove the temporary fixture tenant and its referral data. This is
   destructive; verify the ID before running it:

   ```bash
   printf 'Temporary tenant to remove: %s\n' "$TEMP_TENANT_ID"
   psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -v tenant_id="$TEMP_TENANT_ID" <<'SQL'
   DELETE FROM callcenter.tenants WHERE id=:'tenant_id';
   SQL
   ```

   Expected: the temporary tenant, referral, referral events, KYC profile,
   orders, and tenant-owned rows are removed through their configured foreign
   key cascades. Confirm Tenant A's referral list no longer contains the test
   row.

3. Record the final result:

   | Area | Result | Evidence |
   | --- | --- | --- |
   | Migration/tables | [ ] Pass / [ ] Fail | |
   | Stable tenant code | [ ] Pass / [ ] Fail | |
   | Signup attribution | [ ] Pass / [ ] Fail | |
   | Invalid/disabled code | [ ] Pass / [ ] Fail | |
   | Qualification gates | [ ] Pass / [ ] Fail | |
   | Idempotent qualification | [ ] Pass / [ ] Fail | |
   | Tenant/platform isolation | [ ] Pass / [ ] Fail | |
   | No quota mutation | [ ] Pass / [ ] Fail | |

## Known test boundary

The foundation exposes attribution through public tenant registration and
qualification through a platform-admin endpoint. It does not expose a public
direct-attribution endpoint or a UI for fraud review, self-referral simulation,
circular-referral simulation, or bonus-ledger settlement. Those cases require
the automated store tests or a later affiliate-administration slice.
