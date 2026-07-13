---
id: DES-0003
title: Entity Relationship Diagram
status: approved
updated: 2026-07-12
sprint: SPRINT-019
---

# ER Diagram — Monti Jarvis

Database `monti_jarvis`, Postgres schema `callcenter`. ClickHouse database `monti_jarvis` for vectors/analytics.

## Audit columns (standard)

Every durable Postgres table in `callcenter` carries four audit fields (migration `001_audit_columns_postgres`, `internal/store/audit.go`):

| Column | Type | Notes |
| --- | --- | --- |
| `created_at` | `timestamptz` | Row insert time; default `now()` |
| `updated_at` | `timestamptz` | Auto-set on `UPDATE` via `touch_updated_at()` trigger |
| `created_by` | `text` | Actor user id; default `'system'`; set from JWT via `internal/auditctx` |
| `updated_by` | `text` | Last mutator user id; default `'system'` |

**Audited tables today:** `calls`, `messages`, `call_sessions`, `call_turns`, `knowledge_documents`, `knowledge_chunks`, `tenants`, `users`, `user_roles`, `refresh_tokens`, `package_rule_schemas`, `packages`, `package_limits`, `tenant_entitlements`, `ai_avatars`, `ai_avatar_voices`, `tenant_avatar_assignments`, `tenant_registrations`, `brands`, `tenant_kyc_profiles`, `payment_gateway_configs`, `payment_callback_events`, `payment_orders`. Provider catalog tables (`embedding_models`, `voice_providers`) follow the same pattern when created.

ClickHouse analytics tables use `created_at`, `updated_at`, `created_by`, `updated_by` (`002_audit_columns_clickhouse` + `EnsureAuthEventsSchema`).

## Postgres (`callcenter`)

