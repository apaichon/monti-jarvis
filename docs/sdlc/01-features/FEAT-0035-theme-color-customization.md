# Feature: Multiple Theme Color Customization   (FEAT-0035)
**Sprint:** SPRINT-039   **Owner:** DEV   **Status:** planned

## Problem

Tenant and customer UIs ship a single dark Monti palette (`--ink`, `--cyan`, `--blue`, …). Integrators and brand teams need **configurable color tokens** so the caller desk, embed widget, and tenant admin chrome can match brand guidelines without forking CSS.

Today there is no draft/publish theme model, no contrast check, and no public theme resolve for embed hosts (S14 only passes optional `theme` query as a hint with no server palette).

## Scope

**In:**
- Preset palettes: **light**, **dark**, and **branded** (tenant-custom token set)
- Editable tokens: primary, accent, surface/background, text/ink, muted text, border/line, success, warn, danger
- Per-tenant theme row with **draft** vs **published** state + **reset to preset**
- Optional **embed instance** override: when embed opens with `theme=…` or tenant default published theme, inject CSS variables into embed/customer surfaces
- Tenant admin Theme page (or Settings → Theme): color pickers, live preview, contrast flags, save draft / publish / reset
- APIs for tenant admin + public published theme resolve (for customer/embed bootstrap)
- WCAG-oriented contrast warnings (flag inaccessible text-on-surface / primary-on-surface pairs before publish; soft gate with explicit override optional)
- Platform admin: read-only view of tenant theme summary (support/debug) — no full design-system product
- Design pack DES-0037 + manual UAT checklist

**Out:**
- Per-route micro-themes or arbitrary component-level CSS upload
- Full multi-brand design system / Storybook productization
- Logo/font/typography system beyond colors (logo already on branding paths)
- White-label custom domain CNAME (separate backlog)
- Mobile native theme APIs (may reuse published tokens later via mobile bootstrap)
- Changing package billing or KYC surfaces theme productization
- AI-generated brand palettes

## Acceptance criteria

1. Active tenant admin opens Theme UI, selects a preset (light/dark/branded), edits tokens, saves **draft**, and **publishes** without restarting the server.
2. Published theme applies CSS variables on **customer caller desk** and **embed** surfaces for that tenant within one page load / resolve.
3. **Reset** restores the selected preset defaults and can re-publish.
4. Contrast checker flags failing text/surface and primary/surface combinations before publish (at least AA-ish ratio check documented); admin can still publish only after acknowledging warning **or** fixing colors (pick one in DES — prefer warn + require confirm).
5. Public resolve (or existing embed/customer bootstrap) returns published tokens so clients need no hard-coded Monti blues.
6. Tenant A cannot read or write Tenant B theme; inactive tenant / non-admin → 401/403.
7. Vanilla `monti-embed.js` and framework SDKs still work; optional `theme` prop continues to forward without breaking when theme feature is disabled or unset.
8. Manual UAT checklist under `docs/sdlc/06-manual-tests/SPRINT-039-manual.md` (created at VERIFY).

## Test notes

- Unit: token validation (hex/rgb), contrast ratio helper, draft vs published isolation
- API: GET/PUT draft, POST publish, POST reset, public GET published theme
- Manual: tenant edits primary → publish → open customer desk + embed → colors match; second tenant unchanged
- Regression: default dark preset matches pre-S39 look within documented tolerance

## Dependencies

- **SPRINT-014** Embed to Web (FEAT-0014) — embed surface + optional theme query
- **SPRINT-016** Tenant settings (FEAT-0016) — settings shell / locale patterns
- packages: `internal/store` (theme table), tenant-web Theme UI, customer-web + embed CSS variable application, public resolve path
- blueprint: brand-facing customer portal surfaces

## Notes

- Roadmap #39 · Phase D+ · pull-forward from infra track when brand demand is high.
- Prefer CSS custom properties mapped from a single token JSON document (no per-component hard-coded brand colors in new UI).
