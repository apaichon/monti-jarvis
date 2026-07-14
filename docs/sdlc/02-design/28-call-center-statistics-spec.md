---
id: DES-0028
title: Tenant Call Center Statistics and Quota Usage Specification
status: approved
updated: 2026-07-14
sprint: SPRINT-025
owner: SA
---

# Tenant Call Center Statistics and Quota Usage - Design Spec

**Sprint:** SPRINT-025 · **Release target:** v2.6.0  
**Feature:** [FEAT-0027](../01-features/FEAT-0027-tenant-call-center-statistics.md)  
**Depends on:** [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md), [27-customer-satisfaction-statistics-spec.md](27-customer-satisfaction-statistics-spec.md)

## Goals

- Give active tenant administrators a date-filtered call-center dashboard.
- Project completed conversation usage into ClickHouse without copying contact, transcript, or audio data.
- Show session totals, chat/voice breakdown, duration, AI employee breakdown, and current quota usage.
- Default both dates to today in the tenant deployment timezone.
- Keep Postgres and Redis as source-of-truth systems for operational records and quota enforcement.
- Make projection retries and bounded historical replay idempotent.

## Non-goals

- Platform-wide or cross-tenant analytics; planned for Sprint 29.
- System performance and infrastructure monitoring; planned for Sprint 26.
- Billing, AI infrastructure cost, or overage analytics; planned for Sprint 30.
- Raw transcript, customer contact, rating comment, or audio data in ClickHouse facts.
- A public projection-management API or customer-facing analytics view.

## Environment and configuration

| Variable | Default | Description |
| --- | --- | --- |
| `POSTGRES_SCHEMA` | `callcenter` | Source schema for conversation records, call sessions, and tenant settings. |
| `CLICKHOUSE_URL` | unset | ClickHouse HTTP endpoint. When unset, the projection and dashboard are unavailable. |
| `CLICKHOUSE_DATABASE` | `monti_jarvis` | Analytics database. |
| `REDIS_PREFIX` | `monti_jarvis:` | Existing quota and rate-limit key prefix. No new Sprint 25 key is required. |
| `TZ` | deployment timezone | Date-only range resolution when a tenant setting does not provide a timezone. |
| `AUTH_DISABLED` | `true` | Does not bypass tenant-admin protection for the dashboard. |

## Data model

### Existing Postgres sources

`callcenter.conversation_records` remains the operational source for dashboard facts. A record is eligible when `ended_at` is non-null and its `channel` is `chat` or `voice`; both `archived` and `archive_failed` records remain countable because archive success must not erase a completed call from usage statistics. The projection joins `callcenter.call_sessions.source` to retain `production` or `preview` context without exposing it in the tenant UI.

`tenant_entitlements` and `package_limits.rules` provide monthly package ceilings. Existing Redis quota counters and `tenant_call_limits` provide current monthly and daily usage/limit values. The dashboard never replaces the quota enforcement path.

### ClickHouse `call_center_usage_facts`

Migration/bootstrap contract: `scripts/migrations/025_call_center_analytics.sql` and the ClickHouse schema initializer. The table is a one-row-per-completed-conversation projection:

| Column | Type | Notes |
| --- | --- | --- |
| `fact_id` | `String` | Deterministic `ccf_` + conversation record id; projection idempotency key. |
| `tenant_id` | `String` | Required scope key; first `ORDER BY` dimension. |
| `call_id` | `String` | Source call/session id; no customer contact data. |
| `conversation_record_id` | `String` | Postgres source id. |
| `avatar_id` | `String` | Selected AI employee id, empty only for legacy records. |
| `channel` | `LowCardinality(String)` | `chat` or `voice`. |
| `source` | `LowCardinality(String)` | `production` or `preview`; used for reconciliation, not shown as a customer field. |
| `status` | `LowCardinality(String)` | `archived` or `archive_failed`. |
| `started_at` | `DateTime` | UTC source start time. |
| `ended_at` | `DateTime` | UTC source end time for eligible records. |
| `usage_date` | `Date` | Date bucket resolved from the tenant/deployment timezone. |
| `duration_seconds` | `UInt32` | Non-negative bounded source duration. |
| `source_updated_at` | `DateTime` | Postgres source `updated_at` used for replay ordering. |
| `created_at` | `DateTime` | Projection insert time. |
| `updated_at` | `DateTime` | Projection revision time and `ReplacingMergeTree` version. |
| `created_by` | `String` | `system` or projection actor. |
| `updated_by` | `String` | `system` or projection actor. |

