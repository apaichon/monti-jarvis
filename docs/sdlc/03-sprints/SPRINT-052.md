---
id: SPRINT-052
status: completed
start: 2026-07-28
end: 2026-07-28
closed: 2026-07-28
updated: 2026-07-28
design_pack: shipped
release: v2.24.0
release_target: v2.24.0
roadmap_sprint: 52
feature: FEAT-0043
platform: Tenant / Platform
depends_on: [SPRINT-005, SPRINT-013, SPRINT-015, SPRINT-016, SPRINT-045, SPRINT-050]
goal: "Tenant admins can create unlimited library avatars; only active avatars are capped by package max_ai_employees."
velocity_basis: "Last 3 closed with velocity: S45=13, S46=17, S50=12 → avg 14.0; commit 12"
---

# SPRINT-052 — Tenant Avatar Create & Active Cap

## Goal

Give tenants self-service avatar **creation** without a create-count limit,
while enforcing package **`max_ai_employees` only on activate** so the active
workforce stays within plan. Portraits upload and Gemini AI Studio speaker
voices are selectable in the tenant library UI.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed slices (S45, S46, S50) | 13, 17, 12 → **avg 14.0** |
| **Committed** | **12** |
| **Completed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0185](../04-tasks/TASK-0185.md) Tenant avatar store + API (create library, activate cap) | 5 | completed | dev | Tenant-owned avatars; create free; activate capped |
| [TASK-0186](../04-tasks/TASK-0186.md) Tenant web avatar library UX | 4 | completed | dev | Library UI, portrait upload, speaker voice dropdown |
| [TASK-0187](../04-tasks/TASK-0187.md) Workforce surface + isolation + verification | 3 | completed | tester/dev | Active-only workforce; tests; UAT checklist |
| **Total** | **12** | **12/12** | | **Shipped v2.24.0** |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0043](../01-features/FEAT-0043-tenant-avatar-create-active-cap.md) | shipped |
| Deep spec | **`shipped`** | [DES-0047](../02-design/47-tenant-avatar-create-active-cap-spec.md) |
| Workflow | **`shipped`** | [02-workflow.md](../02-design/02-workflow.md) §118–119 |
| ER | **`shipped`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 52 |
| API | **`shipped`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 52 |
| UX | **`shipped`** | [05-ux-ui.md](../02-design/05-ux-ui.md) T52 |

## Shipped summary (v2.24.0)

- Tenant-owned avatars (`owner_tenant_id`) with inactive-by-default library
- Create without package avatar-count check; activate uses `CheckAIEmployees`
- `GET/POST /api/tenant/avatars`, activate/deactivate, image upload, archive
- `GET /api/tenant/avatar-voices` — 30 Gemini AI Studio speakers
- Tenant UI `/tenant/avatars` — cap meter, create, portrait upload, voice select
- Migration `034_tenant_owned_avatars.sql` + bootstrap ensure
- Manual UAT: [SPRINT-052-tenant-avatars-manual.md](../06-manual-tests/SPRINT-052-tenant-avatars-manual.md)

## Verification

```bash
go test ./internal/store ./cmd/server -count=1
cd apps/tenant-web && npm run check
```
