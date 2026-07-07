# Monti Jarvis — Inbound Call Center

Multi-tenant AI call center platform. **Sprint 1** delivers the customer conversation portal: Svelte + shadcn UI, LiveKit voice rooms, Postgres sessions, Redis 8 state, and NATS lifecycle events.

## Current sprint (SPRINT-001)

- **Customer portal** — `apps/customer-web` (SvelteKit + Tailwind + shadcn-style components)
- **LiveKit calls** — `POST /api/calls`, token issue, mic join, end call
- **Transcript** — SSE stream + turn history
- **Legacy v0.1.0** — set `LEGACY_UI_ENABLED=true` for `/legacy` (Gemini direct voice)

## Run

```bash
cp infra/.env.dev.example infra/.env.dev
make up      # destroy + init infra, start server (:8091)
```

Or step by step:

```bash
make infra-reset   # destroy + init (NATS, LiveKit, DB, Redis, MinIO)
make start         # build + start server
make stop          # stop server
make restart       # stop + start server
make down          # stop server + destroy infra
```

Open http://localhost:8091.

**Dev mode** (hot reload UI):

```bash
make run          # API on :8091
make customer-dev # Svelte on :5173 (proxies /api)
```

## Commands

```bash
make run          # foreground server
make start        # background server, logs in .run/server.log
make stop
make status
make logs
make infra-check  # verify shared Postgres, Redis, MinIO containers/ports
make infra-init   # create monti_jarvis DB/schema and MinIO bucket
make test
```

## API (Sprint 1)

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/healthz` | Liveness + LiveKit/NATS status |
| POST | `/api/calls` | Create call session + LiveKit room |
| GET | `/api/calls/{id}` | Call session status |
| POST | `/api/calls/{id}/token` | LiveKit join token |
| POST | `/api/calls/{id}/end` | End call |
| GET | `/api/calls/{id}/turns` | Transcript turns |
| GET | `/api/calls/{id}/events` | SSE transcript stream |
| GET | `/api/infra` | Postgres / Redis / MinIO / NATS / LiveKit |

## Infra isolation

Reuses the shared local stack but keeps Monti data separate from Jarvis Chat:

| Resource | Monti Jarvis | Jarvis Chat (reference) |
| --- | --- | --- |
| Postgres DB | `monti_jarvis` | `jarvis_chat` |
| Postgres schema | `callcenter` | `chat` |
| Redis DB index | `4` | `3` |
| Redis prefix | `monti_jarvis:` | `jarvis_chat:` |
| MinIO bucket | `monti-jarvis` | `jarvis-chat` |

## Workforce

| Agent | Role | Voice |
| --- | --- | --- |
| Ava | General Support | Aoede |
| Max | Billing Specialist | Charon |
| Luna | Technical Support | Kore |
| Neo | Triage Bot | Puck |

## Planned later

- Caller authentication and account lookup
- Ticket creation and CRM integration
- Knowledge-base search over pgvector
- Outbound campaigns and supervisor dashboard
- Call recording and compliance retention