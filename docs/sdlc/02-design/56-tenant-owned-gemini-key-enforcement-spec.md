---
id: DES-0056
title: Tenant-Owned Gemini Key Enforcement Specification
status: approved
updated: 2026-07-30
sprint: SPRINT-060
owner: SA
feature: FEAT-0052
release_target: v2.33.0
---

# DES-0056 — Tenant-Owned Gemini Key Enforcement

**Sprint:** SPRINT-060 · **Release target:** v2.33.0  
**Feature:** [FEAT-0052](../01-features/FEAT-0052-tenant-owned-gemini-key-enforcement.md)  
**Tasks:** TASK-0216–0219  
**Prior:** [DES-0039](39-tenant-ai-config-extensibility-spec.md) (encrypted key storage)

## 1. Goals

1. Production tenant AI uses only a **validated** tenant Gemini key.
2. AI Settings supports enter / test / replace / delete with masked metadata.
3. Never expose plaintext keys to browser or admin read APIs.
4. Explicit non-prod opt-in for platform env fallback.

## 2. Non-goals

- Multi-provider LLM keys
- Platform-funded shared production Gemini pool
- Billing plan redesign

## 3. Environment

| Variable | Default | Rule |
| --- | --- | --- |
| `GEMINI_API_KEY` | empty | Platform key; **not** used for production tenant runtime after S60 |
| `ALLOW_PLATFORM_GEMINI_FALLBACK` | `false` | When `true` **and** `APP_ENV` ∉ {`prod`,`production`}, allow env key for tenants without valid key |
| `TENANT_SECRET_ENCRYPTION_KEY` | required | AES-256-GCM (S43) |
| `GEMINI_KEY_TEST_MODEL` | platform text model | Lightweight models.list or generateContent probe |

## 4. Data model delta (`tenant_ai_configs`)

Existing ciphertext columns retained. Add:

| Column | Type | Notes |
| --- | --- | --- |
| `gemini_key_status` | text | `none` \| `present` \| `valid` \| `invalid` \| `degraded` |
| `gemini_key_last_validated_at` | timestamptz null | Last successful test |
| `gemini_key_last_error_class` | text null | Safe class: `auth`, `network`, `quota`, `unknown` |
| audit cols | existing | `created_at`, `updated_at`, `created_by`, `updated_by` |

Migration: `scripts/migrations/0XX_tenant_gemini_key_status.sql` (assign next number at implement).

## 5. Runtime resolution

```
resolveTenantGeminiKey(tenantID):
  cfg = load tenant_ai_configs
  if cfg has ciphertext and status == valid:
    return decrypt(cfg)
  if allowPlatformFallback():  # non-prod + flag
    return env GEMINI_API_KEY if set
  return error tenant_gemini_key_required
```

Chat and voice paths must call this resolver; do not read env key directly for
tenant-scoped requests.

## 6. API summary

| Method | Path | Role | Purpose |
| --- | --- | --- | --- |
| GET | `/api/tenant/ai/gemini-key` | tenant_admin | Metadata only: last4, status, last_validated_at |
| PUT | `/api/tenant/ai/gemini-key` | tenant_admin | Save/replace encrypted key; status → `present` until tested |
| DELETE | `/api/tenant/ai/gemini-key` | tenant_admin | Clear key + status `none` |
| POST | `/api/tenant/ai/gemini-key/test` | tenant_admin | Body optional `{ "api_key": "..." }` to test proposed; else test stored |

### POST test — success

```json
{
  "ok": true,
  "status": "valid",
  "last4": "ab12",
  "last_validated_at": "2026-07-30T12:00:00Z"
}
```

### POST test — failure

```json
{
  "ok": false,
  "status": "invalid",
  "error_class": "auth",
  "message": "Gemini rejected the key. Check the key in Google AI Studio."
}
```

### Errors

| Code | When |
| --- | --- |
| 400 | empty key / validation_error |
| 401/403 | not tenant_admin |
| 502 | network/provider unreachable (`error_class=network`) |
| 503 | `tenant_gemini_key_required` at runtime on chat/voice |

## 7. RBAC

Tenant admin of **active** tenant only. Platform admin has no plaintext read.
Audit: `tenant.ai.gemini_key.*` events.

## 8. Verification

```bash
# Missing key production-like
ALLOW_PLATFORM_GEMINI_FALLBACK=0 APP_ENV=production # chat → 503 config error

# Test endpoint
curl -sS -X POST localhost:8091/api/tenant/ai/gemini-key/test \
  -H "Authorization: Bearer $TENANT_JWT" -H 'Content-Type: application/json' \
  -d '{"api_key":"AIza..."}'
```

## See also

Workflow §132–134 · ER Sprint 60 · API Sprint 60 · UX T60
