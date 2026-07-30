---
id: SPRINT-058
status: completed
start: 2026-07-30
end: 2026-08-12
updated: 2026-07-30
design_pack: approved
roadmap_sprint: 58
feature: FEAT-0050
platform: Customer / Tenant / Platform Admin
depends_on: [SPRINT-016, SPRINT-020, SPRINT-039, SPRINT-054, SPRINT-056]
goal: "Ship EN/TH/JA UI language selector and localized labels on customer call page, tenant portal, and platform admin."
velocity_basis: "Last 3 closed: S54=12, S55=12, S56=12 → avg 12; commit 12"
release_target: v2.31.0
release: v2.31.0
closed: 2026-07-30
---

# SPRINT-058 — Portal UI Language Selector (EN · TH · JA)

## Goal

Callers, tenant operators, and platform admins can switch portal UI labels among
English, Thai, and Japanese via a shared language-selector pattern, with
persisted preference and safe EN fallback.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S54, S55, S56) | 12, 12, 12 → **avg 12** |
| **Committed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0208](../04-tasks/TASK-0208.md) Shared i18n runtime + LanguageSelector | 3 | completed | dev | Catalog pattern, storage, selector component |
| [TASK-0209](../04-tasks/TASK-0209.md) Customer call page + directory labels | 3 | completed | dev | customer-web EN/TH/JA chrome |
| [TASK-0210](../04-tasks/TASK-0210.md) Tenant portal UI labels | 3 | completed | dev | tenant-web shell + primary pages |
| [TASK-0211](../04-tasks/TASK-0211.md) Platform admin labels + cross-surface UAT | 3 | completed | tester/dev | admin-web + smoke checklist |
| **Total** | **12** | **12/12** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0050](../01-features/FEAT-0050-portal-ui-language-selector.md) | design_approved |
| Deep spec | **`approved`** | [DES-0054](../02-design/54-portal-ui-language-selector-spec.md) |
| Workflow | **`approved`** | [02-workflow.md](../02-design/02-workflow.md) §130–131 |
| ER | **`approved`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 58 |
| API | **`approved`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 58 |
| UX | **`approved`** | [05-ux-ui.md](../02-design/05-ux-ui.md) C58 / T58 / A58 |

## Scope boundary

### In

- Shared i18n pattern (per-app catalogs; product-web as reference)
- Language selector on customer, tenant, and platform admin surfaces
- Localized primary nav, actions, form labels, statuses, empty/error chrome
- Browser persistence (`localStorage`); optional `?lang=` init like product-web
- EN fallback for missing TH/JA keys
- Document separation from S16 AI/workspace locale

### Out

- Full 100% string coverage of every secondary modal/legacy admin page
- Translating tenant-authored content, KM, transcripts, invoices
- Server preference API / account-profile language field
- product-web rework (already has EN/TH/JA)
- Machine translation pipeline
- Coupling selector to AI spoken language

## Verification

```bash
cd apps/customer-web && npm run check
cd apps/tenant-web && npm run check
cd apps/platform-admin-web && npm run check
# Manual: EN/TH/JA switch + reload on /, /t/{slug}, /tenant, /admin
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Incomplete string inventory | Prioritize shell + primary flows; key map in DES-0054; EN fallback |
| Three apps drift on API | Shared key naming conventions + same storage key prefix |
| Confusing UI lang vs AI locale | Distinct labels (“Display language” vs “AI reply language”) on settings |
| JA quality | Native review of critical chrome; avoid machine-only copy for nav |
| Scope creep into full CMS i18n | Sprint boundary = chrome labels only |

## Worktree

```text
.worktrees/SPRINT-058
branch: feature/sprint-058-portal-ui-i18n
```

## Implementation notes (plan)

- Mirror `apps/product-web/src/lib/i18n/` (`Lang = en|th|ja`, `setLang`, `t`, messages)
- Storage key: `monti_jarvis:ui_lang` (all three apps; product-web may keep own key or migrate later)
- Components: `LanguageSelector.svelte` (or shared pattern) in each app header/rail
- Customer: `TenantPicker`, `CustomerDesk`, auth/account chrome, audio settings labels
- Tenant: `+layout.svelte` nav groups + common buttons/feedback
- Admin: `+layout.svelte` nav + common chrome

## Design section

| Doc | Path | Status |
| --- | --- | --- |
| Deep spec | `docs/sdlc/02-design/54-portal-ui-language-selector-spec.md` | approved |
| Workflow §130–131 | `docs/sdlc/02-design/02-workflow.md` | approved |
| ER Sprint 58 | `docs/sdlc/02-design/03-er-diagram.md` | approved |
| API Sprint 58 | `docs/sdlc/02-design/04-api-spec.md` | approved |
| UX C58/T58/A58 | `docs/sdlc/02-design/05-ux-ui.md` | approved |

## Implementation notes

- Shared pattern from product-web: `lib/i18n` + `LanguageSelector` in customer, tenant, admin
- Storage key `monti_jarvis:ui_lang`; optional `?lang=`
- Tenant settings: display language separate from workspace/AI reply locale
- Manual UAT: [SPRINT-058-portal-ui-language-selector-manual.md](../06-manual-tests/SPRINT-058-portal-ui-language-selector-manual.md)

## Shipped summary (pending release cut)

- EN/TH/JA UI language selector on customer directory/desk, tenant shell, platform admin shell
- Catalog fallback EN; primary chrome localized
- Design: DES-0054 · FEAT-0050 · TASK-0208–0211
- `npm run check` clean on customer-web, tenant-web, platform-admin-web

**Closed:** 2026-07-30
