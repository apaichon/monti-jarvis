---
id: DES-0008
title: Packages and Entitlements Specification
status: shipped
updated: 2026-07-08
sprint: SPRINT-004
---

# Packages and Entitlements — Design Spec

**Sprint:** SPRINT-004 · **Release target:** v0.5.0  
**Depends on:** [06-auth-spec.md](06-auth-spec.md) (JWT + RBAC)

## 1. Goals

- Platform operators define **commercial packages** (catalog) with numeric limits.
- Each tenant has at most one **active entitlement** pointing at a package.
- Downstream features (Sprint 9 purchase, Sprint 13 quotas) read limits via `internal/entitlements`.

## 2. Data model (Postgres `callcenter`)

Limits and feature flags live in **`jsonb`** so new quota keys ship without DDL migrations. A **schema control table** versions the allowed JSON shape.

### `package_rule_schemas` (JSONB structure control)

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | e.g. `rules-v1`, `rules-v2` |
| `version` | int UK | Monotonic schema generation (1, 2, 3…) |
| `name` | text | Human label, e.g. `Sprint 4 base limits` |
| `fields` | jsonb | Field catalog: type, min/max, required, default, description per key |
| `status` | text | `active`, `deprecated` — only one `active` for new packages (enforced in app) |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

**`fields` example (`rules-v1`):**

```json
{
  "max_ai_employees":       { "type": "int",  "min": 0, "required": true,  "description": "Max AI avatars" },
  "max_monthly_call_minutes": { "type": "int",  "min": 0, "required": true,  "description": "Monthly voice minutes" },
  "max_km_documents":       { "type": "int",  "min": 0, "required": true,  "description": "KM documents per tenant" },
  "max_concurrent_calls":   { "type": "int",  "min": 0, "required": true,  "description": "Parallel calls" },
  "voice_enabled":          { "type": "bool", "required": true,  "default": true },
  "rag_enabled":            { "type": "bool", "required": true,  "default": true }
}
```

Future `rules-v2` can add keys (e.g. `max_tokens_per_month`) without altering `package_limits` columns — packages created after v2 ship reference `rules-v2`; existing packages keep their schema id until migrated.

### `packages`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | e.g. `pkg-starter` |
| `slug` | text UK | `starter`, `pro`, `enterprise` |
| `name` | text | Display name |
| `description` | text | Optional |
| `status` | text | `draft`, `active`, `archived` |
| `price_cents` | int | Display/list price; billing in Sprint 10 |
| `currency` | text | Default `USD` |
| `billing_period` | text | `monthly`, `annual`, `one_time` |
| audit cols | | |

### `package_limits`

One row per package (1:1). **Values** in `rules` jsonb; **shape** governed by `package_rule_schemas`.

| Column | Type | Notes |
| --- | --- | --- |
| `package_id` | text PK FK | → `packages.id` |
| `rules_schema_id` | text FK | → `package_rule_schemas.id` |
| `rules` | jsonb | Limit values validated against `fields` for referenced schema |
| audit cols | | |

**`rules` example (Starter, `rules-v1`):**

```json
{
  "max_ai_employees": 2,
  "max_monthly_call_minutes": 500,
  "max_km_documents": 50,
  "max_concurrent_calls": 2,
  "voice_enabled": true,
  "rag_enabled": true
}
```

**Index:** `CREATE INDEX … ON package_limits USING gin (rules)` optional for analytics; resolver reads by `package_id`.

### `tenant_entitlements`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | |
| `tenant_id` | text FK | → `tenants.id` |
| `package_id` | text FK | → `packages.id` |
| `rules_schema_id` | text FK | Schema version **at assignment** (audit / billing) |
| `rules_snapshot` | jsonb | Copy of effective `rules` at assign time; resolver prefers live package unless entitlement suspended |
| `status` | text | `active`, `suspended`, `revoked`, `expired` |
| `valid_from` | timestamptz | Default `now()` |
| `valid_until` | timestamptz | Nullable until Sprint 9 subscriptions |
| audit cols | | |

**Constraint:** partial unique index on `(tenant_id) WHERE status = 'active'` — one active entitlement per tenant.

**Validation:** `internal/packages` loads `package_rule_schemas.fields` for `rules_schema_id`, validates `rules` on create/update. Unknown keys rejected unless schema defines `additionalProperties: true` (default: false).

### Provider catalog stub (Sprint 3–4 carry-over)

