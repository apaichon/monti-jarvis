---
id: DES-0055
title: Call Conversation UX Revamp Specification
status: completed
updated: 2026-07-30
sprint: SPRINT-059
owner: SA
feature: FEAT-0051
release_target: v2.32.0
---

# DES-0055 — Call Conversation UX Revamp

**Sprint:** SPRINT-059 · **Release target:** v2.32.0
**Feature:** [FEAT-0051](../01-features/FEAT-0051-call-conversation-ux-revamp.md)
**Tasks:** [TASK-0212](../04-tasks/TASK-0212.md), [TASK-0213](../04-tasks/TASK-0213.md),
[TASK-0214](../04-tasks/TASK-0214.md), [TASK-0215](../04-tasks/TASK-0215.md)
**Reference mockups:**
`/Users/apaichon/Projects/libra/monti/images/new-design.png`,
`/Users/apaichon/Projects/libra/monti/images/new-design-light.png`

## 1. Goals

1. Replace the current active-call page with a friendlier, clearer conversation
   interface for callers.
2. Keep the avatar and live/listening effect large and visible during active
   calls instead of collapsing into a small rail item.
3. Support both light and dark modes with consistent layout, spacing, contrast,
   and control hierarchy.
4. Preserve current call, chat, transcript, tenant selection, audio-device, and
   EN/TH/JA display-language behavior.

## 2. Non-goals

- Backend voice, Gemini relay, WebRTC, or telephony protocol changes
- Native mobile app
- Queue, ticket, catalog, scheduled-call, or product-rendering features
- New profile/customer history APIs
- Deep tenant theme editor or arbitrary brand palettes
- Real keypad DTMF behavior unless already present in the app

## 3. UX Direction

### 3.1 Desktop layout

```text
+--------------------------------------------------------------------------------+
| Left context rail       | Center conversation stage                 | Right rail |
|                         |                                           |            |
| MONTI brand + status    | top status + new session + more           | About      |
| Tenant info card        |                                           | customer   |
| Customer info card      |        live waveform + large avatar       |            |
| Device settings card    |        greeting + listening state         | Quick      |
| Settings/help rows      |        call controls                      | actions    |
|                         |                                           |            |
|                         | transcript messages                       |            |
|                         | hold-to-talk + text composer + send       |            |
+--------------------------------------------------------------------------------+
```

Desktop keeps the central column dominant. The left rail contains setup/context
as compact icon-led accordions for tenant, customer, and device information.
Each accordion keeps its title and current summary visible when collapsed. The
right rail contains lightweight customer and quick-action data. Neither rail
should visually compete with the avatar stage.

### 3.2 Mobile layout

```text
+----------------------------------+
| menu | logo/avatar | live status |
| large avatar + waveform          |
| greeting + listening state       |
| speaker mute end keypad more     |
| transcript                       |
| hold-to-talk + send              |
| tenant row                       |
| customer row                     |
| device row                       |
+----------------------------------+
```

Mobile prioritizes call confidence: avatar, controls, transcript, then context.
Rows can open existing tenant/customer/device dialogs or panels.

### 3.3 Visual mapping from mockups

| Mockup element | Product mapping |
| --- | --- |
| Large circular avatar with ring | Active avatar stage; never collapse during call |
| Waveform behind avatar | CSS voice-wave/listening visual tied to existing active/listening state |
| Primary end-call button | Existing end-call action, red circular control |
| Speaker/mute controls | Toggle live playback gain and active microphone media track |
| Keypad | Numeric pointer/keyboard entry into the existing message composer; no DTMF |
| More control | Existing avatar/action picker |
| Transcript card | Existing message list with AI/customer bubbles and timestamps |
| Left rail tenant/customer/device cards | Existing selected tenant, customer/session summary, S56 audio devices |
| Left rail section icons and chevrons | Familiar context icon plus independent expand/collapse control |
| Right rail panels | Derived customer context/quick action placeholders using existing safe data |
| Light/dark screenshots | Two token sets applied to the same semantic layout |

## 4. Component Ownership

| Area | Likely files | Notes |
| --- | --- | --- |
| Customer desk shell | `apps/customer-web/src/lib/components/CustomerDesk.svelte` | Main layout and call state |
| Global/customer CSS | `apps/customer-web/src/app.css` | Tokens, responsive layout, wave/avatar styles |
| i18n labels | `apps/customer-web/src/lib/i18n/` | Add any new call-page labels under S58 runtime |
| Audio devices | `apps/customer-web/src/lib/audio/` and desk component | Reuse S56 selection/status |
| Voice/text relay | existing customer-web voice/client modules | No protocol change |

