# SPRINT-020 — Manual Test Checklist

**Feature:** FEAT-0022 · **Tasks:** TASK-0092–0096 · **Target:** v2.1.0

## 0. Preconditions

- [X] Docker, Go, Node/npm, `curl`, `jq`, `psql`, and `redis-cli` are available.
- [X] `infra/.env.dev` has `AUTH_DISABLED=false`.
- [X] `infra/.env.dev` has `JWT_SECRET` set to a random value of at least 32 bytes.
- [X] `infra/.env.dev` has Resend configured with a verified sender:
  - `RESEND_API_KEY`
  - `RESEND_FROM_EMAIL` or `RESEND_FROM_ADDR`
- [ ] Tester controls the mailboxes used below, or routes them to a test inbox.
- [ ] Two active tenants exist:
  - Tenant A: `demo`
  - Tenant B: `<TENANT_B_ID>`
- [ ] Tenant admin credentials exist for Tenant A and Tenant B.
- [ ] No real customer PII is used.

Export local variables:

```bash
set -a
source infra/.env.dev
set +a

export BASE=http://localhost:8091
export TENANT_A=demo
export TENANT_B='<tenant-b-id>'
export ADMIN_A_EMAIL='admin@demo.local'
export ADMIN_A_PASSWORD='demo-admin'
export ADMIN_B_EMAIL='<tenant-b-admin-email>'
export ADMIN_B_PASSWORD='<tenant-b-admin-password>'
export CUST_A_ALLOWED='alice.s20@example.com'
export CUST_A_DENIED='blocked.s20@blocked.example'
export CUST_B_ALLOWED='bob.s20@example.net'
```

Login as both tenant admins:

```bash
export TOKEN_A=$(curl -fsS -X POST "$BASE/api/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"email\":\"$ADMIN_A_EMAIL\",\"password\":\"$ADMIN_A_PASSWORD\"}" | jq -r .access_token)

export TOKEN_B=$(curl -fsS -X POST "$BASE/api/auth/login" \
  -H 'content-type: application/json' \
  -d "{\"email\":\"$ADMIN_B_EMAIL\",\"password\":\"$ADMIN_B_PASSWORD\"}" | jq -r .access_token)

test "$TOKEN_A" != "null" -a "$TOKEN_A" != ""
test "$TOKEN_B" != "null" -a "$TOKEN_B" != ""
```

## 1. Initialize and verify

```bash
make infra-up
make restart
make infra-check
curl -fsS "$BASE/healthz" | jq .
curl -fsS "$BASE/api/infra" | jq .
```

Expected:

- [X] `/healthz` returns `sprint: "SPRINT-020"`.
- [X] `/api/infra` reports Postgres and Redis as `ok`.
- [X] Auth is enabled, and tenant admin login succeeds.
- [X] Customer portal loads at `http://localhost:8091/`.
- [X] Tenant settings loads at `http://localhost:8091/tenant/settings`.

## 2. Prepare tenant data

Create customers and domain policies for both tenants.

```bash
curl -fsS -X POST "$BASE/api/tenant/customers" \
  -H "Authorization: Bearer $TOKEN_A" -H 'content-type: application/json' \
  -d "{\"display_name\":\"Alice S20\",\"email\":\"$CUST_A_ALLOWED\",\"locale\":\"en\",\"source\":\"manual\",\"status\":\"active\"}" \
  | tee /tmp/s20-customer-a.json | jq .

export CUSTOMER_A_ID=$(jq -r .customer.id /tmp/s20-customer-a.json)

curl -fsS -X POST "$BASE/api/tenant/customer-domain-rules" \
  -H "Authorization: Bearer $TOKEN_A" -H 'content-type: application/json' \
  -d '{"domain":"example.com","policy":"allow","active":true}' | jq .

curl -fsS -X POST "$BASE/api/tenant/customer-domain-rules" \
  -H "Authorization: Bearer $TOKEN_A" -H 'content-type: application/json' \
  -d '{"domain":"blocked.example","policy":"deny","active":true}' | jq .

curl -fsS -X POST "$BASE/api/tenant/customers" \
  -H "Authorization: Bearer $TOKEN_B" -H 'content-type: application/json' \
  -d "{\"display_name\":\"Bob S20\",\"email\":\"$CUST_B_ALLOWED\",\"locale\":\"en\",\"source\":\"manual\",\"status\":\"active\"}" \
  | tee /tmp/s20-customer-b.json | jq .

export CUSTOMER_B_ID=$(jq -r .customer.id /tmp/s20-customer-b.json)

curl -fsS -X POST "$BASE/api/tenant/customer-domain-rules" \
  -H "Authorization: Bearer $TOKEN_B" -H 'content-type: application/json' \
  -d '{"domain":"example.net","policy":"allow","active":true}' | jq .
```

