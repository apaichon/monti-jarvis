---
id: DES-0004
title: API Specification
status: approved
updated: 2026-07-11
sprint: SPRINT-014
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

See [06-auth-spec.md](06-auth-spec.md), [08-packages-spec.md](08-packages-spec.md), [10-avatars-spec.md](10-avatars-spec.md), [11-tenant-register-spec.md](11-tenant-register-spec.md), [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md), [17-embed-to-web-spec.md](17-embed-to-web-spec.md), [02-workflow.md](02-workflow.md), [05-ux-ui.md](05-ux-ui.md), and [docs/KM_SETUP.md](../../KM_SETUP.md).