---
id: DES-0004
title: API Specification
status: approved
updated: 2026-07-08
sprint: SPRINT-006
---

# API Specification — Monti Jarvis

**Base URL:** `http://localhost:8091`  
**Auth:** `AUTH_DISABLED=true` (default) — same as v0.3.0 for customer paths. When `AUTH_DISABLED=false`, use `Authorization: Bearer <access_token>` on protected routes. See [06-auth-spec.md](06-auth-spec.md).  
**Packages (Sprint 4):** Platform catalog + entitlements require auth on — see [08-packages-spec.md](08-packages-spec.md).  
**Avatars (Sprint 5):** Platform avatar catalog + tenant assignment — see [10-avatars-spec.md](10-avatars-spec.md).  
**Tenant register (Sprint 6):** Public signup + platform tenant list — see [11-tenant-register-spec.md](11-tenant-register-spec.md).
**CORS:** `*` — methods `GET, POST, PUT, DELETE, OPTIONS`; headers `Content-Type`, `Authorization`

## Health & infra

### `GET /healthz`

Liveness + feature flags.

```json
{
  "ok": true,
  "gemini": true,
  "voice": true,
  "livekit": true,
  "nats": true,
  "rag": true,
  "sprint": "SPRINT-006",
  "auth_disabled": true,
  "customer_web": "apps/customer-web/build",
  "platform_admin_web": "apps/platform-admin-web/build",
  "tenant_web": "apps/tenant-web/build"
}
```

### `GET /api/infra`

Dependency health.

```json
{
  "postgres": "ok",
  "redis": "ok",
  "minio": "ok",
  "clickhouse": "ok",
  "nats": "ok",
  "livekit": "configured",
  "entitlement_cache": "ok"
}
```

`entitlement_cache`: `ok` | `disabled` | `unavailable` *(Sprint 4, when `ENTITLEMENT_CACHE_ENABLED`)*

## Workforce

### `GET /api/workforce`

**Auth:** Public (no Bearer required). **Tenant resolution (Sprint 5):** `X-Tenant-Id` header, JWT `tenant_id` when Bearer present, else `DEMO_TENANT_ID` (`demo`).

Returns active avatars assigned to the resolved tenant from `ai_avatars` + `tenant_avatar_assignments`. If the tenant has **zero** active assignments, falls back to static `internal/workforce` catalog (backward compatible).

| Header | Type | Description |
| --- | --- | --- |
| `X-Tenant-Id` | string | Optional; used when `AUTH_DISABLED=true` or `platform_admin` without tenant in token |

```json
{
  "agents": [
    {
      "id": "ava",
      "name": "Ava",
      "role": "General Support",
      "trait": "Warm & Patient",
      "color": "#008cff",
      "image": "/images/ava.jpg",
      "voice": "Aoede",
      "voice_provider_id": "voice-gemini-live",
      "voice_id": "gemini-2.5-flash-native-audio-latest",
      "popular": true,
      "greeting": "Thank you for calling..."
    }
  ]
}
```

DB field `image_url` is exposed as `image`. `voice`, `voice_provider_id`, `voice_id` come from the **primary** `ai_avatar_voices` row (lowest `priority` among `active`). Optional `flags` keys (`robot`, `skin`, `hair`) merge into agent objects when set.

## Chat (text + RAG)

### `POST /api/chat`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `session_id` | string | no | Reuse id; generated if empty |
| `agent_id` | string | yes | `ava` \| `max` \| `luna` \| `neo` |
| `topic` | string | no | `general` \| `billing` \| `technical` |
| `message` | string | yes | User question |
| `history` | array | no | `{role, content}` prior turns |

**Response 200:**

```json
{
  "session_id": "abc123",
  "agent_id": "max",
  "reply": "Invoices are due within 15 days.",
  "sources": [
    {
      "chunk_id": "...",
      "document_id": "...",
      "scope": "billing",
      "excerpt": "Invoices are issued...",
      "score": 0.82
    }
  ],
  "missing_km": false
}
```