If the current code has split components for avatar cards, tenant cards, or
audio settings, prefer using those local components instead of introducing a new
shared UI framework.

## 5. Theme Model

Use semantic CSS custom properties for the call page:

| Token group | Examples |
| --- | --- |
| Surface | `--call-bg`, `--call-panel`, `--call-panel-strong` |
| Border/shadow | `--call-border`, `--call-border-strong`, `--call-glow` |
| Text | `--call-text`, `--call-muted`, `--call-subtle` |
| Accent | `--call-accent`, `--call-accent-strong`, `--call-success`, `--call-danger` |
| Conversation | `--bubble-ai`, `--bubble-user`, `--bubble-user-text` |
| Controls | `--control-bg`, `--control-border`, `--control-hover` |

Theme selection can initially follow the existing app theme mechanism if present.
If none exists, support a local call-page mode control or system preference with
stable classes such as `theme-dark` / `theme-light`. Avoid tying theme mode to
tenant brand color work beyond using existing accent values.

## 6. Interaction States

| State | Required behavior |
| --- | --- |
| Idle / ready | Avatar visible; primary CTA indicates ready/listening or start path |
| Connecting | Disable duplicate call actions; show connecting text/spinner without layout shift |
| Active call | Avatar remains large; controls enabled; transcript and composer active |
| Muted | Mute control clearly active; mic visual subdued |
| Speaking/listening | Waveform animates or pulses; respect reduced-motion |
| Ended | End control disabled or returns to ready state; transcript remains visible |
| Error | Show readable inline message; do not remove transcript/context |

## 7. Accessibility and Responsive Rules

- Minimum practical tap target: 44px for primary icon buttons.
- Icon-only controls require accessible labels and hover/focus titles where the
  local component pattern supports them.
- Avoid horizontal scroll at 360px, 390px, 768px, 1280px, and 1440px widths.
- Text must not overlap controls or overflow cards in EN, TH, or JA.
- Preserve keyboard focus outlines in both themes.
- Respect `prefers-reduced-motion` by reducing waveform/ring animation.

## 8. Data and APIs

No Postgres, Redis, ClickHouse, MinIO, or REST API changes are required.

Use existing client state for:

- selected tenant and tenant brand
- selected/active avatar
- call state and call duration
- transcript messages
- audio device summary
- display language from S58

Right-rail customer fields can be placeholder-safe when data is unavailable:

| Field | Fallback |
| --- | --- |
| Language | current UI language or `English` |
| Sentiment | neutral/positive placeholder only if already computed |
| Last contact | omit or show unavailable state |
| Total calls | omit or show existing session count if available |

## 9. Task Breakdown

| Task | Points | Scope |
| --- | ---: | --- |
| [TASK-0212](../04-tasks/TASK-0212.md) | 3 | Conversation shell layout and responsive structure |
| [TASK-0213](../04-tasks/TASK-0213.md) | 3 | Light/dark tokens and mode behavior |
| [TASK-0214](../04-tasks/TASK-0214.md) | 3 | Friendly call controls, transcript, and composer UX |
| [TASK-0215](../04-tasks/TASK-0215.md) | 3 | Context panels, mobile rows, screenshots, and UAT |

## 10. Verification

```bash
cd apps/customer-web && npm run check
cd apps/customer-web && npm run build
```

Manual / visual:

1. Desktop dark and light screenshots match approved layout direction.
2. Mobile dark and light screenshots show no horizontal scroll or overlapping
   text.
3. Start an avatar call, send a text message, mute/unmute, end call.
4. Open/refresh audio-device controls; selected mic/speaker summary remains
   accurate.
5. Switch EN/TH/JA labels and confirm controls still fit.
6. Reduced-motion mode keeps the page usable without continuous animation.

## 11. Risks

| Risk | Mitigation |
| --- | --- |
| Visual scope exceeds 12 points | Keep backend unchanged; focus customer-web call conversation page only |
| Mobile transcript/composer crowding | Use a single-column mobile order and fixed tap target sizes |
| Light mode exposes contrast gaps | Use semantic tokens and screenshot both themes before close |
| i18n label expansion breaks buttons | Verify EN/TH/JA and allow labels to wrap or shorten where appropriate |
| Existing call state regressions | Preserve current voice/text state machine and smoke start/end call |
