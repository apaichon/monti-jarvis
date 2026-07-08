---
id: DES-0010
title: Avatar Catalog and Tenant Assignment Specification
status: approved
updated: 2026-07-08
sprint: SPRINT-005
owner: SA
---

# Avatar Catalog and Tenant Assignment — Design Spec

**Sprint:** SPRINT-005 · **Release target:** v0.6.0  
**Feature:** [FEAT-0005](../01-features/FEAT-0005-avatar-catalog.md)  
**Depends on:** [06-auth-spec.md](06-auth-spec.md), [08-packages-spec.md](08-packages-spec.md)

## 1. Goals

- Platform operators manage the **AI avatar catalog** in Postgres (migrate Ava, Max, Luna, Neo from `internal/workforce/workforce.go`).
- **Assign avatars to tenants**; enforce `max_ai_employees` from active entitlement.
- **`GET /api/workforce`** returns tenant-specific active avatars (DB-first, static fallback).
- Platform admin portal screens at `/admin/avatars` and `/admin/tenants/{id}/avatars`.

## 2. Non-goals (Sprint 5)

- Blueprint `ai_employee_versions`, languages, tools, guardrails (Sprint 21).
- MinIO upload pipeline for avatar images (`image_url` text field only).
- Customer portal Svelte changes (still consumes `/api/workforce`).
- Tenant self-service avatar admin (Sprint 15+).
- Runtime voice-provider binding per avatar (`ai_employee_configs` — Sprint 21).

## 3. Data model (Postgres `callcenter`)

### `ai_avatars`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | Stable id, e.g. `ava`, `max` (matches `agent_id` in KM/calls) |
| `slug` | text UK | Lowercase unique slug (same as id for seeds) |
| `name` | text | Display name |
| `role` | text | e.g. `General Support` |
| `trait` | text | Personality label for system prompt |
| `color` | text | Hex accent `#008cff` |
| `voice` | text | Gemini Live voice name `Aoede`, `Charon`, … |
| `image_url` | text | Public path `/images/ava.jpg` or absolute URL |
| `greeting` | text | Opening line for voice/chat |
| `status` | text | `draft`, `active`, `archived` |
| `flags` | jsonb | Optional UI fields: `popular`, `robot`, `skin`, `hair` |
| audit cols | | |

**`flags` example (Ava):**

```json
{ "popular": true, "skin": "#f0bd9b", "hair": "#5a3428" }
```

### `tenant_avatar_assignments`

| Column | Type | Notes |
| --- | --- | --- |
| `tenant_id` | text FK | → `tenants.id` |
| `avatar_id` | text FK | → `ai_avatars.id` |
| `status` | text | `active`, `disabled` |
| audit cols | | |

**Constraint:** `PRIMARY KEY (tenant_id, avatar_id)` — one row per pair; assign upserts to `active`.

**Cap rule:** count rows where `status = 'active'` for tenant ≤ `entitlements.GetEffective(tenant).rules.max_ai_employees`. `409` when at cap.

### Dev seeds

| Avatar | id | voice | image_url |
| --- | --- | --- | --- |
| Ava | `ava` | Aoede | `/images/ava.jpg` |
| Max | `max` | Charon | `/images/max.jpg` |
| Luna | `luna` | Kore | `/images/luna.jpg` |
| Neo | `neo` | Puck | `/images/neo.jpg` |

Assign all four to tenant `demo` (`active`).

DDL ships in `internal/store/avatars.go` `ensureAvatarsSchema` (same pattern as packages).

## 4. API summary

All `/api/platform/avatars*` routes: `platform_admin` + Bearer. See [04-api-spec.md](04-api-spec.md) § Avatars.

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/api/platform/avatars` | List catalog (`?status=`) |
| POST | `/api/platform/avatars` | Create avatar |
| GET | `/api/platform/avatars/{id}` | Get one |
| PUT | `/api/platform/avatars/{id}` | Update |
| DELETE | `/api/platform/avatars/{id}` | Soft-archive |
| GET | `/api/platform/tenants/{tenant_id}/avatars` | List assignments + avatar metadata |
| POST | `/api/platform/tenants/{tenant_id}/avatars` | Assign `{avatar_id}` |
| DELETE | `/api/platform/tenants/{tenant_id}/avatars/{avatar_id}` | Disable assignment |

**Public (unchanged auth):** `GET /api/workforce` — resolves tenant via `X-Tenant-Id` or auth context; returns assigned `active` avatars; if zero assignments, falls back to `workforce.All()`.

## 5. RBAC

| Route | Roles |
| --- | --- |
| `/api/platform/avatars*` | `platform_admin` |
| `/api/platform/tenants/{id}/avatars*` | `platform_admin` |
| `GET /api/workforce` | public (tenant from header/JWT/demo) |

## 6. Workforce JSON shape

Response field `image` maps from DB `image_url` for backward compatibility with customer portal.

```json
{
  "agents": [
    {
      "id": "ava",
      "name": "Ava",
      "role": "General Support",
      "trait": "Warm & Patient",
      "color": "#008cff",
      "voice": "Aoede",
      "image": "/images/ava.jpg",
      "popular": true,
      "greeting": "Thank you for calling..."
    }
  ]
}
```

`internal/workforce.SystemPrompt` continues to use resolved `Agent` struct for chat/voice.

## 7. Verification

```bash
make build && make test
make restart

TOKEN=$(curl -sS -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)

curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/platform/avatars | jq .
curl -sS -H "X-Tenant-Id: demo" http://localhost:8091/api/workforce | jq '.agents | length'
open http://localhost:8091/admin/avatars
```

## 8. Sprint 21 migration note

Sprint 5 table `ai_avatars` is the **platform catalog MVP**. Sprint 21 may introduce `ai_employees` + `ai_employee_configs` for provider bindings; migration path: rename/extend `ai_avatars` or add view — document in ER at Sprint 21 open.

See [02-workflow.md](02-workflow.md) §14–17 · [03-er-diagram.md](03-er-diagram.md) · [05-ux-ui.md](05-ux-ui.md) § P7–P10 · [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md).