```mermaid
erDiagram
  tenants ||--o{ user_roles : scopes
  users ||--o{ user_roles : has
  users ||--o{ refresh_tokens : has
  tenants ||--o{ call_sessions : owns
  tenants ||--o{ knowledge_documents : owns
  tenants ||--o{ tenant_entitlements : entitled
  tenants ||--o{ tenant_avatar_assignments : assigns
  tenants ||--|| tenant_registrations : registered_via
  tenants ||--|| tenant_kyc_profiles : kyc_package
  tenants ||--o{ brands : owns
  ai_avatars ||--o{ tenant_avatar_assignments : enabled_for
  ai_avatars ||--o{ ai_avatar_voices : speaks_with
  voice_providers ||--o{ ai_avatar_voices : provides
  package_rule_schemas ||--o{ package_limits : shapes
  package_rule_schemas ||--o{ tenant_entitlements : snapshot_schema
  packages ||--o{ tenant_entitlements : grants
  packages ||--|| package_limits : defines
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
    note "pending_kyc active suspended"
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  tenant_registrations {
    text id PK
    text tenant_id FK UK
    text company_name
    text admin_email
    text status
    text rejection_reason
    timestamptz reviewed_at
    text reviewed_by
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  tenant_kyc_profiles {
    text tenant_id PK FK
    text contact_name
    text contact_phone
    text contact_address
    text photo_object_key
    jsonb business_doc_keys
    text status
    timestamptz submitted_at
    timestamptz reviewed_at
    text reviewed_by
    text rejection_reason
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  brands {
    text id PK
    text tenant_id FK UK
    text name
    text status
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  users {
    text id PK
    text email UK
    text password_hash
    text display_name
    text status
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  user_roles {
    text user_id FK
    text role
    text tenant_id FK
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  refresh_tokens {
    text id PK
    text user_id FK
    text token_hash UK
    timestamptz expires_at
    timestamptz revoked_at
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  calls {
    text id PK
    text agent_id
    text title
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  messages {
    bigserial id PK
    text call_id FK
    text role
    text content
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
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
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  call_turns {
    bigserial id PK
    text call_id FK
    text role
    text content
    jsonb source_chunk_ids
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  embedding_models {
    text id PK
    text provider
    text model_key
    int dimensions
    text status
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  voice_providers {
    text id PK
    text provider
    text model_key
    text modality
    text status
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  ai_employee_configs {
    text agent_id PK
    text tenant_id FK
    text voice_provider_id FK
    text text_provider_id FK
    text embedding_model_id FK
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
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
    text created_by
    text updated_by
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
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
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
    timestamptz updated_at
    text created_by
    text updated_by
  }

  package_rule_schemas {
    text id PK
    int version UK
    text name
    jsonb fields
    text status
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  packages {
    text id PK
    text slug UK
    text name
    text status
    int price_cents
    text currency
    text billing_period
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  package_limits {
    text package_id PK_FK
    text rules_schema_id FK
    jsonb rules
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  tenant_entitlements {
    text id PK
    text tenant_id FK
    text package_id FK
    text rules_schema_id FK
    jsonb rules_snapshot
    text status
    timestamptz valid_from
    timestamptz valid_until
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  ai_avatars {
    text id PK
    text slug UK
    text name
    text role
    text trait
    text color
    text image_url
    text greeting
    text status
    jsonb flags
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  ai_avatar_voices {
    text id PK
    text avatar_id FK
    text voice_provider_id FK
    text voice_id
    text voice
    int priority
    text status
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  tenant_avatar_assignments {
    text tenant_id FK
    text avatar_id FK
    text status
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  payment_gateway_configs {
    text id PK
    text provider
    text mode
    text status
    text merchant_code
    text api_key
    text md5_key
    text base_url
    int route_no
    text currency
    text callback_url
    text return_url
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  payment_callback_events {
    text id PK
    text provider
    text transaction_id UK
    text order_no
    text payment_status
    text amount
    text customer_id
    text payload_hash
    timestamptz received_at
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  payment_orders {
    text id PK
    text tenant_id FK
    text package_id FK
    text order_no UK
    int amount_cents
    text currency
    text status
    text provider
    text transaction_id
    text payment_url
    timestamptz paid_at
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  tenants ||--o{ payment_orders : places
  packages ||--o{ payment_orders : purchased_as
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
| `package_rule_schemas` | Versioned JSONB field catalog for package `rules` *(Sprint 4)* |
| `packages` | Commercial catalog — Starter/Pro/Enterprise *(Sprint 4)* |
| `package_limits` | `rules` jsonb per package; shape from `package_rule_schemas` *(Sprint 4)* |
| `tenant_entitlements` | Assignment + `rules_snapshot` at bind time *(Sprint 4)* |
| `ai_avatars` | Platform-managed avatar catalog *(Sprint 5)* |
| `ai_avatar_voices` | Ordered voice profiles per avatar (`voice_provider_id`, `voice_id`, `voice`, `priority`) *(Sprint 5)* |
| `tenant_avatar_assignments` | Which avatars each tenant may use *(Sprint 5)* |
| `payment_gateway_configs` | Singleton platform ChillPay/mock gateway settings *(Sprint 8)* |
| `payment_callback_events` | Idempotent ChillPay callback audit log *(Sprint 8)* |
| `payment_orders` | Tenant checkout orders; ChillPay `order_no` + fulfillment *(Sprint 9)* |

### Indexes

- `knowledge_documents (tenant_id, agent_id)`
- `knowledge_chunks (tenant_id, agent_id, embedding_model_id, index_version)`
- `knowledge_index_runs (document_id, embedding_model_id, index_version)`
- `call_sessions (tenant_id, voice_provider_id)` *(planned)*
- `tenant_avatar_assignments (tenant_id) WHERE status = 'active'` *(Sprint 5)*
- `ai_avatar_voices (avatar_id, priority) WHERE status = 'active'` *(Sprint 5)*
- `payment_orders (tenant_id, status)` *(Sprint 9)*
- `payment_orders (order_no)` unique *(Sprint 9)*

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
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  voice_providers {
    text id PK
    text provider
    text model_key
    text modality
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
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
    datetime created_at
    datetime updated_at
    string created_by
    string updated_by
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
    datetime updated_at
    string created_by
    string updated_by
  }

  auth_events {
    string event_id
    string event
    string tenant_id
    string user_id
    string email
    string role
    string ip
    string user_agent
    datetime created_at
    datetime updated_at
    string created_by
    string updated_by
  }

  call_provider_events {
    string event_id
    string call_id
    string tenant_id
    string voice_provider_id
    string event_type
    datetime created_at
    datetime updated_at
    string created_by
    string updated_by
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
| `created_at` / `updated_at` / `created_by` / `updated_by` | Audit fields on all analytics tables (shipped Sprint 3) |
| `auth_events` | NATS auth lifecycle mirror (`auth.user.logged_in`, `logged_out`, etc.) |

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
  kyc/{tenant_id}/          # Sprint 6 — photo + docs (see internal/store/tenantkyc_assets.go)
    photo.{ext}
    docs/{filename}
```

