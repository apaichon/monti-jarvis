---
id: DES-0004
title: API Specification
status: approved
updated: 2026-07-15
sprint: SPRINT-027
---

# API Specification — Monti Jarvis

**Base URL:** `http://localhost:8091`  
**Auth:** `AUTH_DISABLED=true` (default) — same as v0.3.0 for customer paths. When `AUTH_DISABLED=false`, use `Authorization: Bearer <access_token>` on protected routes. See [06-auth-spec.md](06-auth-spec.md).  
**Packages (Sprint 4):** Platform catalog + entitlements require auth on — see [08-packages-spec.md](08-packages-spec.md).  
**Avatars (Sprint 5):** Platform avatar catalog + tenant assignment — see [10-avatars-spec.md](10-avatars-spec.md).  
**Tenant register (Sprint 6):** Public signup + platform tenant list — see [11-tenant-register-spec.md](11-tenant-register-spec.md).  
**Quota (Sprint 13):** Redis quotas + rate limits on hot paths; platform usage read API — see [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md).  
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
  "sprint": "SPRINT-014",
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
  "entitlement_cache": "ok",
  "payment_gateway": {
    "configured": true,
    "provider": "chillpay",
    "mode": "test",
    "status": "active"
  }
}
```

`entitlement_cache`: `ok` | `disabled` | `unavailable` *(Sprint 4, when `ENTITLEMENT_CACHE_ENABLED`)*  
`payment_gateway`: *(Sprint 8)* `configured` bool · `provider` `chillpay`|`mock` · `mode` `test`|`live` · `status` `inactive`|`active`

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
    "active_count": 4,
    "override_allowed": true
  }
}
```

`cap` is informational for UI. `override_allowed` is true only for the configured
demo tenant, whose platform-admin assignment management may exceed the commercial
package cap so admins can promote/demote demo avatars. Other tenants return `409`
when an assignment would exceed the limit.

### `POST /api/platform/tenants/{tenant_id}/avatars`

**Role:** `platform_admin`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `avatar_id` | string | yes | Target catalog avatar |

**Response 200:** assignment + avatar metadata (also reactivates a disabled assignment) · **404** unknown tenant/avatar · **409** at `max_ai_employees` cap except for configured demo-tenant platform administration

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

## Payment Gateway (Sprint 8)

**Role:** `platform_admin` on config routes; callback is **public** (ChillPay server POST).

See [13-payment-gateway-spec.md](13-payment-gateway-spec.md) for ChillPay checksum rules.

### `GET /api/platform/payment-gateway`

Returns singleton platform gateway config. Secrets masked.

**Response 200:**

```json
{
  "provider": "chillpay",
  "mode": "test",
  "status": "active",
  "merchant_code": "M123456",
  "api_key_masked": "****abcd",
  "md5_key_set": true,
  "base_url": "https://sandbox-api.chillpay.co/api/v2/Payment/",
  "route_no": 1,
  "currency": "THB",
  "callback_url": "http://localhost:8091/api/callbacks/chillpay",
  "return_url": "http://localhost:8091/tenant/billing/return",
  "connection_status": "unknown",
  "last_callback_at": null
}
```

When no row exists: `200` with `provider: ""`, `status: "inactive"`, `configured: false`.

**Errors:** `401` · `403`

### `PUT /api/platform/payment-gateway`

**Body:**

```json
{
  "provider": "chillpay",
  "mode": "test",
  "merchant_code": "M123456",
  "api_key": "cp_live_xxx",
  "md5_key": "secret-md5-key",
  "base_url": "https://sandbox-api.chillpay.co/api/v2/Payment/",
  "route_no": 1,
  "currency": "THB",
  "return_url": "http://localhost:8091/tenant/billing/return"
}
```

Omit `api_key` / `md5_key` to keep existing values. Env `CHILLPAY_API_KEY` / `CHILLPAY_MD5_KEY` override at runtime when set.

`callback_url` is server-derived from `APP_PUBLIC_URL`; not writable by client.

**Response 200:** same shape as GET (masked).

**Errors:** `400` invalid provider/mode · `401` · `403`

### `POST /api/platform/payment-gateway/test`

No body. Runs provider `Ping`.

**Response 200:**

```json
{ "ok": true, "provider": "chillpay", "message": "credentials valid" }
```

**Response 502:**

```json
{ "ok": false, "provider": "chillpay", "message": "chillpay status HTTP 401: ..." }
```

**Errors:** `401` · `403` · `503` gateway not configured

### `POST /api/callbacks/chillpay`

**Auth:** None. ChillPay posts `application/x-www-form-urlencoded`.

**Form fields:** `OrderNo`, `Amount`, `TransactionId`, `CustomerId`, `CustomerName`, `BankCode`, `PaymentDate`, `PaymentStatus`, `PaymentDescription`, `BankRefCode`, `Currency`, `CreditCardToken`, `CurrentDate`, `CurrentTime`, `CheckSum`

**PaymentStatus:** `0` success · `1` pending · `2` failed

**Response 200:** empty body (ack)

**Errors:** `400` invalid/missing `CheckSum` · `503` gateway not configured

**Dev:** `PAYMENT_CALLBACK_DEV_BYPASS=true` skips checksum (local only).

### Payment Gateway error codes

| Code | When |
| --- | --- |
| `400` | Callback checksum invalid |
| `502` | Test connection — ChillPay rejected credentials |
| `503` | Gateway not configured / inactive |

**Sprint 9 fulfillment:** On `PaymentStatus=0`, lookup `payment_orders` by `OrderNo`, mark `paid`, assign entitlement. See § Tenant Checkout.

## Tenant Checkout (Sprint 9)

**Role:** `tenant_admin` on **`active`** tenant only. Depends on SPRINT-008 gateway configured.

See [14-buy-package-spec.md](14-buy-package-spec.md).

### `GET /api/tenant/packages`

Returns active commercial packages available for purchase.

**Response 200:**

```json
{
  "packages": [
    {
      "id": "pkg-pro",
      "slug": "pro",
      "name": "Pro",
      "description": "Growing teams",
      "price_cents": 299000,
      "currency": "764",
      "billing_period": "monthly",
      "rules_summary": {
        "max_ai_employees": 5,
        "max_monthly_call_minutes": 2000
      }
    }
  ],
  "current_entitlement": {
    "package_id": "pkg-starter",
    "package_name": "Starter",
    "status": "active"
  }
}
```

`current_entitlement` is `null` when tenant has no active entitlement.

**Errors:** `401` · `403` (not `tenant_admin` or tenant not `active`)

### `POST /api/tenant/checkout`

**Body:**

```json
{ "package_id": "pkg-pro" }
```

Creates `pending` order and starts payment.

**Response 200:**

```json
{
  "order_id": "ord_a1b2c3d4",
  "order_no": "mj_demo_x7k2m9",
  "package_id": "pkg-pro",
  "amount_cents": 299000,
  "currency": "764",
  "status": "pending",
  "payment_url": "https://sandbox-appsrv2.chillpay.co/pay/...",
  "provider": "chillpay"
}
```

Mock provider `payment_url` example: `http://localhost:8091/tenant/billing/mock-pay?order_id=ord_a1b2c3d4`

**Errors:** `400` missing `package_id` · `401` · `403` · `404` package not found · `409` package not active / tenant not active · `503` gateway not configured

### `GET /api/tenant/orders/{id}`

Poll checkout status (own tenant only).

**Response 200:**

```json
{
  "id": "ord_a1b2c3d4",
  "order_no": "mj_demo_x7k2m9",
  "package_id": "pkg-pro",
  "status": "paid",
  "amount_cents": 299000,
  "transaction_id": "123456789",
  "paid_at": "2026-07-09T10:00:00Z",
  "created_at": "2026-07-09T09:55:00Z"
}
```

**Errors:** `401` · `403` · `404`

### `POST /api/callbacks/chillpay` *(Sprint 9 upgrade)*

In addition to SPRINT-008 event log:

| `PaymentStatus` | Order action | Entitlement |
| --- | --- | --- |
| `0` | `pending` → `paid` | Revoke prior active; assign purchased package |
| `1` | unchanged `pending` | none |
| `2` | `pending` → `failed` | none |

Idempotent: replay on already-`paid` order returns `200` without duplicate entitlement.

### `POST /api/dev/mock-pay/{order_id}` *(dev / mock only)*

**Auth:** `tenant_admin` Bearer; order must belong to caller's tenant.

