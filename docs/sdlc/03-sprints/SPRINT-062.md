---
id: SPRINT-062
status: completed
start: 2026-08-29
end: 2026-09-08
updated: 2026-07-30
design_pack: approved
roadmap_sprint: 62
feature: FEAT-0054
platform: Tenant / Growth / Quota
depends_on: [SPRINT-013, SPRINT-045, SPRINT-046, SPRINT-051]
goal: "Tenant can redeem a referral code for bonus quota with validation, ledger tracking, and platform reverse."
velocity_basis: "avg 12; commit 12"
release_target: v2.35.0
release: pending
---

# SPRINT-062 — Referral Code Redemption Adds Bonus Quota

## Goal

Tenant admins enter a referral code, validate eligibility, and receive a bounded
bonus-quota grant through the S46 ledger — without mutating base packages.

## Velocity

| Window | Points |
| --- | ---: |
| Target | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0224](../04-tasks/TASK-0224.md) Validate + apply referral code API | 4 | completed | dev | Idempotent apply + errors |
| [TASK-0225](../04-tasks/TASK-0225.md) Bonus ledger grant + usage display | 3 | completed | dev | Base+bonus in usage UI |
| [TASK-0226](../04-tasks/TASK-0226.md) Tenant redemption UX | 3 | completed | dev | Input, validate, apply, summary |
| [TASK-0227](../04-tasks/TASK-0227.md) Platform reverse + UAT | 2 | completed | tester/dev | Admin revoke/reverse; isolation tests |
| **Total** | **12** | **12/12** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0054](../01-features/FEAT-0054-referral-code-redemption.md) | design_approved |
| Deep spec | **`approved`** | [DES-0058](../02-design/58-referral-code-redemption-spec.md) |
| Workflow | **`approved`** | [02-workflow.md](../02-design/02-workflow.md) §137–139 |
| ER | **`approved`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 62 |
| API | **`approved`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 62 |
| UX | **`approved`** | [05-ux-ui.md](../02-design/05-ux-ui.md) T62 / A62 |

## Scope boundary

### In
- Redemption UX/API, ledger grant, usage display, platform reverse

### Out
- Cash payouts, unlimited concurrency boosts, tax/invoice coupling

## Worktree

```text
.worktrees/SPRINT-060-062
branch: docs/sprint-060-062-plan
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Abuse / self-referral | Server rules + rate limits + audit |
| Double grant on retry | Idempotency key / unique redemption row |
| Bonus vs base confusion | Explicit UI breakdown |

## Verification

```bash
go test ./internal/quota ./internal/store ./cmd/server -count=1
cd apps/tenant-web && npm run check
```
