---
id: DES-0010
title: Avatar Catalog and Tenant Assignment Specification
status: approved
updated: 2026-07-09
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
- Runtime **live failover** during an active call (updates `call_sessions.voice_provider_id` — Sprint 21+); Sprint 5 stores the **ordered voice profile catalog** only.

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
| `image_url` | text | Public path `/images/ava.jpg` or absolute URL |
| `greeting` | text | Opening line for voice/chat |
| `status` | text | `draft`, `active`, `archived` |
| `flags` | jsonb | Optional UI fields: `popular`, `robot`, `skin`, `hair` |
| audit cols | | |

**Voice is not a single column on `ai_avatars`.** Each avatar has **one or more rows** in `ai_avatar_voices` (provider + voice id + persona voice name), ordered by `priority` for reliability failover.

**`flags` example (Ava):**

```json
{ "popular": true, "skin": "#f0bd9b", "hair": "#5a3428" }
```

### `ai_avatar_voices`

One avatar → many voice profiles (primary + alternates). References catalog `voice_providers` from Sprint 4.

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | e.g. `avvoice_ava_gemini` |
| `avatar_id` | text FK | → `ai_avatars.id` ON DELETE CASCADE |
| `voice_provider_id` | text FK | → `voice_providers.id` (e.g. `voice-gemini-live`) |
| `voice_id` | text | Provider-specific voice/model identifier (e.g. `gemini-2.5-flash-native-audio-latest` or vendor voice key) |
| `voice` | text | Persona voice name passed to Live API — `Aoede`, `Charon`, `Kore`, `Puck` |
| `priority` | int | **Lower = preferred**; runtime failover tries next `active` row by ascending priority |
| `status` | text | `active`, `disabled` |
| audit cols | | |

**Index:** `(avatar_id, priority)` where `status = 'active'`.

**Resolver rule:** `PrimaryVoice(avatar)` = lowest `priority` among `active` rows. Voice relay / call start uses primary; on provider error, advance to next profile (Sprint 21 wires live switch; Sprint 5 persists catalog).

**Example (Ava — primary Gemini, optional alternate stub):**

| id | avatar_id | voice_provider_id | voice_id | voice | priority | status |
| --- | --- | --- | --- | --- | ---: | --- |
| `avvoice_ava_gemini` | `ava` | `voice-gemini-live` | `gemini-2.5-flash-native-audio-latest` | `Aoede` | 1 | `active` |
| `avvoice_ava_alt` | `ava` | `voice-grok-stub` | `grok-voice-stub` | `Aoede` | 2 | `disabled` |

Sprint 5 seeds **priority 1** rows only for all four avatars; alternate rows optional `disabled` placeholders for ops documentation.

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

**`ai_avatars` + `ai_avatar_voices` (priority 1, `voice-gemini-live`):**

| Avatar | id | voice | voice_id | image_url |
| --- | --- | --- | --- | --- |
| Ava | `ava` | Aoede | `gemini-2.5-flash-native-audio-latest` | `/images/ava.jpg` |
| Max | `max` | Charon | `gemini-2.5-flash-native-audio-latest` | `/images/max.jpg` |
| Luna | `luna` | Kore | `gemini-2.5-flash-native-audio-latest` | `/images/luna.jpg` |
| Neo | `neo` | Puck | `gemini-2.5-flash-native-audio-latest` | `/images/neo.jpg` |

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

Response field `image` maps from DB `image_url`. Field `voice` is the **primary** profile's `voice` (lowest `priority` active row). Customer portal unchanged.

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
      "voice_provider_id": "voice-gemini-live",
      "voice_id": "gemini-2.5-flash-native-audio-latest",
      "image": "/images/ava.jpg",
      "popular": true,
      "greeting": "Thank you for calling..."
    }
  ]
}
```

**Platform admin avatar detail** includes full `voices[]` array (all profiles). `internal/workforce.SystemPrompt` continues to use resolved `Agent` struct for chat/voice.

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

## 8. Voice failover (phased)

| Phase | Behavior |
| --- | --- |
| **Sprint 5** | Persist ordered `ai_avatar_voices`; expose primary + alternates in admin API/UI |
| **Sprint 21** | Live call path selects primary; on provider error, retry next profile and log `call_provider_events` |
| **Later** | Tenant-level override via `ai_employee_configs` (embedding + voice defaults) |

## 9. Sprint 21 migration note

Sprint 5 `ai_avatars` + `ai_avatar_voices` is the **platform catalog MVP**. Sprint 21 adds runtime failover and may extend with `ai_employee_configs` for embedding bindings — voice profiles remain on `ai_avatar_voices`.

See [02-workflow.md](02-workflow.md) §14–17 · [03-er-diagram.md](03-er-diagram.md) · [05-ux-ui.md](05-ux-ui.md) § P7–P10 · [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md).