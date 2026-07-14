# SPRINT-022 — Manual Test Checklist

**Feature:** FEAT-0024 · **Tasks:** TASK-0102–0106 · **Target:** v2.3.0

## 0. Preconditions

- [ ] SPRINT-021 customer portal build is running.
- [ ] MinIO is configured with bucket `monti-jarvis`.
- [ ] Tenant admin can log in.
- [ ] At least one chat can trigger missing KM / no-source behavior.

```bash
export BASE=http://localhost:8091
export TENANT_ID=libra-tech-co-ltd
export TENANT_TOKEN='<tenant-admin-token>'
export CUSTOMER_TOKEN='<customer-token>'
```

## 1. Init infrastructure

```bash
make infra-up
make restart
curl -fsS "$BASE/healthz" | jq .
curl -fsS "$BASE/api/infra" | jq .
```

Expected:

- [ ] Postgres is `ok`.
- [ ] MinIO is `ok` or the archive record clearly marks `minio_disabled` in local dev.

## 2. Prepare data

1. Sign in to customer portal for `$TENANT_ID`.
2. Select an assigned avatar.
3. Ask at least one normal chat question.
4. Ask one question not covered by tenant KM.
5. Start and end one short call.

## 3. Scenarios

### S1 — Chat creates conversation record and archive metadata (TASK-0102, TASK-0103)

```bash
curl -fsS -H "Authorization: Bearer $TENANT_TOKEN" \
  "$BASE/api/tenant/conversation-records" | jq .
```

Expected:

- [ ] Response includes chat conversation records.
- [ ] Record status is `archived` when MinIO write succeeds or `archive_failed` when MinIO is disabled/unavailable.
- [ ] Object path uses `calls/{tenant_id}/{call_id}/...` and does not include customer email.

### S2 — Tenant records UI works (TASK-0105)

1. Open `/tenant/conversation-records`.
2. Select a record.
3. If a record has `archive_failed`, click **Retry archive**.

Expected:

- [ ] List loads tenant-scoped records.
- [ ] Detail panel shows safe metadata.
- [ ] Retry returns success or a clear infrastructure error.

### S3 — Knowledge gap candidate lifecycle (TASK-0104, TASK-0105)

```bash
curl -fsS -H "Authorization: Bearer $TENANT_TOKEN" \
  "$BASE/api/tenant/knowledge-gaps?status=open" | jq .
```

Then open `/tenant/knowledge-gaps`, add a reviewer note, and resolve/snooze/ignore a gap.

Expected:

- [ ] Missing-KM chat created a gap candidate.
- [ ] Tenant UI updates the lifecycle state.
- [ ] Cross-tenant gap ids return 404/403.

### S4 — Cross-tenant isolation (TASK-0106)

1. Log in as another tenant.
2. Attempt to load a record/gap id from `$TENANT_ID`.

Expected:

- [ ] API returns 404/403.
- [ ] UI does not leak metadata or object paths.

## 4. Teardown

```bash
make down
```

## 5. Sign-off

| Tester | Date | Result | Defects |
| --- | --- | --- | --- |
| | | ☐ Pass / ☐ Fail | |
