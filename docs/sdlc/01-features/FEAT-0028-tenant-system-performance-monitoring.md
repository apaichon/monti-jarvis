---
id: FEAT-0028
title: "Tenant System Performance Monitoring"
status: in_progress
owner: product
created: 2026-07-14
updated: 2026-07-14
sprint: SPRINT-026
---

# FEAT-0028: Tenant System Performance Monitoring

## Purpose

Give tenant administrators a safe operational view of service availability, dependency health, response latency, and call-center analytics freshness.

## Scope

- Probe configured dependencies through the existing server process with bounded timeouts.
- Normalize provider and infrastructure results into tenant-safe status values.
- Expose the active tenant's monitoring snapshot only to an authenticated active tenant administrator.
- Show component status, latency, last checked time, analytics freshness, and retry guidance.
- Preserve the existing customer call, archive, quota, and analytics paths when a dependency is degraded.

## Out of scope

- Platform-wide or cross-tenant monitoring; planned for Sprint 28.
- Audit history and operator action logs; planned for Sprint 27.
- Alert delivery, paging, billing/cost analytics, or persistent time-series metrics.
- Raw provider errors, credentials, network addresses, transcripts, customer contact data, or audio paths.

## Acceptance criteria

1. An active tenant administrator can open a monitoring view and receive a normalized snapshot for the active tenant context.
2. The snapshot distinguishes `operational`, `degraded`, `unavailable`, `disabled`, and `stale` states and includes bounded latency where a probe completed.
3. ClickHouse analytics freshness is shown separately from live dependency health.
4. Missing auth, inactive tenants, invalid tenant context, and cross-tenant access return safe unauthorized/forbidden responses without metadata leakage.
5. A dependency timeout does not block call creation, voice relay, quota enforcement, archive writes, or `/healthz`.
6. The tenant UI provides loading, healthy, degraded, unavailable, stale, and retry states at desktop and mobile widths.

## Design links

- Sprint: [SPRINT-026](../03-sprints/SPRINT-026.md)
- [DES-0029 - Tenant System Performance Monitoring](../02-design/29-tenant-system-performance-spec.md)
- [API specification](../02-design/04-api-spec.md) - Tenant System Performance
