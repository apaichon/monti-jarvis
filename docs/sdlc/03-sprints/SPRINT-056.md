---
id: SPRINT-056
status: design_approved
start: 2026-07-29
end: 2026-08-08
updated: 2026-07-29
design_pack: approved
roadmap_sprint: 56
feature: FEAT-0047
platform: Customer
depends_on: [SPRINT-001, SPRINT-005, SPRINT-039, SPRINT-054]
goal: "Revamp caller desk and tenant list branding (large Monti + company logos) and add mic/speaker device settings for voice."
velocity_basis: "Last 3 closed: S51=14, S53=12, S54=12 → avg 12.7; commit 12"
release_target: v2.28.0
---

# SPRINT-056 — Caller Desk Branding + Mic/Speaker Settings

## Goal

Callers see large Monti and per-company logos on the brand directory and call
desk (per mockups), and can select microphone and speaker devices for voice
calls without technical infra jargon.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed slices (S51, S53, S54) | 14, 12, 12 → **avg 12.7** |
| **Committed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0199](../04-tasks/TASK-0199.md) Tenant list + call desk logo branding revamp | 5 | todo | dev | Large Monti + company logos per mockups |
| [TASK-0200](../04-tasks/TASK-0200.md) Mic/speaker device settings for voice | 4 | todo | dev | Enumerate, select, apply, persist devices |
| [TASK-0201](../04-tasks/TASK-0201.md) Branding + audio UAT and verification | 3 | todo | tester/dev | Manual checklist; logo A/B; permission paths |
| **Total** | **12** | **0/12** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0047](../01-features/FEAT-0047-caller-desk-branding-audio-devices.md) | design_approved |
| Deep spec | **`approved`** | [DES-0051](../02-design/51-caller-desk-branding-audio-devices-spec.md) |
| Workflow | **`approved`** | [02-workflow.md](../02-design/02-workflow.md) §127–129 |
| ER | **`approved`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 56 |
| API | **`approved`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 56 |
| UX | **`approved`** | [05-ux-ui.md](../02-design/05-ux-ui.md) C56 |
| Mockups | reference | [mockups/s56-caller-desk-branding/](../02-design/mockups/s56-caller-desk-branding/) |

## Scope boundary

### In

- Customer-web branding revamp (list + desk) aligned to mockups
- Consume public brand/theme logo fields; monogram fallback
- Browser mic/speaker selection wired into existing voice stack
- Device preference persist (session/localStorage)
- Keep S54 routing and Live · OK status presentation

### Out

- Theme publish CMS redesign
- S55 topic statistics
- Noise suppression / test tone product
- Embed layout redesign (unless shared components force a minimal sync)

## Verification

```bash
cd apps/customer-web && npm run check
# Manual: tenant list logos; desk logos A vs B; mic/speaker select; deny permission
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Browser speaker `setSinkId` support gaps | Feature-detect; fallback default output with clear copy |
| Logo hotlink / CORS | Prefer same-origin or allowlisted public URLs; monogram fallback |
| Scope creep into full marketing hub | Mockup-driven layout only; no new CMS |
| Device permission friction | Prompt only when opening settings or starting voice |

## Worktree

```text
.worktrees/SPRINT-056
branch: feature/sprint-056-caller-desk-branding-audio
```
