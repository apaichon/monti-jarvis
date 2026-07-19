# Feature: Theme Branding & Color Customization   (FEAT-0035)
**Sprint:** SPRINT-039   **Owner:** DEV   **Status:** shipped · **Release:** v2.15.0

## Problem

Caller desk and embed chrome hard-code Monti logo, title, and palette (see production UI: brand mark + name + “AI · text & voice”, dark panels, blue **Start call** / **Send**). Tenants need to brand this surface with their **company name, logo, subtitle, and full color theme** without forking the app.

## UI reference

Customer/embed call chrome (screenshot 2026-07-19):

- Header: **logo** · **brand name** · **subtitle**
- Agent orb / traits (agent product — not theme-owned)
- **Start call** primary CTA, status panel, chat bubbles, **Send**

## Scope

**In:**
- Editable **brand identity**: `brand_name`, `logo` (upload or URL), `subtitle` (+ optional `logo_alt`)
- Editable **color tokens** for full chrome: primary (+ on-primary text), accent, background, surface, surface elevated, text, muted, line, success, warn, danger
- Presets: light / dark / branded
- Draft vs published; reset; contrast warnings on publish
- Apply published branding + colors on **customer desk** and **embed** header + chrome
- Live admin preview that mirrors caller chrome
- Public theme resolve + embed resolve extension
- Platform read-only summary
- Design pack DES-0037 + UAT at VERIFY

**Out:**
- White-label CNAME
- Agent avatar/name/role editing (workforce catalog)
- Arbitrary CSS/HTML injection
- Mobile native theme API
- Marketplace of shared templates

## Acceptance criteria

1. Tenant admin can set **brand name**, **logo**, and **subtitle**, save draft, and **publish**.
2. Published values appear on **embed** and **customer** headers (replacing hard-coded Monti mark/title/sub when set).
3. Tenant admin can edit **all required color tokens**, publish, and see **Start call**, **Send**, panels, borders, and text use those tokens.
4. Empty published fields fall back to documented defaults (Monti logo, workspace name, default subtitle, dark palette).
5. Contrast checker warns on inaccessible pairs; publish requires confirm when warnings remain.
6. Tenant A branding/theme does not affect Tenant B.
7. Inactive / non-admin → 401/403; public never receives draft.
8. Manual UAT checklist covers branding + colors on embed matching screenshot regions.

## Test notes

- Unit: token + branding validation, contrast, fallback chain  
- API: PUT draft with branding+tokens, publish, public GET  
- Manual: set “Libra Tech Co.,Ltd” + custom logo + subtitle + primary blue → embed header and CTAs match  

## Dependencies

- SPRINT-014 embed surface  
- SPRINT-016 settings patterns  
- MinIO for logo objects (`theme/{tenant_id}/`)  
- Existing `brands.logo_url` / company name may seed defaults only  

## Notes

Roadmap #39 · DES-0037 expanded 2026-07-19 to cover full brand chrome, not colors alone.