Recommended engine and order:

```sql
CREATE TABLE IF NOT EXISTS monti_jarvis.call_center_usage_facts (
  fact_id String,
  tenant_id String,
  call_id String,
  conversation_record_id String,
  avatar_id String,
  channel LowCardinality(String),
  source LowCardinality(String),
  status LowCardinality(String),
  started_at DateTime,
  ended_at DateTime,
  usage_date Date,
  duration_seconds UInt32,
  source_updated_at DateTime,
  created_at DateTime DEFAULT now(),
  updated_at DateTime DEFAULT now(),
  created_by String DEFAULT 'system',
  updated_by String DEFAULT 'system'
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (tenant_id, usage_date, call_id, fact_id);
```

Queries must use `FINAL` or an equivalent latest-row strategy when exact idempotent results are required. Dashboard queries always filter `tenant_id` first and only count rows with `ended_at` populated.

### Projection and replay contract

1. On conversation close/archive update, the analytics projector reads the tenant-scoped source record and its call-session source.
2. It computes the deterministic `fact_id` and inserts the complete row into ClickHouse.
3. A retry inserts the same `fact_id` with a newer `updated_at`; it does not create a second logical conversation.
4. A bounded replay reads Postgres records by `updated_at` or date range and applies the same projector. It is an operator job, not a public HTTP endpoint.
5. If ClickHouse is unavailable, the source record remains authoritative and the dashboard returns `analytics_unavailable`; the live call/archive path must not fail solely because analytics projection failed.

## API summary

| Method | Path | Role | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/tenant/call-center/statistics` | `tenant_admin` | Read tenant-scoped date-filtered activity and current quota usage. |

The endpoint accepts optional `start_date` and `end_date` query parameters in `YYYY-MM-DD` format. If omitted, both resolve to today. The date range applies to activity metrics; the quota block reports the current package period and today's operational cap so operators can distinguish historical range totals from current enforcement state.

## RBAC and tenant isolation

| Action | `platform_admin` | `tenant_admin` | `customer` | Anonymous |
| --- | --- | --- | --- | --- |
| Read current tenant dashboard | no Sprint 25 route | yes, active tenant only | no | no |
| Read another tenant's facts | no | no | no | no |
| Trigger projection/replay | operator job only | no | no | no |

The Go handler derives `tenant_id` from the authenticated tenant-admin context. No request parameter may select a tenant. `AUTH_DISABLED=true` preserves the customer demo path but does not make tenant analytics public.

## Verification

```bash
make test
make build
cd apps/tenant-web && npm run build

curl 'http://localhost:8091/api/tenant/call-center/statistics' \
  -H 'Authorization: Bearer <tenant-admin-token>'

curl 'http://localhost:8091/api/tenant/call-center/statistics?start_date=2026-07-14&end_date=2026-07-14' \
  -H 'Authorization: Bearer <tenant-admin-token>'
```

Expected checks:

- Today is the default range and the response includes the resolved timezone.
- Counts and minutes match controlled Postgres conversation fixtures after projection.
- Replaying the same source rows does not increase logical session count.
- A second tenant cannot read the first tenant's facts or quota values.
- ClickHouse outage returns a stable `analytics_unavailable` response without breaking existing calls or `/api/tenant/usage`.

See [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), [04-api-spec.md](04-api-spec.md), and [05-ux-ui.md](05-ux-ui.md).
