---
id: DES-0005
title: UX/UI — ASCII Wireframes
status: review_pending
updated: 2026-07-07
sprint: SPRINT-003
---

# UX/UI — Customer Portal (ASCII)

Primary surface: `apps/customer-web` at `/`. Layout matches legacy dark-neon two-panel design.

## Screen map → API

| UI zone | User action | API / WS |
| --- | --- | --- |
| A1 Brand header | (static) | `GET /images/monti-logo.png` |
| A2 Avatar halo | Select agent | `GET /api/workforce` (on load) |
| A3 Agent cards | Click agent | local state; greeting via chat optional |
| A4 Start/End call | Toggle voice | `POST /api/calls` → `WS /ws/voice?agent&topic` → `POST .../end` |
| A5 Timer | (display) | client timer on call start |
| B1 Topic tabs | General/Billing/Technical | `topic` field on chat + voice WS |
| B2 Chat stream | View messages | SSE turns + chat responses + voice transcripts |
| B3 Composer Send | Text question | `POST /api/chat` |
| B4 Infra line | (display) | `GET /api/infra` |
| B5 Session line | (display) | `session_id` from chat/call |
| Citations | Under assistant bubble | `sources[]` from `/api/chat` |

## Full layout (desktop)

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  http://localhost:8091/                                                          │
├─────────────────────────────┬──────────────────────────────────────────────────┤
│  CONTROL PANEL              │  WORKSPACE — Caller Desk                          │
│                             │                                                  │
│  ┌──┐ MONTI                 │  Caller Desk          Postgres ok · Redis ok …  │
│  │██│ Inbound Call Center   │  ┌────────┐ ┌────────┐ ┌──────────┐            │
│  └──┘                       │  │General*│ │Billing │ │Technical │  ← topic    │
│                             │  └────────┘ └────────┘ └──────────┘            │
│      ╭───────────╮          │                                                  │
│     │  ┌───────┐  │         │  ┌──┐  Welcome to Monti…                       │
│     │  │ AVATAR│  │         │  │A │  Choose an agent…                          │
│     │  │ image │  │         │  └──┘                                             │
│     │  └───────┘  │         │  ┌──┐  ┌─────────────────────────────────────┐  │
│      ╰───────────╮          │  │C │  │ When are invoices due?              │  │
│       waveform    │         │  └──┘  └─────────────────────────────────────┘  │
│                             │  ┌──┐  ┌─────────────────────────────────────┐  │
│       Ava                   │  │M │  │ Invoices are due within 15 days.    │  │
│  General Support · trait    │  └──┘  │ [billing · KB]                      │  │
│                             │        └─────────────────────────────────────┘  │
│  ┌──────────┬──────────┐  │                                                  │
│  │ 00:01:23 │ End call │  │  ┌──────────────────────────────┬────────┐       │
│  └──────────┴──────────┘  │  │ Ask your question...         │ Send   │       │
│  On call with Ava.        │  └──────────────────────────────┴────────┘       │
│                             │  Call abc123 · Max          ← session label        │
│  ┌────────────────────────┐ │                                                  │
│  │[img] Ava      [Select] │ │                                                  │
│  │ General · Popular      │ │                                                  │
│  ├────────────────────────┤ │                                                  │
│  │[img] Max       Select   │ │                                                  │
│  ├────────────────────────┤ │                                                  │
│  │[img] Luna      Select   │ │                                                  │
│  ├────────────────────────┤ │                                                  │
│  │[img] Neo       Select   │ │                                                  │
│  └────────────────────────┘ │                                                  │
│   ↑ GET /api/workforce      │                                                  │
└─────────────────────────────┴──────────────────────────────────────────────────┘
```

## Flow A — Text chat with RAG

```
User types in composer ──► POST /api/chat
                              │
                              ├─ agent_id ← selected card (A3)
                              ├─ topic    ← active tab (B1)
                              └─ history  ← prior turns
                              │
                              ▼
                         Append user bubble (C)
                         Append thinking bubble
                              │
                              ▼
                         Replace with reply + citation chips
                         sources[] → [scope · KB] tags
```

## Flow B — Voice call

```
User clicks Start call (A4)
    │
    ├─► POST /api/calls          → session.id, room_name
    ├─► GET /api/calls/{id}/events (SSE) → live turns
    └─► WS /ws/voice?agent=ava&topic=general
            │
            ├─ mic → audio JSON → Gemini Live
            └─ transcript events → chat stream (B2)

User clicks End call
    └─► POST /api/calls/{id}/end
```

## Flow C — Agent selection

```
Click agent card (A3)
    │
    ├─ Update halo portrait (A2)     image from agent.image
    ├─ Update waveform color         agent.color
    ├─ If live call → hang up first
    └─ Optional greeting in chat     agent.greeting (on select)
```

## Mobile (≤980px)

```
┌─────────────────────┐
│  CONTROL PANEL      │  stacked full width
│  (avatar + agents)  │
├─────────────────────┤
│  WORKSPACE          │
│  chat + composer    │
└─────────────────────┘
```

## Legacy UI (`/legacy/`)

Same API mapping; vanilla JS in `internal/web/public/index.html`. Useful reference for visual parity.

## Design tokens

| Token | Value |
| --- | --- |
| Background | `#020712` + blue/purple radial gradients |
| Panel | glass dark `rgb(5 16 31 / 82%)` |
| Accent | cyan `#00b7ff`, agent `--assistant-color` |
| Radius | panels `26px`, bubbles `16px` |
| Font | Inter |

## Sprint 3 — Auth (no customer UI change)

Customer portal **unchanged** when `AUTH_DISABLED=true` (default). No login screen in Sprint 3.

| Surface | Sprint 3 UX | API |
| --- | --- | --- |
| Customer portal `/` | No login; same as v0.3.0 | Public chat/voice |
| KM admin (ops) | curl / REST client only | `POST /api/auth/login` → Bearer on KM writes |
| Future tenant admin | Deferred Sprint 15+ | — |

```text
┌─────────────────────────────────────────┐
│  Ops / Tester (terminal or REST Client) │
│  POST /api/auth/login                   │
│  → store access_token                   │
│  → Authorization: Bearer on km-seed     │
└─────────────────────────────────────────┘
```

## Component → file

| Component | Path |
| --- | --- |
| Page layout | `apps/customer-web/src/routes/+page.svelte` |
| Portrait | `src/lib/components/Portrait.svelte` |
| Waveform | `src/lib/components/Waveform.svelte` |
| Styles | `src/app.css` |
| Chat API | `src/lib/api/chat.ts` |
| Voice | `src/lib/voice/gemini.ts` |

See [auth-spec.md](auth-spec.md), [api-spec.md](api-spec.md), [workflow.md](workflow.md).