Triggers same fulfillment as successful callback. Disabled when `APP_ENV=prod`.

**Response 200:**

```json
{ "order_id": "ord_a1b2c3d4", "status": "paid" }
```

### Tenant Checkout error codes

| Code | When |
| --- | --- |
| `403` | Not `tenant_admin`, tenant `pending_kyc`, or wrong tenant on order |
| `409` | Package archived or tenant not `active` |
| `503` | Payment gateway not configured |

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

## Quota & rate limit (Sprint 13)

Base: `http://localhost:8091`. Deep spec: [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md).  
Workflows: [02-workflow.md](02-workflow.md) §32–36 · UX: [05-ux-ui.md](05-ux-ui.md) § P14.

### Platform usage

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `GET` | `/api/platform/tenants/{tenant_id}/usage` | `platform_admin` | Effective limits + current usage snapshot |

**Auth:** Bearer required. `tenant_admin` → **403**.

**Path params**

| Field | Type | Notes |
| --- | --- | --- |
| `tenant_id` | string | Existing tenant id (e.g. `demo`) |

**Response 200** — active entitlement

```json
{
  "tenant_id": "demo",
  "package": { "id": "pkg-starter", "slug": "starter", "name": "Starter" },
  "status": "active",
  "period": "2026-07",
  "limits": {
    "max_ai_employees": 2,
    "max_monthly_call_minutes": 500,
    "max_km_documents": 50,
    "max_concurrent_calls": 2,
    "voice_enabled": true,
    "rag_enabled": true
  },
  "usage": {
    "ai_employees": 1,
    "monthly_call_minutes": 12,
    "km_documents": 3,
    "concurrent_calls": 0
  }
}
```

**Response 200** — no entitlement

```json
{
  "tenant_id": "acme",
  "package": null,
  "status": "none",
  "period": "2026-07",
  "limits": null,
  "usage": {
    "ai_employees": 0,
    "monthly_call_minutes": 0,
    "km_documents": 0,
    "concurrent_calls": 0
  }
}
```

**Errors**

| Status | When |
| ---: | --- |
| `401` | Missing/invalid token |
| `403` | Not `platform_admin` |
| `404` | Unknown `tenant_id` (optional; may return `status:none` if tenant exists without package) |

### Enforcement side effects (existing routes)

Tenant for customer paths: JWT tenant, else `DEMO_TENANT_ID` when `AUTH_DISABLED=true`.

| Method | Path | Checks | HTTP / `code` |
| --- | --- | --- | --- |
| `POST` | `/api/chat` | rate `chat`; `rag_enabled` if RAG | 429 `rate_limited` · 403 `feature_disabled` |
| `GET` | `/ws/voice` | rate `voice`; `voice_enabled`; monthly minutes; concurrent slot | 429 `quota_exceeded` / `rate_limited` · 403 `feature_disabled` |
| `POST` | `/api/km/agents/{agent_id}/documents` | rate `km`; `max_km_documents` | 429 |
| `POST` | `/api/platform/tenants/{tenant_id}/avatars` | `max_ai_employees` | 429 `quota_exceeded` |

**Error body**

```json
{
  "error": "KM document limit exceeded",
  "code": "quota_exceeded",
  "dimension": "max_km_documents",
  "limit": 50,
  "usage": 50
}
```

| `code` | HTTP | Headers |
| --- | ---: | --- |
| `quota_exceeded` | 429 | — |
| `rate_limited` | 429 | `Retry-After: <seconds>` preferred |
| `feature_disabled` | 403 | — |
| `no_entitlement` | 403/404 | only if `QUOTA_FAIL_OPEN=false` |

When `QUOTA_ENABLED=false` or (`QUOTA_FAIL_OPEN=true` and Redis error): request proceeds; server logs warning.

### Infra

`GET /api/infra` adds:

```json
{
  "quota": "ok",
  "rate_limit": "ok"
}
```

| Value | Meaning |
| --- | --- |
| `ok` | Enabled and Redis reachable |
| `disabled` | Env master switch off |
| `degraded` | Redis error; fail-open may still serve |

## Error envelope

```json
{"error": "human-readable message"}
```

Sprint 13 quota errors may add `code`, `dimension`, `limit`, `usage` (see above).

## Embed to Web (Sprint 14)

Deep spec: [17-embed-to-web-spec.md](17-embed-to-web-spec.md). Workflows §37–39.

### Public

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| `GET` | `/api/public/embed/{embed_key}` | none | Resolve tenant embed config |
| `GET` | `/embed/monti-embed.js` | none | Loader script |
| `GET` | `/embed` | none | Embed SPA (customer compact UI) |

**GET `/api/public/embed/{embed_key}`**

Headers: `Origin` recommended when allowlist set.

**200**

```json
{
  "tenant_id": "demo",
  "slug": "demo",
  "name": "Demo Workspace",
  "embed_key": "emb_abc…",
  "enabled": true,
  "default_agent_id": "ava",
  "agents": [{ "id": "ava", "name": "Ava", "role": "Receptionist" }]
}
```

| Status | code |
| ---: | --- |
| 404 | `embed_not_found` / `embed_disabled` |
| 403 | `origin_not_allowed` |

Chat/voice from embed: existing `POST /api/chat`, `GET /ws/voice` with `X-Tenant-Id: {tenant_id}` from resolve (quota applies).

### Tenant admin

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `GET` | `/api/tenant/embed` | `tenant_admin` active | Get or lazy-create config |
| `PUT` | `/api/tenant/embed` | `tenant_admin` active | Update enabled, origins, default_agent |
| `POST` | `/api/tenant/embed/rotate-key` | `tenant_admin` active | New embed_key |

**PUT body**

```json
{
  "enabled": true,
  "allowed_origins": ["https://shop.example", "http://localhost:5500"],
  "default_agent_id": "ava"
}
```

**POST rotate-key 200:** `{ "embed_key": "emb_…", "enabled": true, ... }`

## Tenant KM / Scope (Sprint 15)

