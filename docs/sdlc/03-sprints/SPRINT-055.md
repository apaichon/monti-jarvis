---
id: SPRINT-055
status: completed
closed: 2026-07-29
start: 2026-07-29
end: 2026-08-05
updated: 2026-07-29
design_pack: shipped
roadmap_sprint: 55
feature: FEAT-0047
platform: Tenant
depends_on: [SPRINT-022, SPRINT-025, SPRINT-030]
goal: "Tenant call-center statistics can group and filter completed conversations by caller topic."
worktree: .worktrees/SPRINT-055
branch: feature/sprint-055-topic-statistics
release_target: v2.28.0
release: v2.28.0
velocity_basis: "Last 3 closed: S51=14, S53=12, S54=12 -> avg 12.7; commit 12"
---

# SPRINT-055 — Tenant Call-Center Statistics Grouped by Topic

## Goal

Extend the tenant Call Center dashboard with topic-level analytics so tenant
operators can see whether demand is coming from general, billing, technical, or
unknown/unset conversations for a selected date range.

## Worktree

| Item | Value |
| --- | --- |
| Path | `.worktrees/SPRINT-055` |
| Branch | `feature/sprint-055-topic-statistics` |
| Base | `origin/main` |

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S51, S53, S54) | 14, 12, 12 -> **avg 12.7** |
| **Committed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0199](../04-tasks/TASK-0199.md) Project topic dimension into call-center analytics facts | 4 | completed | dev | Topic on ClickHouse facts; unknown fallback |
| [TASK-0200](../04-tasks/TASK-0200.md) Tenant call-center statistics topic API | 4 | completed | dev | `by_topic` aggregate + topic filter |
| [TASK-0201](../04-tasks/TASK-0201.md) Tenant dashboard topic breakdown UX and verification | 4 | completed | dev/tester | Topic UI, empty states, isolation UAT |
| **Total** | **12** | **12 completed** | | |

## Design Pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0047](../01-features/FEAT-0047-tenant-call-center-topic-statistics.md) | shipped |
| Deep spec | **`shipped`** | [DES-0051](../02-design/51-tenant-call-center-topic-statistics-spec.md) |
| Workflow | **`shipped`** | [02-workflow.md](../02-design/02-workflow.md) §127–128 |
| ER | **`shipped`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 55 |
| API | **`shipped`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 55 |
| UX | **`shipped`** | [05-ux-ui.md](../02-design/05-ux-ui.md) T55 |

## Scope Boundary

### In

- Add normalized topic dimension to the call-center analytics projection.
- Return `by_topic` aggregates from the tenant call-center statistics endpoint.
- Add optional topic filter to the existing tenant stats endpoint.
- Update tenant dashboard with topic breakdown and topic filter.
- Preserve existing S25 KPIs, channel/avatar breakdowns, quota panel, and
  unavailable states.

### Out

- ML auto-topic classification.
- Tenant-managed topic taxonomy.
- Platform-wide topic leaderboard.
- Multi-label topics per conversation.
- Customer topic picker redesign.

## Verification Target

```bash
go test ./internal/clickhouse ./cmd/server -count=1 -run 'CallCenter|Topic'
cd apps/tenant-web && npm run check
# Manual: seed billing + technical + missing topic; verify topic rows,
# topic filter, date filter, unknown bucket, and tenant isolation.
```

## Implementation Status

- Backend projects `summary.topic` into `call_center_usage_facts.topic` with
  `unknown` fallback.
- Tenant statistics API returns `topic` and `by_topic`, and supports exact topic
  filtering.
- Tenant dashboard includes topic select, topic KPI, topic breakdown, and share
  of completed conversations.
- Validation completed on 2026-07-29:
  - `go test ./internal/clickhouse ./cmd/server -count=1`
  - `cd apps/tenant-web && npm run check`
  - `cd apps/tenant-web && npm run build`

## Shipped summary (v2.28.0)

- ClickHouse call-center fact table adds `topic String DEFAULT 'unknown'`.
- Conversation analytics projection records normalized topic without changing
  fact idempotency.
- Tenant call-center statistics endpoint returns `topic` and `by_topic`, and
  filters totals/channel/avatar/topic rows by `topic=`.
- Tenant dashboard includes all/topic filter, topic KPI, and topic breakdown
  with completed count, duration, and share.
- Design: DES-0051 · FEAT-0047 · TASK-0199–0201.
- Manual UAT checklist:
  [SPRINT-055-topic-statistics-manual.md](../06-manual-tests/SPRINT-055-topic-statistics-manual.md).

**Closed:** 2026-07-29

## Risks

| Risk | Mitigation |
| --- | --- |
| Historical facts lack topic | Normalize missing values to `unknown`; replay can backfill when source summary has topic |
| ClickHouse schema drift | Add topic to bootstrap and insert query together |
| Dashboard clutter | Keep topic section compact; existing KPI/channel/avatar surfaces remain |
| Topic injection/free text | Normalize to lowercase allowlist-ish code; bound length and fallback invalid values |
