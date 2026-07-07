---
id: DES-0006
title: Auth & RBAC Specification
status: shipped
updated: 2026-07-07
sprint: SPRINT-003
owner: SA
---

# Auth & RBAC Specification — Sprint 3

> **Status: SHIPPED** — v0.4.0 (SPRINT-003)

**Feature:** [FEAT-0003](../01-features/FEAT-0003-auth-rbac.md)  
**Sprint:** [SPRINT-003](../03-sprints/SPRINT-003.md)

## 1. Goals

1. Replace trusted `X-Tenant-Id` header with **JWT-derived tenant** on protected routes when auth is enabled.
2. Introduce **three roles** aligned with blueprint: `platform_admin`, `tenant_admin`, `customer` (stub).
3. Keep **Sprint 1–2 customer demo** working via `AUTH_DISABLED=true` (default in dev).

## 2. Non-goals (Sprint 3)

- Customer login UI, registration, password reset email
- OAuth / SAML / MFA
- Admin Svelte apps
- Asymmetric JWT (RS256) — HS256 dev only
- Per-NATS-subject ACL (enforce in Go handlers)

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `AUTH_DISABLED` | `true` | When `true`, all routes behave as v0.3.0 (demo tenant) |
| `JWT_SECRET` | *(required when auth on)* | HS256 signing secret, ≥32 bytes |
| `JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime (7 days) |
| `DEMO_TENANT_ID` | `demo` | Tenant used when `AUTH_DISABLED=true` |

## 4. Data model (Postgres `callcenter`)

### 4.1 Tables

```sql
CREATE TABLE callcenter.tenants (
  id          text PRIMARY KEY,
  slug        text NOT NULL UNIQUE,
  name        text NOT NULL,
  status      text NOT NULL DEFAULT 'active'
              CHECK (status IN ('active', 'suspended')),
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE callcenter.users (
  id            text PRIMARY KEY,
  email         text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  display_name  text NOT NULL DEFAULT '',
  status        text NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'disabled')),
  created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE callcenter.user_roles (
  user_id    text NOT NULL REFERENCES callcenter.users(id) ON DELETE CASCADE,
  role       text NOT NULL CHECK (role IN ('platform_admin', 'tenant_admin', 'customer')),
  tenant_id  text REFERENCES callcenter.tenants(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, role, tenant_id)
);

