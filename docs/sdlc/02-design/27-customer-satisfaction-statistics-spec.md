---
id: DES-0027
title: Customer Satisfaction Review and Tenant Statistics Specification
status: approved
updated: 2026-07-14
sprint: SPRINT-024
owner: SA
---

# Customer Satisfaction Review and Tenant Statistics - Design Spec

**Sprint:** SPRINT-024 · **Release target:** v2.5.0  
**Feature:** [FEAT-0026](../01-features/FEAT-0026-customer-satisfaction-statistics.md)  
**Depends on:** [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md), [26-tickets-human-escalation-spec.md](26-tickets-human-escalation-spec.md)

## 1. Goals

- Let a completed chat or voice conversation receive one customer satisfaction score from 1 to 5.
- Use star icons in the customer review surface and keep the review independent from call/archive closure.
- Allow a customer who skipped the initial prompt to submit a score from a follow-up prompt without reopening the call.
- Give tenant admins aggregate, date-filtered statistics with today as the default range.
- Keep every review and aggregate query tenant scoped and free of raw transcript or contact data.

## 2. Non-goals (Sprint 24)

- Public reviews or platform-wide cross-tenant analytics.
- Sentiment inference, NPS/CSAT variants, or free-form comment analysis.
- ClickHouse projections; Sprint 24 reads Postgres aggregates and Sprint 25 owns call-center dashboards.
- Reopening, extending, or transferring a completed call to collect a review.

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `AUTH_DISABLED` | `true` | Customer rating follows the existing optional customer path; tenant statistics remain tenant-admin protected. |
| `POSTGRES_SCHEMA` | `callcenter` | Source of truth for ratings and conversation metadata. |
| `TZ` | deployment timezone | Used when resolving the default `today` statistics range. |

No new Redis, NATS, MinIO, or ClickHouse configuration is required.

## 4. Data model (Postgres `callcenter`)

Sprint 24 extends the existing `conversation_ratings` table created with conversation records. The migration must preserve existing rows and make the relationship to the archived conversation explicit.

### `conversation_ratings`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `crat_*` identifier |
| `tenant_id` | text | Required tenant scope |
| `call_id` | text | Required source call/session; unique with `tenant_id` |
| `conversation_record_id` | text FK | Links to `conversation_records.id` when archive exists |
| `customer_id` | text | Optional authenticated customer |
| `avatar_id` | text | Selected AI employee |
| `channel` | text | `chat` or `voice` |
| `score` | integer | Check constraint `1 <= score <= 5` |
| `review` | text | Existing bounded compatibility field; not shown or aggregated in Sprint 24 |
| `created_at` | timestamptz | Audit timestamp |
| `updated_at` | timestamptz | Audit timestamp |
| `created_by` | text | Audit actor, default `system`/customer context |
| `updated_by` | text | Audit actor, default `system`/customer context |

Required constraints and indexes:

- `UNIQUE (tenant_id, call_id)` keeps repeated voice/chat submissions idempotent.
- `CHECK (score BETWEEN 1 AND 5)` rejects invalid ratings at the store boundary and database boundary.
- `INDEX (tenant_id, created_at DESC)` supports date-filtered tenant statistics.
- `INDEX (tenant_id, avatar_id, channel, created_at DESC)` supports dashboard breakdowns.
- `conversation_record_id` is nullable for legacy calls whose archive record was not created.

Migration placeholder: `scripts/migrations/024_customer_satisfaction_reviews.sql`.

## 5. Redis / NATS / ClickHouse

| Store | Contract |
| --- | --- |
| Redis | No new key. Existing call/session state remains authoritative while the call is active. |
| NATS | No new event required for the MVP; review writes are synchronous and idempotent. |
| ClickHouse | No new table in Sprint 24. Aggregate queries use Postgres; Sprint 25 may project them for dashboards. |
| MinIO | Reuse Sprint 22 `calls/{tenant_id}/{call_id}/` archive objects; review submission never rewrites audio/transcript objects. |

## 6. API summary

See [04-api-spec.md](04-api-spec.md) § Customer Satisfaction. Quick list:

| Method | Path | Role |
| --- | --- | --- |
| `POST` | `/api/calls/{id}/rating` | public/optional customer bearer |
| `GET` | `/api/tenant/satisfaction/statistics` | `tenant_admin` |

## 7. RBAC

| Action | `platform_admin` | `tenant_admin` | `customer` |
| --- | --- | --- | --- |
| Submit rating for a call | no | no | yes, only against the call session |
| Read tenant satisfaction statistics | no in Sprint 24 | yes, current tenant only | no |
| Read another tenant's rating or aggregate | no | no | no |

Customer rating is intentionally compatible with the existing no-auth portal when `AUTH_DISABLED=true`; the server resolves the tenant from the call session and never trusts a tenant id supplied in the rating body.

## 8. Verification

```bash
/usr/local/go/bin/go test ./...
/usr/local/go/bin/go build ./cmd/server
curl -X POST http://localhost:8091/api/calls/call_01/rating \
  -H 'Content-Type: application/json' \
  -d '{"score":5}'
curl 'http://localhost:8091/api/tenant/satisfaction/statistics?start_date=2026-07-14&end_date=2026-07-14' \
  -H 'Authorization: Bearer <tenant-admin-token>'
```

## Approver sign-off

Approved by user instruction to build Sprint 24 on 2026-07-14.

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | User instruction | 2026-07-14 | ☑ |
| Dev | | | ☐ |
| Tester | | | ☐ |
