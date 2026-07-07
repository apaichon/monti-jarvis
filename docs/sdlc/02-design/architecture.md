---
id: DES-0001
title: System Architecture
status: active
updated: 2026-07-07
sprint: SPRINT-002
---

# System Architecture — Monti Jarvis

Inbound AI call center: one Go process serves the Svelte customer portal, REST/WS APIs, and optional legacy UI.

## Layer diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Client tier                                                            │
│  ┌──────────────────────┐  ┌──────────────────────┐                       │
│  │ apps/customer-web    │  │ /legacy (optional)   │                       │
│  │ SvelteKit @ /        │  │ vanilla HTML         │                       │
│  └──────────┬───────────┘  └──────────┬───────────┘                       │
└─────────────┼──────────────────────────┼───────────────────────────────┘
              │ REST / WS / SSE            │
┌─────────────▼──────────────────────────▼───────────────────────────────┐
│  Application tier — cmd/server (port 8091)                              │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐         │
│  │ workforce  │ │ calls      │ │ km + rag   │ │ gemini     │         │
│  │ /api/chat  │ │ /api/calls │ │ /api/km/*  │ │ /ws/voice  │         │
│  └─────┬──────┘ └─────┬──────┘ └─────┬──────┘ └─────┬──────┘         │
│        │              │              │              │                   │
│  ┌─────▼──────────────▼──────────────▼──────────────▼──────┐           │
│  │ internal/store · internal/clickhouse · internal/natsbus   │           │
│  └─────┬──────────────┬──────────────┬──────────────┬──────┘           │
└────────┼──────────────┼──────────────┼──────────────┼──────────────────┘
         │              │              │              │
┌────────▼──────┐ ┌─────▼─────┐ ┌─────▼─────┐ ┌────▼─────┐ ┌──────────┐
│ Postgres      │ │ Redis 8   │ │ MinIO     │ │ClickHouse│ │ NATS     │
│ monti_jarvis  │ │ db 4      │ │ monti-    │ │ km_embed │ │ JetStream│
│ callcenter.*  │ │ prefix    │ │ jarvis    │ │ qa_events│ │ optional │
└───────────────┘ └───────────┘ └───────────┘ └──────────┘ └──────────┘
         │                                              │
         │         ┌──────────────┐                     │
         └────────►│ LiveKit      │◄── token API ───────┘
                   │ (optional)   │
                   └──────────────┘
                              │
                   ┌──────────▼──────────┐
                   │ Gemini API          │
                   │ text + live audio   │
                   └─────────────────────┘
```

## Package map

| Package | Responsibility |
| --- | --- |
| `cmd/server` | HTTP routing, handlers |
| `internal/customerweb` | Serve Svelte build at `/` |
| `internal/web` | Legacy static UI at `/legacy/` |
| `internal/workforce` | Agent catalog + system prompts |
| `internal/gemini` | Text chat + embeddings |
| `internal/live` | Gemini Live WebSocket relay |
| `internal/calls` | Call session orchestration |
| `internal/lktoken` | LiveKit JWT |
| `internal/natsbus` | `call.*` events |
| `internal/store` | Postgres, Redis, MinIO |
| `internal/km` | Document ingest + chunking |
| `internal/rag` | Scoped retrieval + prompt augment |
| `internal/clickhouse` | Vector search + qa_events |
| `internal/scope` | Agent/topic → km_scope |

## Isolation (shared dev host)

| Resource | Monti Jarvis | Jarvis Chat (do not share) |
| --- | --- | --- |
| Postgres DB | `monti_jarvis` | `jarvis_chat` |
| Schema | `callcenter` | `chat` |
| Redis DB | `4` | `3` |
| Prefix | `monti_jarvis:` | `jarvis_chat:` |
| Port | `8091` | `8090` |

## Deployment unit (dev)

Single binary `monti-jarvis` + static `apps/customer-web/build`. Infra via `make up` (compose NATS/LiveKit + shared `poc-gml-*`).

## Future (roadmap)

- Sprint 3+: auth middleware, tenant isolation
- Sprint 15: tenant admin KM wizard
- Production: separate customer/tenant/admin Svelte apps; Go API behind load balancer

See [workflow.md](workflow.md), [er-diagram.md](er-diagram.md), [api-spec.md](api-spec.md).