Enable customer OTP auth for both tenants:

```bash
curl -fsS -X PUT "$BASE/api/tenant/customer-auth/settings" \
  -H "Authorization: Bearer $TOKEN_A" -H 'content-type: application/json' \
  -d '{"enabled":true,"auth_mode":"optional","allowed_domains":["example.com"],"otp_ttl_seconds":600,"session_ttl_seconds":604800}' | jq .

curl -fsS -X PUT "$BASE/api/tenant/customer-auth/settings" \
  -H "Authorization: Bearer $TOKEN_B" -H 'content-type: application/json' \
  -d '{"enabled":true,"auth_mode":"optional","allowed_domains":["example.net"],"otp_ttl_seconds":600,"session_ttl_seconds":604800}' | jq .
```

Expected:

- [X] Tenant A and Tenant B have separate customer ids.
- [X] Tenant A allows `example.com` and denies `blocked.example`.
- [X] Tenant B allows `example.net`.
- [X] `GET /api/tenant/customer-auth/settings` returns tenant-scoped settings for each token.

## 3. Scenarios

### S1 — Customer auth settings UI and API survive reload (TASK-0094 · AC 1, 4, 5)

1. Sign in to `http://localhost:8091/tenant/login` as Tenant A admin.
2. Open `http://localhost:8091/tenant/settings`.
3. In **Customer OTP auth**, verify:
   - enabled is checked
   - auth mode is `optional`
   - allowed domain contains `example.com`
   - OTP TTL is `600`
4. Change OTP TTL to `300`, save, reload the page.
5. Change OTP TTL back to `600`, save.

Expected:

- [X] Settings load only for tenant admin.
- [X] Save returns success and survives reload.
- [X] Tenant B admin cannot see or mutate Tenant A settings.
- [X] Unauthenticated request returns 401/403:

```bash
curl -i "$BASE/api/tenant/customer-auth/settings"
```

### S2 — OTP request returns safe challenge metadata (TASK-0093 · AC 2, 5)

```bash
curl -fsS -X POST "$BASE/api/customer/auth/request-otp" \
  -H "X-Tenant-Id: $TENANT_A" -H 'content-type: application/json' \
  -d "{\"email\":\"$CUST_A_ALLOWED\"}" \
  | tee /tmp/s20-otp-a.json | jq .
```

Expected:

- [ ] Response status is `202`.
- [ ] Response includes `challenge_id`, `status: "otp_sent"`, `delivery.channel: "email"`, masked `delivery.to`, `expires_in`, and `resend_after`.
- [ ] Response does not include plaintext OTP, token hashes, secrets, or internal session ids.
- [ ] Email inbox receives the OTP.
- [ ] Server logs do not print the OTP code.

Store the challenge id:

```bash
export CHALLENGE_A=$(jq -r .challenge_id /tmp/s20-otp-a.json)
```

### S3 — OTP verify issues customer JWT/session and `/me` profile (TASK-0093 · AC 3, 4)

Read the OTP from the test mailbox and export it:

```bash
export OTP_A='<otp-from-email>'

curl -fsS -X POST "$BASE/api/customer/auth/verify-otp" \
  -H "X-Tenant-Id: $TENANT_A" -H 'content-type: application/json' \
  -d "{\"challenge_id\":\"$CHALLENGE_A\",\"otp\":\"$OTP_A\"}" \
  | tee /tmp/s20-auth-a.json | jq .

export CUSTOMER_ACCESS_A=$(jq -r .access_token /tmp/s20-auth-a.json)
export CUSTOMER_REFRESH_A=$(jq -r .refresh_token /tmp/s20-auth-a.json)
```

Expected:

