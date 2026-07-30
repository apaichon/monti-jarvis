---
id: SPRINT-061
status: planned
start: 2026-08-22
end: 2026-08-29
updated: 2026-07-30
design_pack: approved
roadmap_sprint: 61
feature: FEAT-0053
platform: Tenant / Platform Admin / AI Operations
depends_on: [SPRINT-026, SPRINT-029, SPRINT-043, SPRINT-060]
goal: "Remove tenant System Performance from tenant nav; show Gemini status in top bar linked to AI Settings."
velocity_basis: "avg 12; commit 12 (UX-focused)"
release_target: v2.34.0
release: pending
---

# SPRINT-061 — Tenant UX Simplification (Gemini Status Top Bar)

## Goal

Tenant admins no longer use a System Performance page. A compact Gemini status
chip in the top bar shows readiness and routes to AI Settings when action is
needed. Platform-admin diagnostics remain.

## Velocity

| Window | Points |
| --- | ---: |
| Target | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0220](../04-tasks/TASK-0220.md) Gemini status API for tenant shell | 3 | todo | dev | Safe status payload for top bar |
| [TASK-0221](../04-tasks/TASK-0221.md) Remove tenant performance nav + page | 3 | todo | dev | Nav/route cleanup |
| [TASK-0222](../04-tasks/TASK-0222.md) Top-bar status UI + link to AI Settings | 3 | todo | dev | Ready/missing/failed/degraded chip |
| [TASK-0223](../04-tasks/TASK-0223.md) Platform preservation + UAT | 3 | todo | tester/dev | Platform perf intact; smoke states |
| **Total** | **12** | **0/12** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0053](../01-features/FEAT-0053-tenant-gemini-status-top-bar.md) | design_approved |
| Deep spec | **`approved`** | [DES-0057](../02-design/57-tenant-gemini-status-top-bar-spec.md) |
| Workflow | **`approved`** | [02-workflow.md](../02-design/02-workflow.md) §135–136 |
| ER | **`approved`** | client + API status only |
| API | **`approved`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 61 |
| UX | **`approved`** | [05-ux-ui.md](../02-design/05-ux-ui.md) T61 |

## Scope boundary

### In
- Nav cleanup, top-bar status, status API reuse of S60 metadata

### Out
- Platform-admin removal, raw infra probes for tenants

## Worktree

```text
.worktrees/SPRINT-060-062
branch: docs/sprint-060-062-plan
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Support still needs S26 API | Keep endpoint; hide UI; document support-only |
| Status lag after key change | Invalidate/refetch after AI Settings save/test |

## Verification

```bash
cd apps/tenant-web && npm run check
cd apps/platform-admin-web && npm run check
# Manual: nav has no performance; top bar states; platform monitoring OK
```
