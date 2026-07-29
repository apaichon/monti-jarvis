---
id: SPRINT-054
status: design_approved
start: 2026-07-29
end: 2026-08-05
updated: 2026-07-29
design_pack: approved
roadmap_sprint: 54
feature: FEAT-0045
platform: Customer / Platform
depends_on: [SPRINT-001, SPRINT-005, SPRINT-006, SPRINT-020, SPRINT-021, SPRINT-027]
goal: "Customer portal lists call-to tenants and removes required ?tenant_id= from the primary entry path."
velocity_basis: "Last 3 closed with velocity: S46=17, S50=12, S52=12 → avg 13.7; commit 12"
release_target: v2.25.0
---

# SPRINT-054 — Customer Portal Tenant List to Call

## Goal

Callers open the customer portal without `?tenant_id=`, pick a public tenant
(brand) from a list, and continue to avatar selection / chat / voice under that
tenant only. Optional deep links may preselect a tenant via path slug.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed slices (S46, S50, S52) | 17, 12, 12 → **avg 13.7** |
| **Committed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0195](../04-tasks/TASK-0195.md) Public directory contract + tenant resolve | 5 | todo | dev | Safe list/resolve APIs; document reuse of public brands |
| [TASK-0196](../04-tasks/TASK-0196.md) Customer portal picker + path context | 4 | todo | dev | Picker UI; `/t/{slug}`; no required query |
| [TASK-0197](../04-tasks/TASK-0197.md) Isolation verification + UAT | 3 | todo | tester/dev | A/B isolation; deep link; manual checklist |
| **Total** | **12** | **0/12** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0045](../01-features/FEAT-0045-customer-portal-tenant-list.md) | design_approved |
| Deep spec | **`approved`** | [DES-0049](../02-design/49-customer-portal-tenant-list-spec.md) |
| Workflow | **`approved`** | [02-workflow.md](../02-design/02-workflow.md) §125–126 |
| ER | **`approved`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 54 |
| API | **`approved`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 54 |
| UX | **`approved`** | [05-ux-ui.md](../02-design/05-ux-ui.md) C54 |

## Scope boundary

### In

- Public tenant/brand list for inbound call entry (reuse `brands` directory)
- Customer portal picker before workforce when no tenant selected
- Path-based context `/t/{slug}` + session storage; optional `?tenant_id=` preselect
- Prefer `X-Tenant-Id` on subsequent client API calls
- Isolation guarantees and verification

### Out

- Full FEAT-0018 marketing hub (facets, ranking, CMS)
- Cross-tenant history
- Embed path changes
- S53 auto-register / app version (separate sprint)
- S55 stats by topic

## Verification

```bash
go test ./internal/store ./cmd/server -count=1
cd apps/customer-web && npm run check
# Manual: open / without query → pick tenant A → call → switch B
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Empty directory (no listed brands) | Empty state + seed/docs for local demo listed tenants |
| Legacy bookmarks with `?tenant_id=` | Accept as preselect; normalize to `/t/{slug}` |
| Client still depends on query for APIs | Prefer header; keep query as transport fallback for WS/SSE |
| Scope creep into full brand hub | Hard out-of-scope list; S54 = list → call only |

## Worktree

```text
.worktrees/SPRINT-054
branch: feature/sprint-054-customer-tenant-list
```