**Errors:** `400` validation · `502` Gemini failure

## Voice

### `GET /ws/voice`

WebSocket upgrade.

| Query | Description |
| --- | --- |
| `agent` | Agent id |
| `topic` | Topic tab for RAG preload |

**Client → server messages:**

```json
{"type": "audio", "data": "<base64 PCM16>"}
{"type": "text", "text": "..."}
{"type": "end"}
```

**Server → client messages:**

```json
{"type": "ready", "agent_id": "ava", "voice": "Aoede"}
{"type": "audio", "data": "<base64>"}
{"type": "transcript", "role": "user|assistant", "text": "..."}
{"type": "text", "text": "..."}
{"type": "interrupted"}
{"type": "turn_complete"}
{"type": "error", "message": "..."}
```

Static assets: `/recorder.js`, `/player.js` (AudioWorklet).

## Call sessions

### `POST /api/calls`

Create session. **Response:** `CallSession` `{id, tenant_id, room_name, status, started_at}`

### `GET /api/calls/{id}`

Get session.

### `POST /api/calls/{id}/token`

LiveKit join token. Body: `{identity?}`. **Response:** `{token, url, identity, room_name}`

### `POST /api/calls/{id}/end`

End session.

### `GET /api/calls/{id}/turns`

List turns. **Response:** `{turns: [{id, role, content, created_at}]}`

### `POST /api/calls/{id}/turns`

Add turn. Body: `{role, content}`

### `GET /api/calls/{id}/events`

SSE stream. Event `turn` with turn JSON payload.

## Auth (Sprint 3 — draft)

### `POST /api/auth/login`

| Field | Type | Required |
| --- | --- | --- |
| `email` | string | yes |
| `password` | string | yes |

**Response 200:** `{access_token, refresh_token, expires_in, token_type, user}`

**Errors:** `401` invalid credentials · `503` auth not configured

### `POST /api/auth/refresh`

Body: `{ "refresh_token": "..." }` → new token pair.

### `POST /api/auth/logout`

Bearer access token or body `{refresh_token}` → revokes refresh.

### `GET /api/auth/me`

Bearer required → `{id, email, display_name, role, tenant_id}`.

### Protected routes (when `AUTH_DISABLED=false`)

| Route | Roles |
| --- | --- |
| `POST /api/km/agents/{id}/documents` | `tenant_admin`, `platform_admin` |
| `POST /api/km/agents/{id}/reset` | `tenant_admin`, `platform_admin` |
| `POST /api/km/seed` | `platform_admin` |
| `/api/platform/rule-schemas` | `platform_admin` |
| `/api/platform/packages*` | `platform_admin` |
| `/api/platform/tenants/{id}/entitlement` | `platform_admin` |
| `/api/platform/avatars*` | `platform_admin` |
| `/api/platform/tenants/{id}/avatars*` | `platform_admin` |
| `/api/platform/tenants/{id}/kyc*` | `platform_admin` |
| `/api/tenant/kyc*` | `tenant_admin` |
| `GET /api/entitlements/me` | `tenant_admin`, `platform_admin` |

Public unchanged: `/api/chat`, `/ws/voice`, `GET /api/km/*`, `GET /api/workforce`.

## Packages & entitlements (Sprint 4)

Requires `AUTH_DISABLED=false`. All routes: `Authorization: Bearer <access_token>`.

Package limits are **`rules` jsonb** validated against **`package_rule_schemas`** (see [08-packages-spec.md](08-packages-spec.md) §2).

### `GET /api/platform/rule-schemas`

**Role:** `platform_admin` · Lists versioned JSONB field catalogs.

**Response 200:**

```json
{
  "schemas": [
    {
      "id": "rules-v1",
      "version": 1,
      "name": "Sprint 4 base limits",
      "status": "active",
      "fields": { "max_ai_employees": { "type": "int", "min": 0, "required": true } }
    }
  ]
}
```

### `GET /api/platform/packages`

**Role:** `platform_admin`

