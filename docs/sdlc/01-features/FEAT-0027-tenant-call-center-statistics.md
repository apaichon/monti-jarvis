---
id: FEAT-0027
title: "Tenant Call Center Statistics and Quota Usage"
status: shipped
owner: product
created: 2026-07-14
updated: 2026-07-14
sprint: SPRINT-025
---

# FEAT-0027: Tenant Call Center Statistics and Quota Usage

## Purpose

Give tenant operators a date-filtered dashboard for call-center activity and package quota consumption, backed by tenant-scoped ClickHouse analytics.

## Scope

- Project tenant-safe conversation and call usage facts into ClickHouse analytics storage.
- Show total conversations, chat/voice breakdown, call minutes, avatar breakdown, and quota usage.
- Default the dashboard range to today in the tenant deployment timezone.
- Allow an explicit start date and end date with clear empty and error states.
- Preserve tenant isolation for every projection, query, and UI response.

## Out of scope

- Platform-wide or cross-tenant dashboards; delivered by FEAT-0032 in Sprint 30.
- System performance monitoring; planned for Sprint 26.
- Billing, AI infrastructure cost, or overage analytics; planned for Sprint 31.
- Raw transcript, contact, or audio data in ClickHouse dashboard facts.

## Acceptance criteria

1. Tenant users can open a dashboard with today's date range selected by default.
2. Tenant users can change the start and end dates and receive consistent statistics.
3. Dashboard metrics include call/session count, channel totals, duration/minutes, avatar breakdown, and quota used/remaining.
4. Analytics queries and projection jobs enforce tenant scope and do not expose customer contact or transcript data.
5. Empty ranges, invalid ranges, ClickHouse unavailability, and unauthorized access produce explicit safe UI states.
6. Existing tenant settings usage and satisfaction statistics remain compatible.

## Design links

- Sprint: [SPRINT-025](../03-sprints/SPRINT-025.md)
- [DES-0028 - Tenant Call Center Statistics and Quota Usage](../02-design/28-call-center-statistics-spec.md)
- [API specification](../02-design/04-api-spec.md) - Tenant Call Center Statistics