## Redis (ephemeral)

| Key pattern | TTL | Fields |
| --- | --- | --- |
| `monti_jarvis:call:{session_id}` | 24h | agent_id, updated_at (legacy chat) |
| `monti_jarvis:call:active:{id}` | 24h | tenant_id, room_name, status, started_at |
| `monti_jarvis:entitlement:{tenant_id}` | 15m (env) | package slug, status, effective limits JSON *(Sprint 4)* |
| `monti_jarvis:register:ip:{ip}` | 1h | registration attempt counter *(Sprint 6)* |

## Workforce (Sprint 5 — DB + fallback)

- **Primary:** `ai_avatars` + `tenant_avatar_assignments` → `GET /api/workforce` per tenant.
- **Fallback:** `internal/workforce/workforce.go` static catalog when tenant has zero active assignments.
- **Sprint 21:** `ai_employee_configs` provider bindings; may extend or alias `ai_avatars`.

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
| 4 ✅ v0.5.0 | `package_rule_schemas`, `packages`, `package_limits`, `tenant_entitlements` |
| 5 ✅ v0.6.0 | `ai_avatars`, `ai_avatar_voices`, `tenant_avatar_assignments` |
| 6 ✅ v0.7.0 | `tenant_registrations`, `brands`, `tenant_kyc_profiles`; `tenants.status` + `pending_kyc` |
| 7 ✅ v0.8.0 | KYC review columns on `tenant_kyc_profiles` + `tenant_registrations`; approve/reject lifecycle |
| 8 ✅ v0.9.0 | `payment_gateway_configs`, `payment_callback_events` |
| 9–12 ✅ v1.3.0 | `payment_orders`, `payment_documents`, `tenant_tax_profiles`, billing ops |
| 13 ✅ v1.4.0 | Redis quota/rate-limit keys (no required DDL) |
| 14 ✅ v1.5.0 | `tenant_embed_configs` |
| 15 ✅ v1.6.0 | Tenant KM store + `km_gaps`; no `km_scope_assignments` table required |
| 16 ✅ v1.7.0 | `tenant_settings`, `tenant_call_limits` |
| 17 ✅ v1.8.0 | `call_sessions.source`; preview uses existing call model |
| 18 ✅ v1.9.0 | `customer_tiers`, `customer_groups` |
| 19 🔄 v2.0.0 | `customers`, `customer_group_members`, `customer_import_jobs`, `customer_domain_rules` |
| 21 | Runtime voice failover + `ai_employee_configs` embedding bindings (voice stays on `ai_avatar_voices`) |
| 22 | `conversation_records` (ClickHouse denorm) |
| backlog | optional `quota_usage_events` audit (not in S13 commitment) |

## Sprint 13 — Quota (Redis + existing Postgres reads)

**Decision:** MVP ships **without** new Postgres tables. Limits come from existing `tenant_entitlements` / `package_limits`; KM and avatar usage from existing tables; concurrent/minutes/rate from Redis.

### Existing entities used for usage (no schema change)

| Entity | Usage dimension |
| --- | --- |
| `tenant_entitlements` + `package_limits.rules` | All limit ceilings |
| KM document store (existing ingest tables) | `max_km_documents` |
| `tenant_avatar_assignments` | `max_ai_employees` |

### Redis entities — Sprint 13

| Key pattern | Type | TTL | Purpose |
| --- | --- | --- | --- |
| `{prefix}quota:{tenant}:concurrent` | string/int | `QUOTA_CONCURRENT_TTL` | Active call slots |
| `{prefix}quota:{tenant}:minutes:{YYYYMM}` | string/int | none (month key) | Monthly voice minutes (UTC) |
| `{prefix}rl:{tenant}:chat:{YYYYMMDDHHMM}` | string/int | ~2m | Chat rate window |
| `{prefix}rl:{tenant}:km:{YYYYMMDDHHMM}` | string/int | ~2m | KM write rate window |
| `{prefix}rl:{tenant}:voice:{YYYYMMDDHHMM}` | string/int | ~2m | Voice open rate window |
| `{prefix}entitlement:{tenant}` | JSON | existing | Entitlement cache (S4) |

`prefix` = `REDIS_PREFIX` default `monti_jarvis:`.

### Explicit non-goal table (do not create in S13)

`quota_usage_events` deferred — would need full audit columns (`created_at`, `updated_at`, `created_by`, `updated_by`) if added later.

