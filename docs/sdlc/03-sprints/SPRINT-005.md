---
id: SPRINT-005
status: in_progress
start: 2026-07-08
end: 2026-07-21
goal: "Platform Admin: Avatars — catalog CRUD and per-tenant assignment backed by Postgres."
roadmap_sprint: 5
platform: Platform Admin
depends_on: [SPRINT-003, SPRINT-004]
release_target: v0.6.0
---

# SPRINT-005 — Platform Admin: Avatars

## Goal

Move the AI avatar catalog from hardcoded Go into **Postgres**, let **platform admins** manage avatars and **assign them to tenants**, and serve tenant-specific lists via **`GET /api/workforce`** (foundation for Sprint 21 workforce picker).

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0020 | 3 | todo | devops | Postgres `ai_avatars` + `tenant_avatar_assignments` + dev seeds |
| TASK-0021 | 5 | todo | dev | Avatar catalog store + platform CRUD API |
| TASK-0022 | 3 | todo | dev | Tenant avatar assign/revoke/list + `max_ai_employees` check |
| TASK-0023 | 2 | todo | dev | DB-backed `GET /api/workforce` tenant resolver |
| TASK-0024 | 3 | todo | dev | Platform admin portal — avatars + tenant assignment UI |

**Committed:** 16 points · **Completed:** 0 points · **Velocity target:** 16

## Scope boundary

**In**
- `ai_avatars` platform catalog (metadata from current Ava/Max/Luna/Neo)
- `tenant_avatar_assignments` — enable/disable avatars per tenant
- Platform APIs under `/api/platform/avatars*`, `/api/platform/tenants/{id}/avatars*`
- Extend `apps/platform-admin-web` — avatars screens + tenant assignment
- Entitlement-aware assign cap (`rules.max_ai_employees`)
- `sprint-tech-specs` design pack before implementation

**Out** (→ backlog / later sprints)
- `ai_employee_versions`, languages, tools, guardrails (Sprint 21)
- MinIO avatar upload (URL field only this sprint)
- Customer portal workforce UI changes (Sprint 21)
- Tenant admin portal (Sprint 15+)
- Live call quota enforcement (Sprint 13)

## Feature

- [FEAT-0005 — Avatar catalog + tenant assignment](../01-features/FEAT-0005-avatar-catalog.md)

## Design pack (`sprint-tech-specs` — run before build)

| Artifact | Path | Status |
| --- | --- | --- |
| Avatars deep spec | `10-avatars-spec.md` (new) | `planned` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §14+ | `planned` |
| ER diagram | [03-er-diagram.md](../02-design/03-er-diagram.md) | `planned` |
| API spec | [04-api-spec.md](../02-design/04-api-spec.md) | `planned` |
| UX/UI ASCII | [05-ux-ui.md](../02-design/05-ux-ui.md) § P7+ | `planned` |
| Portal (prior) | [09-platform-admin-portal-spec.md](../02-design/09-platform-admin-portal-spec.md) | `shipped` |

Run: `/sprint-tech-specs sprint 5`

## Verification

```bash
make build && make test
make infra-init && make restart
# AUTH_DISABLED=false
open http://localhost:8091/admin/avatars
# Assign avatars to demo → curl /api/workforce with X-Tenant-Id: demo
```

- Manual: `docs/sdlc/06-manual-tests/SPRINT-005-manual.md` (Tester, at VERIFY)

## Risks

| Risk | Mitigation |
| --- | --- |
| Schema name drift (`ai_avatars` vs blueprint `ai_employees`) | Document mapping; Sprint 21 migration path in ER |
| Breaking customer portal agent list | Static fallback when tenant has zero assignments |
| Assignment vs package rules | Read entitlement resolver; return 409 on over-cap |

## Definition of done

- Code reviewed · ACs verified by Tester · portal + API UAT · `make build` · tag v0.6.0 at sprint close