Seed rows in `embedding_models` (`emb-gemini-001`) and `voice_providers` (`voice-gemini-live`) if tables exist; idempotent `ON CONFLICT DO NOTHING`.

## 3. Redis cache

| Key | TTL | Payload |
| --- | --- | --- |
| `monti_jarvis:entitlement:{tenant_id}` | `ENTITLEMENT_CACHE_TTL` (default 15m) | JSON: `rules`, `rules_schema_id`, package slug, status |

Invalidate on assign, revoke, suspend, package update affecting active tenants.

Env: `ENTITLEMENT_CACHE_ENABLED` (default `true` when Redis configured).

## 4. API (auth required unless noted)

Base: `http://localhost:8091`. All routes below require `AUTH_DISABLED=false` and Bearer token.

### Rule schemas — `platform_admin` only

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/platform/rule-schemas` | List `package_rule_schemas` (`?status=active`) |

### Package catalog — `platform_admin` only

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/platform/packages` | List packages (`?status=active`) |
| `POST` | `/api/platform/packages` | Create package + limits |
| `GET` | `/api/platform/packages/{id}` | Get package + limits |
| `PUT` | `/api/platform/packages/{id}` | Update metadata/limits |
| `DELETE` | `/api/platform/packages/{id}` | Archive (`status=archived`) |

### Tenant entitlements

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `GET` | `/api/platform/tenants/{tenant_id}/entitlement` | `platform_admin` | Current entitlement + effective limits |
| `POST` | `/api/platform/tenants/{tenant_id}/entitlement` | `platform_admin` | Assign package (`package_id`) |
| `DELETE` | `/api/platform/tenants/{tenant_id}/entitlement` | `platform_admin` | Revoke active entitlement |
| `GET` | `/api/entitlements/me` | `tenant_admin`, `platform_admin` | Entitlement for token tenant |

### Response shape (effective entitlement)

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

## 5. Go packages

```text
internal/packages      Catalog CRUD, rules JSONB validation against package_rule_schemas
internal/entitlements  Resolve effective rules, Redis cache
internal/store         SQL for packages, package_rule_schemas, package_limits, tenant_entitlements
cmd/server/packages.go HTTP handlers
```

## 6. RBAC

| Action | `platform_admin` | `tenant_admin` | `customer` |
| --- | --- | --- | --- |
| Package CRUD | yes | no | no |
| Assign/revoke entitlement | yes | no | no |
| Read own entitlement | yes | yes | no |
| Read any tenant entitlement | yes | no | no |

## 7. Dev defaults (seed)

**Schema:** `rules-v1` (`package_rule_schemas`, `status=active`).

| Package | slug | `rules` (v1) |
| --- | --- | --- |
| Starter | `starter` | `max_ai_employees: 2`, `max_monthly_call_minutes: 500`, `max_km_documents: 50`, `max_concurrent_calls: 2`, voice+rags on |
| Pro | `pro` | `10` / `5000` / `500` / `10` |
| Enterprise | `enterprise` | `50` / `50000` / `5000` / `50` |

`demo` tenant → `pkg-starter` (active), `rules_schema_id=rules-v1`, `rules_snapshot` copied from package.

## 8. Platform admin UI

Package and entitlement management UI lives in [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md) (`/admin/packages`, `/admin/tenants/{id}/entitlement`). This spec defines APIs and data; portal consumes them.

## 9. Out of scope (later sprints)

- Enforcement middleware on `/api/chat`, `/ws/voice`, KM ingest (Sprint 13)
- Proration, trials, coupons (Sprint 9–10)
- ClickHouse `entitlement_events` (optional analytics)

## 10. Verification

```bash
make infra-init && make restart
TOKEN=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'content-type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/platform/packages
curl -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/platform/tenants/demo/entitlement
```

## 11. Related design artifacts

| Artifact | Path |
| --- | --- |
| Workflows | [02-workflow.md](02-workflow.md) §9–11 |
| ER diagram | [03-er-diagram.md](03-er-diagram.md) — `package_rule_schemas`, `packages`, `package_limits`, `tenant_entitlements` |
| API contract | [04-api-spec.md](04-api-spec.md) § Packages & entitlements |
| Platform portal | [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md) |
| UX ASCII | [05-ux-ui.md](05-ux-ui.md) § Platform Admin Portal |
| Auth dependency | [06-auth-spec.md](06-auth-spec.md) |