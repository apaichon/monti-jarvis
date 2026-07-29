---
id: SPRINT-051
status: backlog
implementation_status: verify
release_gate: manual_uat
start: 2026-08-15
end: 2026-08-21
updated: 2026-07-28
design_pack: approved
release_target: v2.25.0
roadmap_sprint: 51
feature: FEAT-0044
platform: Platform / Tenant / Finance
depends_on: [SPRINT-009, SPRINT-010, SPRINT-012, SPRINT-013, SPRINT-025, SPRINT-031, SPRINT-045, SPRINT-048, SPRINT-050]
goal: "Ship distinct Shared Cloud checkout and Dedicated VM quotation flows with authoritative current-plan quota, next-bill, scheduler, receipt, and tax-invoice contracts."
parallel_track: true
---

# SPRINT-051 — Shared Cloud and Dedicated VM Commercial Operations

## Status

Implementation and automated/integration verification are complete in the
isolated Sprint 51 worktree. The sprint remains in the roadmap backlog so it
does not replace the active parallel sprint. Manual browser UAT and production
scheduler enablement are still open; Sprint 51 is not released or closed.

## Commitment

| Task | Points | Status | Outcome |
| --- | ---: | --- | --- |
| [TASK-0190](../04-tasks/TASK-0190.md) | 3 | completed | Catalog modes, immutable versions, authoritative calculator |
| [TASK-0191](../04-tasks/TASK-0191.md) | 3 | completed | Dedicated quote request/review/activation without gateway |
| [TASK-0192](../04-tasks/TASK-0192.md) | 3 | completed | Idempotent renewal, retries, documents, settlement |
| [TASK-0193](../04-tasks/TASK-0193.md) | 3 | completed | Composite current-plan, quota freshness, next bill |
| [TASK-0194](../04-tasks/TASK-0194.md) | 2 | completed | Tenant/platform UX and automated verification |
| **Total** | **14** | **14/14 built** | **UAT pending** |

## Scope boundary

- Shared Cloud: monthly/annual self-service checkout and renewal.
- Dedicated VM: company/capacity quotation only; no tenant gateway.
- Current plan: package, period, next bill, quota, freshness, documents.
- Platform: Dedicated quote operations and safe billing-cycle retry.
- Out: new providers, automatic Dedicated gateway charging, overages, and
  automatic infrastructure provisioning.

## Verification

| Gate | Result |
| --- | --- |
| Go tests | 263 passed / 36 packages |
| Tenant check/build | passed; 0 diagnostics |
| Platform check/build | passed; 29 pre-existing warnings |
| Shared purchase smoke | paid, entitled, receipt + tax invoice |
| Dedicated smoke | full lifecycle; zero payment artifacts on request |
| Scheduler replay | one cycle/order; settled once after documents |
| Temp verification DB | removed after test |

Manual checklist:
[SPRINT-051-commercial-operations-manual.md](../06-manual-tests/SPRINT-051-commercial-operations-manual.md).