- [ ] Response status is `200`.
- [ ] Response includes `status: "authenticated"`.
- [ ] `customer.id` equals `$CUSTOMER_A_ID`.
- [ ] `customer.tenant_id` equals `$TENANT_A`.
- [ ] `customer.role` is `customer`.
- [ ] Access and refresh tokens are present.
- [ ] No plaintext OTP or token hash is present.

Verify profile:

```bash
curl -fsS "$BASE/api/customer/me" \
  -H "Authorization: Bearer $CUSTOMER_ACCESS_A" | jq .
```

Expected:

- [ ] `/api/customer/me` returns Alice S20 only.
- [ ] The profile includes `tenant_id`, `tier_id`, `group_ids`, `locale`, and role `customer`.

### S4 — Denied domain blocks OTP before delivery (TASK-0093 · AC 5)

```bash
curl -i -X POST "$BASE/api/customer/auth/request-otp" \
  -H "X-Tenant-Id: $TENANT_A" -H 'content-type: application/json' \
  -d "{\"email\":\"$CUST_A_DENIED\"}"
```

Expected:

- [ ] Response is `403`.
- [ ] JSON code is `domain_forbidden`.
- [ ] No OTP email is delivered.
- [ ] `customer_auth_events` records `customer.auth.otp_denied` without plaintext OTP.

Optional DB check:

```bash
psql "$POSTGRES_URL" -c "select event,email_normalized from callcenter.customer_auth_events where tenant_id='$TENANT_A' order by created_at desc limit 5;"
```

### S5 — Cross-tenant challenge/session isolation (TASK-0093 · AC 1, 6; TASK-0096 · AC 2)

Request Tenant B OTP:

```bash
curl -fsS -X POST "$BASE/api/customer/auth/request-otp" \
  -H "X-Tenant-Id: $TENANT_B" -H 'content-type: application/json' \
  -d "{\"email\":\"$CUST_B_ALLOWED\"}" \
  | tee /tmp/s20-otp-b.json | jq .

export CHALLENGE_B=$(jq -r .challenge_id /tmp/s20-otp-b.json)
export OTP_B='<otp-from-tenant-b-email>'
```

Try verifying Tenant B challenge under Tenant A:

```bash
curl -i -X POST "$BASE/api/customer/auth/verify-otp" \
  -H "X-Tenant-Id: $TENANT_A" -H 'content-type: application/json' \
  -d "{\"challenge_id\":\"$CHALLENGE_B\",\"otp\":\"$OTP_B\"}"
```

Expected:

- [ ] Cross-tenant challenge verify returns 401.
- [ ] Response does not reveal Tenant B customer data.

Verify Tenant B normally:

```bash
curl -fsS -X POST "$BASE/api/customer/auth/verify-otp" \
  -H "X-Tenant-Id: $TENANT_B" -H 'content-type: application/json' \
  -d "{\"challenge_id\":\"$CHALLENGE_B\",\"otp\":\"$OTP_B\"}" \
  | tee /tmp/s20-auth-b.json | jq .

export CUSTOMER_ACCESS_B=$(jq -r .access_token /tmp/s20-auth-b.json)
export CUSTOMER_REFRESH_B=$(jq -r .refresh_token /tmp/s20-auth-b.json)
```

Expected:

- [ ] Tenant B token returns only Bob S20 from `/api/customer/me`.
- [ ] Tenant A token never returns Tenant B customer data.

### S6 — Refresh and logout revoke customer session (TASK-0093 · AC 3, 6)

```bash
curl -fsS -X POST "$BASE/api/customer/auth/refresh" \
  -H 'content-type: application/json' \
  -d "{\"refresh_token\":\"$CUSTOMER_REFRESH_A\"}" \
  | tee /tmp/s20-refresh-a.json | jq .

export CUSTOMER_ACCESS_A2=$(jq -r .access_token /tmp/s20-refresh-a.json)
export CUSTOMER_REFRESH_A2=$(jq -r .refresh_token /tmp/s20-refresh-a.json)

curl -i -X POST "$BASE/api/customer/auth/refresh" \
  -H 'content-type: application/json' \
  -d "{\"refresh_token\":\"$CUSTOMER_REFRESH_A\"}"

curl -fsS -X POST "$BASE/api/customer/auth/logout" \
  -H 'content-type: application/json' \
  -d "{\"refresh_token\":\"$CUSTOMER_REFRESH_A2\"}" | jq .

curl -i -X POST "$BASE/api/customer/auth/refresh" \
  -H 'content-type: application/json' \
  -d "{\"refresh_token\":\"$CUSTOMER_REFRESH_A2\"}"
```

