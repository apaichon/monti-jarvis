---
id: FEAT-0047
title: "Tenant call-center statistics grouped by topic"
status: shipped
release: v2.28.0
roadmap_sprint: 55
priority: P
depends_on: [SPRINT-022, SPRINT-025, SPRINT-030]
design: DES-0051
design_spec: ../02-design/51-tenant-call-center-topic-statistics-spec.md
updated: 2026-07-29
extends: FEAT-0027
---

# FEAT-0047: Tenant Call-Center Statistics Grouped by Topic

## Purpose

Extend the tenant Call Center dashboard so tenant operators can see which
caller topics drive conversation volume and call minutes for a selected date
range.

## Problem

Sprint 25 shows tenant activity by channel and AI employee, but not by the
caller topic already sent through chat and voice flows. Operators cannot tell
whether demand is billing, technical, general, or unset, so KM investment and
staffing decisions remain guesswork.

## Scope

### In

- Persist/project a normalized `topic` dimension into call-center analytics
  facts.
- Return tenant-scoped topic aggregates from
  `GET /api/tenant/call-center/statistics`.
- Add optional `topic` filtering on the tenant statistics endpoint.
- Show topic breakdown and filter on the tenant dashboard.
- Preserve existing S25 totals, channel breakdown, avatar breakdown, quota
  usage, and ClickHouse outage behavior.
- Bucket missing or invalid topic values as `unknown`.

### Out

- ML topic classification or auto-tagging.
- Tenant-managed custom taxonomy.
- Multi-select topics per conversation.
- Platform-wide topic leaderboard or benchmarking.
- Changing customer topic picker UX beyond ensuring the selected topic is
  captured in analytics.

## Acceptance Criteria

1. Completed chat and voice conversations with a selected topic appear in a
   `by_topic` aggregate for the owning tenant only.
2. Conversations without topic metadata appear under a stable `unknown` bucket.
3. `GET /api/tenant/call-center/statistics?topic=billing` filters totals,
   channel, avatar, and topic rows to billing activity for the date range.
4. Existing non-topic S25 dashboard metrics remain available and consistent.
5. Tenant dashboard renders a topic breakdown and topic filter with empty,
   unavailable, and mobile states.
6. Projection replay stays idempotent; adding topic does not double-count
   historical facts.

## Design Links

- Sprint: [SPRINT-055](../03-sprints/SPRINT-055.md)
- Deep spec: [DES-0051](../02-design/51-tenant-call-center-topic-statistics-spec.md)
- Base feature: [FEAT-0027](FEAT-0027-tenant-call-center-statistics.md)
- Base design: [DES-0028](../02-design/28-call-center-statistics-spec.md)

## Implementation Notes

- Implemented in branch `feature/sprint-055-topic-statistics`.
- Adds ClickHouse `topic` dimension with `unknown` fallback and idempotent
  fact IDs unchanged.
- Tenant API now returns `topic` and `by_topic`; `topic=` filters totals,
  channel, avatar, and topic rows.
- Tenant dashboard now supports All topics/topic filtering and shows topic
  count, duration, and share of completed conversations.
