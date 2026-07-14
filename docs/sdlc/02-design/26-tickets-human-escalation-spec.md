---
id: DES-0026
title: Tickets and Human Escalation Specification
status: approved
updated: 2026-07-14
sprint: SPRINT-023
owner: SA
---

# Tickets and Human Escalation - Design Spec

**Sprint:** SPRINT-023 · **Release target:** v2.4.0  
**Feature:** [FEAT-0025](../01-features/FEAT-0025-tickets-human-escalation.md)  
**Depends on:** [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md), [24-authenticated-workforce-selection-spec.md](24-authenticated-workforce-selection-spec.md)

## 1. Goals

- Let the conversation runtime offer human follow-up when the customer asks for it or an approved escalation signal is present.
- Require explicit customer confirmation before creating a ticket.
- Give tenant admins a tenant-scoped queue, detail view, event timeline, and bounded lifecycle controls.
- Keep ticket creation idempotent and publish safe tenant-context events for future notification workers.

## 2. Non-goals (Sprint 23)

- Real-time human transfer, supervisor listen mode, or agent takeover.
- SLA timers, email/SMS notifications, external CRM adapters, or platform-wide ticket search.
- Customer-facing ticket history or two-way messaging after creation.
- ClickHouse ticket analytics or retention automation.

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `TICKETS_ENABLED` | `true` | Enables customer escalation and tenant ticket APIs. |
| `TICKET_IDEMPOTENCY_TTL` | `24h` | Retention for duplicate-create protection. |
| `TICKET_EVENT_SUBJECT` | `ticket.created,ticket.updated` | NATS subjects emitted with tenant context. |
| `TICKET_MAX_OPEN_PER_CUSTOMER_DAY` | `3` | Soft abuse guard; `0` disables the additional cap. |

## 4. Data model (Postgres `callcenter`)

### `tickets`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `tick_{ulid}` |
| `tenant_id` | text FK | Owning tenant; never taken from a tenant-admin request body |
| `conversation_record_id` | text FK nullable | Source Sprint 22 conversation record |
| `call_id` | text nullable | Source call/session id when a record is not finalized yet |
| `customer_id` | text FK nullable | Nullable for anonymous callers |
| `avatar_id` | text FK nullable | AI employee handling the source conversation |
| `subject` | text | Bounded operator-facing title |
| `description` | text | Bounded customer request summary; no raw transcript dump |
| `category` | text | `general`, `billing`, `technical`, `other` |
| `priority` | text | `low`, `normal`, `high`, `urgent` |
| `status` | text | `open`, `in_progress`, `waiting_customer`, `resolved`, `closed` |
| `source` | text | `customer_request`, `agent_escalation`, `tenant_created` |
| `assignee_user_id` | text FK nullable | Tenant user assigned for follow-up |
| `contact_name` | text | Optional bounded snapshot for anonymous follow-up |
| `contact_email` | text | Required for anonymous follow-up; masked in list responses |
| `resolved_at` | timestamptz nullable | Set on transition to `resolved` |
| `closed_at` | timestamptz nullable | Set on transition to `closed` |
| `last_activity_at` | timestamptz | Queue sort field |
| audit columns | | `created_at`, `updated_at`, `created_by`, `updated_by` |

### `ticket_events`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `tev_{ulid}` |
| `tenant_id` | text FK | Tenant isolation duplicate for query safety |
| `ticket_id` | text FK | Parent ticket |
| `event_type` | text | `created`, `status_changed`, `priority_changed`, `assigned`, `note_added`, `customer_confirmed` |
| `actor_type` | text | `system`, `customer`, `tenant_user` |
| `actor_id` | text nullable | Auth/session actor when available |
| `note` | text | Bounded internal note; empty for non-note events |
| `payload` | jsonb | Safe change metadata, no transcript or secrets |
| audit columns | | `created_at`, `updated_at`, `created_by`, `updated_by` |

### Lifecycle

| Status | Meaning | Allowed next states |
| --- | --- | --- |
| `open` | Confirmed request is waiting for tenant triage | `in_progress`, `closed` |
| `in_progress` | Tenant user is working the request | `waiting_customer`, `resolved`, `closed` |
| `waiting_customer` | Tenant needs more information; no customer reply workflow in this sprint | `in_progress`, `closed` |
| `resolved` | Tenant completed the follow-up work | `closed`, `in_progress` |
| `closed` | Terminal MVP state | none |

## 5. Redis / NATS / ClickHouse / MinIO

| Store | Contract |
| --- | --- |
| Redis | `monti_jarvis:ticket:idempotency:{tenant_id}:{key}` stores the created ticket id for `TICKET_IDEMPOTENCY_TTL`. |
| Redis | `monti_jarvis:rate:{tenant_id}:customer:{customer_id}:ticket:{yyyymmdd}` supports the optional daily abuse guard. |
| NATS | `ticket.created` and `ticket.updated`; event payload includes tenant, ticket, source conversation, status, and actor metadata only. |
| ClickHouse | No new table in Sprint 23. |
| MinIO | No new object contract; ticket links to existing `calls/{tenant_id}/{call_id}/` archive metadata. |

Migration/bootstrap placeholder: `scripts/migrations/023_tickets_human_escalation.sql` or the repository's idempotent `internal/store` schema bootstrap, matching existing deployment conventions.

## 6. API summary

See [04-api-spec.md](04-api-spec.md) § Tickets & Human Escalation. Quick list:

| Method | Path | Role |
| --- | --- | --- |
| `POST` | `/api/customer/tickets` | public or `customer`, subject to tenant customer-auth policy |
| `GET` | `/api/tenant/tickets` | `tenant_admin` |
| `GET` | `/api/tenant/tickets/{id}` | `tenant_admin` |
| `PATCH` | `/api/tenant/tickets/{id}` | `tenant_admin` |
| `POST` | `/api/tenant/tickets/{id}/events` | `tenant_admin` |

## 7. RBAC

| Action | public | customer | tenant_admin | platform_admin |
| --- | --- | --- | --- | --- |
| Receive a structured ticket offer | yes | yes | no | no |
| Create a ticket after explicit confirmation | policy-dependent | yes | no | no |
| List/read current-tenant tickets | no | no | yes | no by default |
| Change priority, assignee, or lifecycle | no | no | yes | no by default |
| Add internal tenant note | no | no | yes | no by default |
| Read another tenant's ticket | no | no | no | no |

## 8. Verification

```bash
make test && make build

curl -X POST "$BASE_URL/api/customer/tickets" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-escalation-01' \
  -d '{"call_id":"call_01","confirm_escalation":true,"subject":"Need a human follow-up","category":"general","description":"Customer requested a human agent","contact_email":"customer@example.com"}'

curl -H "Authorization: Bearer $TENANT_TOKEN" \
  "$BASE_URL/api/tenant/tickets?status=open&start_date=2026-07-15&end_date=2026-07-15"

curl -X PATCH "$BASE_URL/api/tenant/tickets/tick_01" \
  -H "Authorization: Bearer $TENANT_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"in_progress","priority":"high"}'
```

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | | ☐ |
| Dev | | | ☐ |