| Query | Type | Description |
| --- | --- | --- |
| `status` | string | Optional filter: `draft`, `active`, `archived` |

**Response 200:**

```json
{
  "packages": [
    {
      "id": "pkg-starter",
      "slug": "starter",
      "name": "Starter",
      "status": "active",
      "price_cents": 0,
      "currency": "USD",
      "billing_period": "monthly",
      "rules_schema_id": "rules-v1",
      "rules": {
        "max_ai_employees": 2,
        "max_monthly_call_minutes": 500,
        "max_km_documents": 50,
        "max_concurrent_calls": 2,
        "voice_enabled": true,
        "rag_enabled": true
      }
    }
  ]
}
```

### `POST /api/platform/packages`

**Role:** `platform_admin`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `slug` | string | yes | Unique catalog slug |
| `name` | string | yes | Display name |
| `description` | string | no | Marketing copy |
| `status` | string | no | Default `draft` |
| `price_cents` | int | no | List price |
| `currency` | string | no | Default `USD` |
| `billing_period` | string | no | `monthly`, `annual`, `one_time` |
| `rules_schema_id` | string | yes | e.g. `rules-v1` — must reference `active` schema |
| `rules` | object | yes | JSONB values validated against schema `fields` |

**Response 201:** package object (same shape as list item).

**Errors:** `400` validation (unknown key, type mismatch, missing required field) · `409` slug exists · `403` wrong role

### `GET /api/platform/packages/{id}`

**Role:** `platform_admin` · **Response 200:** package object · **404** if missing

### `PUT /api/platform/packages/{id}`

**Role:** `platform_admin` · Body: partial metadata + optional `rules_schema_id` + `rules` · **Response 200:** updated package · **409** if archived with active entitlements and destructive rule change

### `DELETE /api/platform/packages/{id}`

**Role:** `platform_admin` · Soft-archive (`status=archived`) · **409** if active tenant entitlements reference package

### `GET /api/platform/tenants/{tenant_id}/entitlement`

**Role:** `platform_admin`

**Response 200:** effective entitlement (see below) · **404** tenant or no entitlement

### `POST /api/platform/tenants/{tenant_id}/entitlement`

**Role:** `platform_admin`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `package_id` | string | yes | Target package |

Revokes any prior `active` row, inserts new `active` entitlement, invalidates Redis cache.

**Response 200:** effective entitlement · **404** unknown tenant/package · **409** duplicate active

### `DELETE /api/platform/tenants/{tenant_id}/entitlement`

**Role:** `platform_admin` · Revokes active entitlement (`status=revoked`) · **404** no active entitlement

### `GET /api/entitlements/me`

**Role:** `tenant_admin`, `platform_admin` (token `tenant_id`)

**Response 200 — effective entitlement:**

```json
{
  "tenant_id": "demo",
  "package": {
    "id": "pkg-starter",
    "slug": "starter",
    "name": "Starter"
  },
  "status": "active",
  "rules_schema_id": "rules-v1",
  "rules": {
    "max_ai_employees": 2,
    "max_monthly_call_minutes": 500,
    "max_km_documents": 50,
    "max_concurrent_calls": 2,
    "voice_enabled": true,
    "rag_enabled": true
  },
  "valid_from": "2026-07-07T00:00:00Z",
  "valid_until": null
}
```

**Errors:** `401` · `403` customer role · **404** no entitlement (or documented permissive default per resolver — see FEAT-0004 AC5)

### Packages error codes

| Code | When |
| --- | --- |
| `401` | Missing/invalid Bearer |
| `403` | Role not allowed |
| `404` | Package, tenant, or entitlement not found |
| `409` | Slug conflict, archive blocked, duplicate active entitlement |
| `503` | Auth not configured (`AUTH_DISABLED` misconfiguration for these routes) |

## Avatars (Sprint 5)

Requires `AUTH_DISABLED=false`. Platform routes: `Authorization: Bearer <access_token>`. See [10-avatars-spec.md](10-avatars-spec.md).

### `GET /api/platform/avatars`

