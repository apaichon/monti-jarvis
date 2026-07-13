# SPRINT-019 — Manual Test Checklist

**Feature:** FEAT-0021 · **Tasks:** TASK-0087–0091 · **Target:** v2.0.0

## 0. Preconditions

- [x] Docker, Go, Node/npm, `curl`, `jq`, and `psql` are available.
- [x] `infra/.env.dev` contains valid Postgres/Redis settings and `JWT_SECRET`.
- [x] Two active tenants exist with tenant-admin logins: Tenant A and Tenant B.
- [x] Tenant A has an active tier `vip` and group `retail`; Tenant B has separate catalogs.
- [x] No real customer PII is used; use the sample identities below.

Export access tokens obtained from tenant login:

```bash
export TOKEN_A='<tenant-a-access-token>'
export TOKEN_B='<tenant-b-access-token>'
export BASE=http://localhost:8091
```

## 1. Initialize and verify

```bash
make infra-up
make restart
make infra-check
curl -fsS "$BASE/healthz" | jq .
```

- [x] Server and shared infrastructure are healthy.
- [x] `/tenant/customers` loads after Tenant A signs in.
- [x] Customers appears in desktop navigation and mobile bottom navigation.

## 2. Prepare CSV

Create `/tmp/monti-s19-customers.csv`:

```csv
display_name,email,phone,locale,tier_slug,group_slugs,source,external_id
Jane Example,jane.s19@example.com,+66810000001,en,vip,retail,csv,crm-001
Somchai Example,somchai.s19@example.com,+66810000002,th,,,csv,crm-002
Invalid Email,not-an-email,,en,,,,bad-001
```

## 3. Scenarios

### S1 — Manual customer CRUD and deactivation (TASK-0088/0089 · FEAT AC 1)

1. Tenant A opens `http://localhost:8091/tenant/customers`.
2. Add `Manual Customer`, `manual.s19@example.com`, tier `VIP`, group `Retail`.
3. Edit the name and locale, then save.
4. Search by normalized/case-insensitive email.
5. Deactivate the customer.

Expected:

- [x] Create returns one active customer; edit preserves its id.
- [x] Tier/group labels appear in the directory.
- [x] Search finds the record regardless of email case.
- [x] Deactivate changes status to `inactive`; no row is physically exposed as deleted.

### S2 — CSV dry-run performs no customer writes (TASK-0088/0089 · FEAT AC 2)

```bash
BEFORE=$(curl -fsS -H "Authorization: Bearer $TOKEN_A" "$BASE/api/tenant/customers?status=" | jq '.customers|length')
curl -fsS -X POST -H "Authorization: Bearer $TOKEN_A" \
  -F dry_run=true -F file=@/tmp/monti-s19-customers.csv \
  "$BASE/api/tenant/customer-imports" | tee /tmp/s19-dry.json | jq .
AFTER=$(curl -fsS -H "Authorization: Bearer $TOKEN_A" "$BASE/api/tenant/customers?status=" | jq '.customers|length')
test "$BEFORE" = "$AFTER"
```

Expected:

- [x] Status is `validated`, total is 3, accepted is 2, rejected is 1.
- [x] Row error identifies the invalid email.
- [x] Customer count is unchanged.
- [x] UI commit button becomes enabled only for the validated file.

### S3 — CSV commit and idempotent reimport (TASK-0088/0090 · FEAT AC 3)

```bash
for run in 1 2; do
  curl -fsS -X POST -H "Authorization: Bearer $TOKEN_A" \
    -F dry_run=false -F file=@/tmp/monti-s19-customers.csv \
    "$BASE/api/tenant/customer-imports" | jq .
done
curl -fsS -H "Authorization: Bearer $TOKEN_A" \
  "$BASE/api/tenant/customers?q=crm-001&status=" | jq .
```

Expected:

- [x] First commit creates two valid customers and rejects one row.
- [x] Second commit reports updates, not duplicate creates.
- [x] Exactly one Tenant A customer has external id `crm-001` and its id is stable.

### S4 — Explicit assignment and domain-default precedence (TASK-0090 · FEAT AC 4–5)

1. Create Tenant A domain rule `example.com`, policy `allow`, with a default tier/group different from `VIP/Retail`.
2. Import the sample CSV where Jane explicitly names `vip/retail` and Somchai has no assignment.

Expected:

- [x] Jane retains explicit VIP/Retail assignment.
- [x] Somchai receives the matching domain defaults.
- [x] The UI explains that allow/deny enforcement is deferred to SPRINT-020.
- [x] Duplicate `example.com` rule returns `409 domain_rule_exists`.

### S5 — Tenant isolation (TASK-0087/0088/0090 · FEAT AC 6)

1. Capture a Tenant A customer id, import id, and domain-rule id.
2. Request or mutate each id with `TOKEN_B`.

```bash
curl -i -H "Authorization: Bearer $TOKEN_B" "$BASE/api/tenant/customers/<A_CUSTOMER_ID>"
curl -i -H "Authorization: Bearer $TOKEN_B" "$BASE/api/tenant/customer-imports/<A_IMPORT_ID>"
curl -i -X PUT -H "Authorization: Bearer $TOKEN_B" -H 'Content-Type: application/json' \
  -d '{"domain":"example.com","policy":"allow"}' \
  "$BASE/api/tenant/customer-domain-rules/<A_RULE_ID>"
```

Expected:

- [x] All three cross-tenant operations return 404 without revealing Tenant A data.
- [x] Tenant B list endpoints contain no Tenant A customers, imports, or domain rules.

### S6 — Import limits and malformed input (TASK-0088 · FEAT AC 2)

- [x] Missing file returns `400 import_invalid`.
- [x] Unknown/duplicate headers return `400 import_invalid`.
- [x] A file above `CUSTOMER_IMPORT_MAX_BYTES` returns 413.
- [x] Rows above `CUSTOMER_IMPORT_MAX_ROWS` are rejected.
- [x] Invalid tier/group slugs appear as row errors before commit.

### S7 — Customer authentication remains unavailable (TASK-0091 · FEAT AC 7)

- [x] Public customer conversation portal still works without customer login.
- [x] No customer password, OAuth, invitation, verification, or JWT endpoint was added.
- [x] Domain `allow`/`deny` policy does not block the public portal in SPRINT-019.

### S8 — Automated/build regression

```bash
GOFLAGS=-vet=off make test
npm --prefix apps/tenant-web run check
make build
```

- [x] Go tests pass.
- [x] Tenant Svelte check reports zero errors/warnings.
- [x] All three web applications and the Go binary build.
- [x] Existing platform-admin accessibility warnings, if present, are recorded but do not fail the build.

## 4. Teardown

```bash
make down
rm -f /tmp/monti-s19-customers.csv /tmp/s19-dry.json
```

## 5. Sign-off

Executed locally on 2026-07-13 against Tenant A `demo` and isolated Tenant B `uat-s19-b`. Scripted curl UAT verified customer CRUD/deactivation, dry-run and commit summaries, idempotent re-import, explicit-vs-domain assignment precedence, duplicate-domain conflict, and cross-tenant 404 behavior. Full Go tests, tenant checks, all web builds, Go build, infrastructure health, and runtime smoke passed. Parser/handler automated tests cover malformed headers, byte/row limits, and invalid catalog slugs.

| Tester | Date | Result | Defects |
| --- | --- | --- | --- |
| Codex release verification | 2026-07-13 | ☑ Pass | None |

Any failed checkbox requires a defect task; SPRINT-019 must remain open until the defect is resolved and the scenario reruns successfully.
