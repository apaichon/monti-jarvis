---
id: FEAT-0045
title: "Customer portal tenant list to call (no required tenant_id query)"
status: implemented
roadmap_sprint: 54
priority: C
depends_on: [SPRINT-001, SPRINT-005, SPRINT-006, SPRINT-020, SPRINT-021, SPRINT-027, SPRINT-038]
design: DES-0049
design_spec: ../02-design/49-customer-portal-tenant-list-spec.md
updated: 2026-07-29
---

# FEAT-0045: Customer Portal Tenant List to Call

## Purpose

Let callers open the customer portal **without** `?tenant_id=` and choose a
call-to tenant (public brand) from a list, then continue to workforce selection
and conversation under that tenant only.

## Problem

The customer portal currently scopes the desk from the query string:

```text
/?tenant_id=acme
```

Callers must know or be handed a tenant id. That blocks a clean multi-brand
entry surface and couples bookmarks/share links to internal ids.

Public brand listing already exists for mobile (`GET /api/public/brands`), but
the web portal does not use it as a first-class picker.

## Scope

### In

- Customer portal **tenant/brand picker** when no tenant is selected
- Reuse (and document) **public directory** of active + listed brands for
  inbound call entry (safe fields only)
- After pick: bind tenant context in-app via **path segment** (`/t/{slug}`)
  and/or session storage — not a required page query string
- Optional deep link: `?tenant_id=` or `/t/{slug}` preselects, then normalizes
  away from a required query-only UX
- Subsequent chat/voice/workforce calls use selected tenant context
  (prefer `X-Tenant-Id` header; API query remains transport fallback)
- Tenant isolation: never mix workforce/KM/history across tenants after switch
- Align with FEAT-0018 central brand portal intent as the **minimal list → call**
  slice (not the full marketing hub)

### Out

- Full multi-brand marketing portal polish (search facets, paid ranking, CMS)
- Cross-tenant conversation history
- Platform admin brand CMS redesign
- Changing OTP / auto-register policy beyond consuming selected tenant
- Embed widget path (still keyed by embed key)
- Removing demo single-tenant env defaults entirely

## Acceptance criteria

1. Opening the customer portal **without** a selected tenant shows a tenant
   list (or empty state), not a broken desk that assumes a tenant.
2. Selecting a tenant loads that tenant’s workforce and conversation only.
3. Primary docs and CTAs do not require `?tenant_id=` on the page URL.
4. Optional deep link may preselect a tenant (`/t/{slug}` preferred;
   `?tenant_id=` accepted then normalized).
5. Chat/voice/workforce requests after pick use the selected tenant context
   and never leak tenant A data after selecting tenant B.
6. Only **active + listed + platform-listed** (and KYC-eligible) tenants appear
   in the public directory (same rules as existing public brands).
7. Directory responses expose only safe public fields (no secrets, admin
   emails, or internal ops metadata).

## Test notes

- Functional: open `/` with no query → picker → select A → agents for A only
- Switch A → B: workforce/theme reset; no residual A messages in new call
- Deep link `/t/{slug}` and legacy `?tenant_id=` both land on desk for that tenant
- Unlisted / pending_kyc tenants never appear
- Empty directory shows empty state with recovery guidance
- Thai + English labels on picker

## Dependencies

- packages: `apps/customer-web`, `cmd/server` public brands, `internal/store`
  brands/tenants, existing workforce/theme APIs
- related: [FEAT-0018](FEAT-0018-central-brand-call-portal.md) (broader hub),
  [DES-0030](../02-design/30-mobile-call-api-sdk-spec.md) public brands,
  [DES-0049](../02-design/49-customer-portal-tenant-list-spec.md)