**Role:** `platform_admin`

| Query | Type | Description |
| --- | --- | --- |
| `status` | string | Optional: `draft`, `active`, `archived` |

**Response 200:** each avatar includes `voices[]` (ordered by `priority`).

```json
{
  "avatars": [
    {
      "id": "ava",
      "slug": "ava",
      "name": "Ava",
      "role": "General Support",
      "trait": "Warm & Patient",
      "color": "#008cff",
      "image_url": "/images/ava.jpg",
      "greeting": "Thank you for calling...",
      "status": "active",
      "flags": { "popular": true, "skin": "#f0bd9b", "hair": "#5a3428" },
      "voices": [
        {
          "id": "avvoice_ava_gemini",
          "voice_provider_id": "voice-gemini-live",
          "voice_id": "gemini-2.5-flash-native-audio-latest",
          "voice": "Aoede",
          "priority": 1,
          "status": "active"
        }
      ]
    }
  ]
}
```

### `POST /api/platform/avatars`

**Role:** `platform_admin`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `slug` | string | yes | Unique lowercase id |
| `name` | string | yes | Display name |
| `role` | string | yes | Role label |
| `trait` | string | no | Personality |
| `color` | string | no | Hex accent |
| `image_url` | string | no | Default `/images/{slug}.jpg` |
| `greeting` | string | yes | Opening line |
| `status` | string | no | Default `draft` |
| `flags` | object | no | `popular`, `robot`, `skin`, `hair` |
| `voices` | array | yes | ≥1 voice profile (see below) |

**Voice profile object (`voices[]`):**

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `voice_provider_id` | string | yes | FK → `voice_providers.id` |
| `voice_id` | string | yes | Provider model/voice key |
| `voice` | string | yes | Persona name (`Aoede`, `Charon`, …) |
| `priority` | int | no | Default next integer; **lower = preferred** |
| `status` | string | no | Default `active` |

**Response 201:** avatar object with `voices[]` · **400** no voices / invalid provider · **409** slug exists

### `GET /api/platform/avatars/{id}`

**Role:** `platform_admin` · **200** avatar + `voices[]` · **404** missing

### `PUT /api/platform/avatars/{id}`

**Role:** `platform_admin` · Partial metadata + optional full `voices[]` replace · **200** updated · **409** if archived with active tenant assignments · **400** if `voices` empty when provided

### `DELETE /api/platform/avatars/{id}`

**Role:** `platform_admin` · Soft-archive · **409** if active tenant assignments exist

### `GET /api/platform/tenants/{tenant_id}/avatars`

**Role:** `platform_admin`

**Response 200:**

```json
{
  "tenant_id": "demo",
  "assignments": [
    {
      "avatar_id": "ava",
      "status": "active",
      "avatar": {
        "id": "ava",
        "name": "Ava",
        "role": "General Support",
        "status": "active"
      }
    }
  ],
  "cap": {
    "max_ai_employees": 2,
    "active_count": 4
  }
}
```

`cap` is informational for UI; assign still returns `409` when over limit.

### `POST /api/platform/tenants/{tenant_id}/avatars`

**Role:** `platform_admin`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `avatar_id` | string | yes | Target catalog avatar |

**Response 200:** assignment + avatar metadata · **404** unknown tenant/avatar · **409** at `max_ai_employees` cap

### `DELETE /api/platform/tenants/{tenant_id}/avatars/{avatar_id}`

**Role:** `platform_admin` · Sets assignment `disabled` · **404** no assignment

### Avatars error codes

| Code | When |
| --- | --- |
| `401` | Missing/invalid Bearer |
| `403` | Role not allowed |
| `404` | Avatar, tenant, or assignment not found |
| `409` | Slug conflict, archive blocked, `max_ai_employees` exceeded |
| `503` | Auth not configured |

## Tenant registration (Sprint 6)

Public onboarding. Works when `TENANT_REGISTER_ENABLED=true` (default). Independent of `AUTH_DISABLED` on customer paths.

### `POST /api/public/tenant/register`

