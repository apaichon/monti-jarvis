# SPRINT-021 — Manual Test Checklist

**Feature:** FEAT-0023 · **Tasks:** TASK-0097–0101 · **Target:** v2.2.0

## 0. Preconditions

- [ ] `AUTH_DISABLED=false`, `JWT_SECRET` is set, and tenant/customer auth from SPRINT-020 works.
- [ ] Tenant A is optional-auth; Tenant B is required-auth for workforce selection.
- [ ] Each tenant has at least one active assigned avatar.
- [ ] Tester controls customer OTP mailbox.

```bash
export BASE=http://localhost:8091
export TENANT_A=demo
export TENANT_B=libra-tech-co-ltd
export TOKEN_A='<tenant-a-admin-token>'
export TOKEN_B='<tenant-b-admin-token>'
export CUSTOMER_EMAIL='apaichon@gmail.com'
```

## 1. Init infrastructure

```bash
make infra-up
make restart
curl -fsS "$BASE/healthz" | jq .
curl -fsS "$BASE/api/infra" | jq .
```

Expected:

- [ ] Server is reachable.
- [ ] Postgres and Redis are `ok`.
- [ ] Customer portal opens at `$BASE/?tenant_id=$TENANT_B`.

## 2. Prepare tenant policy

Optional tenant:

```bash
curl -fsS -X PUT "$BASE/api/tenant/customer-auth/settings" \
  -H "Authorization: Bearer $TOKEN_A" -H 'content-type: application/json' \
  -d '{"enabled":true,"auth_mode":"optional","require_auth_for_workforce":false,"customer_daily_call_seconds":0,"customer_max_call_seconds":0}' | jq .
```

Required tenant with a small quota:

```bash
curl -fsS -X PUT "$BASE/api/tenant/customer-auth/settings" \
  -H "Authorization: Bearer $TOKEN_B" -H 'content-type: application/json' \
  -d '{"enabled":true,"auth_mode":"required","require_auth_for_workforce":true,"customer_daily_call_seconds":60,"customer_max_call_seconds":30}' | jq .
```

## 3. Scenarios

### S1 — Optional tenant preserves no-auth flow (TASK-0097, TASK-0098)

1. Open `$BASE/?tenant_id=$TENANT_A`.
2. Do not sign in.
3. Confirm avatar picker is usable.
4. Send a chat message and start/end a call.

Expected:

- [ ] Workforce loads.
- [ ] Chat/call are not blocked by OTP.

### S2 — Required tenant blocks workforce until OTP (TASK-0097, TASK-0098)

1. Open `$BASE/?tenant_id=$TENANT_B`.
2. Confirm sign-in card states customer sign-in is required.
3. Attempt to use picker/chat/start call before OTP.
4. Complete OTP sign-in.

Expected:

- [ ] Picker/chat/call are disabled before OTP.
- [ ] After OTP, assigned avatars load and selected customer is shown.

### S3 — Customer quota state is visible and enforced (TASK-0099, TASK-0100)

1. With Tenant B signed in, confirm quota text is visible.
2. Set `customer_daily_call_seconds` low enough to exhaust quickly.
3. Send chat/call until quota is exhausted.

Expected:

- [ ] `/api/customer/quota` returns remaining seconds.
- [ ] Exhausted quota blocks new chat/call with `customer_quota_exhausted`.
- [ ] UI shows quota exhausted state.

### S4 — Tenant settings persist S21 fields (TASK-0100)

1. Open `/tenant/settings`.
2. Toggle **Require OTP before AI workforce selection**.
3. Set customer daily/max call minutes.
4. Save and reload.

Expected:

- [ ] Values persist after reload.
- [ ] Customer portal behavior changes immediately after refresh.

## 4. Teardown

```bash
make down
```

## 5. Sign-off

| Tester | Date | Result | Defects |
| --- | --- | --- | --- |
| | | ☐ Pass / ☐ Fail | |
