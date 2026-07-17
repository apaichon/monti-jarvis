---
id: FEAT-0031
title: "Platform System Performance Monitoring"
status: completed
owner: product
created: 2026-07-16
updated: 2026-07-17
sprint: SPRINT-029
---

# FEAT-0031: Platform System Performance Monitoring

## Purpose

Give platform administrators a bounded cross-tenant view of Monti service health, dependency latency, tenant analytics freshness, and audit delivery state without exposing raw infrastructure details or changing customer operations.

## Scope

- Normalize shared Postgres, Redis, MinIO, ClickHouse, NATS, LiveKit, and Gemini health into allowlisted statuses.
- Show tenant-safe analytics freshness and audit delivery state for active tenants.
- Provide platform-admin filtering by tenant and derived health state with bounded pagination.
- Add a responsive platform monitoring screen with loading, partial failure, stale, unavailable, unauthorized, and retry states.
- Reuse the existing tenant monitoring probe and analytics contracts where compatible.

## Out of scope

- Alerting, paging, email, SMS, webhook, or SIEM integrations.
- Persistent time-series metrics, historical SLO charts, billing, quota, or AI infrastructure cost dashboards.
- Customer-facing or tenant-admin monitoring surfaces.
- Raw provider errors, dependency URLs, credentials, customer data, transcripts, audio paths, or local audit spool contents.

## Acceptance criteria

1. Platform administrators can query a safe cross-tenant snapshot with shared dependency health and paginated tenant rows.
2. Tenant rows expose only tenant identity allowed to platform admins, derived status, analytics freshness, and audit delivery summary.
3. Probe timeout and concurrency are bounded; monitoring never runs from customer call, voice, quota, archive, or chat critical paths.
4. Dependency failures and ClickHouse freshness states are normalized without returning provider messages or infrastructure topology.
5. Non-platform callers receive the existing unauthorized/forbidden contract and no monitoring data.
6. The platform UI renders operational, degraded, unavailable, disabled, stale, empty, loading, and retry states without overlap at desktop or mobile widths.

## Design links

- [DES-0032 - Platform System Performance Monitoring Specification](../02-design/32-platform-system-performance-spec.md)
- [04-api-spec.md](../02-design/04-api-spec.md) Sprint 29 section
- [02-workflow.md](../02-design/02-workflow.md) Sprint 29 section
- [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 29 storage contract
- [05-ux-ui.md](../02-design/05-ux-ui.md) Sprint 29 platform surface