**Auth:** None (public). **Rate limit:** Redis `monti_jarvis:register:ip:{ip}` → `429` when exceeded.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `company_name` | string | yes | Legal/display company name (2–120 chars) |
| `slug` | string | yes | Tenant id + URL slug; lowercase `a-z0-9-`, 2–32 chars |
| `admin_email` | string | yes | First tenant_admin email (unique) |
| `admin_password` | string | yes | min 8 characters |
| `admin_display_name` | string | yes | Shown in profile (1–80 chars) |

**Response 201:**

```json
{
  "tenant_id": "acme",
  "slug": "acme",
  "registration_id": "reg_a1b2c3d4e5f67890",
  "access_token": "eyJ...",
  "refresh_token": "rt_...",
  "expires_in": 900,
  "token_type": "Bearer",
  "user": {
    "id": "usr_acme_admin",
    "email": "admin@acme.test",
    "display_name": "Acme Admin",
    "role": "tenant_admin",
    "tenant_id": "acme"
  }
}
```

**Errors:** `400` validation · `409` slug/email conflict · `429` rate limit · `503` registration disabled or Postgres down

### `GET /api/platform/tenants`

**Role:** `platform_admin`

| Query | Type | Description |
| --- | --- | --- |
| `status` | string | Optional filter: `pending_kyc`, `active`, `suspended` |
| `kyc_status` | string | Optional filter: `draft`, `submitted`, `approved`, `rejected` (joins `tenant_kyc_profiles`) *(Sprint 7)* |
| `limit` | int | Default `50`, max `100` |
| `offset` | int | Pagination offset |

**Response 200:**

```json
{
  "tenants": [
    {
      "id": "acme",
      "slug": "acme",
      "name": "Acme Corp",
      "status": "pending_kyc",
      "registration_id": "reg_a1b2c3d4e5f67890",
      "admin_email": "admin@acme.test",
      "kyc_status": "submitted",
      "created_at": "2026-07-08T01:00:00Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

## Tenant KYC — tenant portal (Sprint 6, shipped v0.7.0)

**Role:** `tenant_admin` on own tenant. Routes under `guard.RequireBearer` + tenant scope.

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/tenant/kyc` | Current KYC profile + asset URLs |
| `PUT` | `/api/tenant/kyc` | Update contact fields |
| `POST` | `/api/tenant/kyc/photo` | Upload portrait (multipart `file`) |
| `POST` | `/api/tenant/kyc/documents` | Upload business document (multipart `file`) |
| `POST` | `/api/tenant/kyc/submit` | Set status `submitted` |

**Assets:** `GET /api/assets/kyc/{tenant_id}/{kind}/{file}` — `kind` = `photo` | `docs`.

## Platform KYC review (Sprint 7)

**Role:** `platform_admin` on all routes below.

### `GET /api/platform/tenants/{tenant_id}/kyc`

Returns tenant metadata, registration row, and KYC package for review.

**Response 200:**

```json
{
  "tenant": {
    "id": "acme",
    "slug": "acme",
    "name": "Acme Corp",
    "status": "pending_kyc",
    "created_at": "2026-07-08T01:00:00Z"
  },
  "registration": {
    "id": "reg_a1b2c3d4e5f67890",
    "company_name": "Acme Corp",
    "admin_email": "admin@acme.test",
    "status": "submitted"
  },
  "kyc": {
    "status": "submitted",
    "contact_name": "Jane Doe",
    "contact_phone": "+66 2 000 0000",
    "contact_address": "Bangkok, TH",
    "photo_url": "/api/assets/kyc/acme/photo/photo.jpg",
    "documents": [
      { "object_key": "kyc/acme/docs/license.pdf", "url": "/api/assets/kyc/acme/docs/license.pdf" }
    ],
    "submitted_at": "2026-07-09T08:00:00Z",
    "reviewed_at": null,
    "reviewed_by": "",
    "rejection_reason": ""
  }
}
```

**Errors:** `401` · `403` · `404` tenant not found

