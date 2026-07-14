---
id: SPRINT-021
status: completed
start: 2026-07-13
end: 2026-07-14
updated: 2026-07-14
design_pack: approved
release_target: v2.2.0
goal: "Customer: require OTP before AI workforce selection when tenant policy demands it, then enforce customer-aware call time and quota limits."
roadmap_sprint: 21
platform: Customer
depends_on: [SPRINT-001, SPRINT-005, SPRINT-013, SPRINT-016, SPRINT-020]
---

# SPRINT-021 - Customer: Authenticated Workforce Selection and Quota Limits

## Goal

Make the customer conversation entry production-safe by allowing tenants to require OTP before AI workforce selection and by enforcing per-customer time/quota limits across chat and voice.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S18-S20) | 16, 16, 16 -> **avg 16** |
| **Commitment** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0097](../04-tasks/TASK-0097.md) | 3 | completed | dev | Tenant policy for required customer auth before workforce selection |
| [TASK-0098](../04-tasks/TASK-0098.md) | 4 | completed | dev | Customer portal OTP gate and workforce picker flow |
| [TASK-0099](../04-tasks/TASK-0099.md) | 4 | completed | dev | Customer-aware call/chat quota and duration enforcement |
| [TASK-0100](../04-tasks/TASK-0100.md) | 3 | completed | dev | Tenant quota/settings UI and API refinements |
| [TASK-0101](../04-tasks/TASK-0101.md) | 2 | completed | tester | Optional/required auth and quota UAT evidence |

**Committed:** 16 points

## Scope boundary

**In**

- Required-auth tenant mode for customer portal workforce selection.
- Customer session propagation through workforce selection, chat, and voice call creation.
- Available workforce filtering for tenant-assigned active avatars.
- Per-customer daily call time and per-call duration limit enforcement.
- Customer-facing quota/blocked states.

**Out**

- Password/OAuth customer auth.
- Ticketing, satisfaction surveys, conversation history, or human handoff.
- New package SKUs or billing rules.

## Feature

- [FEAT-0023 - Authenticated Workforce Selection and Customer Quota Enforcement](../01-features/FEAT-0023-authenticated-workforce-selection.md)

## Design pack

| Artifact | Path | Status |
| --- | --- | --- |
| Deep spec | [24-authenticated-workforce-selection-spec.md](../02-design/24-authenticated-workforce-selection-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §64–65 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 21 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Authenticated Workforce Selection & Customer Quota | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § Sprint 21 | `approved` |

> **Closed:** 2026-07-14 as release v2.2.0 scope. Frontend checks/builds passed; Go validation was blocked by a local incomplete Go toolchain (`go: no such tool "vet"`).

## Verification

```bash
make test && make build
# Tenant A optional auth: no-auth conversation still works.
# Tenant B required auth: workforce selection blocked until OTP sign-in.
# Exhausted customer quota blocks chat/voice with safe errors.
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-021-manual.md
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Regressing public no-auth tenants | Keep auth mode tenant-scoped and test optional mode |
| Customer quota charged to wrong tenant/customer | Use session-derived tenant/customer context for counters |
| Workforce picker hides valid avatars | Test active/disabled/assigned/unassigned avatar permutations |
| Poor UX on short screens | Keep popup workforce picker and quota status visible at 100% scale |

## Links

- Depends: [SPRINT-001](SPRINT-001.md), [SPRINT-005](SPRINT-005.md), [SPRINT-013](SPRINT-013.md), [SPRINT-016](SPRINT-016.md), [SPRINT-020](SPRINT-020.md)
- Next: SPRINT-022 Conversation Records and Knowledge Gaps
- Target: **v2.2.0**
