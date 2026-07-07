---
id: DES-0003
title: Entity Relationship Diagram
status: review_pending
updated: 2026-07-08
sprint: SPRINT-003
---

# ER Diagram — Monti Jarvis

Database `monti_jarvis`, Postgres schema `callcenter`. ClickHouse database `monti_jarvis` for vectors/analytics.

## Postgres (`callcenter`)

```mermaid
erDiagram
  tenants ||--o{ user_roles : scopes
  users ||--o{ user_roles : has
  users ||--o{ refresh_tokens : has
  tenants ||--o{ call_sessions : owns
  tenants ||--o{ knowledge_documents : owns
  embedding_models ||--o{ knowledge_index_runs : indexes_with
  voice_providers ||--o{ call_sessions : uses
  voice_providers ||--o{ ai_employee_configs : default_voice
  knowledge_documents ||--o{ knowledge_index_runs : versioned_by
  knowledge_index_runs ||--o{ knowledge_chunks : produces
  calls ||--o{ messages : has
  call_sessions ||--o{ call_turns : has

  tenants {
    text id PK
    text slug UK
    text name
    text status
    timestamptz created_at
  }

  users {
    text id PK
    text email UK
    text password_hash
    text display_name
    text status
    timestamptz created_at
  }

  user_roles {
    text user_id FK
    text role
    text tenant_id FK
    timestamptz created_at
  }

  refresh_tokens {
    text id PK
    text user_id FK
    text token_hash UK
    timestamptz expires_at
    timestamptz revoked_at
    timestamptz created_at
  }

  calls {
    text id PK
    text agent_id
    text title
    timestamptz created_at
    timestamptz updated_at
  }

  messages {
    bigserial id PK
    text call_id FK
    text role
    text content
    timestamptz created_at
  }

  call_sessions {
    text id PK
    text tenant_id
    text room_name UK
    text voice_provider_id FK
    text status
    timestamptz started_at
    timestamptz ended_at
    text recording_key
  }

  call_turns {
    bigserial id PK
    text call_id FK
    text role
    text content
    jsonb source_chunk_ids
    timestamptz created_at
  }

  embedding_models {
    text id PK
    text provider
    text model_key
    int dimensions
    text status
    timestamptz created_at
  }

  voice_providers {
    text id PK
    text provider
    text model_key
    text modality
    text status
    timestamptz created_at
  }

  ai_employee_configs {
    text agent_id PK
    text tenant_id FK
    text voice_provider_id FK
    text text_provider_id FK
    text embedding_model_id FK
    timestamptz updated_at
  }

  knowledge_documents {
    text id PK
    text tenant_id
    text agent_id
    text filename
    text object_key
    text mime
    text status
    text km_scope
    int content_version
    int chunk_count
    timestamptz created_at
    timestamptz updated_at
  }

  knowledge_index_runs {
    text id PK
    text document_id FK
    text tenant_id
    text embedding_model_id FK
    int index_version
    text status
    timestamptz started_at
    timestamptz completed_at
  }

  knowledge_chunks {
    text id PK
    text document_id FK
    text index_run_id FK
    text tenant_id
    text agent_id
    int chunk_index
    text content
    text km_scope
    text embedding_model_id FK
    int index_version
    timestamptz created_at
  }
```

### Table notes

| Table | Purpose |
| --- | --- |
| `calls` | Legacy chat session ids from `/api/chat` |
| `messages` | Text chat transcript pairs (caller/agent) |
| `call_sessions` | Voice call sessions; `voice_provider_id` for Gemini/Grok switch *(planned)* |
| `call_turns` | Voice/text turns per call session |
| `knowledge_documents` | KM upload metadata; `content_version` bumps on file replace |
| `knowledge_index_runs` | One embed/index pass per document + **embedding model** + `index_version` *(planned)* |
| `knowledge_chunks` | Chunk text; FK to `index_run_id` and `embedding_model_id` |
| `embedding_models` | Catalog: `gemini-embedding-001`, future models; drives vector compatibility |
| `voice_providers` | Catalog: Gemini Live, Grok Voice, etc. — switchable per agent/tenant |
| `ai_employee_configs` | Per-agent binding of voice, text, and embedding providers *(Sprint 21)* |
| `tenants` | SaaS tenant registry *(Sprint 3)* |
| `users` | Login identities *(Sprint 3)* |
| `user_roles` | RBAC role per user/tenant *(Sprint 3)* |
| `refresh_tokens` | Hashed refresh tokens *(Sprint 3)* |

### Indexes

- `knowledge_documents (tenant_id, agent_id)`
- `knowledge_chunks (tenant_id, agent_id, embedding_model_id, index_version)`
- `knowledge_index_runs (document_id, embedding_model_id, index_version)`
- `call_sessions (tenant_id, voice_provider_id)` *(planned)*

### KM versioning model (planned)

Two version axes keep provider switches safe:

| Field | Layer | Bumps when |
| --- | --- | --- |
| `content_version` | Postgres `knowledge_documents` | Operator uploads/replaces source file |
| `index_version` | `knowledge_index_runs` + chunks + ClickHouse | Re-embed with same or new **embedding model** |