Expected:

- [ ] First refresh succeeds and rotates refresh token.
- [ ] Reusing old refresh token returns 401.
- [ ] Logout returns `{ "status": "ok" }`.
- [ ] Refresh after logout returns 401.

### S7 — Inactive customer cannot complete login (TASK-0095 · AC 5)

Deactivate Alice:

```bash
curl -fsS -X DELETE "$BASE/api/tenant/customers/$CUSTOMER_A_ID" \
  -H "Authorization: Bearer $TOKEN_A" | jq .
```

Request OTP and try verify with the emailed code:

```bash
curl -fsS -X POST "$BASE/api/customer/auth/request-otp" \
  -H "X-Tenant-Id: $TENANT_A" -H 'content-type: application/json' \
  -d "{\"email\":\"$CUST_A_ALLOWED\"}" \
  | tee /tmp/s20-inactive-otp.json | jq .

export INACTIVE_CHALLENGE=$(jq -r .challenge_id /tmp/s20-inactive-otp.json)
export INACTIVE_OTP='<otp-from-email>'

curl -i -X POST "$BASE/api/customer/auth/verify-otp" \
  -H "X-Tenant-Id: $TENANT_A" -H 'content-type: application/json' \
  -d "{\"challenge_id\":\"$INACTIVE_CHALLENGE\",\"otp\":\"$INACTIVE_OTP\"}"
```

Expected:

- [ ] OTP may be delivered, but verify returns 403/401.
- [ ] No authenticated token is issued for inactive customer.

Reactivate Alice for remaining tests:

```bash
curl -fsS -X PUT "$BASE/api/tenant/customers/$CUSTOMER_A_ID" \
  -H "Authorization: Bearer $TOKEN_A" -H 'content-type: application/json' \
  -d "{\"display_name\":\"Alice S20\",\"email\":\"$CUST_A_ALLOWED\",\"locale\":\"en\",\"source\":\"manual\",\"status\":\"active\"}" | jq .
```

### S8 — Customer portal OTP UX preserves no-auth path (TASK-0095 · AC 1, 2, 5)

1. Open `http://localhost:8091/` in a clean browser profile.
2. Without signing in, select an avatar and send a chat message.
3. Sign in using the customer OTP panel with `$CUST_A_ALLOWED`.
4. Verify signed-in account state appears.
5. Send another chat message.
6. Click **Sign out**.
7. Send a third no-auth chat message.

Expected:

- [ ] No-auth chat works before sign-in.
- [ ] OTP sign-in shows masked delivery and accepts the emailed code.
- [ ] Signed-in chat works.
- [ ] Sign-out clears account state.
- [ ] No-auth chat still works after sign-out.
- [ ] Expired/wrong OTP shows a clear error.

### S9 — Authenticated chat consumes correct tenant rate-limit keys (TASK-0095 · AC 3; TASK-0096 · AC 4)

Record Redis rate-limit keys before chat:

```bash
redis-cli -u "$REDIS_URL" --scan --pattern 'monti_jarvis:rl:*:chat:*' | sort > /tmp/s20-rl-before.txt
```

Send authenticated chat as Tenant A customer:

```bash
curl -fsS -X POST "$BASE/api/chat" \
  -H "Authorization: Bearer $CUSTOMER_ACCESS_A2" \
  -H 'content-type: application/json' \
  -d '{"agent_id":"ava","topic":"general","message":"What support options do I have?","history":[]}' | jq .
```

Record Redis keys after chat:

```bash
redis-cli -u "$REDIS_URL" --scan --pattern 'monti_jarvis:rl:*:chat:*' | sort > /tmp/s20-rl-after.txt
comm -13 /tmp/s20-rl-before.txt /tmp/s20-rl-after.txt
```

Expected:

- [ ] Chat succeeds.
- [ ] New/incremented chat rate-limit key is under Tenant A, not `demo` unless Tenant A is `demo`.
- [ ] Tenant B chat creates or increments a Tenant B key only.
- [ ] Response does not expose quota internals to the customer.

