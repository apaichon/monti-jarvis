---
id: DES-0033
title: Platform Call Center Statistics by Tenant Specification
status: shipped
updated: 2026-07-17
sprint: SPRINT-030
owner: SA
---

# Platform Call Center Statistics by Tenant - Design Spec

**Sprint:** SPRINT-030 · **Release target:** v2.11.0  
**Feature:** [FEAT-0032](../01-features/FEAT-0032-platform-call-center-statistics.md)  
**Depends on:** [28-call-center-statistics-spec.md](28-call-center-statistics-spec.md), [32-platform-system-performance-spec.md](32-platform-system-performance-spec.md)

## 1. Goals

- Give platform administrators a date-filtered overview of completed AI conversation activity across active tenants.
- Reuse Sprint 25 ClickHouse facts and metric definitions so platform totals can be reconciled with tenant statistics.
- Provide a bounded tenant breakdown with aggregate-only values, safe tenant identity, package labels, and satisfaction summaries.
- Distinguish current, stale, empty, unavailable, and partially unavailable enrichment states.
- Keep the endpoint read-only and outside customer call, voice, archive, quota, audit-delivery, and chat critical paths.

## 2. Non-goals

- Billing reconciliation, invoice reporting, quota enforcement changes, or AI infrastructure cost allocation; these remain Sprint 31 scope.
- Customer-level drill-down, transcripts, contact data, ticket notes, rating comments, audio paths, or archive playback.
- Persistent dashboard snapshots, time-series history, scheduled reports, exports, alerts, or mobile API changes.
- A second usage authority that replaces Redis quota counters or tenant entitlements.

## 3. Environment and runtime contract

| Variable | Default | Contract |
| --- | --- | --- |
| `POSTGRES_SCHEMA` | `callcenter` | Source schema for tenant metadata, completed records, ratings, entitlements, and package labels. |
| `CLICKHOUSE_URL` | unset | Required for activity statistics. Unset or unavailable produces `analytics_unavailable`, never zero activity. |
| `CLICKHOUSE_DATABASE` | `monti_jarvis` | Existing analytics database containing `call_center_usage_facts`. |
| `TZ` | deployment timezone | Resolves omitted date filters and the returned range timezone. |
| `MONITORING_PROBE_TIMEOUT` | `2s` | Existing dependency probe timeout; Sprint 30 does not reuse monitoring probes for statistics queries. |

The statistics request uses a bounded request context. An implementation may use a dedicated constant or configuration for the analytics query timeout, but it must not allow an unbounded ClickHouse or Postgres request. No new cache or background polling loop is introduced.

## 4. Data sources and metric definitions

### 4.1 ClickHouse activity facts

The existing `monti_jarvis.call_center_usage_facts` table is the activity source. Queries must use `FINAL` or an equivalent latest-row strategy and apply:

```sql
status = 'archived'
AND usage_date BETWEEN toDate(:start_date) AND toDate(:end_date)
```

The query returns aggregate rows grouped by `tenant_id`, `avatar_id`, and `channel`. It must never select customer identifiers, transcripts, rating comments, ticket notes, or audio paths.

| Metric | Definition |
| --- | --- |
| `completed_conversations` | `count()` of eligible archived facts. |
| `total_duration_seconds` | `sum(duration_seconds)` across eligible facts. |
| `average_duration_seconds` | Total duration divided by completed conversations; zero when count is zero. |
| `chat_conversations` / `voice_conversations` | Count grouped by `channel`. |
| `range_call_minutes` | `ceil(total_duration_seconds / 60)` for reporting; separate from live quota counters. |
| `last_projected_at` | Maximum `updated_at` of the eligible facts, nullable for an empty range. |

The default source behavior matches Sprint 25 and includes all eligible archived facts. `source` remains available in the fact for reconciliation but is not a Sprint 30 filter.

### 4.2 Postgres enrichment

Postgres is used for small, redacted enrichment queries:

- Active tenant identity: `tenant_id`, slug, display name, and optional safe `logo_url`.
- Satisfaction: completed records joined to `conversation_ratings`, returning only reviewed count, average score, completion rate, and distribution keys `1` through `5`.
- Package label: active entitlement status and package name, without exposing rules JSON or quota configuration.

Satisfaction enrichment is scoped to the same date range and completed-record definition. If it cannot be read, activity totals remain valid and the response marks satisfaction `unavailable`.

Package usage is a reporting value derived from ClickHouse range duration. Current monthly/daily Redis enforcement counters are not aggregated by this endpoint and are never mutated.

### 4.3 Freshness

Freshness uses the same five-minute threshold as tenant system monitoring:

| Status | Condition |
| --- | --- |
| `current` | Activity query succeeds and `now - last_projected_at <= 5m`. |
| `stale` | Activity query succeeds but `last_projected_at` is older than 5m. |
| `empty` | Valid range has no eligible facts and no projection timestamp. |
| `unavailable` | ClickHouse is disabled, unreachable, or the bounded query fails. |

The API must not turn `unavailable` into a valid empty response.

## 5. Ephemeral response model

```text
PlatformCallCenterStatistics
├── range { start_date, end_date, timezone }
├── freshness { source, status, generated_at, last_projected_at? }
├── totals { activity, duration, channel, range minutes, satisfaction }
├── by_channel[]
├── by_avatar[]
├── package_usage { active_package_tenants, range_call_minutes, enforcement_counters }
├── tenants[]
└── pagination { total, limit, offset }
```

Tenant rows contain:

| Field | Contract |
| --- | --- |
| `tenant_id`, `slug`, `name` | Existing platform-visible active tenant identity. |
| `logo_url` | Optional existing brand image URL; no new branding storage. |
| `package` | Active package name/status only. |
| `analytics_status` | `current`, `stale`, `empty`, or `unavailable`. |
| Activity fields | Completed count, duration, average duration, and range call minutes. |
| Satisfaction fields | Reviewed count, average score, completion rate; nullable/unavailable when enrichment fails. |

The response is assembled per request and is not persisted, cached, or returned to tenant-admin/customer roles.

## 6. API contract

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/platform/call-center/statistics` | `platform_admin` | Read aggregate platform activity and a bounded tenant page. |

Query fields:

| Field | Type | Default | Validation |
| --- | --- | --- | --- |
| `start_date` | date | today | `YYYY-MM-DD`, valid calendar date. |
| `end_date` | date | today | `YYYY-MM-DD`, valid and on/after start. |
| `tenant_id` | string | empty | Exact filter, maximum 128 characters. |
| `limit` | integer | 50 | `1..100`. |
| `offset` | integer | 0 | `0..1000000`. |

Errors use the existing JSON shape and stable codes:

| HTTP | Code | Meaning |
| ---: | --- | --- |
| 400 | `validation_error` | Invalid date, date order, tenant id, limit, or offset. |
| 401 | `unauthorized` / `session_expired` | Missing or expired platform session. |
| 403 | `forbidden` | Caller is not a platform administrator. |
| 503 | `analytics_unavailable` | Activity source cannot provide a safe result. |
| 500 | `statistics_unavailable` | Unexpected safe-response failure. |

Postgres enrichment failure is represented inside the `enrichment`/satisfaction state in a `200` response when activity data is available. Raw errors are logged server-side only.

## 7. Authorization and privacy

1. Apply the existing platform-admin guard before querying tenant or analytics data.
2. Treat `tenant_id` as a filter only; never use it to establish authorization.
3. Read active tenant metadata through the existing platform store path and apply the same bounded page limit.
4. Keep all ClickHouse queries aggregate-only and use parameter escaping/validated date values.
5. Do not return customer ids, emails, phone numbers, transcripts, rating comments, ticket notes, audio paths, package rules, Redis keys, SQL, provider URLs, credentials, or local audit spool details.
6. Do not emit a new audit event for the read-only request solely because the dashboard was opened; existing protected-route audit policy remains authoritative.

## 8. Workflow and performance bounds

1. Normalize the date range once using `TZ` and include the resolved timezone in the response.
2. Query ClickHouse once for overall and grouped activity aggregates where possible; avoid one ClickHouse request per tenant row.
3. Query active tenant metadata with bounded pagination and merge aggregate rows by tenant id.
4. Run the rating/package enrichment in bounded queries or bounded batches; never create an unbounded goroutine per tenant.
5. Return a complete page atomically. A retry repeats the same read contract and does not mutate source facts.
6. Keep the route out of customer call, voice relay, archive, quota, audit spool, and chat handlers.

## 9. UX/UI contract

The platform-admin route is `/admin/call-center` and uses the existing platform shell. It must provide:

- Today-default start/end date controls with Apply and Today actions.
- Aggregate KPI summary before detailed tenant rows.
- Separate display for range call minutes versus live quota enforcement counters.
- Channel, avatar, satisfaction, freshness, empty, stale, unavailable, loading, retry, and session-expired states.
- Bounded tenant pagination and optional exact tenant filter.
- Responsive layout below 700px with stacked filters and readable two-line tenant rows.

No customer-level drill-down or raw source record link is part of this sprint.

## 10. Verification

```bash
make test
make build
cd apps/platform-admin-web && npm run check && npm run build
git diff --check

curl -i 'http://localhost:8091/api/platform/call-center/statistics' \
  -H 'Authorization: Bearer <platform-admin-token>'

curl -i 'http://localhost:8091/api/platform/call-center/statistics?start_date=2026-07-17&end_date=2026-07-17&limit=50&offset=0' \
  -H 'Authorization: Bearer <platform-admin-token>'
```

Expected checks:

- Platform totals equal the sum of returned tenant rows for controlled fixtures.
- Metrics match Sprint 25 tenant facts and inclusive date boundaries.
- Empty and stale ranges are distinguishable from ClickHouse unavailable.
- Satisfaction and package enrichment never leak comments, rules, or customer data.
- Tenant filter and pagination are bounded and do not change platform authorization.
- Tenant-admin, customer, and anonymous callers receive safe unauthorized/forbidden responses.
- Existing tenant statistics, platform monitoring, audit log, calls, quota, and archive paths remain compatible.
- Manual UAT: [SPRINT-030-manual.md](../06-manual-tests/SPRINT-030-manual.md).

## 11. Risks

| Risk | Mitigation |
| --- | --- |
| Platform totals drift from tenant dashboard totals | Share the same fact definition, `FINAL` semantics, date range, and controlled fixture assertions. |
| Large tenant count causes N+1 analytics queries | Use grouped ClickHouse aggregates and bounded Postgres enrichment batches. |
| Stale facts appear as zero activity | Return explicit freshness status; only valid empty ranges return zero totals. |
| Satisfaction/package source fails independently | Keep activity visible and mark only the affected enrichment unavailable. |
| Admin surface exposes cross-tenant details | Centralize allowlisted response types and test platform/non-platform RBAC. |

See [02-workflow.md](02-workflow.md) §84, [03-er-diagram.md](03-er-diagram.md), [04-api-spec.md](04-api-spec.md), and [05-ux-ui.md](05-ux-ui.md) A22.
