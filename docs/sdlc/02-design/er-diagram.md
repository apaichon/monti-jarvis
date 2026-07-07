---
id: DES-0003
title: Entity Relationship Diagram
status: active
updated: 2026-07-07
sprint: SPRINT-002
---

# ER Diagram — Monti Jarvis

Database `monti_jarvis`, Postgres schema `callcenter`. ClickHouse database `monti_jarvis` for vectors/analytics.

## Postgres (`callcenter`)

```mermaid
erDiagram
  calls ||--o{ messages : has
  call_sessions ||--o{ call_turns : has
  knowledge_documents ||--o{ knowledge_chunks : has

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

  knowledge_documents {
    text id PK
    text tenant_id
    text agent_id
    text filename
    text object_key
    text mime
    text status
    text km_scope
    int km_version
    int chunk_count
    timestamptz created_at
    timestamptz updated_at
  }

  knowledge_chunks {
    text id PK
    text document_id FK
    text tenant_id
    text agent_id
    int chunk_index
    text content
    text km_scope
    timestamptz created_at
  }
```

### Table notes

| Table | Purpose |
| --- | --- |
| `calls` | Legacy chat session ids from `/api/chat` |
| `messages` | Text chat transcript pairs (caller/agent) |
| `call_sessions` | Sprint 1 voice call sessions + tenant |
| `call_turns` | Voice/text turns per call session |
| `knowledge_documents` | KM upload metadata per agent |
| `knowledge_chunks` | Chunk text; links to ClickHouse `chunk_id` |

### Indexes

- `knowledge_documents (tenant_id, agent_id)`
- `knowledge_chunks (tenant_id, agent_id)`

## ClickHouse (`monti_jarvis`)

```mermaid
erDiagram
  km_embeddings {
    string tenant_id
    string agent_id
    string document_id
    string chunk_id
    string km_scope
    uint32 km_version
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
    datetime created_at
  }
```

Vectors are keyed by `(tenant_id, agent_id, document_id, chunk_id)`. Search filters `km_scope` per topic/agent rules in `internal/scope`.

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

## Future entities (roadmap)

| Sprint | Tables |
| --- | --- |
| 3 Auth | `users`, `sessions`, `tenant_users` |
| 6+ | `tenants`, `brands`, `packages` |
| 15 | `km_scope_assignments` (tenant admin) |
| 22 | `conversation_records` (ClickHouse denorm) |

See [architecture.md](architecture.md).