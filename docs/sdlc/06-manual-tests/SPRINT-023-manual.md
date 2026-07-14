# SPRINT-023 - Manual Test Checklist

**Feature:** FEAT-0025 · **Tasks:** TASK-0107–0111 · **Target:** v2.4.0

## 0. Preconditions

- [ ] Postgres, NATS, and the Monti server are running.
- [ ] At least one tenant admin and one customer portal session are available.
- [ ] A customer call or chat has created a conversation record.

```bash
export BASE=http://localhost:8091
export TENANT_ID=libra-tech-co-ltd
export TENANT_TOKEN='<tenant-admin-token>'
export CUSTOMER_TOKEN='<customer-token>'
```

## 1. Customer offer and confirmation

1. Start a customer chat or voice call with an assigned AI employee.
2. Ask to speak with a human agent.
3. Verify a follow-up offer appears and no ticket is created yet.
4. Choose **No thanks** and verify the conversation continues without a ticket.
5. Trigger the offer again and confirm it with a valid email when anonymous.

Expected:

- [ ] The offer contains a bounded reason and category.
- [ ] Declining or ignoring the offer does not create a ticket.
- [ ] Confirmation requires the existing session/call id.
- [ ] Anonymous confirmation requires a contact email.
- [ ] The success message exposes only the ticket reference.

## 2. Idempotency and source safety

Repeat the confirmation with the same idempotency key:

```bash
curl -fsS -X POST "$BASE/api/customer/tickets" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: sprint23-repeat-01' \
  -H "X-Tenant-Id: $TENANT_ID" \
  -d '{"call_id":"<call-id>","confirm_escalation":true,"subject":"Human follow-up requested","description":"The customer asked for human assistance.","category":"general","contact_email":"customer@example.com"}' | jq .
```

Expected:

- [ ] The first request creates one open ticket.
- [ ] The repeated request returns the same ticket with `idempotent: true`.
- [ ] A different subject with the same key returns a conflict.
- [ ] A missing or cross-tenant source call returns `not_found`.

## 3. Tenant queue and lifecycle

1. Open `/tenant/tickets` as the tenant admin.
2. Verify the default filter is today plus open tickets.
3. Filter by date range, status, priority, and category.
4. Open a ticket, change status/priority, assign an active tenant admin, and add an internal note.
5. Attempt to assign a user from another tenant.

Expected:

- [ ] Queue and detail are tenant-scoped.
- [ ] Avatar name and source call context are visible when available.
- [ ] Status transitions follow the bounded lifecycle.
- [ ] The event timeline records status, priority, assignment, and notes.
- [ ] Cross-tenant ticket ids return `404 not_found` without metadata.
- [ ] Cross-tenant assignees are rejected.

## 4. Verification

```bash
/usr/local/go/bin/go test ./...
cd apps/customer-web && npm run check && npm run build
cd ../tenant-web && npm run check && npm run build
```

## 5. Sign-off

| Tester | Date | Result | Defects |
| --- | --- | --- | --- |
| | | ☐ Pass / ☐ Fail | |