## Sprint 14 — tenant_embed_configs

```mermaid
erDiagram
  tenants ||--o| tenant_embed_configs : has
  tenant_embed_configs {
    text tenant_id PK
    text embed_key UK
    boolean enabled
    jsonb allowed_origins
    text default_agent_id
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }
```

## Sprint 15 — KM lifecycle + km_gaps

S15 productizes **tenant-admin** access to existing S2 KM entities and adds **`km_gaps`** for unanswered customer questions (FAQ backlog).

```mermaid
erDiagram
  tenants ||--o{ knowledge_documents : owns
  tenants ||--o{ km_gaps : records
  knowledge_documents ||--o{ knowledge_chunks : contains
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
    text created_by
    text updated_by
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
    timestamptz updated_at
    text created_by
    text updated_by
  }
  km_gaps {
    text id PK
    text tenant_id
    text agent_id
    text topic
    text question
    text question_hash
    text session_id
    text call_id
    text source
    text status
    int occurrence_count
    timestamptz last_seen_at
    timestamptz resolved_at
    text resolved_document_id
    text notes
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }
```

### `km_gaps` (new — S15)

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | |
| `tenant_id` | text NOT NULL | Isolation |
| `agent_id` | text | Workforce agent at ask time |
| `topic` | text | Caller topic tab |
| `question` | text | Customer question (original) |
| `question_hash` | text | sha256(lower(trim(question))) for dedupe |
| `session_id` / `call_id` | text | Optional correlation |
| `source` | text | `chat` \| `voice` \| `embed` |
| `status` | text | `open` \| `resolved` \| `dismissed` \| `converted` |
| `occurrence_count` | int | Bumped on repeat same hash |
| `last_seen_at` | timestamptz | Last missing_km hit |
| `resolved_at` | timestamptz null | When closed |
| `resolved_document_id` | text | Optional link after KM upload |
| `notes` | text | Tenant notes |
| audit | | `created_at`, `updated_at`, `created_by`, `updated_by` |

**Unique:** `(tenant_id, agent_id, question_hash)`  
**Index:** `(tenant_id, status, last_seen_at DESC)`

**Write path:** when chat RAG sets `missing_km=true` → `RecordKMGap` (Postgres) **and** existing ClickHouse `qa_events` (`event_type=missing_km`).

**Tenant APIs:** `GET /api/tenant/km/gaps`, `PATCH /api/tenant/km/gaps/{id}`

### ClickHouse (existing)

| Table | S15 delete filter |
| --- | --- |
| `monti_jarvis.km_embeddings` | `tenant_id` + `document_id` (single doc) or `tenant_id` + `agent_id` (reset) |

Columns (shipped S2): `tenant_id`, `agent_id`, `document_id`, `chunk_id`, `km_scope`, `km_version`, `content`, `embedding`.

### MinIO object key layout

```text
km/{tenant_id}/{agent_id}/{document_id}/original/{filename}
```

### Delete cascade (document-level) — implement in `internal/km`

1. Load doc by id; require `doc.tenant_id == jwt.tenant_id` else 404  
2. ClickHouse: `ALTER … DELETE WHERE tenant_id AND document_id` (existing client helper)  
3. MinIO: delete `object_key`  
4. Postgres: delete document (chunks via `ON DELETE CASCADE`)

### Document status (code constants)

| status | UI label | Meaning |
| --- | --- | --- |
| `uploaded` | Processing | Row created, not yet indexed |
| `indexing` | Processing | Chunking/embedding in progress |
| `indexed` | Ready | Available for RAG |
| `failed` | Failed | Ingest error |

**Scope values (`km_scope`):** `general` | `billing` | `technical`

### Explicit non-goals (do not create in S15)

| Entity | Why deferred |
| --- | --- |
| `km_versions` / publish logs | Approval workflow later |
| `tenant_agent_scope_overrides` | Use hard-map + document scope |
| `knowledge_folders` | Flat list per agent is enough |

## Sprint 16 — tenant_settings + tenant_call_limits

```mermaid
erDiagram
  tenants ||--o| tenant_settings : has
  tenants ||--o| tenant_call_limits : has
  tenant_settings {
    text tenant_id PK
    text locale
    text timezone
    text display_name
    text ai_reply_locale
    text user_tier_label
    text user_group_label
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }
  tenant_call_limits {
    text tenant_id PK
    int max_minutes_per_call
    int max_call_minutes_per_day
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }
```

