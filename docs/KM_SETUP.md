# Knowledge Base Setup — Sprint 2

Per-avatar knowledge bases ground chat and voice answers in approved documents. Each workforce agent has its own KB and default scope.

**Related SDLC:** [UAT checklist](sdlc/06-manual-tests/SPRINT-002-manual.md) · [test matrix](sdlc/05-test-scenarios/TEST-MATRIX.md) · [local deploy](sdlc/07-deployment/LOCAL-DEV.md)

| Agent | Scope | Sample file |
| --- | --- | --- |
| Ava | `general` | `docs/samples/km/ava.md` |
| Max | `billing` | `docs/samples/km/max.md` |
| Luna | `technical` | `docs/samples/km/luna.md` |
| Neo | `general`, `billing`, `technical` (triage) | `docs/samples/km/neo.md` |

## Prerequisites

```bash
# Shared datastores + Monti compose
docker start poc-gml-postgres poc-gml-redis poc-gml-minio poc-gml-clickhouse 2>/dev/null || true
make infra-up
make infra-init

# Env
cp infra/.env.dev.example infra/.env.dev
# Set GEMINI_API_KEY (required for embed + chat)
```

`infra/.env.dev` should include:

```env
CLICKHOUSE_URL=http://localhost:8123
CLICKHOUSE_DB=monti_jarvis
CLICKHOUSE_USER=monti
CLICKHOUSE_PASSWORD=monti
DEMO_TENANT_ID=demo
```

## Quick start — seed all avatars

```bash
make start
curl -X POST http://localhost:8091/api/km/seed
```

This ingests the four sample Markdown files under `docs/samples/km/`.

## API reference

Canonical spec: [`docs/sdlc/02-design/api-spec.md`](sdlc/02-design/api-spec.md) · UX mapping: [`ux-ui.md`](sdlc/02-design/ux-ui.md)

**REST client:** [`docs/km-setup.http`](km-setup.http) — runnable requests for VS Code REST Client or IntelliJ HTTP Client (health, seed, upload, reset, RAG chat).

```text
# VS Code: install "REST Client" (humao.rest-client), open km-setup.http, click "Send Request"
# IntelliJ: open km-setup.http, click green gutter icons
# Suggested flow: infra → kmSeedAll → chatMaxBillingGrounded (expect sources[])
```

Base URL: `http://localhost:8091`  
Tenant header (optional): `X-Tenant-Id: demo` (default)

### Get agent KB status

```http
GET /api/km/agents/{agent_id}
```

Example:

```bash
curl http://localhost:8091/api/km/agents/max
```

Response:

```json
{
  "agent_id": "max",
  "tenant_id": "demo",
  "scope": "billing",
  "document_count": 1,
  "chunk_count": 4,
  "documents": [...]
}
```

### List documents for an agent

```http
GET /api/km/agents/{agent_id}/documents
```

### Upload a document for an agent

```http
POST /api/km/agents/{agent_id}/documents
Content-Type: multipart/form-data
```

Form fields:

| Field | Required | Description |
| --- | --- | --- |
| `file` | yes | `.md` or `.txt` knowledge file |
| `scope` | no | Defaults to agent scope (`general`, `billing`, `technical`) |

Example — upload billing FAQ for Max:

```bash
curl -X POST http://localhost:8091/api/km/agents/max/documents \
  -F "file=@docs/samples/km/max.md" \
  -F "scope=billing"
```

Pipeline: **MinIO** (`km/demo/{agent}/…`) → chunk → **Gemini embed** → **ClickHouse** `km_embeddings` + Postgres metadata.

### Reset an agent knowledge base

Clears Postgres rows, MinIO objects, and ClickHouse embeddings for that agent.

```http
POST /api/km/agents/{agent_id}/reset
```

Examples:

```bash
# Reset Max only
curl -X POST http://localhost:8091/api/km/agents/max/reset

# Reset all four avatars
for a in ava max luna neo; do
  curl -X POST "http://localhost:8091/api/km/agents/$a/reset"
done
```

### Seed demo knowledge (all agents)

```http
POST /api/km/seed
```

## Manual setup per avatar

### 1. Ava (general support)

```bash
curl -X POST http://localhost:8091/api/km/agents/ava/documents \
  -F "file=@docs/samples/km/ava.md" \
  -F "scope=general"
```

Ask on the **General** tab: *"What are your business hours?"*

### 2. Max (billing)

```bash
curl -X POST http://localhost:8091/api/km/agents/max/documents \
  -F "file=@docs/samples/km/max.md" \
  -F "scope=billing"
```

Ask on the **Billing** tab: *"When are invoices due?"*

### 3. Luna (technical)

```bash
curl -X POST http://localhost:8091/api/km/agents/luna/documents \
  -F "file=@docs/samples/km/luna.md" \
  -F "scope=technical"
```

Ask on the **Technical** tab: *"My microphone is not working on calls"*

### 4. Neo (triage)

```bash
curl -X POST http://localhost:8091/api/km/agents/neo/documents \
  -F "file=@docs/samples/km/neo.md" \
  -F "scope=general"
```

## Verify RAG is working

```bash
# Infra health
curl http://localhost:8091/api/infra
# expect clickhouse: ok

# Grounded chat
curl -X POST http://localhost:8091/api/chat \
  -H 'content-type: application/json' \
  -d '{"agent_id":"max","topic":"billing","message":"When are invoices due?","history":[]}'
# expect sources[] with billing excerpts
```

Portal: http://localhost:8091 — citation chips appear under assistant replies when KM matches.

## Reset and rebuild everything

```bash
# Clear all avatar KBs
for a in ava max luna neo; do curl -X POST "http://localhost:8091/api/km/agents/$a/reset"; done

# Re-seed
curl -X POST http://localhost:8091/api/km/seed
```

Full infra reset (includes DB drop — not KM-only):

```bash
make down && make up
curl -X POST http://localhost:8091/api/km/seed
```

## Scope rules

- **Topic tab** filters which `km_scope` tags are searched.
- **Agent** defines which scopes they may access (see `internal/scope/scope.go`).
- Ava on the Technical tab still searches `general` scope only.

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| `clickhouse: disabled` | Run `make infra-up` (starts `monti-clickhouse` on `:8123`); set `CLICKHOUSE_*` in `.env.dev` |
| No `sources` in chat | Run seed/upload; check agent + topic match scope |
| Embed errors | Set `GEMINI_API_KEY`; verify API quota |
| Upload 502 | Ensure MinIO bucket `monti-jarvis` exists (`make infra-init`) |