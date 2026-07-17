---
id: SPRINT-030
status: completed
start: 2026-07-17
end: 2026-07-18
updated: 2026-07-17
closed: 2026-07-17
design_pack: approved
release_target: v2.11.0
release: v2.11.0
goal: "Platform: provide date-filtered overall call-center statistics with a tenant breakdown from ClickHouse analytics."
roadmap_sprint: 30
feature: FEAT-0032
platform: Platform
depends_on: [SPRINT-025, SPRINT-029]
---

# SPRINT-030 - Platform: Overall Call Center Statistics by Tenant

## Goal

Give platform operators a reliable, date-filtered view of completed AI conversation activity across tenants while preserving the aggregate-only and tenant-safe contract established by Sprint 25.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S27, S28, S29) | 16, 16, 16 -> **avg 16** |
| **Committed** | **16** |
| **Completed** | **16** |

## Commitment

No unassigned `proposed`/`approved` task files were available. The roadmap item is decomposed into four packages totaling the established 16-point average:

| Work package | Points | Owner | Outcome |
| --- | ---: | --- | --- |
| [TASK-0136](../04-tasks/TASK-0136.md) Platform analytics projection and aggregate contract | 5 | devops | Reuse/extend ClickHouse call-center facts for platform-wide and per-tenant aggregates with date and freshness semantics |
| [TASK-0137](../04-tasks/TASK-0137.md) Platform statistics API | 4 | dev | Platform-admin authorization, inclusive date filters, bounded tenant pagination, redaction, and safe analytics errors |
| [TASK-0138](../04-tasks/TASK-0138.md) Platform statistics dashboard | 5 | dev | `/admin/call-center` dashboard with KPI summary, tenant breakdown, filters, empty/unavailable/retry states, and responsive layout |
| [TASK-0139](../04-tasks/TASK-0139.md) Verification and operations | 2 | tester | Aggregate correctness, tenant isolation, date boundaries, monitoring smoke carry-over, build checks, and UAT checklist |

**Committed:** 16 points · **Implementation completed:** 16 points · **Task IDs:** TASK-0136–TASK-0139.
Manual browser, responsive-layout, and dependency-failure UAT remains deferred to the next tester run.

## Scope boundary

**In**

- Platform-admin read-only aggregate dashboard backed by existing ClickHouse conversation facts.
- Today-default inclusive start/end date filters and bounded tenant breakdown pagination.
- Completed conversation count, total talk time, average conversation duration, chat/voice totals, avatar totals, satisfaction summary, and daily package usage.
- Tenant-safe aggregate response fields, freshness metadata, explicit ClickHouse unavailable/stale handling, and responsive platform UI.
- Regression coverage for Sprint 29 monitoring/audit health and Sprint 25 tenant statistics compatibility.

**Out**

- Billing reconciliation, invoice data, quota enforcement changes, or AI infrastructure cost allocation.
- Raw conversation records, transcripts, customer identifiers, audio paths, ticket notes, or per-customer drill-down.
- Persistent dashboard snapshots, scheduled reports, exports, alerts, or new mobile APIs.

## Dependencies and design gates

1. Reuse the Sprint 25 fact definition and inclusive date boundary contract; do not create a second usage authority.
2. Define aggregate query shapes, tenant pagination, freshness semantics, and ClickHouse failure behavior in the Sprint 30 design pack.
3. Define platform RBAC and redaction for aggregate and tenant rows before implementation.
4. Carry Sprint 29 monitoring smoke coverage into verification so analytics unavailability remains diagnosable without raw infrastructure leakage.

## Design pack

The Sprint 30 technical design pack was approved for implementation and shipped with this release.

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0032 - Platform Call Center Statistics by Tenant](../01-features/FEAT-0032-platform-call-center-statistics.md) | `completed` |
| Deep spec | [33-platform-call-center-statistics-spec.md](../02-design/33-platform-call-center-statistics-spec.md) | `shipped` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §84 | `shipped` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 30 | `shipped` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Platform Call Center Statistics | `shipped` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) A22 | `shipped` |

Implementation is gated on approval of the deep spec, API contract, and aggregate metric definitions.

## Verification

```bash
make test
make build
cd apps/platform-admin-web && npm run check && npm run build
git diff --check
# Platform admins receive correct aggregate and tenant-level totals for a controlled date range.
# Today defaults correctly and date boundaries are inclusive in the platform timezone.
# Tenant pagination and filters do not leak records or customer-level data.
# Empty, stale, unavailable, unauthorized, and retry states are explicit and safe.
# Existing tenant statistics, platform monitoring, audit log, calls, and quota paths remain compatible.
# Automated verification passed; manual browser, responsive-layout, and dependency-failure UAT is deferred:
# docs/sdlc/06-manual-tests/SPRINT-030-manual.md
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Platform totals disagree with tenant dashboard totals | Reuse the same facts, date boundaries, and metric definitions; verify against controlled fixtures |
| Cross-tenant aggregate leaks customer data | Keep ClickHouse queries aggregate-only and enforce an allowlisted API response model |
| Large tenant counts slow the dashboard | Use bounded pagination and separate aggregate queries from tenant-page queries |
| ClickHouse lag is mistaken for zero activity | Return freshness state and render stale/unavailable separately from an empty result |

## Links

- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Sprint 30
- Feature: [FEAT-0032](../01-features/FEAT-0032-platform-call-center-statistics.md)
- Depends: [SPRINT-025](SPRINT-025.md), [SPRINT-029](SPRINT-029.md)
- Next: Sprint 31 platform billing, quota, and AI infrastructure cost usage

## Release close

Automated Go tests, server build, platform-admin type checks/build, aggregate contract tests, authorization/error checks, and `git diff --check` passed. Manual browser, responsive-layout, and dependency-failure UAT remains deferred and is tracked in [SPRINT-030-manual.md](../06-manual-tests/SPRINT-030-manual.md).

**Closed:** 2026-07-17 as release `v2.11.0`.