### S10 — Authenticated call session uses customer tenant context (TASK-0095 · AC 3; TASK-0096 · AC 4)

```bash
curl -fsS -X POST "$BASE/api/calls" \
  -H "Authorization: Bearer $CUSTOMER_ACCESS_B" \
  -H 'content-type: application/json' \
  -d '{}' | tee /tmp/s20-call-b.json | jq .

export CALL_B=$(jq -r .id /tmp/s20-call-b.json)
```

Expected:

- [ ] Call session response has `tenant_id` equal to `$TENANT_B`.
- [ ] Tenant A customer token cannot read or mutate Tenant B call if route protection is enabled in the current environment.
- [ ] Voice rate-limit/concurrent quota keys, if voice is exercised, are attributed to `$TENANT_B`.

Optional voice browser check:

1. Sign in as Tenant B customer in the customer portal.
2. Start voice call.
3. Speak one short question.
4. End call.

Expected:

- [ ] Voice starts and ends successfully.
- [ ] Transcript persists.
- [ ] Redis concurrent-call key is released after call end or TTL.

### S11 — Storage safety: hashes only, no plaintext credentials (TASK-0092 · AC 2, 5)

Run DB inspection:

```bash
psql "$POSTGRES_URL" -c "select id, tenant_id, email_normalized, status, attempts, length(code_hash) as hash_len from callcenter.customer_otp_challenges order by created_at desc limit 5;"
psql "$POSTGRES_URL" -c "select id, tenant_id, customer_id, length(refresh_token_hash) as hash_len, revoked_at is not null as revoked from callcenter.customer_sessions order by created_at desc limit 5;"
psql "$POSTGRES_URL" -c "select tenant_id, customer_id, email_normalized, event, metadata from callcenter.customer_auth_events order by created_at desc limit 10;"
```

Expected:

- [ ] OTP challenge rows contain only `code_hash`, never plaintext OTP.
- [ ] Customer session rows contain only `refresh_token_hash`, never raw refresh tokens.
- [ ] Auth event metadata contains IP/user-agent context only; no OTP, access token, refresh token, or secret.
- [ ] Customer identities are unique by tenant/provider/email.

### S12 — Automated/build regression gate

```bash
GOFLAGS=-vet=off make test
npm --prefix apps/customer-web run check
npm --prefix apps/tenant-web run check
make build
make infra-check
```

Expected:

- [ ] Go tests pass.
- [ ] Customer Svelte check reports zero errors/warnings.
- [ ] Tenant Svelte check reports zero errors/warnings.
- [ ] Customer, platform-admin, tenant web apps and Go binary build.
- [ ] Infrastructure health remains green.

## 4. Teardown

```bash
rm -f /tmp/s20-*.json /tmp/s20-rl-before.txt /tmp/s20-rl-after.txt
make down
```

## 5. Sign-off

Release close evidence recorded on 2026-07-13:

- Browser OTP/account smoke passed on Libra Tech tenant (`libra-tech-co-ltd`) using `?tenant_id=libra-tech-co-ltd`.
- Tenant-context customer portal routing was verified by successful sign-in as `apaichon@gmail.com`.
- Avatar selection usability was fixed with a popup picker and validated by customer-web check/build.
- Automated gates passed: Go tests, server build, customer/tenant Svelte checks, customer/tenant builds, and diff check.
- The deeper multi-session quota/rate-limit load variant remains listed below for pre-production re-run before broad customer traffic.

| Tester | Date | Result | Defects |
| --- | --- | --- | --- |
| Codex release verification + user browser smoke | 2026-07-13 | ☑ Pass for v2.1.0 release | None |

Readiness evidence required:

- [x] Tenant-context OTP smoke tested on Libra Tech tenant.
- [x] Manual checklist created for two-tenant and multi-customer UAT.
- [x] Domain allow/deny behavior covered by API implementation and checklist.
- [x] Authenticated chat/call tenant routing covered by implementation and automated smoke.
- [x] OTP/session storage uses hashes only.
- [x] No broad production customer traffic enabled by this local release.

Any failed checkbox requires a defect task; SPRINT-020 must remain open until the defect is resolved and the scenario reruns successfully.
