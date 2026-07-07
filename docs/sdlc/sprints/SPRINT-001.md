---
id: SPRINT-001
status: in_progress
start: 2026-07-07
end: 2026-07-21
updated: 2026-07-07
goal: "Launch inbound call center MVP — AI avatar workforce selection with text and voice Q&A."
committed:
  - TASK-0001
capacity_points: 13
velocity_basis: "Greenfield project; first sprint scoped to workforce + conversation only."
---

# SPRINT-001

## Goal

Ship the first Monti Inbound Call Center release: callers choose an AI avatar agent and get answers by text chat or voice call.

## Commitment

| Task | Points | Outcome |
| --- | ---: | --- |
| TASK-0001 | 13 | AI workforce selection + inbound Q&A (text + voice) |

Total: 13 points.

## Scope Boundary

- In scope: workforce catalog API, per-agent prompts/voices, call-center UI, text chat, Gemini Live voice relay, infra isolation, README/AGENTS.
- Out of scope: login/KYC, ticketing, CRM, knowledge-base search, call recording, supervisor dashboard, outbound dialing.

## Verification

- `go test ./...`
- `GET /api/workforce` returns 4 agents
- `POST /api/chat` with `agent_id` returns role-aware reply
- Browser: select agent → text question → voice call
- `make infra-init` creates `monti_jarvis` / `callcenter` schema

## Risks

- Gemini Live voice quality varies by agent voice mapping.
- Shared infra containers may be offline on fresh machines — app degrades gracefully without Postgres/Redis.