# Feature: Tenant Embed to Web   (FEAT-0014)
**Sprint:** SPRINT-014   **Owner:** DEV   **Status:** shipped (v1.5.0)

## Problem

Active tenants can use Monti conversation only on the hosted customer portal (`/`). To go live on **their** website they need a **copy-paste embed** (script or iframe) that opens chat/voice for **their** tenant, without giving platform admin access to the host page.

## Scope

**In:**
- Per-tenant **embed config**: public `embed_key`, enable/disable, **allowed origins** allowlist, optional default agent
- Public resolve API: `GET /api/public/embed/{embed_key}` (no auth)
- Embed **loader** (JS snippet) that mounts an iframe (or floating launcher) pointing at Monti embed surface
- Customer **embed mode** UI (minimal chrome) using resolved tenant for chat/voice/workforce
- Tenant portal **`/tenant/embed`**: show snippet, rotate key, origins, preview
- CORS / `frame-ancestors` / origin checks for public embed paths
- Design: [17-embed-to-web-spec.md](../02-design/17-embed-to-web-spec.md)
- **Integrator guide:** [EMBED_WEB_INTEGRATION.md](../../EMBED_WEB_INTEGRATION.md)

**Out:**
- Tenant KM/scope admin (→ SPRINT-015)
- Locale / customer-facing limits UI (→ SPRINT-016)
- Test & preview sandbox (→ SPRINT-017)
- Customer register/login (→ SPRINT-019–020)
- White-label custom domain CNAME
- Mobile SDK; framework packages (Vue/React/Svelte/Web Component) → **SPRINT-036 / FEAT-0017** (loader remains first-party static asset in S14)
- Platform multi-brand directory portal → **SPRINT-037 / FEAT-0018** (central hub; embed remains per-tenant site)
- Billing for embed views

## Acceptance criteria

1. Active tenant can enable embed and copy a snippet that loads Monti on a third-party HTML page.
2. Embed resolves only by **embed_key** (not forgeable tenant_id alone); disabled or unknown key → 404.
3. Requests from origins **not** in allowlist are rejected (when allowlist non-empty); empty allowlist = any origin in dev (documented).
4. Embed surface can list workforce agents and run **text chat** for that tenant; voice works when Gemini configured.
5. Quota enforcement (S13) applies to embed traffic under the resolved tenant id.
6. Tenant admin can regenerate embed key (old key invalid) and toggle enabled.
7. Customer full portal at `/` unchanged for demo; embed does not require tenant JWT.

## Test notes

- Static HTML fixture with snippet + localhost origin
- Curl public resolve + negative cases (disabled, bad origin)
- Tenant UI copy + preview
- Thai + English labels on tenant embed screen

## Dependencies

- SPRINT-001 conversation · SPRINT-006 tenant register · active tenant after KYC
- SPRINT-013 quotas (enforcement reuses resolved tenant)
- packages: `apps/customer-web`, `apps/tenant-web`, `cmd/server`
