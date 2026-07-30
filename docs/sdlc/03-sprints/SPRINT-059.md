---
id: SPRINT-059
status: completed
start: 2026-07-30
end: 2026-07-30
closed: 2026-07-30
updated: 2026-07-30
design_pack: approved
roadmap_sprint: 59
feature: FEAT-0051
platform: Customer / Conversation UX
depends_on: [SPRINT-001, SPRINT-021, SPRINT-039, SPRINT-054, SPRINT-056, SPRINT-058]
goal: "Revamp the customer call conversation page into a friendly desktop/mobile experience with clear call controls, prominent live avatar, transcript, context panels, and light/dark mode."
velocity_basis: "Last 3 closed: S56=12, S57=12, S58=12 -> avg 12; commit 12"
release_target: v2.32.0
release: v2.32.0
git_tag: v2.32.0
---

# SPRINT-059 — Call Conversation UX Revamp

## Goal

Ship a friendlier customer call conversation page that keeps the live avatar
prominent, makes speaking and ending a call obvious, presents transcript context
clearly, and works in both light and dark modes on desktop and mobile.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S56, S57, S58) | 12, 12, 12 -> **avg 12** |
| **Committed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0212](../04-tasks/TASK-0212.md) Conversation shell layout + responsive structure | 3 | completed | dev | Desktop/mobile call page scaffold |
| [TASK-0213](../04-tasks/TASK-0213.md) Light/dark theme tokens + mode behavior | 3 | completed | dev | Semantic token sets and theme application |
| [TASK-0214](../04-tasks/TASK-0214.md) Friendly call controls + transcript/composer UX | 3 | completed | dev | Clear controls, transcript bubbles, input area |
| [TASK-0215](../04-tasks/TASK-0215.md) Context panels + visual/UAT verification | 3 | completed | tester/dev | Context panels, mobile rows, screenshots, smoke |
| **Total** | **12** | **12/12** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0051](../01-features/FEAT-0051-call-conversation-ux-revamp.md) | approved |
| Deep spec | **`approved`** | [DES-0055](../02-design/55-call-conversation-ux-revamp-spec.md) |
| Workflow | **`planned`** | Customer active-call journey only; no backend workflow change |
| ER | **`n/a`** | No schema change |
| API | **`completed`** | Avatar image upload accepts default/dark/light variants |
| UX | **`approved`** | Desktop/mobile call conversation mapped from supplied mockups |

## Reference mockups

- Dark mode: `/Users/apaichon/Projects/libra/monti/images/new-design.png`
- Light mode: `/Users/apaichon/Projects/libra/monti/images/new-design-light.png`

## Scope boundary

### In

- Customer-web call conversation page layout revamp.
- Prominent live avatar area with waveform/listening state.
- Desktop layout: left context rail, center conversation stage, right context
  panels, transcript, and composer.
- Mobile layout: compact header, avatar stage, controls, transcript, composer,
  and tenant/customer/device rows.
- Light/dark theme tokens and behavior for the call page.
- Tenant-owned avatars can carry separate dark-mode and light-mode portraits,
  with default portrait fallback.
- Preserve existing call, chat, device settings, tenant selection, and S58 i18n.

### Out

- Backend voice/Gemini protocol changes.
- Native mobile app.
- Ticket, product catalog, queue, scheduled call, or backup features.
- Server-side customer context APIs.
- Full theme-brand editor changes outside call page.

## Verification

```bash
cd apps/customer-web && npm run check
cd apps/customer-web && npm run build
```

Completed:

- Customer-web `npm run check` clean.
- Customer-web `npm run build` clean.
- Tenant avatar API/UI supports Default, Dark, and Light portrait uploads.
- Desktop context rail uses compact icon-led tenant/customer/device accordions
  with localized EN/TH/JA labels and persistent collapsed summaries.
- Context accordions default closed, and the desktop conversation workspace
  scrolls vertically instead of clipping the subtitle or call/end controls.
- Speaker and microphone controls now toggle real call media state; Keypad
  provides numeric pointer/keyboard input to the text composer without DTMF.
- Playwright visual QA with mocked customer/tenant APIs:
  - `.playwright-mcp/sprint59-desktop-dark.png`
  - `.playwright-mcp/sprint59-desktop-light.png`
  - `.playwright-mcp/sprint59-mobile-dark.png`
  - Desktop/mobile horizontal overflow check passed.

Manual:

- User screenshot review covered desktop layout, collapsed rails, avatar
  prominence, center scrolling, and composer placement.
- Automated customer and tenant checks/builds cover dark/light portraits,
  icon-only theme control, responsive structure, and EN/TH/JA labels.
- Target-device live microphone/speaker smoke remains a documented post-release
  follow-up in the [manual UAT](../06-manual-tests/SPRINT-059-call-conversation-ux-manual.md).

## Shipped summary

Released as **v2.32.0** on 2026-07-30 with 12/12 points completed:

- Friendly responsive call conversation workspace with a prominent live avatar.
- Collapsed icon-led tenant, customer, and device context rails.
- Light/dark presentation with icon-only theme switching.
- Tenant avatar default, dark, and light portrait variants.
- Real call microphone/speaker toggles and numeric keypad composer input.
- Scroll-safe transcript and bottom-aligned conversation composer.

## Risks

| Risk | Mitigation |
| --- | --- |
| Visual scope grows beyond sprint velocity | Keep work in customer-web and preserve backend behavior |
| Mobile layout becomes crowded | Prioritize avatar, controls, transcript, composer; move context to rows |
| Light/dark drift | Use semantic tokens and verify both modes before close |
| Existing call flow regression | Smoke voice/text/end-call against current state machine |

## Worktree

```text
.worktrees/SPRINT-059
branch: feature/sprint-059-call-conversation-ux
```
