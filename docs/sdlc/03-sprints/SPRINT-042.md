---
id: SPRINT-042
status: completed
start: 2026-07-21
end: 2026-07-23
updated: 2026-07-19
design_pack: shipped
release_target: v2.16.0
release: v2.16.0
closed: 2026-07-19
goal: "Quality: fix tenant session expiry, first-login menu, nav grouping/scroll, and KM document scope."
roadmap_sprint: 42
feature: FEAT-0036
platform: Quality / Tenant
depends_on: [SPRINT-003, SPRINT-015, SPRINT-020]
---

# SPRINT-042 — Bug Fix (Quality / Tenant UX)

## Goal

Close four production-blocking **tenant console** defects — session expiry UX, first-login missing menu, non-scrollable flat nav, and incomplete KM document scope — without shipping new product features from S43.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S32, S37, S39) | 3, 14, 14 → **avg ~10.3** |
| Last 3 full product slices (S31, S37, S39) | 16, 14, 14 → **avg ~14.7** |
| **Committed** | **12** |

Commitment sits near the mixed average: four focused bug slices, no Twilio/AI product risk.

## Commitment

| Work package | Points | Owner | Status | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0154](../04-tasks/TASK-0154.md) Session expired redirect + next | 3 | dev | completed | Consistent 401 → login with reason + deep-link restore |
| [TASK-0155](../04-tasks/TASK-0155.md) First-login menu paints without refresh | 3 | dev | completed | Shell/nav reactive to session after login |
| [TASK-0156](../04-tasks/TASK-0156.md) Tenant nav groups + scroll | 3 | dev | completed | Grouped IA + overflow scroll sidebar |
| [TASK-0157](../04-tasks/TASK-0157.md) KM document scope assign/filter | 3 | dev | completed | Scope on document CRUD + list; RAG respects scope |

**Committed:** 12 points · **Task IDs:** TASK-0154–TASK-0157.

UAT checklist is produced at VERIFY ([manual-test-doc](../../.claude/skills/manual-test-doc/SKILL.md)) — not a separate commitment row.

## Scope boundary

**In**

- Tenant web auth session expiry UX (`session_expired`, `next`).
- Fix first paint of tenant layout after login.
- Nav grouping + scrollable sidebar.
- Document ↔ scope on tenant KM (API + UI).
- Regression: existing login, KM upload, and scope list still work.

**Out**

- S43 features (embed auth flag, env groups, tenant Gemini key, tools/skills).
- Platform admin or customer portal redesign.
- Changing package/quota product rules.
- Full design-system rewrite.

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0036](../01-features/FEAT-0036-tenant-ux-bugfix.md) | `planned` |
| Deep spec | [DES-0038](../02-design/38-tenant-ux-bugfix-spec.md) | `shipped` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §91–92 | `shipped` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 42 (document scope) | `shipped` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 42 | `shipped` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) Sprint 42 T21 | `shipped` |

Implementation starts when DES-0038 + API deltas are **approved** (nav/session can start under UX approval alone if store schema unchanged).

## Context

| Sprint | Capability reused |
| --- | --- |
| 3 | JWT / session / 401 semantics |
| 15 | Tenant KM + scopes |
| 20 | Customer auth redirect patterns (parity notes) |
| 39 | Long tenant nav (Theme link added — scroll bug more visible) |

## Verification target

```bash
make test
make build
cd apps/tenant-web && npm run check && npm run build
# Manual: SPRINT-042-manual.md (create at VERIFY)
# Browser: expire token; first login; short viewport scroll; document scope round-trip
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Over-fixing auth breaks remember-me / refresh | Keep refresh path; only hard-logout when refresh fails |
| Nav regroup confuses power users | Keep same routes; group labels only; no URL renames |
| Scope change breaks existing docs | Default scope = tenant-wide / existing behavior when unset |

## Release

Target **v2.16.0** (patch-like quality release; minor bump if schema/API surface for document scope expands).
