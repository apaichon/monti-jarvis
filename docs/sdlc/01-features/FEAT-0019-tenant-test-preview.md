# Feature: Tenant Test and Preview Sandbox   (FEAT-0019)
**Sprint:** SPRINT-017   **Owner:** DEV   **Status:** shipped (v1.8.0)

## Problem

Tenants can configure KM (S15), settings/limits (S16), and embed (S14), but have **no in-portal way** to try the full caller experience **as their own tenant** before sending traffic to customers. They currently use the public demo desk or a live embed — which either uses the wrong tenant or **consumes production package minutes/quotas**.

## Scope

**In:**
- Tenant admin **Preview** surface (`/tenant/preview`): pick agent + topic, **text chat** and optional **voice** against **their** tenant KM / locale / workforce
- Sessions tagged **`preview`** so ops can distinguish from production calls
- **Quota policy for preview (same as production package):**
  - Enforce **rate limits**, package concurrent slots, S13 monthly minutes, S16 daily/per-call caps
  - Sessions logged with `source=preview` for ops visibility (still charged)
- **Embed-like UI:** avatar portrait, agent picker, chat + voice matching customer embed
- **Scenario checklist** (static): suggested questions by topic to validate KM coverage
- Link to **live embed** when enabled
- Design pack DES-0020 + manual UAT

**Out:**
- Separate staging infrastructure / dual databases
- A/B script experimentation product
- Customer identity (S19–20)
- Customer tiers (S18)
- Full conversation archive product (S22) — preview may write lightweight session rows only
- Changing production embed security model

## Acceptance criteria

1. Active tenant admin opens `/tenant/preview`, selects an assigned agent, sends a text question, and receives a reply grounded in **that tenant’s** KM when available.
2. Preview chat requests carry tenant from JWT (not forgeable tenant id alone).
3. Preview voice (when Gemini configured) works under the same tenant scope; captions use mono-language rules from S16.
4. Completing a preview voice session **does** increase package monthly call minutes and daily operational counters (same as production).
5. Preview UI shows **agent avatar** (portrait) like real embed, with chat + voice.
6. UI shows clear **“Preview mode — uses package rate limits & call minutes”** banner.
7. Inactive / non-admin users cannot access preview APIs.

## Test notes

- Tenant A KM answer vs tenant B isolation
- Compare Redis monthly minutes before/after preview voice
- Embed link opens when embed enabled; disabled embed shows CTA to `/tenant/embed`

## Dependencies

- SPRINT-001 conversation · SPRINT-015 KM · SPRINT-016 settings/locale · SPRINT-014 embed (optional link)
- packages: `apps/tenant-web`, `cmd/server`, `internal/quota`, `internal/rag`

## Production note (carry from S16)

Before **customer production launch** after tenant customer-user auth (S19–20), re-verify rate limit + package quota under multi-user load. Preview mode must remain distinguishable so it never masks production quota bugs.