**Redis:** `{prefix}call_daily:{tenant}:{YYYYMMDD}` minutes used (tenant timezone day).

## Sprint 17 — preview sessions

```mermaid
erDiagram
  tenants ||--o{ call_sessions : owns
  call_sessions {
    text id PK
    text tenant_id
    text source "production|preview"
    text status
    timestamptz started_at
    timestamptz ended_at
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }
```

| Change | Notes |
| --- | --- |
| `call_sessions.source` | Default `production`; preview APIs set `preview` |
| Redis `preview:concurrent:{tenant}` | Soft concurrent voice preview slots |

**No new tables required** if column add on existing `call_sessions` is preferred over separate `preview_sessions`.

## Sprint 18 — customer_tiers + customer_groups

```mermaid
erDiagram
  tenants ||--o{ customer_tiers : defines
  tenants ||--o{ customer_groups : defines
  customer_tiers {
    text id PK
    text tenant_id
    text name
    text slug
    int priority
    text description
    text default_agent_id
    text ai_reply_locale
    int max_minutes_per_call
    int max_call_minutes_per_day
    boolean active
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }
  customer_groups {
    text id PK
    text tenant_id
    text name
    text slug
    text description
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }
```

**Future (S19+):** `customers.tier_id` → `customer_tiers`, optional group membership.

## Sprint 19 — customer accounts, imports, and domain rules

```mermaid
erDiagram
  tenants ||--o{ customers : owns
  customer_tiers ||--o{ customers : classifies
  tenants ||--o{ customer_import_jobs : runs
  tenants ||--o{ customer_domain_rules : defines
  customer_tiers ||--o{ customer_domain_rules : defaults
  customer_groups ||--o{ customer_domain_rules : defaults
  customers ||--o{ customer_group_members : joins
  customer_groups ||--o{ customer_group_members : joins

  customers {
    text id PK
    text tenant_id FK
    text email
    text email_normalized
    text phone
    text display_name
    text locale
    text tier_id FK
    text source
    text external_id
    text status
    jsonb metadata
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  customer_group_members {
    text customer_id PK_FK
    text group_id PK_FK
    text tenant_id FK
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  customer_import_jobs {
    text id PK
    text tenant_id FK
    text filename
    text mode
    text status
    int total_rows
    int created_rows
    int updated_rows
    int rejected_rows
    jsonb errors
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  customer_domain_rules {
    text id PK
    text tenant_id FK
    text domain
    text policy
    text default_tier_id FK
    text default_group_id FK
    boolean active
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }
```

### Constraints and indexes

| Entity | Rule |
| --- | --- |
| `customers` | unique `(tenant_id, email_normalized)` when email exists |
| `customers` | unique `(tenant_id, source, external_id)` when external id exists |
| `customers` | `status IN ('active','inactive')`; delete API sets inactive |
| `customer_group_members` | PK `(customer_id, group_id)`; service verifies same tenant |
| `customer_import_jobs` | `mode IN ('dry_run','commit')`; state follows §56 workflow |
| `customer_domain_rules` | unique `(tenant_id, domain)`; policy allow/deny |

TASK-0087 uses the idempotent `internal/store` ensure-schema mechanism. No Redis, ClickHouse, or MinIO entity is added. CSV bytes are not retained.

**Future SPRINT-020:** customer credential/user binding and domain-rule enforcement. No password, OAuth, or session fields belong in the SPRINT-019 tables.

See [01-architecture.md](01-architecture.md) · [08-packages-spec.md](08-packages-spec.md) · [10-avatars-spec.md](10-avatars-spec.md) · [11-tenant-register-spec.md](11-tenant-register-spec.md) · [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md) · [13-payment-gateway-spec.md](13-payment-gateway-spec.md) · [14-buy-package-spec.md](14-buy-package-spec.md) · [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md) · [17-embed-to-web-spec.md](17-embed-to-web-spec.md) · [18-tenant-scope-km-spec.md](18-tenant-scope-km-spec.md) · [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md) · [20-tenant-test-preview-spec.md](20-tenant-test-preview-spec.md) · [21-customer-tier-spec.md](21-customer-tier-spec.md) · [22-customer-account-import-spec.md](22-customer-account-import-spec.md) · [02-workflow.md](02-workflow.md) · [04-api-spec.md](04-api-spec.md) · [05-ux-ui.md](05-ux-ui.md).