Search and RAG **must** filter on `(tenant_id, embedding_model_id, index_version)` so vectors from `gemini-embedding-001` are never compared to queries from `grok-embed` or another model.

## Provider catalog (Postgres — planned)

```mermaid
erDiagram
  embedding_models ||--o{ knowledge_index_runs : model
  embedding_models ||--o{ knowledge_chunks : embedded_by
  voice_providers ||--o{ call_sessions : voice
  voice_providers ||--o{ ai_employee_configs : voice_default

  embedding_models {
    text id PK
    text provider
    text model_key
    int dimensions
  }

  voice_providers {
    text id PK
    text provider
    text model_key
    text modality
  }
```

Example seed rows:

| `id` | `provider` | `model_key` | Role |
| --- | --- | --- | --- |
| `emb-gemini-001` | `google` | `gemini-embedding-001` | KM embed + query |
| `voice-gemini-live` | `google` | `gemini-2.5-flash-native-audio-latest` | Voice relay |
| `voice-grok` | `xai` | `grok-voice-*` | Future voice failover |

Tenant/agent config (`ai_employee_configs`) points at active `voice_provider_id` and `embedding_model_id`. Switching voice provider does **not** invalidate KM index; switching **embedding model** requires a new `index_version` and re-embed.

## ClickHouse (`monti_jarvis`)

```mermaid
erDiagram
  knowledge_chunks ||--o| km_embeddings : "chunk_id + model + version"
  embedding_models ||--o{ km_embeddings : indexed_as

  km_embeddings {
    string tenant_id
    string agent_id
    string document_id
    string chunk_id
    string embedding_model_id
    string embedding_provider
    uint32 content_version
    uint32 index_version
    string km_scope
    string content
    array_float32 embedding
    datetime updated_at
  }

  qa_events {
    string event_id
    string tenant_id
    string agent_id
    string topic
    string question
    string event_type
    string embedding_model_id
    datetime created_at
  }

  call_provider_events {
    string event_id
    string call_id
    string tenant_id
    string voice_provider_id
    string event_type
    datetime created_at
  }
```

### ClickHouse notes

| Column | Purpose |
| --- | --- |
| `embedding_model_id` | FK to Postgres catalog; **required** on search filter |
| `embedding_provider` | Denormalized vendor (`google`, `xai`) for analytics |
| `content_version` | Matches Postgres document content generation |
| `index_version` | Matches `knowledge_index_runs`; invalidate old vectors on re-index |
| `chunk_id` | Join key to Postgres `knowledge_chunks.id` |

**Search rule:** `WHERE tenant_id = ? AND agent_id = ? AND embedding_model_id = ? AND index_version = active AND km_scope IN (?)` — never mix models in cosine ranking.

**Sprint 2 today:** `km_embeddings` uses `km_version` as a single field; migration path renames/splits into `content_version` + `index_version` + `embedding_model_id` (default `emb-gemini-001` for existing rows).

## Cross-store relationships

```text
Postgres knowledge_chunks.id  ──►  ClickHouse km_embeddings.chunk_id
Postgres embedding_models.id  ──►  km_embeddings.embedding_model_id
Postgres voice_providers.id   ──►  call_sessions.voice_provider_id
Postgres index_run.id         ──►  knowledge_chunks.index_run_id
```

Voice path (Gemini → Grok failover) updates `call_sessions.voice_provider_id` and logs `call_provider_events`; KM vectors stay tied to **embedding model**, not voice provider.

## MinIO (object keys)

```
monti-jarvis/
  calls/                    # future recordings
  km/{tenant_id}/{agent_id}/{doc_id}/original/{filename}
```

## Redis (ephemeral)

| Key pattern | TTL | Fields |
| --- | --- | --- |
| `monti_jarvis:call:{session_id}` | 24h | agent_id, updated_at (legacy chat) |
| `monti_jarvis:call:active:{id}` | 24h | tenant_id, room_name, status, started_at |

## Workforce (in-memory, not DB)

Agents `ava`, `max`, `luna`, `neo` defined in `internal/workforce/workforce.go` — Sprint 21 will move to Postgres catalog.

## Implementation phases

| Phase | KM / provider scope |
| --- | --- |
| **v0.3.0 (shipped)** | Single embed model env var; `km_version` column; Gemini voice only |
| **Sprint 3–4** | Provider catalog tables stub; default `emb-gemini-001` / `voice-gemini-live` seeds |
| **Sprint 15** | Tenant selects embedding model; re-index workflow bumps `index_version` |
| **Sprint 21** | `ai_employee_configs` binds voice + embed providers per avatar |
| **Failover** | Runtime voice provider switch via `voice_providers`; KM unchanged until embed model changes |

## Future entities (roadmap)

| Sprint | Tables |
| --- | --- |
| 4+ | `packages`, `entitlements` |
| 6+ | `tenant_registrations`, `brands` |
| 15 | `km_scope_assignments`, tenant-driven re-index |
| 21 | `ai_employees`, `ai_employee_configs` (full catalog) |
| 22 | `conversation_records` (ClickHouse denorm) |

See [architecture.md](architecture.md) · blueprint §15.3 Embedding Provider · §16.4 KM domains.