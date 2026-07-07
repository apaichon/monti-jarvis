---
id: DEPLOY-LOCAL
status: active
updated: 2026-07-07
environment: local-dev
---

# Local Development Deployment

Single-machine setup for Monti Jarvis on port **8091**.

## Architecture (dev)

```text
Browser → monti-jarvis Go server (:8091)
            ├── Svelte customer-web (embedded build)
            ├── /api/chat, /ws/voice (Gemini)
            ├── /api/km/* (ingest + RAG)
            └── /legacy/ (embedded HTML)

Datastores (shared Docker):
  poc-gml-postgres   → DB monti_jarvis, schema callcenter
  poc-gml-redis      → DB index 4, prefix monti_jarvis:
  poc-gml-minio      → bucket monti-jarvis (calls/, km/)
  poc-gml-clickhouse → DB monti_jarvis (km_embeddings, qa_events)

Monti compose (infra/docker-compose.yml):
  monti-nats, monti-livekit
```

## Prerequisites

| Tool | Version |
| --- | --- |
| Go | 1.22+ |
| Node | 20+ |
| Docker | Desktop or engine with compose |
| curl | any |

## First-time setup

```bash
cd /path/to/monti-jarvis

# Environment
cp infra/.env.dev.example infra/.env.dev
# Edit infra/.env.dev — set GEMINI_API_KEY (required)

# Start shared datastores (if not already running)
docker start poc-gml-postgres poc-gml-redis poc-gml-minio poc-gml-clickhouse

# Full reset + init + start server
make up
```

`make up` runs: `infra-destroy` → `infra-up` → `infra-init` → `start`.

## Day-to-day commands

| Command | Purpose |
| --- | --- |
| `make start` | Build portal + Go binary; run server in background |
| `make stop` | Stop background server |
| `make restart` | Stop then start |
| `make status` | PID + `/healthz` |
| `make logs` | Tail `.run/server.log` |
| `make run` | Foreground server (debug) |
| `make customer-dev` | Vite on :5173 (proxies API to :8091) |
| `make test` | `go test ./...` |
| `make km-seed` | Ingest sample KB for all avatars |
| `make infra-check` | Health probe all dependencies |
| `make down` | Stop server + destroy Monti-managed infra |

## Environment variables (`infra/.env.dev`)

| Variable | Required | Default / notes |
| --- | --- | --- |
| `GEMINI_API_KEY` | **Yes** | Chat, voice, embeddings |
| `PORT` | No | `8091` |
| `DATABASE_URL` | No | Points to `monti_jarvis` on shared Postgres |
| `REDIS_URL` | No | DB index `4` |
| `MINIO_*` | No | Bucket `monti-jarvis` |
| `CLICKHOUSE_URL` | Sprint 2+ | `http://localhost:8123` |
| `CLICKHOUSE_DB` | No | `monti_jarvis` |
| `DEMO_TENANT_ID` | No | `demo` |
| `NATS_URL` | No | Optional lifecycle events |
| `LIVEKIT_*` | No | Optional token API |

See `infra/.env.example` for the full list.

## Verify deployment

```bash
make status
curl -fsS http://localhost:8091/healthz | python3 -m json.tool
curl -fsS http://localhost:8091/api/infra | python3 -m json.tool
open http://localhost:8091
```

Sprint 2 additionally:

```bash
make km-seed
curl http://localhost:8091/api/km/agents/ava
```

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| `health: unreachable` | `make logs`; check port conflict on 8091 |
| `postgres container not found` | `docker start poc-gml-postgres` |
| `clickhouse: disabled` | Start `poc-gml-clickhouse`; set `CLICKHOUSE_URL` |
| Portal 404 on `/` | `make customer-web` then `make restart` |
| Embed/chat 401/403 | Valid `GEMINI_API_KEY` in `infra/.env.dev` |
| Stale server after code change | `make restart` |

## Clean teardown

```bash
make down
# Optionally stop shared containers:
docker stop poc-gml-postgres poc-gml-redis poc-gml-minio poc-gml-clickhouse
```