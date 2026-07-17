---
id: DES-0032
title: Platform System Performance Monitoring Specification
status: shipped
updated: 2026-07-17
sprint: SPRINT-029
owner: SA
---

# Platform System Performance Monitoring - Design Spec

Sprint: SPRINT-029 · Release target: v2.10.0
Feature: FEAT-0031 - Platform System Performance Monitoring
Depends on: 29-tenant-system-performance-spec.md, 31-cross-tenant-audit-log-spec.md

## 1. Goals

- Give platform administrators a cross-tenant view of shared dependency health and tenant analytics freshness.
- Reuse existing bounded probes and normalized status values from Sprint 26.
- Keep audit delivery health visible without exposing spool paths, marker contents, or raw infrastructure errors.
- Keep monitoring request-time and read-only so it cannot affect customer call, voice relay, quota, archive, or chat paths.
- Make platform authorization, tenant filtering, pagination, and redaction explicit.

## 2. Non-goals

- Persistent monitoring history or time-series storage.
- Alert delivery, paging, webhooks, SIEM, billing, quota, or AI cost reporting.
- Changing tenant-admin monitoring behavior or customer-facing UI.
- Returning credentials, dependency URLs, raw provider errors, customer identifiers, transcripts, audio paths, or audit file contents.

## 3. Environment and runtime contract

| Variable | Default | Contract |
| --- | --- | --- |
| `MONITORING_PROBE_TIMEOUT` | `2s` | Shared upper bound for one platform snapshot's dependency probes and analytics reads. Invalid or non-positive values use the default. |
| `CLICKHOUSE_URL` | existing configuration | Enables analytics freshness checks when configured; disabled ClickHouse is represented as `disabled`. |
| `AUDIT_LOG_MODE` | `spool` | Existing Sprint 28 mode used to report audit delivery state; no new audit configuration is introduced. |

Snapshots are request-time values. Sprint 29 does not add a Redis cache, Postgres table, ClickHouse table, MinIO object, or background polling loop. A future cache may be added only with an explicit freshness contract.

## 4. Snapshot model

### Shared components

The shared component list is sorted by stable name and uses the Sprint 26 contract:

| Field | Type | Notes |
| --- | --- | --- |
| `name` | string | `postgres`, `redis`, `minio`, `clickhouse`, `nats`, `livekit`, or `gemini`. |
| `status` | string | `operational`, `degraded`, `unavailable`, or `disabled`. |
| `latency_ms` | integer or null | Only for probes that actually run; configuration-only checks return null. |
| `checked_at` | RFC3339 timestamp | Snapshot check time; no provider timestamp is exposed. |

### Tenant row

| Field | Type | Notes |
| --- | --- | --- |
| `tenant_id` | string | Platform-admin-visible tenant identifier. |
| `slug` | string | Tenant slug for operator navigation. |
| `name` | string | Tenant display name. |
| `status` | string | Derived `operational`, `degraded`, or `unavailable`; not the tenant lifecycle status. |
| `analytics` | object | `current`, `stale`, `unavailable`, or `disabled`, with bounded freshness timestamps. |
| `audit_status` | string | Derived global audit delivery label for this tenant row; detailed delivery counts are top-level only. |

Tenant status is `unavailable` when a required shared component is unavailable or the tenant analytics read fails; `degraded` when analytics is stale, audit delivery is degraded, or an optional component is disabled; otherwise `operational`.

### Aggregate response

The response includes `overall_status`, `checked_at`, shared `components`, aggregate counts, a top-level redacted `audit` health object, and a bounded `tenants` page. `overall_status` is `unavailable` if any required shared component is unavailable, `degraded` if any tenant row or optional dependency is degraded, and `operational` only when all returned state is operational.

## 5. Workflow and execution bounds

1. The platform-admin guard authenticates the caller before any tenant or dependency query.
2. The handler validates filters and clamps `limit` to 1–100.
3. A shared snapshot context is created with `MONITORING_PROBE_TIMEOUT`.
4. Shared dependency probes run concurrently through the existing `internal/observability` service.
5. Active tenant metadata is read from Postgres with a bounded list query.
6. Tenant analytics freshness is read from the existing ClickHouse call-center projection, without transcript or call record details.
7. Audit state is read from the existing in-process audit writer health contract.
8. Only allowlisted fields are serialized; raw errors are logged server-side and mapped to normalized statuses.

No snapshot work is invoked from customer request handlers, live audio relay, quota checks, archive writes, or chat orchestration.

## 6. API summary

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/platform/system-performance` | `platform_admin` | Read a bounded cross-tenant health snapshot. |

Query parameters:

| Field | Type | Default | Contract |
| --- | --- | --- | --- |
| `tenant_id` | string | empty | Exact tenant filter; no fuzzy matching. |
| `status` | string | empty | Derived tenant status: `operational`, `degraded`, or `unavailable`. |
| `limit` | integer | `50` | Maximum `100`. |
| `offset` | integer | `0` | Non-negative bounded offset. |

Response shape and errors are defined in [04-api-spec.md](04-api-spec.md) Sprint 29.

## 7. RBAC and redaction

| Action | `platform_admin` | `tenant_admin` | `customer` | Anonymous |
| --- | --- | --- | --- | --- |
| Read cross-tenant snapshot | yes | no | no | no |
| Filter by tenant | yes | no | no | no |
| Configure probes | no | no | no | no |

`401` and `403` responses use the existing safe error shape. The API never returns provider messages, credentials, URLs, stack traces, customer data, transcripts, audio paths, audit event metadata, or local spool file names.

## 8. Verification

```bash
go test ./...
go build ./cmd/server
cd apps/platform-admin-web && npm run check && npm run build
git diff --check

curl -i 'http://localhost:8091/api/platform/system-performance?limit=50' \
  -H 'Authorization: Bearer <platform-admin-token>'
```

Expected checks:

- Platform admins receive shared component health and bounded tenant rows.
- Tenant filters and pagination do not leak records outside the requested result.
- Probe and analytics work completes within the configured timeout.
- Disabled, degraded, unavailable, and stale states remain distinct.
- Audit state contains only safe counts/status and no local paths or event payloads.
- Tenant-admin and anonymous requests receive unauthorized/forbidden responses.
- Existing `/healthz`, tenant monitoring, customer calls, quota, archive, and audit delivery remain compatible.

See [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), and [05-ux-ui.md](05-ux-ui.md).
