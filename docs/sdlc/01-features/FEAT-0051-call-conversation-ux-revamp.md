---
id: FEAT-0051
title: "Call conversation UX revamp with light and dark modes"
status: completed
roadmap_sprint: 59
priority: A+
depends_on: [SPRINT-001, SPRINT-021, SPRINT-039, SPRINT-054, SPRINT-056, SPRINT-058]
design: DES-0055
design_spec: ../02-design/55-call-conversation-ux-revamp-spec.md
updated: 2026-07-30
---

# FEAT-0051: Call Conversation UX Revamp

## Purpose

Make the customer call conversation page feel friendly, clear, and easy to use
on desktop and mobile, with a centered avatar conversation experience, readable
transcript, obvious call controls, useful context panels, and first-class light
and dark themes.

## Problem

The current caller desk has the required call, text, audio-device, tenant, and
i18n capabilities, but the live conversation screen still feels operational and
can collapse important avatar context during active calls. Callers need a clear
visual focus for "who I am speaking with", "how I talk", "how I end or control
the call", and "where conversation history and context live".

## Scope

### In

- Revamp the customer call conversation page layout for desktop and mobile.
- Desktop: left context rail, central avatar/call stage, transcript/composer,
  and right-side customer/quick-action panels.
- Mobile: compact header, large avatar stage, call controls, transcript,
  hold-to-talk/send composer, and tenant/customer/device rows.
- Light and dark mode tokens for the call conversation page.
- Theme-specific avatar portraits for tenant-owned agents, with default image
  fallback when a dark or light portrait is not uploaded.
- Friendly controls for speaker, mute, end call, keypad, more actions, new
  session, hold-to-talk, send, and listening/live states.
- Transcript bubbles with timestamps, participant identity, readable spacing,
  and no horizontal overflow.
- Preserve current tenant selection, call start/end, text chat, audio-device
  settings, i18n, and transcript behavior.
- Compact icon-led desktop context sections that can collapse independently
  while retaining tenant, customer, and device summaries.

### Out

- Backend voice transport, Gemini protocol, or telephony routing changes.
- Native mobile application.
- New queue, ticket, catalog, scheduled call, or product recommendation
  features.
- Full tenant theme builder work beyond call-page light/dark tokens.
- New authenticated customer profile APIs.

## Acceptance criteria

1. Desktop call conversation page follows the approved friendly layout direction
   from the dark/light references and keeps the avatar live effect prominent.
2. Mobile layout remains usable without horizontal scroll and keeps avatar,
   controls, transcript, and composer visible in a natural order.
3. Light mode and dark mode both render with readable text, consistent spacing,
   and accessible focus/tap targets.
4. Existing voice call, text chat, end call, mute/speaker controls, microphone
   and speaker selection, tenant selection, i18n labels, and transcript behavior
   continue to work.
5. Transcript bubbles, buttons, and panels do not overlap or resize
   unpredictably at desktop, tablet, or phone widths.
6. Customer and tenant context is available without distracting from the active
   call stage.
7. Tenant-owned avatars can upload dark-mode and light-mode portraits, and the
   customer call page renders the image matching the selected mode.

## Test notes

- Run `npm run check` and `npm run build` in `apps/customer-web`.
- Capture desktop and mobile screenshots in both light and dark modes.
- Smoke active call: select tenant/avatar, start call, send text, mute, end.
- Verify existing audio-device picker behavior from S56 still opens and updates
  the displayed mic/speaker summary.
- Verify S58 language labels still render and do not overflow in the new layout.

## Build notes

- Implemented in `apps/customer-web/src/lib/components/CustomerDesk.svelte` and
  `apps/customer-web/src/app.css`.
- Verified with `npm run check`, `npm run build`, and mocked Playwright desktop
  dark/light plus mobile dark screenshots.

## Dependencies

- Prior customer call foundation: SPRINT-001, SPRINT-021
- Theme foundation: SPRINT-039
- Tenant list / selected tenant: SPRINT-054
- Caller desk branding and audio-device settings: SPRINT-056
- UI language selector: SPRINT-058
- Design: [DES-0055](../02-design/55-call-conversation-ux-revamp-spec.md)