CREATE TABLE callcenter.refresh_tokens (
  id          text PRIMARY KEY,
  user_id     text NOT NULL REFERENCES callcenter.users(id) ON DELETE CASCADE,
  token_hash  text NOT NULL UNIQUE,
  expires_at  timestamptz NOT NULL,
  revoked_at  timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX refresh_tokens_user_idx ON callcenter.refresh_tokens (user_id);
CREATE INDEX refresh_tokens_expires_idx ON callcenter.refresh_tokens (expires_at);
```

### 4.2 Role rules

| Role | `tenant_id` in `user_roles` | Scope |
| --- | --- | --- |
| `platform_admin` | `NULL` | Any tenant; cross-tenant KM ops allowed |
| `tenant_admin` | required | Own tenant only |
| `customer` | required | Own tenant; **no KM write** in Sprint 3 |

Constraint (app-enforced): `platform_admin` rows use `tenant_id IS NULL`; other roles require non-null `tenant_id`.

### 4.3 Dev seed (idempotent)

| Email | Password (dev) | Role | Tenant |
| --- | --- | --- | --- |
| `platform@monti.local` | `monti-platform` | `platform_admin` | — |
| `admin@demo.local` | `demo-admin` | `tenant_admin` | `demo` |

Passwords are **dev-only**; document in `LOCAL-DEV.md`, never use in production.

## 5. JWT design

### 5.1 Access token claims

```json
{
  "sub": "usr_abc123",
  "email": "admin@demo.local",
  "role": "tenant_admin",
  "tenant_id": "demo",
  "iat": 1710000000,
  "exp": 1710000900,
  "typ": "access"
}
```

- `platform_admin`: `tenant_id` may be empty; handlers accept optional `X-Tenant-Id` **only for platform_admin** to select target tenant on write ops.
- `tenant_admin` / `customer`: `tenant_id` claim is authoritative; `X-Tenant-Id` **ignored** when `AUTH_DISABLED=false`.

### 5.2 Refresh token

- Opaque random 32-byte value, sent to client once.
- Stored as SHA-256 hash in `refresh_tokens`.
- Rotation: each refresh issues new refresh token and revokes previous hash.
- Logout sets `revoked_at` on matching row.

### 5.3 Algorithms

- Sign: **HS256** with `JWT_SECRET`
- Password: **bcrypt** cost 12

## 6. HTTP API

Base: `http://localhost:8091`. See [api-spec.md](api-spec.md) § Auth for full contract.

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| `POST` | `/api/auth/login` | none | Email + password → tokens |
| `POST` | `/api/auth/refresh` | none | Refresh body → new tokens |
| `POST` | `/api/auth/logout` | Bearer or refresh body | Revoke refresh |
| `GET` | `/api/auth/me` | Bearer | Current user profile |

### 6.1 Login request/response

```json
// POST /api/auth/login
{ "email": "admin@demo.local", "password": "demo-admin" }

// 200
{
  "access_token": "eyJ...",
  "refresh_token": "rt_...",
  "expires_in": 900,
  "token_type": "Bearer",
  "user": {
    "id": "usr_...",
    "email": "admin@demo.local",
    "display_name": "Demo Admin",
    "role": "tenant_admin",
    "tenant_id": "demo"
  }
}
```

### 6.2 Errors

| Code | Body | When |
| --- | --- | --- |
| `401` | `{"error":"invalid credentials"}` | Bad email/password |
| `401` | `{"error":"unauthorized"}` | Missing/invalid/expired access token |
| `403` | `{"error":"forbidden"}` | Valid token, insufficient role |
| `503` | `{"error":"auth is not configured"}` | `AUTH_DISABLED=false` but no `JWT_SECRET` |

## 7. Route policy matrix

When `AUTH_DISABLED=true`: all routes public; tenant = `DEMO_TENANT_ID`; optional `X-Tenant-Id` for KM reads.

When `AUTH_DISABLED=false`:

| Route | Method | Policy |
| --- | --- | --- |
| `/healthz`, `/api/infra` | GET | public |
| `/api/workforce` | GET | public |
| `/api/chat` | POST | public (customer inbound demo) |
| `/ws/voice` | GET | public |
| `/api/calls`, `/api/calls/*` | * | public (Sprint 3); `tenant_id` from token if present else demo |
| `/api/km/agents/*` | GET | public read |
| `/api/km/agents/*/documents` | POST | `tenant_admin` \| `platform_admin` |
| `/api/km/agents/*/reset` | POST | `tenant_admin` \| `platform_admin` |
| `/api/km/seed` | POST | `platform_admin` only |
| `/api/auth/*` | * | per endpoint above |

**Future (Sprint 4+):** tighten `/api/calls` and KM reads to authenticated customer/tenant roles.

## 8. Middleware (`internal/auth`)

```text
Request
  → CORS / common headers
  → AuthMiddleware (optional skip if AUTH_DISABLED)
       → parse Bearer
       → validate JWT → AuthContext{UserID, Email, Role, TenantID}
       → attach to context.Context
  → RBAC wrapper on protected handlers
  → handler
```

```go
type AuthContext struct {
    UserID   string
    Email    string
    Role     string
    TenantID string // resolved effective tenant for this request
}

func ResolveTenant(ctx context.Context, headerTenant string) string
// AUTH_DISABLED → DEMO_TENANT_ID or X-Tenant-Id
// platform_admin → headerTenant if set else claim
// tenant_admin/customer → claim only
```

## 9. Package layout

```text
internal/auth/
  jwt.go          issue, parse, validate
  password.go     bcrypt hash/verify
  middleware.go   HTTP middleware + RBAC helpers
  context.go      AuthContext on context
  refresh.go      refresh token CRUD
internal/store/
  auth.go         users, tenants, roles, refresh_tokens queries
cmd/server/
  auth.go         /api/auth/* handlers
  main.go         wire middleware; wrap km routes
```

## 10. Security notes

- Generic login failure message (no email enumeration).
- Rate-limit login (future Sprint 13); log failures in Sprint 3.
- Refresh tokens never stored plaintext.
- CORS: add `Authorization` to allowed headers when auth enabled.
- KM/ClickHouse queries always filter by resolved `tenant_id`.

## 11. Migration & rollback

- `infra-init.sh` adds auth DDL + seed after existing tables.
- Rollback: set `AUTH_DISABLED=true` — no code path change for customer demo.
- No drop of existing KM/call data.

## 12. Test plan (design)

| Layer | Tests |
| --- | --- |
| Unit | JWT round-trip, bcrypt, RBAC `RequireRole`, `ResolveTenant` |
| HTTP | `httptest` login 200/401, protected KM 401/403/201 |
| Manual | `SPRINT-003-manual.md` at VERIFY |
| Regression | `AUTH_DISABLED=true` full v0.3.0 smoke |

## 13. Open questions (for review)

1. **`/api/km/seed`**: restrict to `platform_admin` only — OK?
2. **`POST /api/calls`**: keep public for customer demo when auth on — OK?
3. **Refresh storage**: Postgres only (proposed) vs Redis — prefer Postgres for audit?
4. **Platform admin tenant override**: via `X-Tenant-Id` on write ops — OK?

---

**Approver sign-off**

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | | ☐ |
| Dev | | | ☐ |
| DevOps | | | ☐ |