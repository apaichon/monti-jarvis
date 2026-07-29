---
id: DES-0051
title: Tenant Call-Center Topic Statistics Specification
status: shipped
release: v2.28.0
updated: 2026-07-29
sprint: SPRINT-055
owner: SA
feature: FEAT-0047
---

# DES-0051 — Tenant Call-Center Topic Statistics

**Sprint:** SPRINT-055 · **Release target:** v2.28.0  
**Feature:** [FEAT-0047](../01-features/FEAT-0047-tenant-call-center-topic-statistics.md)  
**Tasks:** [TASK-0199](../04-tasks/TASK-0199.md),
[TASK-0200](../04-tasks/TASK-0200.md), [TASK-0201](../04-tasks/TASK-0201.md)  
**Extends:** [DES-0028](28-call-center-statistics-spec.md)

## 1. Goals

- Add a tenant-scoped topic dimension to existing call-center analytics.
- Let tenant admins group and filter completed conversations by topic.
- Preserve existing S25 totals, quota, channel, avatar, and freshness behavior.
- Keep raw transcript, customer contact, and audio out of analytics facts.

## 2. Non-Goals

- ML auto-classification or LLM topic inference.
- Tenant-defined taxonomy management.
- Multi-topic tagging per conversation.
- Platform-wide topic ranking.
- Customer-facing topic UX redesign.

## 3. Topic Policy

| Concern | Rule |
| --- | --- |
| Source | Prefer conversation summary `topic`; otherwise call/session topic if available; otherwise `unknown` |
| Normalization | trim, lowercase, replace spaces with `_`, max 48 chars |
| Known values | `general`, `billing`, `technical`; future values are allowed if normalized |
| Missing / invalid | `unknown` |
| Display label | UI maps common codes; unknown displays `Unknown / unset` |

## 4. Data Model

Extend ClickHouse `call_center_usage_facts` with:

| Column | Type | Default | Notes |
| --- | --- | --- | --- |
| `topic` | `String` | `'unknown'` | Normalized caller topic dimension |

Updated insert/query order keeps `tenant_id` first:

```sql
ALTER TABLE monti_jarvis.call_center_usage_facts
  ADD COLUMN IF NOT EXISTS topic String DEFAULT 'unknown';
```

Recommended order remains:

```text
ORDER BY (tenant_id, usage_date, call_id, fact_id)
```

No new Postgres table is required. The operational source remains
`conversation_records.summary.topic` where present.

## 5. API Summary

Extend existing tenant endpoint:

| Method | Path | Auth | Change |
| --- | --- | --- | --- |
| GET | `/api/tenant/call-center/statistics` | tenant_admin | Add `topic` query and `by_topic` response |

### Query

| Param | Default | Notes |
| --- | --- | --- |
| `start_date` | today | Existing `YYYY-MM-DD` |
| `end_date` | today | Existing `YYYY-MM-DD` |
| `topic` | all | Optional normalized code; filters all aggregate totals |

### Response Additions

```json
{
  "topic": "all",
  "by_topic": [
    {
      "topic": "billing",
      "label": "Billing",
      "completed": 12,
      "total_duration_seconds": 1840,
      "average_duration_seconds": 153.3
    },
    {
      "topic": "unknown",
      "label": "Unknown / unset",
      "completed": 2,
      "total_duration_seconds": 120,
      "average_duration_seconds": 60
    }
  ]
}
```

Existing `by_avatar`, `by_channel`, KPI totals, quota, call limits, and
freshness stay in the same response.

### Errors

| HTTP | Code | When |
| ---: | --- | --- |
| 400 | `validation_error` | invalid date range or topic code |
| 401 | `unauthorized` | no tenant admin session |
| 503 | `analytics_unavailable` | ClickHouse disabled/unavailable |
| 502 | `quota_unavailable` | quota snapshot failure, unchanged from S25 |

## 6. RBAC and Isolation

The handler derives `tenant_id` from tenant-admin auth. Requests cannot select a
tenant. ClickHouse queries must filter `tenant_id` before date/topic and return
aggregate-only rows.

## 7. UX Summary

- Add a topic select beside the date controls.
- Add a `By topic` breakdown card/table.
- Show share of total conversations and total duration per topic.
- Preserve channel/avatar cards and quota panel.
- Mobile stacks: KPIs, topic, channel, avatar, quota.

## 8. Verification

```bash
# API
curl -sS 'http://localhost:8091/api/tenant/call-center/statistics?start_date=2026-07-29&end_date=2026-07-29' \
  -H 'Authorization: Bearer <tenant-admin-token>' | jq '.by_topic'

curl -sS 'http://localhost:8091/api/tenant/call-center/statistics?topic=billing' \
  -H 'Authorization: Bearer <tenant-admin-token>' | jq .

# Tests
go test ./internal/clickhouse ./cmd/server -count=1 -run 'CallCenter|Topic'
cd apps/tenant-web && npm run check
```

Manual verification seeds two tenants with billing, technical, and missing-topic
facts, then confirms each tenant only sees its own topic totals.

## 9. Implementation Sequence

1. **TASK-0199** — schema/bootstrap, projection, and ClickHouse insert/query
   support for topic.
2. **TASK-0200** — tenant statistics handler/API response and validation.
3. **TASK-0201** — tenant dashboard UI and manual verification checklist.

## See Also

- Workflow §127–128
- ER Sprint 55
- API Sprint 55
- UX T55
- [SPRINT-055](../03-sprints/SPRINT-055.md)
