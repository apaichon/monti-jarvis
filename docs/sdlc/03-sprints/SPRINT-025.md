---
id: SPRINT-025
status: completed
start: 2026-07-14
end: 2026-07-14
updated: 2026-07-14
closed: 2026-07-14
design_pack: approved
release_target: v2.6.0
release: v2.6.0
goal: "Tenant: provide date-filtered call-center statistics and quota usage from tenant-scoped ClickHouse analytics."
roadmap_sprint: 25
platform: Tenant
depends_on: [SPRINT-022]
---

# SPRINT-025 - Tenant: Call Center Statistics and Quota Usage

## Goal

Give tenant operators a reliable dashboard for call-center activity and package quota consumption, with today as the default date range and ClickHouse-backed tenant isolation.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S20, S23, S24) | 16, 16, 16 -> **avg 16** |
| **Commitment** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0116](../04-tasks/TASK-0116.md) | 5 | completed | devops | ClickHouse tenant call-center analytics projection |
| [TASK-0117](../04-tasks/TASK-0117.md) | 5 | completed | dev | Tenant call-center statistics and quota usage API |
| [TASK-0118](../04-tasks/TASK-0118.md) | 4 | completed | dev | Tenant call-center dashboard and quota usage UI |
| [TASK-0119](../04-tasks/TASK-0119.md) | 2 | completed | tester | Dashboard UAT, tenant isolation, and performance checks |

**Committed:** 16 points · **Completed:** 16 points

## Scope boundary

**In**

- Tenant-scoped ClickHouse analytics facts and replay/backfill contract.
- Date-filtered call/session counts, chat/voice totals, duration/minutes, and avatar breakdowns.
- Tenant quota used/remaining view joined to the existing entitlement and usage contract.
- Today as the default date range, explicit date range controls, and safe empty/error states.
- Tenant isolation, query performance checks, and manual UAT evidence.

**Out**

- Platform-wide or cross-tenant dashboards; planned for Sprint 30.
- System performance monitoring; planned for Sprint 26.
- Billing, AI infrastructure cost, and overage analytics; planned for Sprint 31.
- Raw transcript, customer contact, or audio data in ClickHouse dashboard facts.

## Feature

- [FEAT-0027 - Tenant Call Center Statistics and Quota Usage](../01-features/FEAT-0027-tenant-call-center-statistics.md)

## Design

The Sprint 25 technical design pack is approved for implementation.

| Artifact | Planned scope | Status |
| --- | --- | --- |
| Deep spec | [28-call-center-statistics-spec.md](../02-design/28-call-center-statistics-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §73-75 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 25 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Tenant Call Center Statistics | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § Sprint 25 | `approved` |

## Verification

```bash
make test
make build
cd apps/tenant-web && npm run build
# Tenant dashboard defaults to today and supports start/end filters.
# Metrics match controlled conversation and quota fixtures.
# Cross-tenant access returns 403/404 without metadata leakage.
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-025-manual.md
```

Release validation passed on 2026-07-14 after analytics decoding, completed-call metadata preservation, and expired-session redirect hardening were verified against the tenant dashboard flow.

## Risks

| Risk | Mitigation |
| --- | --- |
| ClickHouse projection lags source records | Expose freshness metadata and retain replay/backfill support |
| Usage totals disagree with existing quota counters | Define one authoritative calculation per metric and verify against fixtures |
| Date boundaries differ by tenant timezone | Resolve dates using the deployment timezone contract and test boundary timestamps |
| Analytics facts leak customer data | Keep facts aggregate-oriented and reject raw transcript/contact fields |

## Links

- Depends: [SPRINT-022](SPRINT-022.md)
- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Sprint 25
- Target: **v2.6.0**
