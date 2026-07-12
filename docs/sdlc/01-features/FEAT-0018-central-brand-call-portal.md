# Feature: Centralized Multi-Tenant Brand Call Portal   (FEAT-0018)
**Sprint:** SPRINT-037   **Owner:** DEV   **Status:** backlog

## Problem

Today Monti has:

| Surface | Audience | Scope |
| --- | --- | --- |
| `/` demo conversation | Anyone | Platform demo workforce (not multi-brand discovery) |
| `/embed` + tenant embed key | Per-tenant website | Single tenant only |
| `/tenant/*` | Tenant admin | Own tenant only |
| `/admin/*` | Platform admin | Ops / catalog |

End customers who do **not** land on a tenant’s own site need a **platform-hosted call center portal** that lists **all active tenant brands**, lets them search/browse a brand, pick language + AI employee, and start chat/voice — the blueprint §5.1 **Customer Portal** experience.

Without this, Monti is only “embed on my site” or “demo desk,” not a **centralized multi-brand call center**.

## Scope

**In:**
- Public **brand directory** portal (e.g. `/brands` or dedicated host path) listing **active** tenants with brand profile (name, logo, short description, languages, channels)
- **Search / filter** brands (name, category tag if available)
- **Brand detail** page: profile, languages, assigned AI workforce (avatars), CTA Start chat / Start call
- Session binds to **selected tenant + brand + agent** (reuse conversation/RAG/quota under that `tenant_id`)
- Platform controls: opt-in “list on central portal” per tenant (privacy — not every tenant must be public)
- Tenant admin: brand profile fields used on the central portal (name, logo, blurb, category, listed flag)
- Platform admin: moderate / unlist abusive brands; feature flags
- Deep link: `/brands/{slug}` → conversation pre-selected
- Design pack + manual UAT; align with blueprint §5.1 wireframe

**Out:**
- Per-tenant custom domain CNAME (separate white-label track)
- Full marketplace payments between brands
- Human agent / SIP PSTN (later ops phases)
- Replacing tenant-owned embed (S14) — both coexist: **own site** vs **central Monti hub**
- Full customer account history (→ S19–20; guest path first)
- Embed framework SDKs (→ S36)

## Acceptance criteria

1. Guest opens the central portal and sees only **active + listed** brands (not pending KYC / unlisted).
2. Search by brand name returns matching brands; empty query lists featured/default sort.
3. Selecting a brand shows profile + available AI employees for that tenant.
4. Start **text chat** (and voice when configured) runs under that brand’s `tenant_id` with correct KM/RAG and S13 quotas.
5. Tenant admin can set **listed on central portal** on/off and edit public brand blurb/logo used by the hub.
6. Platform admin can force-unlist a brand.
7. Demo `/` path remains available; central portal is clearly multi-brand (not single hardcoded tenant).

## Test notes

- Seed ≥2 active listed tenants with different KM; confirm answers stay tenant-scoped
- Unlisted / pending tenant never appears in public directory
- Quota exhaustion on tenant A does not block tenant B
- Thai + English brand names and UI labels

## Dependencies

- **SPRINT-006** brands + tenant register · **SPRINT-007** active after KYC
- **SPRINT-001** conversation · **SPRINT-005** avatars · **SPRINT-015** tenant KM (quality)
- Optional: **SPRINT-016** locale · **SPRINT-020** customer auth for saved history
- packages: `apps/customer-web` (or new public brand routes), `cmd/server`, `brands` APIs

## Notes

- Blueprint: Customer Portal “Search brand → select language → AI employee → Start Call”
- Complements embed (S14): central hub for Monti-operated multi-brand call center; embed for tenant-owned websites
- Schedule as **Sprint 37** (Phase J); pull forward after go-live (S14–18) if product wants hub-first distribution
