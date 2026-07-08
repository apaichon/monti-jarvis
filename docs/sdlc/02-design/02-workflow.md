---
id: DES-0002
title: Workflows
status: approved
updated: 2026-07-07
sprint: SPRINT-005
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

## 9. Package catalog CRUD (Sprint 4)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant M as auth middleware
  participant P as internal/packages
  participant DB as Postgres

  Op->>G: POST /api/platform/packages {slug, name, limits, ...}
  G->>M: validate JWT + platform_admin
  alt forbidden
    M-->>Op: 403
  else ok
    M->>P: Create(ctx, input)
    P->>DB: validate rules vs package_rule_schemas
    P->>DB: INSERT packages + package_limits (rules jsonb)
    G-->>Op: 201 {id, slug, limits, ...}
  end
```

## 10. Assign tenant entitlement (Sprint 4)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant E as internal/entitlements
  participant DB as Postgres
  participant R as Redis

  Op->>G: POST /api/platform/tenants/demo/entitlement {package_id}
  G->>G: RBAC platform_admin
  G->>DB: revoke prior active row (if any)
  G->>DB: INSERT tenant_entitlements + rules_snapshot + rules_schema_id
  G->>R: DEL monti_jarvis:entitlement:demo
  G-->>Op: 200 effective entitlement JSON
```

## 11. Entitlement resolve + cache (Sprint 4)

```mermaid
sequenceDiagram
  participant C as Client (tenant_admin)
  participant G as Go :8091
  participant E as internal/entitlements
  participant R as Redis
  participant DB as Postgres

  C->>G: GET /api/entitlements/me + Bearer
  G->>G: tenant_id from JWT
  G->>E: Resolve(tenant_id)
  E->>R: GET monti_jarvis:entitlement:{tenant_id}
  alt cache hit
    R-->>E: cached JSON
  else cache miss
    E->>DB: tenant_entitlements JOIN packages JOIN package_limits (rules jsonb)
    E->>R: SETEX key TTL payload
  end
  E-->>G: effective limits
  G-->>C: 200 {tenant_id, package, limits, status}
```

## State: package (Sprint 4)

| Status | Meaning |
| --- | --- |
| `draft` | Not assignable; hidden from default list |
| `active` | Assignable to tenants |
| `archived` | No new assignments; existing entitlements honored until revoked |

## State: tenant entitlement (Sprint 4)

| Status | Meaning |
| --- | --- |
| `active` | Tenant receives package limits (at most one per tenant) |
| `suspended` | Limits withheld; row kept for audit |
| `revoked` | Operator ended entitlement; resolver returns fallback |
| `expired` | `valid_until` passed (Sprint 9+ subscriptions) |

## 12. Platform admin login (Sprint 4)

```mermaid
sequenceDiagram
  participant B as Browser /admin
  participant G as Go :8091
  participant A as internal/auth

  B->>G: POST /api/auth/login {email, password}
  G->>A: Login + issue tokens
  G-->>B: {access_token, refresh_token, user}
  B->>B: sessionStorage tokens
  B->>G: GET /admin/packages (SPA)
  G-->>B: platform-admin-web index.html
  B->>G: GET /api/platform/packages + Bearer
  G-->>B: packages[]
```

## 13. Platform admin logout (Sprint 4)

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go :8091

  B->>G: POST /api/auth/logout + Bearer
  G-->>B: 200
  B->>B: clear sessionStorage
  B->>B: navigate /admin/login
```

## 14. Avatar catalog CRUD (Sprint 5)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant M as auth middleware
  participant S as internal/store avatars
  participant DB as Postgres

  Op->>G: POST /api/platform/avatars {slug, name, role, voice, ...}
  G->>M: validate JWT + platform_admin
  alt forbidden
    M-->>Op: 403
  else ok
    M->>S: CreateAvatar(ctx, input)
    S->>DB: INSERT ai_avatars (flags jsonb)
    G-->>Op: 201 {id, slug, name, status, ...}
  end
```

## 15. Assign tenant avatar (Sprint 5)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant E as internal/entitlements
  participant S as internal/store avatars
  participant DB as Postgres

  Op->>G: POST /api/platform/tenants/demo/avatars {avatar_id}
  G->>G: RBAC platform_admin
  G->>E: GetEffective(demo) → rules.max_ai_employees
  G->>S: CountActiveAssignments(demo)
  alt at cap
    G-->>Op: 409 max_ai_employees exceeded
  else ok
    S->>DB: UPSERT tenant_avatar_assignments status=active
    G-->>Op: 200 {tenant_id, avatar, status}
  end
```

## 16. Workforce resolve (Sprint 5)

```mermaid
sequenceDiagram
  participant B as Browser (customer /)
  participant G as Go :8091
  participant W as internal/workforce
  participant S as internal/store avatars
  participant A as internal/auth

  B->>G: GET /api/workforce
  Note over B,G: Optional X-Tenant-Id or Bearer tenant
  G->>A: ResolveTenant(ctx, header, authDisabled, demo)
  G->>S: ListAssignedAvatars(tenant_id, active)
  alt has assignments
    S-->>G: ai_avatars + primary ai_avatar_voices row
    G->>W: map to Agent JSON (image_url → image, voice from priority 1)
  else no assignments
    G->>W: All() static fallback
  end
  G-->>B: 200 {agents: [...]}
```

## State: avatar (Sprint 5)

| Status | Meaning |
| --- | --- |
| `draft` | Not assignable; hidden from default platform list |
| `active` | Assignable to tenants; eligible for workforce when assigned |
| `archived` | No new assignments; existing assignments may be disabled by operator |

## State: tenant avatar assignment (Sprint 5)

| Status | Meaning |
| --- | --- |
| `active` | Avatar appears in tenant `/api/workforce` list |
| `disabled` | Assignment revoked; avatar hidden from tenant workforce |

## State: avatar voice profile (Sprint 5)

| Status | Meaning |
| --- | --- |
| `active` | Eligible for primary selection or failover (by `priority`) |
| `disabled` | Skipped by resolver; kept for audit / future enable |

**Failover order:** ascending `priority` among `active` rows for the same `avatar_id`. Sprint 21 applies this during live calls.

## 17. Customer portal agent pick (unchanged UI, Sprint 5 data)

Customer portal still calls `GET /api/workforce` on load. Sprint 5 only changes **data source** when tenant has assignments; UI components unchanged.

See [06-auth-spec.md](06-auth-spec.md), [08-packages-spec.md](08-packages-spec.md), [10-avatars-spec.md](10-avatars-spec.md), [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md), [04-api-spec.md](04-api-spec.md), [05-ux-ui.md](05-ux-ui.md).