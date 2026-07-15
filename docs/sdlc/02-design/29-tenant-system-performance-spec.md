---
id: DES-0029
title: Tenant System Performance Monitoring Specification
status: approved
updated: 2026-07-14
sprint: SPRINT-026
owner: SA
---

# Tenant System Performance Monitoring - Design Spec

**Sprint:** SPRINT-026 · **Release target:** v2.7.0  
**Feature:** [FEAT-0028](../01-features/FEAT-0028-tenant-system-performance-monitoring.md)  
**Depends on:** [28-call-center-statistics-spec.md](28-call-center-statistics-spec.md)

## 1. Goals

- Give an active tenant administrator a safe read-only view of the service path used by the tenant.
- Normalize dependency status and bounded probe latency into a stable response contract.
- Show ClickHouse analytics freshness separately from live dependency health.
- Keep monitoring out of customer call, voice relay, quota, archive, and liveness critical paths.
- Preserve tenant isolation and redact provider details, credentials, URLs, customer data, transcripts, and audio paths.

## 2. Non-goals (Sprint 26)

- Platform-wide or cross-tenant monitoring; planned for Sprint 28.
- Audit history; planned for Sprint 27.
- Alert delivery, paging, webhooks, billing, AI infrastructure cost, or persistent time-series storage.
- Replacing `/healthz`, `/api/infra`, ClickHouse analytics, or existing quota enforcement.
- Customer-facing health details.

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `MONITORING_PROBE_TIMEOUT` | `2s` | Maximum time for one dependency probe; invalid or non-positive values use the default. |
| `CLICKHOUSE_URL` | unset | Existing ClickHouse configuration; unset is reported as `disabled`. |
| `AUTH_DISABLED` | `true` | Does not bypass the tenant-admin guard on the monitoring route. |
| `TZ` | deployment timezone | Used only when formatting freshness timestamps for the tenant UI. |

The implementation may use the existing environment configuration and clients. No new infrastructure service is introduced.

## 4. Data model

Sprint 26 adds no Postgres table, migration, ClickHouse table, MinIO object, or Redis key. A monitoring snapshot is an ephemeral in-process value generated per request under a bounded probe timeout.

### Snapshot shape

| Field | Type | Notes |
| --- | --- | --- |
| `overall_status` | string | `operational`, `degraded`, or `unavailable`. |
| `checked_at` | RFC3339 timestamp | Time the snapshot was generated. |
| `components` | array | One normalized entry per configured dependency. |
| `components[].name` | string | Stable allowlisted name such as `postgres` or `clickhouse`. |
| `components[].status` | string | `operational`, `degraded`, `unavailable`, or `disabled`. |
| `components[].latency_ms` | integer or null | Present only when a bounded probe completed. |
| `components[].checked_at` | RFC3339 timestamp | Probe completion time. |
| `analytics` | object | Existing ClickHouse call-center freshness summary. |
| `analytics.status` | string | `current`, `stale`, `unavailable`, or `disabled`. |
| `analytics.last_projected_at` | RFC3339 timestamp or null | Existing projection freshness value; never includes source record details. |
| `analytics.generated_at` | RFC3339 timestamp or null | Existing statistics generation value when available. |

Status aggregation:

| Condition | Overall status |
| --- | --- |
| All configured components operational and analytics not stale/unavailable | `operational` |
| One or more components degraded, disabled, or analytics stale | `degraded` |
| A required configured dependency is unavailable or the snapshot cannot be produced | `unavailable` |

`disabled` means the dependency is intentionally not configured. It is not a probe failure. Raw errors stay in server logs and are converted to the normalized status contract before serialization.

## 5. Redis / NATS / ClickHouse / MinIO

| Store or service | Sprint 26 contract |
| --- | --- |
| Postgres | Use a bounded `Ping` probe only; no new table or query path. |
| Redis | Use a bounded `PING` probe only; do not create monitoring keys. |
| MinIO | Use the existing bucket existence probe; do not list objects or expose bucket names. |
| ClickHouse | Use the existing `Ping` and call-center freshness result; no new table. |
| NATS | Report configured/enabled state through the existing bus; no new subject. |
| LiveKit | Report configured state; no customer room or token data. |
| Gemini | Report configured/enabled state; no prompt, transcript, or provider response data. |
| MinIO objects | No new object path. |

Probes run concurrently with a shared deadline, have no more than one in-flight probe per component for a snapshot, and are never called from the live call close/archive path.

## 6. API summary

See [04-api-spec.md](04-api-spec.md) section `Tenant System Performance (Sprint 26)`.

| Method | Path | Role |
| --- | --- | --- |
| `GET` | `/api/tenant/system-performance` | `tenant_admin` on an active tenant |

The request has no tenant selector and no query parameters in Sprint 26. The handler derives tenant context from the authenticated session. `AUTH_DISABLED=true` keeps customer/demo routes compatible but does not make this endpoint public.

## 7. RBAC

| Action | `platform_admin` | `tenant_admin` | `customer` | Anonymous |
| --- | --- | --- | --- | --- |
| Read active tenant performance snapshot | no Sprint 26 route | yes, active tenant only | no | no |
| Read another tenant snapshot | no | no | no | no |
| Trigger or configure probes | operator process only | no | no | no |

The response must not include a tenant id, dependency URL, provider error, customer identifier, transcript content, rating comment, ticket note, or audio object path. A missing or invalid tenant context returns the existing `401`/`403` shape without revealing whether another tenant exists.

## 8. Verification

```bash
make test
make build
cd apps/tenant-web && npm run check && npm run build

curl -i 'http://localhost:8091/api/tenant/system-performance' \
  -H 'Authorization: Bearer <tenant-admin-token>'
```

Expected checks:

- The response contains only the allowlisted component names and normalized statuses.
- Probe timeout completes within the configured bounded deadline and returns `degraded` or `unavailable` without blocking calls.
- ClickHouse freshness is `stale` when projections lag the documented threshold and is separate from live dependency status.
- Tenant A cannot read Tenant B state, and unauthorized callers receive no monitoring metadata.
- Existing `/healthz`, `/api/infra`, customer calls, archive writes, quota checks, and call-center statistics remain compatible.

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | 2026-07-14 | [x] |
| Dev | | 2026-07-14 | [x] |
