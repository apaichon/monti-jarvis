---
id: SPRINT-032
status: planned
start: 2026-07-18
end: 2026-07-20
updated: 2026-07-18
design_pack: existing
release_target: v2.13.0
goal: "Close Sprint 31 reporting readiness gaps with controlled reconciliation fixtures and deferred platform billing usage UAT."
roadmap_sprint: 32
feature: FEAT-0033
platform: Platform
depends_on: [SPRINT-031]
---

# SPRINT-032 - Platform Billing Usage Readiness and Reconciliation

## Goal

Finish the operational readiness work carried over from Sprint 31: execute the deferred platform billing usage UAT and make Postgres, Redis, and ClickHouse reconciliation fixtures repeatable.

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §87 | `review_pending` |
| ER / fixture boundary | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 32 | `review_pending` |
| API verification contract | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 32 | `review_pending` |
| UX / UAT operator surface | [05-ux-ui.md](../02-design/05-ux-ui.md) Sprint 32 | `review_pending` |
| Billing usage deep spec | [DES-0034](../02-design/34-platform-billing-quota-ai-cost-spec.md) | `approved` |
| Production tuning deep spec | [DES-0035](../02-design/35-production-transport-cache-tuning-spec.md) | `review_pending` |

Sprint 32 introduces no new public endpoint, durable entity, or customer UI. DES-0035 defines the roadmap tuning track as design-only; the 5-point implementation commitment remains TASK-0144/TASK-0145.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S29, S30, S31) | 16, 16, 16 -> **avg 16** |
| **Committed** | **5** |

## Commitment

Sprint planning found no unassigned `proposed`/`approved` task files. The commitment is limited to the two explicit Sprint 31 carry-over tasks, preserving capacity for evidence-driven verification rather than adding new product scope:

| Work package | Points | Owner | Outcome |
| --- | ---: | --- | --- |
| [TASK-0144](../04-tasks/TASK-0144.md) Execute deferred Sprint 31 billing usage UAT | 2 | tester | Browser, responsive, source-failure, session-expiry, and two-tenant usage scenarios produce evidence in UAT-031 |
| [TASK-0145](../04-tasks/TASK-0145.md) Add controlled billing usage reconciliation fixtures | 3 | devops | Repeatable Postgres/Redis/ClickHouse fixtures verify date boundaries, mismatch, quota divergence, duplicate events, and measurement states |

**Committed:** 5 points · **Task IDs:** TASK-0144–TASK-0145.

## Scope boundary

**In**

- Deferred Sprint 31 UAT for `/admin/billing/usage`.
- Controlled, resettable cross-store fixtures and aggregate reconciliation evidence.
- Follow-up defects required to make the shipped reporting contract testable and operationally verifiable.

**Out**

- New billing, quota, payment, entitlement, AI metering, or dashboard product behavior.
- Customer-facing cost display, invoice behavior, quota policy changes, or production traffic expansion.

## Verification target

```bash
make test
make build
cd apps/platform-admin-web && npm run check && npm run build
git diff --check
```

Manual evidence must be attached to [SPRINT-031-manual.md](../06-manual-tests/SPRINT-031-manual.md), and fixture reset/scope safety must be documented with TASK-0145.

## Risks

| Risk | Mitigation |
| --- | --- |
| UAT remains deferred after another release | Make TASK-0144 the tester's primary commitment and require evidence links for completion. |
| Cross-store fixtures contaminate shared data | Use isolated tenant IDs, explicit reset commands, and read-only production-like authorities. |
| Reconciliation differences are mistaken for defects | Label historical ClickHouse range data separately from current Redis enforcement snapshots. |
