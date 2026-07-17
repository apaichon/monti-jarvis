---
id: FEAT-0032
title: "Platform Call Center Statistics by Tenant"
status: completed
owner: product
created: 2026-07-17
updated: 2026-07-17
sprint: SPRINT-030
---

# FEAT-0032: Platform Call Center Statistics by Tenant

## Purpose

Give platform administrators a tenant-safe overview of completed AI conversations, duration, channels, avatars, satisfaction, and package usage across a selected date range.

## Scope

- Reuse the tenant call-center facts and ClickHouse projection from Sprint 25.
- Provide aggregate totals across all active tenants and a paginated tenant breakdown.
- Support inclusive start/end date filters with today as the default range in the platform timezone.
- Show completed conversation count, total talk time, average conversation time, channel totals, avatar totals, satisfaction summary, and daily package usage.
- Preserve platform-admin authorization, tenant visibility rules, bounded pagination, and safe empty/unavailable states.

## Out of scope

- Billing reconciliation, quota enforcement changes, or AI infrastructure cost allocation; planned for Sprint 31.
- Raw transcripts, customer contact data, audio paths, ticket notes, or individual customer records.
- Historical time-series charts, alerting, scheduled exports, or tenant-admin changes.

## Acceptance criteria

1. Platform administrators can open a date-filtered dashboard with today selected by default.
2. Aggregate KPIs match the tenant-scoped call-center facts for the selected inclusive date range.
3. A tenant breakdown supports bounded pagination and exposes only allowlisted tenant identity and aggregate fields.
4. Empty, invalid-range, stale, unavailable, unauthorized, and retry states are explicit and do not leak infrastructure details.
5. Existing tenant statistics, platform monitoring, audit log, calls, and quota enforcement remain compatible.
6. The dashboard remains readable at desktop and mobile widths without overlapping metrics or controls.

## Design links

- Sprint: [SPRINT-030](../03-sprints/SPRINT-030.md)
- Deep spec: [DES-0033 - Platform Call Center Statistics](../02-design/33-platform-call-center-statistics-spec.md)
- [API specification](../02-design/04-api-spec.md) - Platform Call Center Statistics