Auth: **Bearer** `tenant_admin` + tenant **active** (`RequireTenantAdminActive`).  
**Tenant id always from JWT** — never accept `tenant_id` from body/query for these routes.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/tenant/km/scopes` | Allowed scope tags |
| `GET` | `/api/tenant/km/agents` | Agents for tenant + doc counts |
| `GET` | `/api/tenant/km/agents/{agent_id}/documents` | List documents for agent |
| `POST` | `/api/tenant/km/agents/{agent_id}/documents` | Upload + ingest |
| `PATCH` | `/api/tenant/km/documents/{id}` | Update `km_scope` |
| `DELETE` | `/api/tenant/km/documents/{id}` | Cascade delete one doc |
| `POST` | `/api/tenant/km/agents/{agent_id}/reset` | Clear all KM for agent |
| `GET` | `/api/tenant/km/gaps` | List unanswered questions (`km_gaps`) |
| `PATCH` | `/api/tenant/km/gaps/{id}` | Update gap status / notes |

### GET `/api/tenant/km/gaps`

Query: `status` (optional), `agent_id` (optional), default limit 100.

**200**

```json
{
  "gaps": [
    {
      "id": "…",
      "tenant_id": "t_…",
      "agent_id": "ava",
      "topic": "billing",
      "question": "Do you offer student discounts?",
      "source": "chat",
      "status": "open",
      "occurrence_count": 3,
      "last_seen_at": "2026-07-12T10:00:00Z"
    }
  ]
}
```

### PATCH `/api/tenant/km/gaps/{id}`

```json
{ "status": "dismissed", "notes": "Out of scope", "resolved_document_id": "" }
```

`status`: `open` | `resolved` | `dismissed` | `converted`

### GET `/api/tenant/km/scopes`

**200**

```json
{
  "scopes": [
    { "id": "general", "label": "General" },
    { "id": "billing", "label": "Billing" },
    { "id": "technical", "label": "Technical" }
  ]
}
```

### GET `/api/tenant/km/agents`

Prefer assigned workforce agents for the tenant (S5); if none assigned, fall back to catalog agents with `assigned: false`.

**200**

```json
{
  "agents": [
    {
      "id": "ava",
      "name": "Ava",
      "role": "General Support",
      "doc_count": 3,
      "chunk_count": 40,
      "by_scope": { "general": 2, "billing": 0, "technical": 1 },
      "default_scopes": ["general"],
      "assigned": true
    }
  ]
}
```

`default_scopes` is **read-only** help from `internal/scope.AgentScopes` (S2).

### GET `/api/tenant/km/agents/{agent_id}/documents`

| Query | Required | Notes |
| --- | --- | --- |
| — | — | Optional future: `?scope=general` filter |

**200**

```json
{
  "agent_id": "ava",
  "documents": [
    {
      "id": "doc_…",
      "tenant_id": "t_…",
      "agent_id": "ava",
      "filename": "faq.md",
      "mime": "text/markdown",
      "status": "indexed",
      "km_scope": "general",
      "km_version": 1,
      "chunk_count": 12,
      "created_at": "2026-07-12T00:00:00Z",
      "updated_at": "2026-07-12T00:00:00Z"
    }
  ]
}
```

Do **not** return `object_key` to the browser (internal storage path).

### POST `/api/tenant/km/agents/{agent_id}/documents`

`Content-Type: multipart/form-data`

| Field | Required | Notes |
| --- | --- | --- |
| `file` | yes | text/markdown/plain; max ~8MB (existing server limit) |
| `scope` | no | default `scope.DefaultScope(agent_id)` |

**S13:** `AllowRate(BucketKM)` + `CheckKMDocument` before ingest.

**201** — document object (status usually `indexed` on sync success, or `failed` with 502 if index fails after create — prefer atomic failure: 502 without orphan when possible).

### PATCH `/api/tenant/km/documents/{id}`

```json
{ "km_scope": "billing" }
```

| Field | Required | Notes |
| --- | --- | --- |
| `km_scope` | yes | one of general \| billing \| technical |

**200** — updated document. Updates PG chunks + CH `km_scope` for that document (re-embed optional if vectors independent of scope tag).

### DELETE `/api/tenant/km/documents/{id}`

**200**

```json
{ "deleted": true, "id": "doc_…" }
```

Or **204** No Content. Prefer **200** with body for UI.

Cascade: CH embeddings → MinIO object → PG document (chunks CASCADE).

### POST `/api/tenant/km/agents/{agent_id}/reset`

**200**

```json
{
  "agent_id": "ava",
  "status": "reset",
  "message": "knowledge base cleared for agent"
}
```

### Error codes

| HTTP | error / code | When |
| ---: | --- | --- |
| 400 | `unknown agent_id` / `invalid scope` / `file is required` / `file is empty` | Validation |
| 401 | `unauthorized` | Missing/invalid JWT |
| 403 | inactive tenant / not tenant_admin | RBAC |
| 404 | `document not found` | Wrong id or other tenant |
| 429 | S13 rate limit codes | KM write rate |
| 403 | S13 `max_km_documents` etc. | Quota (existing writeQuotaError) |
| 502 | ingest/embed failure | Downstream KM pipeline |

### Dual surface (legacy)

| Path | Audience |
| --- | --- |
| `/api/tenant/km/*` | **Product** — tenant portal S15 |
| `/api/km/agents/*` | Ops / platform seed / OptionalBearer (S2) |

Do not remove legacy routes in S15.

### Packages

| Package | Role |
| --- | --- |
| `internal/km` | Ingest, DeleteDocument, ResetAgent, UpdateScope |
| `internal/store` | PG + MinIO KM methods |
| `internal/clickhouse` | Embedding insert/delete |
| `internal/scope` | ValidAgent, DefaultScope, scope tags |
| `internal/quota` | KM rate + document limits |
| `internal/auth` | RequireTenantAdminActive |

## Tenant Settings & Call Limits (Sprint 16)

Auth: **Bearer** `tenant_admin` + active. Tenant id from JWT only.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/tenant/settings` | Lazy-create defaults |
| `PUT` | `/api/tenant/settings` | locale, timezone, display_name, ai_reply_locale, labels |
| `GET` | `/api/tenant/usage` | Package limits + usage snapshot (own tenant) |
| `GET` | `/api/tenant/call-limits` | Per-call / daily caps |
| `PUT` | `/api/tenant/call-limits` | Update caps (`0` = unset) |

**PUT settings body:**

```json
{
  "locale": "th",
  "timezone": "Asia/Bangkok",
  "display_name": "Libra Support",
  "ai_reply_locale": "th",
  "user_tier_label": "",
  "user_group_label": ""
}
```

**PUT call-limits body:**

```json
{ "max_minutes_per_call": 15, "max_call_minutes_per_day": 120 }
```

| HTTP | error |
| ---: | --- |
| 400 | invalid locale / timezone / negative limits |
| 401/403 | not tenant_admin / inactive |
| 403 | `daily_call_limit` / `per_call_limit` on voice |

## Tenant Test & Preview (Sprint 17)

Auth: **Bearer** `tenant_admin` + active. Tenant id from JWT only.

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/api/tenant/preview/chat` | Preview text chat (RAG + locale; **no** minute quota) |
| `GET` | `/ws/tenant/preview/voice` | Preview voice WS (**no** minute quota; rate limit + concurrent cap) |

**Policy:** rate limits apply; package monthly minutes and S16 daily/per-call **do not** apply. Soft concurrent preview cap (env `PREVIEW_MAX_CONCURRENT`, default 2).

**POST body:** same as public chat (`agent_id`, `topic`, `message`, `history`, optional `session_id`).

**200 response:** public chat shape + `"mode": "preview"`.

| HTTP | code |
| ---: | --- |
| 401/403 | not tenant_admin / inactive |
| 429 | `rate_limited` / `preview_concurrent` |

## Customer Tiers & Groups (Sprint 18)

Auth: **Bearer** `tenant_admin` + active. Tenant id from JWT only.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/tenant/tiers` | List tiers |
| `POST` | `/api/tenant/tiers` | Create tier |
| `GET` | `/api/tenant/tiers/{id}` | Get one |
| `PUT` | `/api/tenant/tiers/{id}` | Update |
| `DELETE` | `/api/tenant/tiers/{id}` | Delete |
| `GET` | `/api/tenant/groups` | List groups |
| `POST` | `/api/tenant/groups` | Create group |
| `GET` | `/api/tenant/groups/{id}` | Get one |
| `PUT` | `/api/tenant/groups/{id}` | Update |
| `DELETE` | `/api/tenant/groups/{id}` | Delete |

**POST tier body:** see [21-customer-tier-spec.md](21-customer-tier-spec.md).

**Preview:** `POST /api/tenant/preview/chat` may include `tier_id`; voice WS may pass `tier_id` query for cap/locale overrides.

| HTTP | code |
| ---: | --- |
| 400 | invalid slug / locale / negative caps |
| 404 | tier not found (or other tenant) |
| 401/403 | not tenant_admin / inactive |

## Customer Accounts & Imports (Sprint 19)

**Auth:** Bearer, active `tenant_admin`. Tenant id is read only from JWT. `AUTH_DISABLED` does not expose these routes.

### Endpoint summary

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/tenant/customers` | Search/list tenant customers |
| `POST` | `/api/tenant/customers` | Create or idempotently upsert a customer |
| `GET` | `/api/tenant/customers/{id}` | Get one tenant customer |
| `PUT` | `/api/tenant/customers/{id}` | Update profile and assignments |
| `DELETE` | `/api/tenant/customers/{id}` | Deactivate customer |
| `POST` | `/api/tenant/customer-imports` | CSV dry-run or commit |
| `GET` | `/api/tenant/customer-imports/{id}` | Read import summary |
| `GET` | `/api/tenant/customer-domain-rules` | List domain rules |
| `POST` | `/api/tenant/customer-domain-rules` | Create rule |
| `PUT` | `/api/tenant/customer-domain-rules/{id}` | Update rule |
| `DELETE` | `/api/tenant/customer-domain-rules/{id}` | Delete rule |

### `GET /api/tenant/customers`

| Query | Type | Required | Description |
| --- | --- | --- | --- |
| `q` | string | no | Search display name, normalized email, phone, external id |
| `status` | string | no | `active`, `inactive`, or empty for all |
| `tier_id` | string | no | Tenant-owned tier filter |
| `limit` | int | no | Default 50, maximum 200 |
| `cursor` | string | no | Opaque pagination cursor |

**Response `200`**

```json
{
  "customers": [
    {
      "id": "cust_01",
      "email": "jane@example.com",
      "phone": "+66812345678",
      "display_name": "Jane Doe",
      "locale": "th",
      "tier_id": "tier_vip",
      "group_ids": ["grp_retail"],
      "source": "csv",
      "external_id": "crm-42",
      "status": "active",
      "metadata": {},
      "created_at": "2026-07-12T12:00:00Z",
      "updated_at": "2026-07-12T12:00:00Z"
    }
  ],
  "next_cursor": ""
}
```

`tenant_id` and `email_normalized` are never writable and need not be returned.

### `POST /api/tenant/customers`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `display_name` | string | yes | Trimmed, 1–200 chars |
| `email` | string | conditional | Required unless external id supplied |
| `phone` | string | no | Max 40 chars |
| `locale` | string | no | Empty, `en`, or `th` |
| `tier_id` | string | no | Same-tenant tier |
| `group_ids` | string[] | no | Same-tenant groups |
| `source` | string | no | Default `manual`; reserved values validated |
| `external_id` | string | conditional | Stable source id |
| `metadata` | object | no | Max serialized 16 KiB; credentials forbidden |

**Request**

```json
{
  "display_name": "Jane Doe",
  "email": "jane@example.com",
  "locale": "th",
  "tier_id": "tier_vip",
  "group_ids": ["grp_retail"],
  "source": "api",
  "external_id": "crm-42"
}
```

**Response `201`** for create or `200` for idempotent update: customer shape plus `"outcome":"created|updated"`.

### `GET|PUT|DELETE /api/tenant/customers/{id}`

`GET` returns the customer shape. `PUT` accepts the mutable POST fields; source/external id may be set only when they do not conflict. `DELETE` performs soft deactivation.

**DELETE response `200`**

```json
{ "id": "cust_01", "status": "inactive" }
```

### `POST /api/tenant/customer-imports`

`Content-Type: multipart/form-data`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `file` | CSV file | yes | UTF-8, max `CUSTOMER_IMPORT_MAX_BYTES` |
| `dry_run` | boolean | yes | Must be true before UI enables commit |
| `source` | string | no | Default `csv` |

CSV columns: `display_name,email,phone,locale,tier_slug,group_slugs,source,external_id`. `group_slugs` uses `|` as the within-cell separator.

**Response `200` dry-run / `201` commit**

```json
{
  "id": "cimp_01",
  "mode": "dry_run",
  "status": "validated",
  "total_rows": 3,
  "accepted_rows": 2,
  "created_rows": 0,
  "updated_rows": 0,
  "rejected_rows": 1,
  "errors": [
    { "row": 3, "field": "email", "code": "invalid_email", "message": "Invalid email" }
  ]
}
```

Dry-run writes only the import summary, not customers or memberships. Commit reparses/revalidates the uploaded CSV; the client must send the file again.

### `GET /api/tenant/customer-imports/{id}`

Returns the import summary above. Other-tenant ids return 404. Raw CSV content is never returned because it is not retained.

### Customer domain rules

**POST/PUT body**

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `domain` | string | yes | Normalized lower-case domain, no scheme/path |
| `policy` | string | yes | `allow` or `deny` |
| `default_tier_id` | string | no | Same-tenant tier |
| `default_group_id` | string | no | Same-tenant group |
| `active` | boolean | no | Default true |

```json
{
  "domain": "example.com",
  "policy": "allow",
  "default_tier_id": "tier_standard",
  "default_group_id": "grp_retail",
  "active": true
}
```

`GET` returns `{ "rules": [...] }`. `POST` returns 201. `PUT` returns 200. `DELETE` returns `{ "deleted": true }`. SPRINT-019 stores `policy`; public registration/login enforcement begins in SPRINT-020.

### Errors

| HTTP | Code | When |
| ---: | --- | --- |
| 400 | `validation_error` | Invalid body, locale, domain, references, CSV row, or metadata size |
| 400 | `import_invalid` | Encoding, header, byte/row cap, or parser failure |
| 401 | `unauthorized` | Missing/invalid token |
| 403 | `forbidden` | Wrong role or inactive tenant |
| 404 | `not_found` | Missing or cross-tenant customer/import/rule/reference |
| 409 | `customer_conflict` | Email and external id point to different existing customers |
| 409 | `domain_rule_exists` | Duplicate normalized domain in tenant |
| 413 | `import_too_large` | Upload exceeds configured byte/row cap |

See [22-customer-account-import-spec.md](22-customer-account-import-spec.md), [02-workflow.md](02-workflow.md) §55–58, [03-er-diagram.md](03-er-diagram.md) § Sprint 19, and [05-ux-ui.md](05-ux-ui.md) § T12.

## Customer Authentication & Domain Enforcement (Sprint 20)

SPRINT-020 introduces customer-scoped email OTP authentication and sessions. Tenant id is resolved from trusted tenant context: host, embed key, or existing tenant routing metadata. API bodies must not be allowed to switch tenants.

### Endpoint summary

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/tenant/customer-auth/settings` | active `tenant_admin` | Read auth settings |
| `PUT` | `/api/tenant/customer-auth/settings` | active `tenant_admin` | Enable/disable auth and domain enforcement |
| `POST` | `/api/customer/auth/request-otp` | public tenant context | Validate policy, create challenge, send OTP email |
| `POST` | `/api/customer/auth/verify-otp` | public tenant context | Verify OTP and issue customer session |
| `POST` | `/api/customer/auth/refresh` | refresh token | Rotate customer session tokens |
| `POST` | `/api/customer/auth/logout` | `customer` | Revoke current session |
| `GET` | `/api/customer/me` | `customer` | Current customer profile/context |

### `GET /api/tenant/customer-auth/settings`

Returns a lazy default if no row exists.

```json
{
  "enabled": false,
  "mode": "disabled",
  "domain_enforcement": "off",
  "allow_public_no_auth": true,
  "session_ttl_minutes": 60,
  "refresh_ttl_minutes": 43200
}
```

### `PUT /api/tenant/customer-auth/settings`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `enabled` | boolean | yes | Enables customer OTP auth endpoints |
| `mode` | string | yes | `disabled`, `optional`, or `required` |
| `domain_enforcement` | string | yes | `off`, `allowlist`, `denylist`, or `allowlist_and_denylist` |
| `allow_public_no_auth` | boolean | yes | Preserve public no-auth conversation path |
| `session_ttl_minutes` | int | no | 5-240, default 60 |
| `refresh_ttl_minutes` | int | no | 60-43200, default 43200 |

**Response `200`** returns settings. Invalid combinations return `400 validation_error`; for example `mode=disabled` forces `enabled=false`.

### `POST /api/customer/auth/request-otp`

Creates an OTP challenge for an existing imported customer or a policy-allowed self-claim customer, then sends a short OTP to the customer email.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `email` | string | yes | Normalized and matched to tenant/domain rules |
| `purpose` | string | yes | `login`, `register`, or `claim` |
| `display_name` | string | no | Required only when no imported customer exists |
| `external_id` | string | no | Optional claim hint for imported identities |

```json
{
  "email": "jane@example.com",
  "purpose": "login",
  "display_name": "Jane Doe"
}
```

**Response `202`**

```json
{
  "challenge_id": "cotp_01",
  "status": "otp_sent",
  "delivery": {
    "channel": "email",
    "to": "j***@example.com",
    "expires_in": 600,
    "resend_after": 60
  },
  "customer_hint": {
    "matched_existing_customer": true,
    "requires_profile_completion": false
  }
}
```

The OTP email contains only the OTP code, expiry, tenant-facing brand name, and safety copy. It never contains tokens or internal ids except the user-visible masked email context.

### `POST /api/customer/auth/verify-otp`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `challenge_id` | string | yes | OTP challenge id from request response |
| `otp` | string | yes | Numeric OTP from email |

**Request**

```json
{
  "challenge_id": "cotp_01",
  "otp": "123456"
}
```

**Response `200`**

```json
{
  "status": "authenticated",
  "access_token": "<jwt>",
  "refresh_token": "<opaque>",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_expires_in": 2592000,
  "customer": {
    "id": "cust_01",
    "tenant_id": "demo",
    "display_name": "Jane Doe",
    "email": "jane@example.com",
    "tier_id": "tier_vip",
    "group_ids": ["grp_retail"],
    "locale": "th",
    "role": "customer"
  }
}
```

Customer access tokens include `sub`, `tenant_id`, `customer_id`, `role=customer`, and `session_id`. They must not include OTP hashes, refresh token hashes, or domain-rule internals.

### `POST /api/customer/auth/refresh`

```json
{ "refresh_token": "<opaque>" }
```

**Response `200`** returns a rotated access/refresh token pair. Reuse of a revoked refresh token returns `401 session_revoked` and should revoke the session family when implemented.

### `POST /api/customer/auth/logout`

Requires `Authorization: Bearer <customer_access_token>`.

```json
{ "refresh_token": "<optional current refresh token>" }
```

**Response `200`**

```json
{ "revoked": true }
```

### `GET /api/customer/me`

Requires customer token.

```json
{
  "customer": {
    "id": "cust_01",
    "tenant_id": "demo",
    "display_name": "Jane Doe",
    "email": "jane@example.com",
    "status": "active",
    "tier_id": "tier_vip",
    "group_ids": ["grp_retail"],
    "locale": "th"
  },
  "auth": {
    "session_id": "csess_01",
    "expires_at": "2026-07-13T12:00:00Z"
  }
}
```

### Public chat/voice with customer token

Existing `POST /api/chat`, `POST /api/calls`, and voice WebSocket paths accept an optional customer Bearer token. When present and valid, quota, rate-limit, RAG, locale, tier, and group resolution use the session customer context. When absent, the existing public no-auth flow remains available unless tenant settings require auth.

### Domain policy behavior

| `domain_enforcement` | Behavior |
| --- | --- |
| `off` | Domain rules ignored for auth; existing S19 defaults still apply on imports |
| `allowlist` | Email domain must match active `allow` rule |
| `denylist` | Email domain must not match active `deny` rule |
| `allowlist_and_denylist` | Deny wins, otherwise require active allow |

### Errors

| HTTP | Code | When |
| ---: | --- | --- |
| 400 | `validation_error` | Invalid email, invalid OTP format, invalid settings mode |
| 401 | `otp_invalid_or_expired` | Challenge missing, expired, locked, or OTP mismatch |
| 401 | `session_expired` | Access/refresh token expired |
| 401 | `session_revoked` | Session was revoked or refresh token reused |
| 403 | `customer_auth_disabled` | Tenant has auth disabled |
| 403 | `domain_denied` | Active deny rule matched |
| 403 | `domain_not_allowed` | Allowlist mode and no active allow rule matched |
| 403 | `customer_inactive` | Linked S19 customer is inactive |
| 404 | `not_found` | Cross-tenant customer/session reference |
| 409 | `otp_already_pending` | Pending challenge exists and resend cooldown has not elapsed |
| 409 | `identity_conflict` | Email and claim hint resolve to different tenant customers |
| 429 | `rate_limited` | Auth attempt or chat/voice rate limit exceeded |
| 502 | `otp_delivery_failed` | Mail provider rejected or failed delivery |

`AUTH_DISABLED=true` does not mint customer sessions. It may continue to allow legacy public chat/voice for local development, but tenant-admin and customer-auth settings routes still require their explicit auth checks when `AUTH_DISABLED=false`.

See [23-customer-auth-spec.md](23-customer-auth-spec.md), [02-workflow.md](02-workflow.md) §59–63, [03-er-diagram.md](03-er-diagram.md) § Sprint 20, and [05-ux-ui.md](05-ux-ui.md) § T13.

## Authenticated Workforce Selection & Customer Quota (Sprint 21)

**Auth:** public policy read, optional/required `customer` token depending on tenant setting, `tenant_admin` for settings.

### `GET /api/customer/portal-policy`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `tenant_id` | string | yes | Tenant slug/id resolved from portal URL or embed context |

**Response `200`**

```json
{
  "tenant_id": "libra-tech-co-ltd",
  "customer_auth": {
    "enabled": true,
    "mode": "required",
    "require_auth_for_workforce": true,
    "allow_public_no_auth": false
  },
  "quota": {
    "daily_call_seconds": 1800,
    "max_call_seconds": 300
  }
}
```

### `GET /api/customer/workforce`

Returns active tenant-assigned avatars available for customer selection.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `tenant_id` | string | yes | Tenant context |

**Response `200`**

```json
{
  "avatars": [
    {
      "id": "ava",
      "name": "Ava",
      "role": "General Support",
      "status": "active",
      "portrait_url": "/api/assets/avatars/ava/portrait.png",
      "quota_state": "available"
    }
  ],
  "selected_avatar_id": "ava"
}
```

### `GET /api/customer/quota`

Requires customer token when tenant auth is required; optional otherwise.

**Response `200`**

```json
{
  "tenant_id": "libra-tech-co-ltd",
  "customer_id": "cust_01",
  "daily_remaining_seconds": 1200,
  "max_call_seconds": 300,
  "reset_at": "2026-07-14T00:00:00+07:00",
  "state": "quota_available"
}
```

### `PUT /api/tenant/customer-auth/settings` additions

SPRINT-021 extends the S20 request body:

```json
{
  "mode": "required",
  "require_auth_for_workforce": true,
  "allow_public_no_auth": false,
  "customer_daily_call_seconds": 1800,
  "customer_max_call_seconds": 300
}
```

### Chat/call errors added in Sprint 21

| HTTP | Code | When |
| ---: | --- | --- |
| 401 | `customer_auth_required` | Tenant requires customer auth before workforce/chat/voice |
| 403 | `avatar_unavailable` | Avatar disabled, unassigned, or unavailable for tenant/customer |
| 403 | `call_duration_limit_exceeded` | Requested/active call exceeds per-call duration limit |
| 429 | `customer_quota_exhausted` | Daily customer quota exhausted |

## Conversation Records & Knowledge Gaps (Sprint 22)

**Auth:** `tenant_admin`; all resources are tenant-scoped from token context.

### `GET /api/tenant/conversation-records`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `from` | datetime | no | Start filter |
| `to` | datetime | no | End filter |
| `avatar_id` | string | no | Filter by workforce avatar |
| `customer_id` | string | no | Filter by customer |
| `status` | string | no | `recording`, `archived`, `archive_failed` |

**Response `200`**

```json
{
  "records": [
    {
      "id": "crec_01",
      "call_id": "call_01",
      "channel": "voice",
      "customer_id": "cust_01",
      "avatar_id": "ava",
      "status": "archived",
      "started_at": "2026-07-13T10:00:00Z",
      "duration_seconds": 180,
      "archive_object_count": 2,
      "knowledge_gap_count": 1
    }
  ],
  "next_cursor": null
}
```

### `GET /api/tenant/conversation-records/{id}`

Returns safe metadata, transcript preview when permitted, archive object metadata, and linked gap ids. Object keys are not returned unless the caller has explicit archive-access permission.

### `POST /api/tenant/conversation-records/{id}/archive/retry`

Retries failed archive writes.

**Response `202`**

```json
{ "status": "retry_queued" }
```

### `GET /api/tenant/knowledge-gaps`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `status` | string | no | `open`, `snoozed`, `resolved`, `ignored` |
| `avatar_id` | string | no | Filter by avatar |
| `reason` | string | no | `no_source`, `low_confidence`, `fallback`, `tenant_flag` |

**Response `200`**

```json
{
  "gaps": [
    {
      "id": "kgap_01",
      "conversation_record_id": "crec_01",
      "question": "What is Libra Tech warranty policy?",
      "gap_reason": "no_source",
      "status": "open",
      "confidence": 0.22,
      "created_at": "2026-07-13T10:02:00Z"
    }
  ]
}
```

### `PATCH /api/tenant/knowledge-gaps/{id}`

```json
{
  "status": "resolved",
  "reviewer_note": "Added warranty policy to tenant KM"
}
```

**Errors**

| HTTP | Code | When |
| ---: | --- | --- |
| 400 | `validation_error` | Invalid filters, status, or review note |
| 401 | `unauthorized` | Missing/invalid tenant token |
| 403 | `forbidden` | Wrong role |
| 404 | `not_found` | Cross-tenant or missing record/gap |
| 409 | `archive_not_retryable` | Archive object is already stored/deleted |
| 502 | `archive_write_failed` | MinIO write failed |

## Tickets & Human Escalation (Sprint 23)

**Auth:** customer ticket creation is public or `customer` according to the existing tenant policy. Tenant queue APIs require `tenant_admin`; tenant id always comes from the token context.

### `POST /api/customer/tickets`

Creates a tenant ticket only after the customer explicitly confirms the structured `ticket_offer`. For anonymous callers, `contact_email` is required for follow-up. Send the same `Idempotency-Key` on retries.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `call_id` | string | conditional | Source call/session id when available |
| `conversation_record_id` | string | no | Finalized Sprint 22 record, when available |
| `confirm_escalation` | boolean | yes | Must be `true`; an offer alone never creates a ticket |
| `subject` | string | yes | 1-160 characters |
| `description` | string | yes | 1-2000 character bounded summary |
| `category` | string | yes | `general`, `billing`, `technical`, `other` |
| `contact_name` | string | no | Bounded anonymous requester name |
| `contact_email` | string | conditional | Required when no customer session identifies follow-up contact |

**Request**

```json
{
  "call_id": "call_01",
  "conversation_record_id": "crec_01",
  "confirm_escalation": true,
  "subject": "Need a human follow-up",
  "description": "Customer requested human help with a billing question.",
  "category": "billing",
  "contact_name": "PUP",
  "contact_email": "customer@example.com"
}
```

**Response `201`**

```json
{
  "ticket": {
    "id": "tick_01",
    "status": "open",
    "priority": "normal",
    "category": "billing",
    "source": "customer_request",
    "conversation_record_id": "crec_01",
    "created_at": "2026-07-15T10:02:00Z"
  }
}
```

### `GET /api/tenant/tickets`

Returns the current tenant's queue. The default view is recent tickets; clients may pass date filters for historical review.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `start_date` | date | no | Inclusive `YYYY-MM-DD` start filter |
| `end_date` | date | no | Inclusive `YYYY-MM-DD` end filter |
| `status` | string | no | `open`, `in_progress`, `waiting_customer`, `resolved`, `closed` |
| `priority` | string | no | `low`, `normal`, `high`, `urgent` |
| `category` | string | no | `general`, `billing`, `technical`, `other` |
| `avatar_id` | string | no | Filter by source AI employee |
| `customer_id` | string | no | Filter by known customer |
| `assignee_user_id` | string | no | Filter by assigned tenant user |

**Response `200`**

```json
{
  "tickets": [
    {
      "id": "tick_01",
      "subject": "Need a human follow-up",
      "category": "billing",
      "priority": "normal",
      "status": "open",
      "customer_label": "PUP",
      "avatar_id": "ava",
      "source": "customer_request",
      "last_activity_at": "2026-07-15T10:02:00Z"
    }
  ],
  "next_cursor": null
}
```

### `GET /api/tenant/tickets/{id}`

Returns safe ticket detail, masked requester contact, source conversation summary/link, and the ticket event timeline. A missing or cross-tenant id returns `404 not_found`.

**Response `200`**

```json
{
  "ticket": {
    "id": "tick_01",
    "subject": "Need a human follow-up",
    "description": "Customer requested human help with a billing question.",
    "category": "billing",
    "priority": "normal",
    "status": "open",
    "assignee_user_id": null,
    "contact_email_masked": "c***@example.com",
    "conversation_record_id": "crec_01"
  },
  "events": [
    { "type": "created", "actor_type": "customer", "created_at": "2026-07-15T10:02:00Z" }
  ]
}
```

### `PATCH /api/tenant/tickets/{id}`

Updates operator-controlled fields and validates the lifecycle transition.

**Request**

```json
{ "status": "in_progress", "priority": "high", "assignee_user_id": "usr_01" }
```

**Response `200`**

```json
{ "ticket": { "id": "tick_01", "status": "in_progress", "priority": "high", "assignee_user_id": "usr_01" } }
```

### `POST /api/tenant/tickets/{id}/events`

Adds a bounded internal note to the tenant ticket timeline.

**Request**

```json
{ "type": "note", "note": "Asked billing team to review the invoice." }
```

**Response `201`**

```json
{ "event": { "id": "tev_01", "type": "note_added", "created_at": "2026-07-15T10:05:00Z" } }
```

### Ticket errors

| HTTP | Code | When |
| ---: | --- | --- |
| 400 | `validation_error` | Missing confirmation, invalid filter, field, or lifecycle transition |
| 401 | `customer_auth_required` / `unauthorized` | Tenant policy requires customer auth or Bearer is invalid |
| 403 | `forbidden` | Wrong role or ticket creation disabled |
| 404 | `not_found` | Missing/cross-tenant source or ticket id |
| 409 | `ticket_already_open` | Same source conversation already has an open ticket without a matching idempotency key |
| 409 | `idempotency_conflict` | Idempotency key reused with a different request body |
| 429 | `ticket_rate_limited` | Optional per-customer daily ticket guard exceeded |

See [26-tickets-human-escalation-spec.md](26-tickets-human-escalation-spec.md), [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md), [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), and [05-ux-ui.md](05-ux-ui.md).

See [06-auth-spec.md](06-auth-spec.md), [08-packages-spec.md](08-packages-spec.md), [10-avatars-spec.md](10-avatars-spec.md), [11-tenant-register-spec.md](11-tenant-register-spec.md), [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md), [17-embed-to-web-spec.md](17-embed-to-web-spec.md), [18-tenant-scope-km-spec.md](18-tenant-scope-km-spec.md), [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md), [20-tenant-test-preview-spec.md](20-tenant-test-preview-spec.md), [21-customer-tier-spec.md](21-customer-tier-spec.md), [22-customer-account-import-spec.md](22-customer-account-import-spec.md), [23-customer-auth-spec.md](23-customer-auth-spec.md), [24-authenticated-workforce-selection-spec.md](24-authenticated-workforce-selection-spec.md), [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md), [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), [05-ux-ui.md](05-ux-ui.md), and [docs/KM_SETUP.md](../../KM_SETUP.md).

## Customer Satisfaction (Sprint 24)

The customer rating endpoint extends the existing call contract. The tenant is resolved from the call session; clients must not choose the tenant in the request body.

### `POST /api/calls/{id}/rating`

**Auth:** public/optional customer bearer, matching the existing call routes. When customer authentication is required by tenant policy, the normal customer session must be present.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `score` | integer | yes | Whole number from 1 to 5; rendered as star icons |
| `review` | string | no | Existing bounded compatibility field; Sprint 24 UI does not expose free-form comments |

**Request**

```json
{ "score": 5 }
```

**Response `201`**

```json
{ "status": "saved" }
```

The operation is idempotent for `(tenant_id, call_id)` and updates the existing score on repeated submission.

**Errors**

| Code | When |
| --- | --- |
| `400 validation_error` | Invalid JSON, score outside 1-5, or compatibility review exceeds the bounded length |
| `401 customer_auth_required` | Tenant policy requires a customer session |
| `404 not_found` | Call id is missing or does not belong to the resolved tenant |
| `502 storage_unavailable` | Postgres is unavailable while saving the rating |

### `GET /api/tenant/satisfaction/statistics`

**Auth:** `tenant_admin` with an active tenant. `AUTH_DISABLED` does not bypass tenant-admin protection.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `start_date` | date | no | Inclusive `YYYY-MM-DD`; defaults to today in deployment timezone |
| `end_date` | date | no | Inclusive `YYYY-MM-DD`; defaults to today |
| `avatar_id` | string | no | Filter to one AI employee |
| `channel` | string | no | `chat` or `voice` |

**Response `200`**

```json
{
  "range": { "start_date": "2026-07-14", "end_date": "2026-07-14" },
  "total_completed_conversations": 18,
  "reviewed_conversations": 15,
  "unrated_conversations": 3,
  "review_completion_rate": 83.33,
  "average_score": 4.47,
  "distribution": { "1": 0, "2": 1, "3": 2, "4": 4, "5": 8 },
  "by_avatar": [
    { "avatar_id": "ava", "avatar_name": "Ava", "completed": 18, "reviewed": 15, "average_score": 4.47 }
  ],
  "by_channel": [
    { "channel": "voice", "completed": 12, "reviewed": 10, "average_score": 4.4 },
    { "channel": "chat", "completed": 6, "reviewed": 5, "average_score": 4.6 }
  ]
}
```

**Errors**

| Code | When |
| --- | --- |
| `400 validation_error` | Invalid date range or unsupported avatar/channel filter |
| `401 unauthorized` | Missing or invalid tenant-admin bearer |
| `403 forbidden` | Caller is not an active tenant admin |
| `404 not_found` | Requested avatar is outside the tenant assignment |
| `500 statistics_unavailable` | Aggregate query failed |

See [27-customer-satisfaction-statistics-spec.md](27-customer-satisfaction-statistics-spec.md), [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), and [05-ux-ui.md](05-ux-ui.md).

## Tenant Call Center Statistics and Quota Usage (Sprint 25)

The Sprint 25 endpoint is tenant-admin only. It reads activity from the ClickHouse projection and current quota state from the existing entitlement/Redis path. It does not replace `GET /api/tenant/usage` or expose raw conversation records.

### `GET /api/tenant/call-center/statistics`

**Auth:** `tenant_admin` with an active tenant. `AUTH_DISABLED` does not bypass tenant-admin protection.

| Query field | Type | Required | Description |
| --- | --- | --- | --- |
| `start_date` | date | no | Inclusive `YYYY-MM-DD`; defaults to today in the resolved tenant/deployment timezone. |
| `end_date` | date | no | Inclusive `YYYY-MM-DD`; defaults to `start_date` or today. |

When only one date is supplied, the missing date is set to the supplied date. `start_date` must be less than or equal to `end_date`. The range is applied to `usage_date`; quota values describe the current package period and today's operational cap.

**Response `200`**

```json
{
  "tenant_id": "libra-tech-co-ltd",
  "range": {
    "start_date": "2026-07-14",
    "end_date": "2026-07-14",
    "timezone": "Asia/Bangkok"
  },
  "freshness": {
    "source": "clickhouse",
    "generated_at": "2026-07-14T10:20:00Z",
    "last_projected_at": "2026-07-14T10:19:48Z"
  },
  "totals": {
    "sessions": 18,
    "chat_sessions": 6,
    "voice_sessions": 12,
    "voice_minutes": 32,
    "average_duration_seconds": 106
  },
  "quota": {
    "period": "2026-07",
    "monthly_used_minutes": 32,
    "monthly_limit_minutes": 5000,
    "monthly_remaining_minutes": 4968,
    "daily_used_minutes": 18,
    "daily_limit_minutes": 180,
    "daily_timezone": "Asia/Bangkok"
  },
  "by_avatar": [
    { "avatar_id": "ava", "avatar_name": "Ava", "sessions": 18, "voice_minutes": 32 }
  ],
  "by_channel": [
    { "channel": "voice", "sessions": 12, "minutes": 32 },
    { "channel": "chat", "sessions": 6, "minutes": 0 }
  ]
}
```

`monthly_used_minutes` and `daily_used_minutes` come from the existing quota snapshot/counter path. `totals.voice_minutes` is the selected date range's completed voice duration from ClickHouse. The two values may differ when the requested range is not the current quota period; the response labels them separately.

**Empty range `200`**

```json
{
  "tenant_id": "libra-tech-co-ltd",
  "range": { "start_date": "2026-07-01", "end_date": "2026-07-01", "timezone": "Asia/Bangkok" },
  "freshness": { "source": "clickhouse", "generated_at": "2026-07-14T10:20:00Z", "last_projected_at": null },
  "totals": { "sessions": 0, "chat_sessions": 0, "voice_sessions": 0, "voice_minutes": 0, "average_duration_seconds": 0 },
  "quota": { "period": "2026-07", "monthly_used_minutes": 32, "monthly_limit_minutes": 5000, "monthly_remaining_minutes": 4968, "daily_used_minutes": 18, "daily_limit_minutes": 180, "daily_timezone": "Asia/Bangkok" },
  "by_avatar": [],
  "by_channel": []
}
```

**Errors**

| Status | Code | When |
| --- | --- | --- |
| `400` | `validation_error` | Invalid date format or `start_date` after `end_date`. |
| `401` | `unauthorized` | Missing or invalid bearer token. |
| `403` | `forbidden` | Caller is not an active tenant administrator. |
| `502` | `quota_unavailable` | Existing entitlement or Redis quota snapshot cannot be read. |
| `503` | `analytics_unavailable` | ClickHouse projection/query is unavailable. |
| `500` | `statistics_unavailable` | Unexpected aggregate or response construction failure. |

The error response uses the existing shape:

```json
{ "error": "analytics unavailable", "code": "analytics_unavailable" }
```

### Projection operations

The replay/backfill path is an operator job and is not exposed as an HTTP endpoint in Sprint 25. It reads Postgres `conversation_records` and `call_sessions`, then writes the deterministic ClickHouse fact through `internal/clickhouse`. Tenant users only consume the read endpoint above.

See [28-call-center-statistics-spec.md](28-call-center-statistics-spec.md), [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), and [05-ux-ui.md](05-ux-ui.md).

## Tenant System Performance (Sprint 26)

Sprint 26 adds one tenant-admin read endpoint. It returns a normalized, redacted snapshot assembled from existing dependency clients. It does not replace `/healthz`, `/api/infra`, `/api/tenant/usage`, or the call-center statistics endpoint.

### `GET /api/tenant/system-performance`

**Auth:** `tenant_admin` with an active tenant. `AUTH_DISABLED` does not bypass this guard. The request has no query parameters and cannot select a tenant.

**Response `200`**

```json
{
  "overall_status": "degraded",
  "checked_at": "2026-07-14T10:30:00Z",
  "components": [
    { "name": "postgres", "status": "operational", "latency_ms": 4, "checked_at": "2026-07-14T10:29:59Z" },
    { "name": "redis", "status": "operational", "latency_ms": 2, "checked_at": "2026-07-14T10:29:59Z" },
    { "name": "minio", "status": "operational", "latency_ms": 7, "checked_at": "2026-07-14T10:29:59Z" },
    { "name": "clickhouse", "status": "operational", "latency_ms": 9, "checked_at": "2026-07-14T10:29:59Z" },
    { "name": "nats", "status": "disabled", "latency_ms": null, "checked_at": "2026-07-14T10:29:59Z" },
    { "name": "livekit", "status": "operational", "latency_ms": null, "checked_at": "2026-07-14T10:29:59Z" },
    { "name": "gemini", "status": "degraded", "latency_ms": null, "checked_at": "2026-07-14T10:29:59Z" }
  ],
  "analytics": {
    "status": "current",
    "generated_at": "2026-07-14T10:29:58Z",
    "last_projected_at": "2026-07-14T10:29:48Z"
  }
}
```

Only the allowlisted component names and normalized statuses are serialized. `latency_ms` is null for configuration-only checks. `analytics.status` is `current`, `stale`, `unavailable`, or `disabled`; it is independent from the live dependency status.

**Status values**

| Field | Values | Meaning |
| --- | --- | --- |
| `overall_status` | `operational`, `degraded`, `unavailable` | Aggregate tenant-safe state. |
| `components[].status` | `operational`, `degraded`, `unavailable`, `disabled` | State of one configured dependency. |
| `analytics.status` | `current`, `stale`, `unavailable`, `disabled` | Freshness of the existing analytics projection. |

**Errors**

| Status | Code | When |
| --- | --- | --- |
| `401` | `unauthorized` | Missing or invalid bearer token. |
| `403` | `forbidden` | Caller is not an active tenant administrator. |
| `503` | `monitoring_unavailable` | A snapshot cannot be produced at all; partial probe failures should return a `200` normalized snapshot. |
| `500` | `monitoring_error` | Unexpected response construction failure. |

The error response uses the existing shape:

```json
{ "error": "monitoring unavailable", "code": "monitoring_unavailable" }
```

Raw provider messages, credentials, dependency URLs, tenant ids, customer data, transcripts, ticket notes, rating comments, and audio paths are never returned. Probe timeouts use the configured bounded deadline and do not block customer operations.

See [29-tenant-system-performance-spec.md](29-tenant-system-performance-spec.md), [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), and [05-ux-ui.md](05-ux-ui.md).

## Mobile Call API and SDK (Sprint 27)

Sprint 27 adds a versioned mobile facade. It reuses existing call sessions, customer authentication, avatar assignment, quota, rate-limit, transcript, archive, and rating behavior. It does not replace /api/calls or /ws/voice.

### Public brand directory

`GET /api/public/brands` is a public JSON directory for mobile tenant selection. It supports `q`, `limit` (default 50, max 100), and `offset`, and returns `{items,total,limit,offset}`. Results include only active tenants with an active brand, approved KYC (or no KYC row in local/demo deployments), `listed=true`, and `platform_listed=true`.

`GET /api/public/brands/{slug}` returns `{item}` for one public brand. Tenant administrators update profile fields with `PUT /api/tenant/brand`; platform administrators can force-unlist with `PUT /api/platform/tenants/{tenant_id}/brand-listing`.

### Endpoint summary

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| POST | /api/mobile/v1/calls | customer Bearer when required by tenant policy | Create a mobile call session. |
| GET | /api/mobile/v1/calls/{call_id} | same caller and tenant context | Read safe call state. |
| POST | /api/mobile/v1/calls/{call_id}/end | same caller and tenant context | Idempotently end a call. |
| POST | /api/mobile/v1/calls/{call_id}/rating | same caller and tenant context | Save or update a 1-5 star score. |
| GET | /ws/mobile/v1/calls/{call_id} | Bearer on upgrade | Exchange mobile audio/control frames and receive events. |

Tenant resolution is trusted context only. The request body cannot switch tenants.

### GET /api/mobile/v1/bootstrap

Returns the mobile-safe configuration for the resolved tenant. For an authenticated customer, tenant context comes from the Bearer token. For an optional/public flow, the existing trusted portal context is required.

Response 200:

~~~json
{
  "tenant": { "display_name": "Libra Tech Co.,Ltd", "slug": "libra-tech-co-ltd" },
  "auth": {
    "enabled": true,
    "mode": "required",
    "require_auth_for_workforce": true,
    "allow_public_no_auth": false,
    "otp": { "channel": "email", "ttl_seconds": 600, "resend_after_seconds": 60 }
  },
  "locale": {
    "default": "th",
    "customer": "th",
    "ai_reply": "th",
    "supported": ["en", "th"],
    "timezone": "Asia/Bangkok"
  },
  "avatars": [
    {
      "id": "ava",
      "name": "Ava",
      "role": "General Support",
      "trait": "Warm & Patient",
      "image_url": "/api/assets/avatars/ava/portrait.png",
      "status": "active",
      "is_default": true
    }
  ],
  "default_avatar_id": "ava",
  "limits": {
    "max_call_seconds": 300,
    "daily_limit_seconds": 1800,
    "daily_remaining_seconds": 1200,
    "warning_at_seconds": 10,
    "reset_at": "2026-07-16T00:00:00+07:00",
    "state": "quota_available"
  }
}
~~~

Only active tenant-assigned avatars are returned. Locale resolution is customer.locale, then tenant ai_reply_locale, then tenant locale, then en. Tenant timezone defaults to Asia/Bangkok. max_call_seconds and daily_remaining_seconds are the minimum applicable server-side limits; reset_at is the next tenant-local day boundary.

### Mobile OTP notification additions

`POST /api/customer/auth/request-otp` remains email OTP authentication. The mobile client may include:

~~~json
{
  "email": "jane@example.com",
  "locale": "th",
  "notification": {
    "platform": "ios|android",
    "push_token": "<apns-or-fcm-token>",
    "app_version": "1.0.0"
  }
}
~~~

Email is required for the current customer-auth contract. When email delivery succeeds and a valid iOS or Android token is supplied, the server may queue a non-sensitive APNs/FCM notification telling the customer to check email. The OTP itself must never be included in a push payload.

Response additions:

~~~json
{
  "notifications": { "push": "queued|disabled|rejected" },
  "expires_in": 600,
  "resend_after": 60
}
~~~

Push failures do not invalidate a successful email OTP delivery. APNs/FCM credentials stay server-side.

### POST /api/mobile/v1/calls

Headers:

| Header | Required | Description |
| --- | --- | --- |
| Authorization | policy-dependent | Customer Bearer token when tenant customer auth requires it. |
| Idempotency-Key | yes | 1-128 printable characters; same caller/key returns the original create response for the bounded TTL. |
| X-Monti-SDK-Version | no | SDK version for diagnostics; not an authorization input. |

Request:

~~~json
{
  "avatar_id": "ava",
  "locale": "th"
}
~~~

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| avatar_id | string | no | Active avatar assigned to the resolved tenant; omitted uses default_avatar_id from mobile bootstrap. |
| locale | string | no | Supported locale hint; server default applies when absent. |
| notification | object | no | Optional platform and push token for background call notifications. |
| tenant_id | string | forbidden | Never accepted as an authority for tenant resolution. |

`notification` accepts `platform` (`ios` or `android`), `push_token`, and `app_version`. The token is used only for non-sensitive lifecycle notifications and is never returned to the client or included in persisted conversation content.

Response 201:

~~~json
{
  "call_id": "call_01",
  "avatar": {
    "id": "ava",
    "name": "Ava",
    "role": "General Support"
  },
  "status": "active",
  "started_at": "2026-07-15T10:00:00Z",
  "expires_at": "2026-07-15T10:03:00Z",
  "time_limits": {
    "max_call_seconds": 300,
    "daily_remaining_seconds": 1200,
    "warning_at_seconds": 10,
    "reset_at": "2026-07-16T00:00:00+07:00"
  },
  "ws_path": "/ws/mobile/v1/calls/call_01"
}
~~~

No LiveKit token, room name, provider id, provider credential, tenant selector, or customer contact data is returned.

### GET /api/mobile/v1/calls/{call_id}

Response 200:

~~~json
{
  "call_id": "call_01",
  "status": "active",
  "avatar_id": "ava",
  "started_at": "2026-07-15T10:00:00Z",
  "ended_at": null,
  "remaining_seconds": 164,
  "last_event_at": "2026-07-15T10:00:12Z"
}
~~~

Allowed status values are creating, active, ending, ended, and failed. remaining_seconds is null when no per-call limit is configured. The response never includes a tenant selector, customer contact details, room name, recording path, or provider metadata.

### POST /api/mobile/v1/calls/{call_id}/end

Headers:

| Header | Required | Description |
| --- | --- | --- |
| Authorization | policy-dependent | Same caller context used to create the call. |
| Idempotency-Key | recommended | Repeated end requests return the same terminal state. |

Request:

~~~json
{ "reason": "customer" }
~~~

reason is customer, timeout, network, or omitted.

Response 200:

~~~json
{
  "call_id": "call_01",
  "status": "ended",
  "ended_at": "2026-07-15T10:02:11Z"
}
~~~

Calling end after the session is already ended returns the existing terminal response. It must not duplicate archive, quota, or rating side effects.

### POST /api/mobile/v1/calls/{call_id}/rating

Request:

~~~json
{ "score": 5 }
~~~

Response 201:

~~~json
{ "status": "saved" }
~~~

score is an integer from 1 through 5. The existing conversation_ratings uniqueness and upsert behavior makes repeated submission safe. Rating comments remain outside the mobile SDK v1 surface.

### GET /ws/mobile/v1/calls/{call_id}

The client sends a Bearer token during the WebSocket upgrade and the optional X-Monti-SDK-Version header. No query-string tenant selector is accepted.

Client frame types:

| Type | Fields | Meaning |
| --- | --- | --- |
| start_audio | sample_rate, channels, encoding | Declare PCM16 mono 16 kHz capture. |
| audio | data_base64 | One bounded PCM16 chunk. |
| text | text | Optional typed input through the same call session. |
| end | none | Request server-side end and terminal status. |

Server event types:

| Type | Fields | Meaning |
| --- | --- | --- |
| ready | call_id, avatar_id, encoding | Session accepted for media. |
| audio | data_base64 | AI audio chunk. |
| transcript | role, text, final | Caller or assistant transcript. |
| call_status | status | active, ending, ended, or failed. |
| turn_complete | none | Current turn completed. |
| error | code, retryable | Stable client-facing failure. |

### Error contract

| HTTP / close | Code | Client behavior |
| --- | --- | --- |
| 400 | validation_error | Fix request and do not retry automatically. |
| 401 | unauthorized, session_expired, customer_auth_required | Refresh token or show sign-in. |
| 403 | forbidden, avatar_not_available, feature_disabled | Stop and show policy-safe message. |
| 404 | not_found | Treat the call as inaccessible or already gone. |
| 409 | idempotency_conflict, call_not_active | Re-read status before creating another call. |
| 413 / WS error | audio_frame_too_large | Split frames; do not retry the same oversized frame. |
| 429 | quota_exceeded, concurrent_limit, rate_limited | Stop automatic create retries and expose retry timing. |
| 502/503 | voice_unavailable, storage_unavailable, service_unavailable | Retry only when retryable is true. |

The mobile contract maps provider and infrastructure failures to these stable codes. It never returns provider messages, credentials, URLs, tenant ids, customer email, transcript archive paths, or raw stack traces.

See 30-mobile-call-api-sdk-spec.md, 02-workflow.md §77–79, 03-er-diagram.md, and 05-ux-ui.md M1.
