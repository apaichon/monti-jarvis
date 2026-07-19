---
id: DES-0038
title: Tenant UX Bug Fix Specification
status: review_pending
updated: 2026-07-19
sprint: SPRINT-042
owner: SA
release_target: v2.16.0
---

# Tenant UX Bug Fix — Design Spec

**Sprint:** SPRINT-042 · **Release target:** v2.16.0  
**Feature:** [FEAT-0036](../01-features/FEAT-0036-tenant-ux-bugfix.md)  
**Depends on:** Auth (S3), Tenant KM/scope (S15)

## 1. Goals

1. Reliable **session-expired** UX for tenant admin console.  
2. **First-login** always shows full shell + nav.  
3. **Grouped, scrollable** tenant navigation.  
4. **Document scope** assignable and enforced for KM.

## 2. Non-goals

| Out | Notes |
| --- | --- |
| S43 product features | Separate sprint |
| Customer portal nav redesign | Only touch if shared helpers |
| JWT TTL product change | Document only |

## 3. Session expired

### 3.1 Client flow

```text
apiFetch → 401 + had token
  → try refresh once (if refresh endpoint/helper exists)
  → on failure: clearSession()
  → location.replace(`${base}/login?reason=session_expired&next=${encodeURIComponent(path)}`)
```

### 3.2 Login page

| Query | Behavior |
| --- | --- |
| `reason=session_expired` | Banner: “Your session expired. Please sign in again.” |
| `next` | Must be relative path under tenant app; reject `//`, `http:`, external |

### 3.3 Server

No new endpoints required unless refresh is missing. Ensure API returns **401** (not 500) on expired JWT.

## 4. First-login menu

### Root cause class

Layout reads session only once at module init, or login navigates before session write flushes.

### Fix direction

- Single reactive session source (`$state` / store) written **before** `goto`.
- Layout subscribes to session; when `hasSession` flips true, render shell + nav.
- Avoid `window.location.href` full reload as the only fix (prefer SPA-correct reactivity).

## 5. Nav grouping + scroll

### Recommended groups (labels editable in UI)

| Group | Links (existing routes) |
| --- | --- |
| Operations | Overview, Call center, Monitoring, Tickets, Satisfaction, Preview |
| Knowledge | Knowledge, Gaps, Conversation records |
| Commerce | Billing, Documents, Tax |
| Channels | Embed, Theme |
| Directory | Customers, Tiers |
| Settings | Settings, (KM remains Knowledge) |

Exact membership may be tuned in implementation; **no URL changes**.

### Layout CSS

```text
aside.sidebar
  brand (shrink 0)
  nav.tenant-nav (flex 1; min-height 0; overflow-y: auto)
  foot (shrink 0)
```

## 6. Document scope

### Data

Prefer existing KM document table column or metadata:

| Field | Type | Notes |
| --- | --- | --- |
| `scope_id` | text null | FK/logical id to tenant scope; null = **tenant default / unscoped** (current behavior) |

If column missing: `ALTER … ADD COLUMN IF NOT EXISTS scope_id text`.

### API (extend existing tenant KM)

| Method | Path | Change |
| --- | --- | --- |
| `POST` …/documents | multipart/json | accept `scope_id` |
| `PATCH` …/documents/{id} | body | update `scope_id` |
| `GET` …/documents | query | `scope_id` filter optional |

### RAG

When resolving agent context, filter embeddings/docs by agent allowed scopes ∪ unscoped policy documented in S15. Default: unscoped docs remain available to all agents of the tenant (non-breaking).

## 7. Verification

```bash
# Session
# 1) Login, delete access token in Application storage, click nav → login?reason=session_expired
# Menu
# 2) Incognito login → nav full
# Scroll
# 3) DevTools device mode short height → scroll to Theme
# Scope
# 4) Upload doc with scope S → list filter S → chat agent without S does not cite it (if scoped-only)
```

## 8. See also

- [02-workflow.md](02-workflow.md) §91–92  
- [03-er-diagram.md](03-er-diagram.md) Sprint 42  
- [04-api-spec.md](04-api-spec.md) Sprint 42  
- [05-ux-ui.md](05-ux-ui.md) Sprint 42  
- [SPRINT-042](../03-sprints/SPRINT-042.md)
