---
id: SPRINT-001
status: in_progress
start: 2026-07-07
end: 2026-07-21
updated: 2026-07-07
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

Total: 16 points.

## Stack delivered

```text
apps/customer-web     SvelteKit + Tailwind + shadcn-style components + livekit-client
internal/calls        Call orchestration service
internal/lktoken      LiveKit JWT tokens
internal/natsbus      call.started / call.ended / call.turn.created
internal/customerweb  Serves built portal at /
```

## Scope boundary

**In scope:** Sprint 1 ACs above. Legacy UI behind `LEGACY_UI_ENABLED=true` at `/legacy`.

**Out of scope:** KM/RAG (Sprint 2), Auth (Sprint 3), workforce picker (Sprint 21).

## Verification

```bash
# Infra (Postgres, Redis, optional NATS + LiveKit containers)
make infra-init

# Terminal 1 — optional LiveKit dev server
docker run --rm -p 7880:7880 -e LIVEKIT_KEYS="devkey: secret" livekit/livekit-server --dev

# Terminal 2 — optional NATS
docker run --rm -p 4222:4222 nats:2.10 -js

# Build + run
cp infra/.env.dev.example infra/.env.dev
make start
open http://localhost:8091
```

- `go test ./...`
- `cd apps/customer-web && npm run build`
- Manual: Start call → mic permission → End call → check Postgres `call_sessions`

## Risks

- LiveKit and NATS are optional at dev time; API degrades with warnings if offline.
- AI agent audio participant is stubbed until Sprint 21 workforce integration.