### `POST /api/platform/tenants/{tenant_id}/kyc/approve`

No body. Atomically activates tenant.

**Response 200:**

```json
{
  "tenant_id": "acme",
  "tenant_status": "active",
  "registration_status": "approved",
  "kyc_status": "approved",
  "reviewed_at": "2026-07-09T09:00:00Z",
  "reviewed_by": "usr_platform_admin"
}
```

**Errors:** `401` · `403` · `404` · `409` tenant not `pending_kyc` or KYC not `submitted`

### `POST /api/platform/tenants/{tenant_id}/kyc/reject`

**Body:**

```json
{ "reason": "Business license document is illegible. Please re-upload." }
```

**Response 200:**

```json
{
  "tenant_id": "acme",
  "tenant_status": "pending_kyc",
  "registration_status": "rejected",
  "kyc_status": "rejected",
  "rejection_reason": "Business license document is illegible. Please re-upload.",
  "reviewed_at": "2026-07-09T09:05:00Z",
  "reviewed_by": "usr_platform_admin"
}
```

**Errors:** `400` missing/empty `reason` · `401` · `403` · `404` · `409` KYC not `submitted`

### Platform KYC error codes

| Code | When |
| --- | --- |
| `400` | Reject without `reason` |
| `409` | Approve/reject when prerequisites not met (wrong tenant/KYC status) |

### Post-approve route policy

`tenant_admin` on `active` tenant: `POST /api/km/*` writes allowed (`RequireKMWrite` passes `IsTenantActive`). See [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md) §6.

### Tenant registration error codes

| Code | When |
| --- | --- |
| `400` | Invalid slug, password, reserved slug, missing field |
| `409` | Slug or email already exists |
| `429` | IP rate limit exceeded |
| `503` | `TENANT_REGISTER_ENABLED=false` or store unavailable |

### Pending tenant route policy

`tenant_admin` on `pending_kyc` tenant: login + `/api/auth/me` OK; `POST /api/km/*` writes → `403 tenant not active`. After Sprint 7 approve → `active` → KM writes OK. See [11-tenant-register-spec.md](11-tenant-register-spec.md) §7 · [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md).

## Knowledge base (per avatar)

### `GET /api/km/agents/{agent_id}`

KB summary: `{agent_id, tenant_id, scope, document_count, chunk_count, documents[]}`

### `GET /api/km/agents/{agent_id}/documents`

List documents for agent.

### `POST /api/km/agents/{agent_id}/documents`

Multipart: `file` (required), `scope` (optional, defaults per agent).

**Response 201:** document record `{id, status, chunk_count, ...}`

### `POST /api/km/agents/{agent_id}/reset`

Clear all KB for agent (Postgres + MinIO + ClickHouse).

**Response:** `{agent_id, status: "reset"}`

### `POST /api/km/seed`

Ingest sample files from `docs/samples/km/{agent}.md` for all four agents.

## Static routes

| Route | Handler |
| --- | --- |
| `/` | Svelte customer portal (`apps/customer-web/build`) |
| `/admin/` | Svelte platform admin portal (`apps/platform-admin-web/build`) *(Sprint 4)* |
| `/admin/*` | SPA fallback → `index.html` |
| `/tenant/` | Svelte tenant portal (`apps/tenant-web/build`) *(Sprint 6)* |
| `/tenant/*` | SPA fallback → `index.html` |
| `/legacy/` | Legacy HTML UI |
| `/images/*` | Static assets |

Platform admin UI calls JSON APIs on same origin (`8091`); tokens in `sessionStorage` (see [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md)).

## Error envelope

```json
{"error": "human-readable message"}
```

See [06-auth-spec.md](06-auth-spec.md), [08-packages-spec.md](08-packages-spec.md), [10-avatars-spec.md](10-avatars-spec.md), [11-tenant-register-spec.md](11-tenant-register-spec.md), [02-workflow.md](02-workflow.md), [05-ux-ui.md](05-ux-ui.md), and [docs/KM_SETUP.md](../../KM_SETUP.md).