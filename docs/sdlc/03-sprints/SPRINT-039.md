---
id: SPRINT-039
status: planned
start: 2026-07-21
end: 2026-07-24
updated: 2026-07-19
design_pack: review_pending
release_target: v2.15.0
goal: "Tenant / Platform: configurable caller brand chrome (name, logo, subtitle) and full color theme with draft/publish."
roadmap_sprint: 39
feature: FEAT-0035
platform: Tenant / Platform
depends_on: [SPRINT-014, SPRINT-016]
---

# SPRINT-039 — Multiple Theme Color Customization

## Goal

Let tenants configure and publish **brand name, logo, subtitle, and full color theme** for the customer caller desk and web embed chrome (screenshot-aligned header + CTAs), with draft/publish, contrast warnings, and reset — without forking the app.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S31, S32, S37) | 16, 3, 14 → **avg 11.0** |
| Trailing note | S32 was readiness-only (3 pts); S31/S37 were full feature slices at 16/14 |
| **Committed** | **14** |

Commitment sits slightly above the 11-pt average because token model + multi-surface application + editor/preview are a cohesive unit; no Twilio/provider risk in this slice.

## Commitment

| Work package | Points | Owner | Status | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0149](../04-tasks/TASK-0149.md) Theme schema, branding+tokens APIs | 4 | dev | todo | `tenant_themes` branding+tokens, draft/publish/reset, logo upload, contrast, RBAC |
| [TASK-0150](../04-tasks/TASK-0150.md) Apply branding + CSS on customer + embed | 4 | dev | todo | Header logo/name/subtitle + CSS vars; public/embed resolve; dark default parity |
| [TASK-0151](../04-tasks/TASK-0151.md) Tenant Theme editor (brand + colors + preview) | 4 | dev | todo | Brand fields, logo upload, color pickers, screenshot-like preview, publish/reset |
| [TASK-0152](../04-tasks/TASK-0152.md) Platform read view, UAT, docs | 2 | tester/dev | todo | Support summary, UAT for brand+colors, docs |

**Committed:** 14 points · **Task IDs:** TASK-0149–TASK-0152.

## Scope boundary

**In**

- Brand chrome: **brand name**, **logo** (upload/URL), **subtitle** (caller/embed header).
- Presets: light, dark, branded.
- Full color tokens (primary, primary_text, accent, background, surface, surface_elevated, text, muted, line, success, warn, danger) — DES-0037.
- Draft vs published; reset; contrast soft-gate on publish.
- Apply published branding + colors on customer desk + embed (header + CTAs + panels).
- Tenant Theme admin with live preview matching caller chrome screenshot.
- Public / embed resolve of published payload only.
- Platform admin read-only summary.

**Out**

- White-label CNAME; arbitrary CSS/HTML injection; Storybook DS productization.
- Agent portrait/name/role editing (workforce product).
- Mobile native theme APIs (may consume published payload later).
- Template marketplace; changing S14 embed security model.

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0035](../01-features/FEAT-0035-theme-color-customization.md) | `planned` |
| Deep spec | [DES-0037](../02-design/37-theme-color-customization-spec.md) | `review_pending` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §89–90 | `review_pending` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 39 | `review_pending` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 39 Theme | `review_pending` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) Sprint 39 T20 | `review_pending` |

Implementation starts when DES-0037 + API sections are **approved**.

## Context

| Sprint | Capability reused |
| --- | --- |
| 14 | Embed surface, public resolve, optional `theme` query |
| 16 | Tenant settings shell, locale, active `tenant_admin` APIs |
| 37 | Framework embed SDKs — must keep `theme` prop harmless |

## Verification target

```bash
make test
make build
cd apps/tenant-web && npm run check && npm run build
cd apps/customer-web && npm run check && npm run build
# Unit: contrast + token validation
# Manual: draft → publish → customer/embed colors; tenant isolation; reset preset
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Low-contrast brand colors harm accessibility | Soft gate: contrast warnings + confirm-to-publish |
| Embed cache shows stale theme | Resolve published theme on iframe load; short cache headers if any |
| CSS var drift across apps | Single token map documented in DES-0037; shared naming (`--mj-primary`, …) |
| Parallel S40 task-id reuse in other worktree | This sprint uses TASK-0149+ on main lineage after S37 TASK-0148 |

## Release

Target **v2.15.0** when verified (minor feature).
