---
id: DES-0016
title: Quota and Rate Limit Specification
status: shipped
updated: 2026-07-11
sprint: SPRINT-013
release: v1.4.0
owner: SA
---

# Quota & Rate Limit — Design Spec

**Sprint:** SPRINT-013 · **Release target:** v1.4.0  
**Feature:** [FEAT-0013](../01-features/FEAT-0013-quota-rate-limit.md)  
**Depends on:** [08-packages-spec.md](08-packages-spec.md), [06-auth-spec.md](06-auth-spec.md)  
**Tasks:** TASK-0057 … TASK-0061

## 1. Goals

- Enforce package `rules` as **hard quotas** on live product paths.
- Apply **rate limits** to protect Gemini/upstream from burst abuse.
- Expose **usage vs limits** to platform admins (read-only).
- Keep customer demo (`AUTH_DISABLED`) usable via fail-open + `demo` tenant resolution.

## 2. Non-goals (Sprint 13)

- Tenant self-service quota UI (→ SPRINT-016)
- Monetary overage billing / auto-upgrade
- ClickHouse usage warehouse or cost dashboards
- Login brute-force rate limit (auth Phase F)
- NATS usage events
- Per-end-customer quotas (S19+)
- New `rules-v2` schema keys (use existing rules-v1 only)

## 3. Concepts

| Term | Meaning |
| --- | --- |
| **Limit** | Ceiling from effective entitlement `rules` (jsonb) |
| **Usage** | Current consumption (Redis counter and/or Postgres count) |
| **Quota check** | Deny when `usage >= limit` for numeric dims (strict ceiling) |
| **Rate limit** | Max requests per wall-clock minute window (not monthly package cap) |
| **Fail-open** | On Redis/transport error, allow request and log warning |

### Comparison semantics

| Dimension | Deny when |
| --- | --- |
| Numeric max_* | `usage >= limit` (cannot add one more unit) |
| Boolean `*_enabled` | rule is `false` |
| Rate limit | `count > per_min` after INCR |

`limit == 0` means **hard block** for that dimension (no capacity).

### Dimensions (rules-v1)

| Rule key | Enforcement point | Usage source |
| --- | --- | --- |
| `max_ai_employees` | `POST /api/platform/tenants/{id}/avatars` | `COUNT` active `tenant_avatar_assignments` |
| `max_monthly_call_minutes` | Voice start (pre-check) + end (add elapsed) | Redis `minutes:{YYYYMM}` UTC |
| `max_km_documents` | KM document ingest | Postgres count of tenant KM docs |
| `max_concurrent_calls` | Voice WS / session open | Redis INCR/DECR concurrent |
| `voice_enabled` | Voice path | boolean |
| `rag_enabled` | Chat when RAG would run | boolean |

## 4. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `QUOTA_ENABLED` | `true` | Master quota switch |
| `QUOTA_FAIL_OPEN` | `true` | On Redis error → allow (dev-friendly) |
| `RATE_LIMIT_ENABLED` | `true` | Master rate-limit switch |
| `RATE_LIMIT_CHAT_PER_MIN` | `60` | Per-tenant chat window |
| `RATE_LIMIT_KM_PER_MIN` | `30` | Per-tenant KM write window |
| `RATE_LIMIT_VOICE_PER_MIN` | `20` | Per-tenant voice open attempts |
| `QUOTA_CONCURRENT_TTL` | `2h` | Safety TTL if WS dies without release |
| `REDIS_PREFIX` | `monti_jarvis:` | Existing prefix |

Document in `infra/.env.dev.example` + LOCAL-DEV (TASK-0057).

## 5. Data model

### Postgres (`callcenter`)

**MVP: no new required tables.** Usage from:

| Source | What |
| --- | --- |
| `tenant_entitlements` + `package_limits` | Limits via `internal/entitlements` |
| KM document tables (existing) | `max_km_documents` usage |
| `tenant_avatar_assignments` | `max_ai_employees` usage |

**Out of MVP (do not implement unless TASK-0057 stretch explicitly chosen):** `quota_usage_events` audit log. Prefer Redis-only counters for S13.

### Redis keys

```text
{prefix}quota:{tenant_id}:concurrent              # INT · EXPIRE QUOTA_CONCURRENT_TTL refreshed on acquire
{prefix}quota:{tenant_id}:minutes:{YYYYMM}        # INT · no short TTL (month natural key)
{prefix}rl:{tenant_id}:chat:{YYYYMMDDHHMM}        # INT · EXPIRE 2m
{prefix}rl:{tenant_id}:km:{YYYYMMDDHHMM}          # INT · EXPIRE 2m
{prefix}rl:{tenant_id}:voice:{YYYYMMDDHHMM}       # INT · EXPIRE 2m
{prefix}entitlement:{tenant_id}                   # existing S4 cache
```

`YYYYMM` / minute buckets use **UTC**.

## 6. Go design

```text
internal/quota/
  service.go       New(entitlements, redis, store, cfg)
  errors.go        ErrLimitExceeded, ErrRateLimited, ErrFeatureDisabled
  types.go         Snapshot, Limits, Usage

internal/entitlements/   GetEffective (unchanged contract)
cmd/server/              thin hooks in chat, voice, km, avatar handlers
apps/platform-admin-web  usage panel + lib/api/usage.ts
```

