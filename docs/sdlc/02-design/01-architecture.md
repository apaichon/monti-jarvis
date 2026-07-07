---
id: DES-0001
title: System Architecture
status: review_pending
updated: 2026-07-07
sprint: SPRINT-003
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
│  ┌──────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────┐│
│  │ auth     │ │ workforce  │ │ calls      │ │ km + rag   │ │ gemini ││
│  │/api/auth │ │ /api/chat  │ │ /api/calls │ │ /api/km/*  │ │/ws/voice│
│  └────┬─────┘ └─────┬──────┘ └─────┬──────┘ └─────┬──────┘ └───┬────┘│
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
| `internal/auth` | JWT, middleware, RBAC *(Sprint 3 — planned)* |

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

## Auth layer (Sprint 3 — draft)

```text
AUTH_DISABLED=true  → AuthContext{TenantID: demo} — v0.3.0 behavior
AUTH_DISABLED=false → Bearer JWT → middleware → RBAC on KM writes
```

Customer portal (`/`) stays **unauthenticated** for inbound demo. Admin KM ops use Bearer token (curl / future admin UI).

## Platform admin portal (Sprint 4 — in progress)

```text
apps/platform-admin-web/build  →  GET /admin/*  (internal/platformweb)
  /admin/login     → POST /api/auth/login
  /admin/profile   → GET /api/auth/me
  /admin/packages  → GET/POST/PUT/DELETE /api/platform/packages*
  /admin/tenants/{id}/entitlement → entitlement APIs
```

## Packages layer (Sprint 4)

```text
AUTH_DISABLED=false → platform_admin via portal or API
                    → rules jsonb + package_rule_schemas versions
                    → entitlements.Resolve + Redis cache (no enforcement yet)
```

Customer portal (`/`) stays public demo. Tenant admin UI deferred Sprint 15+.

## Future (roadmap)
- Sprint 15: tenant admin KM wizard
- Sprint 19–20: customer register/auth UI
- Production: Go API behind load balancer; three Svelte apps (customer / tenant / platform)

See [06-auth-spec.md](06-auth-spec.md), [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), [04-api-spec.md](04-api-spec.md).