---
id: SPRINT-056
status: completed
start: 2026-07-29
end: 2026-08-08
updated: 2026-07-29
design_pack: shipped
roadmap_sprint: 56
feature: FEAT-0048
platform: Customer
depends_on: [SPRINT-001, SPRINT-005, SPRINT-039, SPRINT-054]
goal: "Revamp caller desk and tenant list branding (large Monti + company logos) and add mic/speaker device settings for voice."
velocity_basis: "Last 3 closed: S51=14, S53=12, S54=12 → avg 12.7; commit 12"
release_target: v2.29.0
release: v2.29.0
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
| [TASK-0202](../04-tasks/TASK-0202.md) Tenant list + call desk logo branding revamp | 5 | completed | dev | Large Monti + company logos per mockups |
| [TASK-0203](../04-tasks/TASK-0203.md) Mic/speaker device settings for voice | 4 | completed | dev | Enumerate, select, apply, persist devices |
| [TASK-0204](../04-tasks/TASK-0204.md) Branding + audio UAT and verification | 3 | completed | tester/dev | Manual checklist; logo A/B; permission paths |
| **Total** | **12** | **12/12** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0048](../01-features/FEAT-0048-caller-desk-branding-audio-devices.md) | design_approved |
| Deep spec | **`approved`** | [DES-0052](../02-design/52-caller-desk-branding-audio-devices-spec.md) |
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


## Implementation notes

- `TenantPicker` large Monti hero + logo cards
- `CustomerDesk` Monti hero + company logo card + audio settings
- `lib/audio/devices.ts` enumerate/persist; GeminiVoice deviceId options
- Manual UAT: [SPRINT-056-caller-desk-branding-audio-manual.md](../06-manual-tests/SPRINT-056-caller-desk-branding-audio-manual.md)


## Shipped summary (v2.29.0)

- Large Monti + company logos on tenant list and call desk (theme logo coalesce)
- Mic/speaker device selection, live levels, audio test, refresh devices
- Call desk UX: quick actions, collapsible audio, account (last active), secure footer
- Design: DES-0052 · FEAT-0048 · TASK-0202–0204
- Manual UAT: [SPRINT-056-caller-desk-branding-audio-manual.md](../06-manual-tests/SPRINT-056-caller-desk-branding-audio-manual.md)

**Closed:** 2026-07-29
