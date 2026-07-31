---
id: SPRINT-063
status: completed
start: 2026-07-31
end: 2026-07-31
updated: 2026-07-31
design_pack: approved
roadmap_sprint: 63
feature: FEAT-0055
platform: Platform Admin / Product Web / Sales
depends_on: [SPRINT-048, SPRINT-051]
goal: "Render Product Web Book Demo leads returned by the platform API in the Platform Admin Leads inbox with accurate counts."
velocity_basis: "Focused P1 regression fix; commit 3 points below normal 12-point velocity."
release_target: v2.35.1
release: v2.35.1
---

# SPRINT-063 - Platform Admin Leads Inbox Rendering Fix

## Goal

Fix the response-contract mismatch that hides valid Product Web lead records
from Platform Admin sales operators.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0228](../04-tasks/TASK-0228.md) Align Leads list client contract and rendering | 2 | completed | dev | API `items[]` renders as inbox rows |
| [TASK-0229](../04-tasks/TASK-0229.md) Contract regression tests and UAT | 1 | completed | tester/dev | Counts, empty state, check/build verified |
| **Total** | **3** | **3/3** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0055](../01-features/FEAT-0055-platform-admin-leads-inbox-rendering.md) | completed |
| Deep spec | n/a | Existing Sprint 48 marketing-leads domain |
| Workflow | approved | [02-workflow.md](../02-design/02-workflow.md), section 140 |
| ER | approved | [03-er-diagram.md](../02-design/03-er-diagram.md), no schema delta |
| API | approved | [04-api-spec.md](../02-design/04-api-spec.md), Sprint 63 |
| UX | approved | [05-ux-ui.md](../02-design/05-ux-ui.md), A63 |

## Scope boundary

### In

- Platform Admin Leads list response contract.
- Safe list normalization and accurate shown/total counts.
- Book Demo lead rendering.
- Contract tests, Svelte check/build, and browser UAT checklist.

### Out

- Backend endpoint or Postgres changes.
- Product Web form changes.
- Lead lifecycle or CRM redesign.

## Verification

```bash
cd apps/platform-admin-web
npm test
npm run check
npm run build
```

Manual: [SPRINT-063-leads-inbox-manual.md](../06-manual-tests/SPRINT-063-leads-inbox-manual.md)

## Review - PASS

The client response type now matches the server `items[]` contract. A tested
normalizer preserves rows and count metadata, and the Leads page no longer
reads the nonexistent `res.leads` field. Automated tests and web gates pass;
credentialed target-environment confirmation remains in the manual checklist.

## Release

- Version: v2.35.1
- Closed: 2026-07-31
- Type: patch bug fix
- Merge: `504aaad` on `main`
- Tag: `v2.35.1`

## Worktree

```text
.worktrees/SPRINT-063
branch: feature/sprint-063-leads-inbox
```
