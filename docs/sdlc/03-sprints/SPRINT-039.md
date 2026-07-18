---
id: SPRINT-039
status: planned
start: 2026-07-21
end: 2026-07-24
updated: 2026-07-18
design_pack: review_pending
release_target: v2.15.0
goal: "Tenant / Platform: configurable theme color tokens (presets, draft/publish, contrast, customer+embed application)."
roadmap_sprint: 39
feature: FEAT-0035
platform: Tenant / Platform
depends_on: [SPRINT-014, SPRINT-016]
---

# SPRINT-039 — Multiple Theme Color Customization

## Goal

Let tenants configure and publish brand color palettes that apply consistently to the **customer caller desk** and **web embed**, with draft/publish safety, contrast warnings, and reset-to-preset — without forking CSS per host site.

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
| [TASK-0149](../04-tasks/TASK-0149.md) Theme schema, validation, tenant APIs | 4 | dev | todo | `tenant_themes` (+ audit), draft/publish/reset, contrast helper, RBAC |
| [TASK-0150](../04-tasks/TASK-0150.md) CSS token runtime on customer + embed | 4 | dev | todo | Published tokens → CSS variables; public resolve; default dark parity |
| [TASK-0151](../04-tasks/TASK-0151.md) Tenant Theme editor UI + live preview | 4 | dev | todo | Presets, pickers, preview pane, warn-on-contrast, publish/reset |
| [TASK-0152](../04-tasks/TASK-0152.md) Platform read view, UAT, docs | 2 | tester/dev | todo | Support summary, manual UAT, integrator notes |

**Committed:** 14 points · **Task IDs:** TASK-0149–TASK-0152.

## Scope boundary

**In**

- Presets: light, dark, branded (custom tokens).
- Tokens: primary, accent, surface, text, muted, line, success, warn, danger (exact list locked in DES-0037).
- Per-tenant draft vs published document; reset to preset.
- Apply published theme on customer portal + embed iframe surfaces.
- Tenant Theme admin UI with preview and contrast flags.
- Public published-theme resolve for client bootstrap.
- Platform admin read-only tenant theme summary (support).

**Out**

- Logo/font packs, full white-label CNAME, Storybook DS productization.
- Mobile native theme APIs (may consume tokens later).
- Per-page or per-component CSS upload.
- Bulk multi-tenant theme templates marketplace.
- Changing S14 security model or embed key semantics.

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
