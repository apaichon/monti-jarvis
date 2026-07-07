---
id: DES-0002
title: Workflows
status: review_pending
updated: 2026-07-07
sprint: SPRINT-003
---

# Workflows — Monti Jarvis

## 1. Portal load

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go :8091
  participant W as workforce

  B->>G: GET /
  G-->>B: Svelte SPA
  B->>G: GET /api/workforce
  G->>W: All()
  G-->>B: agents[]
  B->>G: GET /api/infra
  G-->>B: postgres, redis, minio, clickhouse
```

## 2. Text chat (with RAG)

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go server
  participant R as rag
  participant CH as ClickHouse
  participant AI as Gemini text

  B->>G: POST /api/chat {agent_id, topic, message, history}
  G->>R: Retrieve(agent, topic, message)
  R->>CH: SearchScoped (embed query)
  CH-->>R: top-k chunks
  R-->>G: sources + context block
  G->>AI: Reply(augmented prompt, history)
  AI-->>G: reply text
  G->>G: SaveExchange (Postgres)
  G-->>B: {reply, sources[], missing_km}
```

## 3. Voice call

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go server
  participant R as rag
  participant L as Gemini Live
  participant P as Postgres

  B->>G: POST /api/calls
  G->>P: CreateCallSession
  G-->>B: session {id, room_name}
  B->>G: WS /ws/voice?agent=&topic=
  G->>R: Retrieve(agent, topic, "") preload KB
  G->>L: setup(systemInstruction + KB)
  loop conversation
    B->>G: audio frames (JSON)
    G->>L: realtimeInput
    L-->>G: audio + transcript
    G-->>B: transcript events
    G->>P: addTurn (optional)
  end
  B->>G: POST /api/calls/{id}/end
```

## 4. KM ingest (per avatar)

```mermaid
sequenceDiagram
  participant Op as Operator/curl
  participant G as Go server
  participant M as MinIO
  participant P as Postgres
  participant AI as Gemini embed
  participant CH as ClickHouse

  Op->>G: POST /api/km/agents/{id}/documents (multipart)
  G->>M: Put km/demo/{agent}/...
  G->>P: knowledge_documents row
  G->>G: ChunkText
  loop each chunk
    G->>AI: Embed(chunk)
    AI-->>G: vector
  end
  G->>P: knowledge_chunks rows
  G->>CH: km_embeddings upsert
  G-->>Op: document {status: indexed, chunk_count}
```

## 5. KM reset (per avatar)

```mermaid
sequenceDiagram
  participant Op as Operator
  participant G as Go server
  participant P as Postgres
  participant M as MinIO
  participant CH as ClickHouse

  Op->>G: POST /api/km/agents/{id}/reset
  G->>P: DELETE documents + chunks
  G->>M: DELETE object keys
  G->>CH: DELETE embeddings for agent
  G-->>Op: {status: reset}
```

## 6. Call events (SSE)

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go server
  participant N as NATS

  B->>G: GET /api/calls/{id}/events (SSE)
  Note over G,N: turn persisted
  G->>N: call.turn.created (optional)
  G-->>B: event: turn {role, content}
```

## 6. Auth login (Sprint 3 — draft)

```mermaid
sequenceDiagram
  participant C as Client (curl/admin)
  participant G as Go :8091
  participant DB as Postgres
  participant A as internal/auth

  C->>G: POST /api/auth/login {email, password}
  G->>DB: lookup user + role
  G->>A: verify bcrypt
  A->>A: issue access JWT + refresh opaque
  G->>DB: insert refresh_tokens (hash)
  G-->>C: {access_token, refresh_token, user}
```

## 7. Protected KM upload (auth enabled)

```mermaid
sequenceDiagram
  participant C as Client
  participant G as Go server
  participant M as auth middleware
  participant KM as internal/km

  C->>G: POST /api/km/.../documents + Bearer
  G->>M: validate JWT, check tenant_admin
  alt forbidden
    M-->>C: 403
  else ok
    M->>KM: Ingest(tenant_id from context)
    KM-->>C: 201 document
  end
```

## 8. Dev bypass (`AUTH_DISABLED=true`)

No login required. All handlers use `tenant_id = DEMO_TENANT_ID`. Identical to v0.3.0 flows above.

## State: call session

| Status | Meaning |
| --- | --- |
| `active` | Call in progress; Redis key `monti_jarvis:call:active:{id}` |
| `ended` | `ended_at` set; Redis key removed |

## State: knowledge document

| Status | Meaning |
| --- | --- |
| `uploaded` | MinIO object stored |
| `indexing` | Chunk + embed in progress |
| `indexed` | Postgres + ClickHouse ready |
| `failed` | Embed or index error |

See [auth-spec.md](auth-spec.md), [api-spec.md](api-spec.md), [ux-ui.md](ux-ui.md).