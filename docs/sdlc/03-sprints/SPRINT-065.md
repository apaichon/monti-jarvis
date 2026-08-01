---
id: SPRINT-065
status: in_progress
start: 2026-08-01
end: 2026-08-02
updated: 2026-08-01
design_pack: review_pending
roadmap_sprint: 65
feature: FEAT-0057
platform: Customer / Tenant / Quota
depends_on: [SPRINT-013, SPRINT-016, SPRINT-021, SPRINT-045, SPRINT-051, SPRINT-056]
goal: "Callers over a tenant's concurrent-call package limit wait in a bounded queue and are admitted when capacity opens."
velocity_basis: "Last 3 closed: S61=12, S62=12, S63=3 -> avg 9; commit 9"
release_target: v2.37.0
---

# SPRINT-065 - Queued Concurrent-Call Admission

## Goal

Replace immediate over-concurrent-call rejection with a tenant-scoped waiting
queue that admits the next caller when another customer finishes, disconnects,
or times out.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S61-S63) | 12, 12, 3 -> **avg 9** |
| **Committed** | **9** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0234](../04-tasks/TASK-0234.md) Redis concurrent-call queue admission service | 3 | done | dev | Reserve slot or enqueue with FIFO position |
| [TASK-0235](../04-tasks/TASK-0235.md) Voice WS promotion and release hooks | 3 | done | dev | Queued callers promote on release without over-admission |
| [TASK-0236](../04-tasks/TASK-0236.md) Waiting UX, tenant visibility, and UAT | 3 | review | dev/tester | Customer queue states plus active/queued counts; manual UAT remains |
| **Total** | **9** | **6/9 done + 3 review** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0057](../01-features/FEAT-0057-queued-concurrent-call-admission.md) | in_progress |
| Deep spec | **`review_pending`** | [DES-0060](../02-design/60-queued-concurrent-call-admission-spec.md) |
| Workflow | **`review_pending`** | [02-workflow.md](../02-design/02-workflow.md) section 141 |
| ER | **`review_pending`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 65 |
| API | **`review_pending`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 65 |
| UX | **`review_pending`** | [05-ux-ui.md](../02-design/05-ux-ui.md) C65/T65 |

## Scope boundary

### In

- Tenant-scoped Redis queue for public/customer voice calls when
  `max_concurrent_calls` is full.
- Admission service that either reserves a concurrent slot or returns queue
  metadata.
- Promotion on call release, disconnect, timeout, and failed start, safe for
  multiple app instances.
- Customer waiting state, queue position, timeout, cancel, and automatic call
  start when admitted.
- Tenant/admin visibility for active count, queued count, package limit, and
  recent timeouts.
- Metrics/audit events for queued, admitted, cancelled, timed out, rejected, and
  promoted states.

### Out

- Exceeding a tenant package concurrent-call limit.
- Priority queues, paid queue tiers, callback scheduling, or human-agent routing.
- Cross-tenant capacity sharing.
- Replacing monthly-minute, daily-limit, per-call, mobile, or rate-limit
  enforcement.
- A new durable Postgres queue table in this MVP.

## Worktree

```text
.worktrees/SPRINT-065
branch: feature/sprint-065-concurrent-call-queue
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Over-admission across app instances | Redis atomic admission, promotion lock, and active-slot lease checks |
| Stale queued callers | Entry TTL, browser close cleanup, timeout worker, and cancel-on-WS-close |
| Poor caller experience | Live queue position, wait timeout, cancel action, and clear fallback errors |
| Capacity hidden from tenants | Current-plan/usage response includes active, queued, and limit fields |

## Verification

```bash
go test ./...
cd apps/customer-web && npm run prebuild && npm run check && npm run build
cd apps/tenant-web && npm run prebuild && npm run check && npm run build
```

Manual UAT: two tenants, package limit 1, caller B waits, caller A ends, caller
B auto-starts, and active calls never exceed 1.
