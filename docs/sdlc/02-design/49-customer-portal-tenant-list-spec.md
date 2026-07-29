---
id: DES-0049
title: Customer Portal Tenant List to Call Specification
status: approved
updated: 2026-07-29
sprint: SPRINT-054
owner: SA
feature: FEAT-0045
release_target: v2.25.0
---

# DES-0049 — Customer Portal Tenant List to Call

**Sprint:** SPRINT-054 · **Release target:** v2.25.0  
**Feature:** [FEAT-0045](../01-features/FEAT-0045-customer-portal-tenant-list.md)  
**Tasks:** [TASK-0195](../04-tasks/TASK-0195.md), [TASK-0196](../04-tasks/TASK-0196.md),
[TASK-0197](../04-tasks/TASK-0197.md)  
**Related:** [FEAT-0018](../01-features/FEAT-0018-central-brand-call-portal.md) (full hub),
[30-mobile-call-api-sdk-spec.md](30-mobile-call-api-sdk-spec.md) (public brands),
[23-customer-auth-spec.md](23-customer-auth-spec.md)

## 1. Goals

1. Customer portal entry works **without** a required `?tenant_id=` query string.
2. Callers **pick a call-to tenant** from a public directory, then enter the desk.
3. Tenant context after pick is bound via **path** (`/t/{slug}`) and client state.
4. Directory and desk remain **tenant-isolated** (list safe; calls scoped).
5. Reuse existing `brands` public listing — do not invent a parallel catalog.

## 2. Non-goals

- Full multi-brand marketing hub (FEAT-0018 polish, paid ranking, CMS)
- Cross-tenant conversation history
- Changing embed-key resolution
- S53 auto-register / version display
- New commercial packaging

## 3. Policy

| Concern | Rule |
| --- | --- |
| Who appears in list | `tenants.status=active` AND `brands.status=active` AND `brands.listed=true` AND `brands.platform_listed=true` AND KYC approved (or no KYC row for demo) |
| Safe fields | `id`, `slug`, `name`, `blurb`, `logo_url`, `category`, `languages`, `status` |
| Forbidden fields | admin email, secrets, KYC dossier, entitlement, internal registration ids |
| Page URL | Primary: `/` picker or `/t/{slug}` desk; `?tenant_id=` optional preselect only |
| API transport | Prefer `X-Tenant-Id` header after selection; query/body allowed for WS/SSE/backcompat |
| Embed | Unchanged — still `embed_key` / public embed path |

## 4. Data model

**No new tables for MVP.** Reuse:

| Entity | Use |
| --- | --- |
| `tenants` | id, slug, status |
| `brands` | name, blurb, logo_url, category, languages, listed, platform_listed, status |
| `tenant_kyc_profiles` | filter on approved (existing public brands query) |

Optional seed/docs only: ensure local demo tenants have `listed=true` and
`platform_listed=true` so the picker is non-empty in dev.

### Redis / MinIO

| Store | Key / path | Notes |
| --- | --- | --- |
| Redis | none required | Optional short TTL cache for public list later |
| sessionStorage | `monti_jarvis:selected_tenant` | Client-only `{ id, slug, name }` |
| MinIO | brand logos via existing `logo_url` | No new prefix |

## 5. API summary

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| GET | `/api/public/brands` | none | Paginated public directory (`q`, `limit`, `offset`) |
| GET | `/api/public/brands/{slug}` | none | Resolve one listable brand by slug or tenant id |
| GET | `/api/public/tenants` | none | **Optional alias** of brands list (same payload) |
| GET | `/api/public/theme/{tenant_id}` | none | Existing published theme after pick |
| GET | `/api/customer/workforce` | optional bearer | Existing; scoped by selected tenant |

### List response (existing shape)

```json
{
  "items": [
    {
      "id": "tenant-acme",
      "slug": "acme",
      "name": "Acme Support",
      "blurb": "We help with orders",
      "logo_url": "https://…",
      "category": "retail",
      "languages": ["en", "th"],
      "listed": true,
      "status": "active"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

### Errors

| HTTP | Code / message | When |
| ---: | --- | --- |
| 404 | brand not found | slug not listable or missing |
| 502 | public brand directory unavailable | store/Postgres failure |

## 6. Client routing

```text
GET /                          → if no selected tenant → Picker
                               → if session has tenant → redirect /t/{slug}
GET /t/{slug}                  → resolve brand → desk for that tenant
GET /?tenant_id={id|slug}      → preselect → replace to /t/{slug}
GET /embed                     → unchanged (embed key)
```

Desk APIs after selection:

1. `GET /api/public/theme/{tenant_id}`
2. `GET /api/customer/workforce` with `X-Tenant-Id`
3. Chat/voice/call start with same tenant binding

## 7. RBAC

| Actor | Capability |
| --- | --- |
| Anonymous caller | List public brands; enter desk; guest chat per existing policy |
| Authenticated customer | Same list; OTP/session still tenant-scoped after pick |
| Tenant admin | `PUT /api/tenant/brand` listed flag (existing) |
| Platform admin | `PUT /api/platform/tenants/{id}/brand-listing` (existing) |

## 8. Isolation invariants

1. Directory query already filters to public-listable rows only.
2. After pick B, no client cache of A workforce/theme remains on desk.
3. Server still enforces tenant scope on workforce/RAG/quota (existing).
4. Switching tenant clears active call/transcript client state.

## 9. Verification

```bash
# Directory
curl -s 'http://127.0.0.1:8091/api/public/brands?limit=20' | jq .
curl -s 'http://127.0.0.1:8091/api/public/brands/acme' | jq .

# Unlisted / unknown
curl -s -o /dev/null -w '%{http_code}\n' \
  'http://127.0.0.1:8091/api/public/brands/not-a-real-tenant'

# Portal (manual)
# open http://127.0.0.1:5173/  → picker
# open http://127.0.0.1:5173/t/acme → desk
```

```bash
go test ./internal/store ./cmd/server -count=1 -run 'PublicBrand|Brand'
```

## 10. Implementation sequence

1. **TASK-0195** — confirm/harden public directory + tests; optional alias
2. **TASK-0196** — picker UI + `/t/{slug}` + session + header preference
3. **TASK-0197** — isolation UAT + manual checklist

## See also

- Workflow §125–126 · ER Sprint 54 · API Sprint 54 · UX C54
- [SPRINT-054](../03-sprints/SPRINT-054.md)
