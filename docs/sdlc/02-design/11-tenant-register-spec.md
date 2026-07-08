---
id: DES-0011
title: Tenant Self-Registration Specification
status: approved
updated: 2026-07-08
sprint: SPRINT-006
owner: SA
---

# Tenant Self-Registration — Design Spec

**Sprint:** SPRINT-006 · **Release target:** v0.7.0  
**Feature:** [FEAT-0006](../01-features/FEAT-0006-tenant-register.md)  
**Depends on:** [06-auth-spec.md](06-auth-spec.md)

## 1. Goals

- **Public registration** creates a new SaaS tenant without platform operator action.
- Atomic provisioning: `tenants` (`pending_kyc`) + default `brands` row + `users` + `user_roles` (`tenant_admin`) + `tenant_registrations` audit row.
- Return **JWT tokens** immediately so the registrant can authenticate (same `TokenPair` shape as login).
- **`apps/tenant-web`** at `/tenant/register` for prospect-facing signup.
- **`GET /api/platform/tenants`** for `platform_admin` to see `pending_kyc` queue (Sprint 7 KYC builds on this).

## 2. Non-goals (Sprint 6)

- KYC approve/reject, document upload, status transitions to `active` (Sprint 7)
- Email verification, SMTP, invitations, password reset
- Auto-assign Starter package on signup (Sprint 9)
- Full tenant admin dashboard (Sprint 15+)
- `brands` channels, locales, search index (blueprint §16.2)
- CAPTCHA — Redis IP rate-limit stub only
- Customer self-registration (Sprint 19)

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `TENANT_REGISTER_ENABLED` | `true` | When `false`, `POST /api/public/tenant/register` returns `503` |
| `TENANT_REGISTER_RATE_LIMIT` | `5` | Max registrations per client IP per hour (Redis stub) |
| `TENANT_WEB_DIR` | `apps/tenant-web/build` | Static tenant portal served at `/tenant/` |
| `AUTH_DISABLED` | `true` | Does **not** disable registration API; only affects demo bypass on other routes |
| `JWT_SECRET` | — | Required when issuing tokens (`AUTH_DISABLED=false` or registration always uses auth service when configured) |

## 4. Data model (Postgres `callcenter`)

### 4.1 `tenants.status` migration

Extend check constraint:

| Status | Meaning |
| --- | --- |
| `pending_kyc` | Registered; awaiting platform KYC (Sprint 7) |
| `active` | Fully onboarded; commerce + production use |
| `suspended` | Blocked by operator |

Existing seeds (`demo`) remain `active`.

### 4.2 `tenant_registrations`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `reg_{hex}` |
| `tenant_id` | text FK | → `tenants.id` ON DELETE CASCADE |
| `company_name` | text | As submitted |
| `admin_email` | text | Denormalized from `users.email` at signup |
| `status` | text | `submitted` (Sprint 6 only; Sprint 7 adds `approved`/`rejected`) |
| audit cols | | `created_by` = `'system'` for public signup |

**Index:** `(tenant_id)` unique — one registration row per tenant in Sprint 6.

### 4.3 `brands` (stub)

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `brand_{tenant_id}` |
| `tenant_id` | text FK | → `tenants.id` ON DELETE CASCADE |
| `name` | text | Initial = `company_name` |
| `status` | text | `active` |
| audit cols | | |

**Constraint:** one default brand per tenant in Sprint 6 (unique on `tenant_id`).

### 4.4 ID conventions

| Entity | Pattern | Example |
| --- | --- | --- |
| Tenant | slug = id | `acme` |
| User | `usr_{slug}_admin` | `usr_acme_admin` |
| Brand | `brand_{tenant_id}` | `brand_acme` |
| Registration | `reg_{16hex}` | `reg_a1b2c3d4e5f67890` |

## 5. API summary

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| `POST` | `/api/public/tenant/register` | Public | Create tenant + admin + tokens |
| `GET` | `/api/platform/tenants` | `platform_admin` | List tenants; filter `?status=` |
| `GET` | `/api/auth/me` | Bearer | Unchanged; reflects new `tenant_id` |

Full contract: [04-api-spec.md](04-api-spec.md) § Tenant registration.

## 6. Validation rules

| Field | Rule |
| --- | --- |
| `company_name` | 2–120 chars, trimmed |
| `slug` | `^[a-z0-9]([a-z0-9-]{1,30}[a-z0-9])?$`; reserved: `demo`, `admin`, `platform`, `api`, `www` → `400` |
| `admin_email` | RFC5322-ish regex; lowercased; unique in `users` |
| `admin_password` | min 8 chars |
| `admin_display_name` | 1–80 chars, trimmed |

Collisions: `409` with `{ "error": "slug already taken" }` or `{ "error": "email already registered" }`.

## 7. RBAC — `pending_kyc` tenant access

| Route class | `pending_kyc` tenant_admin |
| --- | --- |
| `POST /api/public/tenant/register` | N/A (creates tenant) |
| `POST /api/auth/login`, `GET /api/auth/me` | ✅ Allowed |
| `GET /api/workforce` | ✅ Allowed (likely empty assignments until platform assigns avatars) |
| `POST /api/chat`, `GET /ws/voice` | ✅ Allowed (demo-style usage; no entitlement gate yet) |
| `POST /api/km/*` write | ❌ `403` — tenant not active |
| `GET /api/platform/*` | ❌ `403` |
| `POST /api/platform/tenants/*/entitlement` | ❌ `403` — Sprint 9 purchase |

Platform admin assigns avatars/entitlements to `pending_kyc` tenants in Sprint 6 (existing platform APIs unchanged).

## 8. Redis (rate limit stub)

| Key | TTL | Value |
| --- | --- | --- |
| `monti_jarvis:register:ip:{client_ip}` | 1h | increment counter |

When counter > `TENANT_REGISTER_RATE_LIMIT` → `429 Too Many Requests`. If Redis unavailable, log warning and allow (dev-friendly).

## 9. NATS event (optional)

Subject: `monti.auth.tenant.registered` (or reuse auth stream)

```json
{
  "event": "tenant.registered",
  "tenant_id": "acme",
  "registration_id": "reg_...",
  "admin_user_id": "usr_acme_admin",
  "status": "pending_kyc"
}
```

Publish when bus enabled; no-op otherwise.

## 10. Verification

```bash
# Register
curl -s -X POST http://localhost:8091/api/public/tenant/register \
  -H 'Content-Type: application/json' \
  -d '{
    "company_name": "Acme Corp",
    "slug": "acme",
    "admin_email": "admin@acme.test",
    "admin_password": "secret1234",
    "admin_display_name": "Acme Admin"
  }' | jq .

# Me
curl -s -H "Authorization: Bearer $ACCESS" http://localhost:8091/api/auth/me | jq .

# Platform queue
PLATFORM=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)
curl -s -H "Authorization: Bearer $PLATFORM" \
  'http://localhost:8091/api/platform/tenants?status=pending_kyc' | jq .
```

See [02-workflow.md](02-workflow.md) §18–20 · [05-ux-ui.md](05-ux-ui.md) § T1–T3 · [03-er-diagram.md](03-er-diagram.md).