### Service API (contract for TASK-0058)

```go
// Snapshot returns limits + usage for platform UI.
Snapshot(ctx, tenantID) (*Snapshot, error)

// Rate limit bucket: "chat" | "km" | "voice"
AllowRate(ctx, tenantID, bucket string) error

CheckFeature(ctx, tenantID, flag string) error // voice_enabled | rag_enabled
CheckKMDocument(ctx, tenantID) error
CheckAIEmployees(ctx, tenantID, nextCount int) error
CheckMonthlyMinutes(ctx, tenantID, additional int) error

// AcquireConcurrent returns release func; caller MUST defer release on success path.
AcquireConcurrent(ctx, tenantID) (release func(), err error)
AddCallMinutes(ctx, tenantID, minutes int) error
```

### Errors → HTTP

| error | HTTP | `code` |
| --- | ---: | --- |
| `ErrLimitExceeded` | 429 | `quota_exceeded` |
| `ErrRateLimited` | 429 | `rate_limited` |
| `ErrFeatureDisabled` | 403 | `feature_disabled` |
| no entitlement + fail-closed | 404/403 | `no_entitlement` |
| Redis error + fail-open | — | allow + log |

JSON body:

```json
{
  "error": "KM document limit exceeded",
  "code": "quota_exceeded",
  "dimension": "max_km_documents",
  "limit": 50,
  "usage": 50
}
```

`rate_limited` responses SHOULD set `Retry-After: <seconds>` (seconds left in current minute window, or `60`).

## 7. Handler integration (TASK-0059)

| Handler | Order of checks |
| --- | --- |
| `POST /api/chat` | `AllowRate(chat)` → resolve tenant → if RAG: `CheckFeature(rag_enabled)` → chat |
| `GET /ws/voice` | `AllowRate(voice)` → `CheckFeature(voice_enabled)` → `CheckMonthlyMinutes(0)` → `AcquireConcurrent` → on close: `release` + `AddCallMinutes` |
| `POST /api/km/agents/{id}/documents` | `AllowRate(km)` → `CheckKMDocument` → ingest |
| `POST /api/platform/tenants/{id}/avatars` | `CheckAIEmployees(count+1)` → assign |

### Tenant resolution

| Mode | Tenant id for quota |
| --- | --- |
| `AUTH_DISABLED=true` | `DEMO_TENANT_ID` (default `demo`) |
| Authenticated request with tenant claim | JWT `tenant_id` |
| Platform avatar assign | path `{tenant_id}` |
| KM with Bearer | existing guard tenant resolution |

If no active entitlement:

| `QUOTA_FAIL_OPEN` | Behavior |
| --- | --- |
| `true` | Allow live paths; Snapshot returns `status: none`, empty limits |
| `false` | Deny live paths with `no_entitlement` |

## 8. API summary

See [04-api-spec.md](04-api-spec.md) § Quota & rate limit.

| Method | Path | Role |
| --- | --- | --- |
| `GET` | `/api/platform/tenants/{tenant_id}/usage` | `platform_admin` |
| (side effect) | existing chat / voice / km / avatar routes | enforced |

`GET /api/entitlements/me` unchanged (no usage fields required this sprint).

## 9. RBAC

| Action | `platform_admin` | `tenant_admin` | public / AUTH_DISABLED |
| --- | --- | --- | --- |
| GET any tenant usage | yes | no | no |
| Live path enforcement | by target tenant | own tenant | `demo` tenant |

## 10. Platform UI (TASK-0060)

- Prefer **Usage** section on existing tenant detail route if present; else `/admin/tenants/[id]` panel.
- Read-only; Refresh button re-GETs usage.
- Empty entitlement: show “No active package” (not blank crash).
- Labels EN (TH optional secondary).

Wireframe: [05-ux-ui.md](05-ux-ui.md) § P14.

## 11. Infra

`GET /api/infra`:

| Field | Values |
| --- | --- |
| `quota` | `ok` \| `disabled` \| `degraded` |
| `rate_limit` | `ok` \| `disabled` \| `degraded` |

`disabled` when env off; `degraded` when Redis ping fails but fail-open still serves traffic.

## 12. Verification

```bash
make test && go test ./internal/quota/...
make infra-init && make restart

TOKEN=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'content-type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)

curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8091/api/platform/tenants/demo/usage | jq .

curl -s http://localhost:8091/api/infra | jq '.quota, .rate_limit'

# Exhaust KM on Starter → expect 429 quota_exceeded
# Open concurrent voice = limit, next → 429
```

Manual: `docs/sdlc/06-manual-tests/SPRINT-013-manual.md` (TASK-0061).

## 13. Related artifacts

| Artifact | Path | Status |
| --- | --- | --- |
| Workflow | [02-workflow.md](02-workflow.md) §32–36 | approved |
| ER / Redis | [03-er-diagram.md](03-er-diagram.md) | approved |
| API | [04-api-spec.md](04-api-spec.md) § Quota | approved |
| UX | [05-ux-ui.md](05-ux-ui.md) § P14 | approved |
| Packages | [08-packages-spec.md](08-packages-spec.md) | shipped |
| Sprint | [SPRINT-013](../03-sprints/SPRINT-013.md) | in_progress |

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM / SA | sprint-tech-specs | 2026-07-11 | ☑ |
| Dev | | | ☐ |
