---
id: SPRINT-001
status: completed
start: 2026-07-07
end: 2026-07-07
closed: 2026-07-07
updated: 2026-07-07
release: v0.2.0
goal: "Customer: Conversation — Svelte + LiveKit inbound voice with NATS lifecycle events."
roadmap_sprint: 1
platform: Customer
depends_on: []
---

# SPRINT-001 — Customer: Conversation

## Goal

An end customer can start and end an inbound AI voice conversation through the **Svelte + shadcn** customer portal, with audio carried by **LiveKit** and call lifecycle published on **NATS**.

## Commitment

| Task | Points | Status | Outcome |
| --- | ---: | --- | --- |
| TASK-0002 | 5 | completed | Svelte + shadcn customer portal shell |
| TASK-0003 | 5 | completed | LiveKit room/token API + join flow |
| TASK-0004 | 3 | completed | Call session (Postgres + Redis 8) + NATS events |
| TASK-0005 | 3 | completed | Transcript stream UI |

**Committed:** 16 points · **Completed:** 16 points · **Velocity:** 16

## Shipped (v0.2.0)

- Svelte customer portal at `/` with legacy dark-neon two-panel UI
- Gemini voice relay (`/ws/voice`) wired into call flow
- Text chat via `/api/chat` and infra status line
- Avatar photos for Ava, Max, Luna, Neo
- Call sessions API, Postgres persistence, Redis active state, NATS events
- LiveKit token API and local compose (NATS + LiveKit)
- Legacy UI always available at `/legacy/`
- `make up` / `make down` lifecycle for dev infra

## Stack delivered

```text
apps/customer-web     SvelteKit + Tailwind + legacy-style UI + Gemini voice
internal/calls        Call orchestration service
internal/lktoken      LiveKit JWT tokens
internal/natsbus      call.started / call.ended / call.turn.created
internal/customerweb  Serves built portal at /
```

## Scope boundary

**In scope:** Sprint 1 ACs above.

**Out of scope:** KM/RAG (Sprint 2), Auth (Sprint 3), full workforce catalog admin (Sprint 21).

## Verification

```bash
make customer-web && make start
open http://localhost:8091
```

- `go test ./...`
- `cd apps/customer-web && npm run build`
- Manual: [SPRINT-001 UAT checklist](../06-manual-tests/SPRINT-001-manual.md)
- Readiness: [RELEASE-READINESS](../08-readiness/RELEASE-READINESS.md) (SPRINT-001 section)

## Retrospective notes

- Gemini voice was required for a usable conversation; LiveKit join alone was insufficient for Sprint 1 demo.
- Main portal UI was aligned to legacy design for visual continuity while keeping Svelte architecture.
- Shared `poc-gml-*` containers remain the default Postgres/Redis/MinIO target in dev.

## Risks (closed)

- LiveKit and NATS degrade gracefully when offline; compose targets are documented in Makefile.
- AI audio uses Gemini WebSocket relay; dedicated LiveKit agent participant deferred to later sprints.