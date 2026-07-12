# Feature: Tenant Customer Tier   (FEAT-0020)
**Sprint:** SPRINT-018   **Owner:** DEV   **Status:** shipped (v1.9.0)

## Problem

S16 only stored free-text **tier/group labels** on settings. Tenants need a real **customer tier catalog** (e.g. VIP / Standard / Guest) with rules that will drive benefits when customer accounts arrive (S19–20): default agent, language hint, tighter/looser call caps, notes for ops.

Without this, go-live operators cannot model “who gets what” before identity is online.

## Scope

**In:**
- Postgres `customer_tiers` (per tenant): name, slug, priority, description, optional `default_agent_id`, optional override caps (`max_minutes_per_call`, `max_call_minutes_per_day` — `0` = inherit tenant/package), `ai_reply_locale` override, active flag, audit
- Optional `customer_groups` (name, slug, notes) for ops labels — assignment to customers deferred to S19+
- Tenant APIs (`tenant_admin` active):
  - `GET/POST /api/tenant/tiers`
  - `GET/PUT/DELETE /api/tenant/tiers/{id}`
  - `GET/POST /api/tenant/groups` (minimal CRUD)
  - `GET/PUT/DELETE /api/tenant/groups/{id}`
- Tenant UI `/tenant/tiers` — list/create/edit tiers + groups; show inherit vs override caps
- Settings page: replace free-text tier/group scaffold help with link to Tiers admin; keep legacy label fields read-only or remove from primary UX
- Preview (optional): pass `tier_id` on preview chat/voice to apply tier locale/cap overrides for realistic testing
- Design DES-0021 + manual UAT

**Out:**
- End-customer register/login (→ **SPRINT-019–020**)
- Binding real customers to tiers/groups (needs customer identity)
- Monetary benefits / discounts
- Platform-global tier templates
- Full RBAC matrix per tier

## Acceptance criteria

1. Active tenant admin opens `/tenant/tiers` and creates at least one tier (e.g. `vip`) with name + slug.
2. Tier optional caps cannot be negative; `0` means inherit tenant call limits / package.
3. List/get/update/delete tiers scoped to JWT tenant only (no cross-tenant leak).
4. Groups can be created as free-form ops labels (no customer assignment required).
5. Preview or chat path can accept optional `tier_id` and apply that tier’s `ai_reply_locale` (and document cap overrides for voice when set).
6. Inactive / non-admin → 401/403.
7. Manual UAT under `06-manual-tests/SPRINT-018-manual.md`.

## Test notes

- Two tenants: tier of A not visible to B
- Delete tier in use: soft-block if referenced later; S18 may hard-delete if no customer FKs yet
- Thai + English labels on tiers UI chrome

## Dependencies

- SPRINT-016 settings + call limits foundation
- SPRINT-017 preview (optional tier_id wiring)
- packages: `apps/tenant-web`, `cmd/server`, `internal/store`

## Notes

- Blueprint Phase D “customer tier rules”
- Prepares S19–20: `customers.tier_id` FK will land when identity ships
