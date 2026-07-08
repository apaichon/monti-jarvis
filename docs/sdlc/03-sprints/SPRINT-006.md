---
id: SPRINT-006
status: in_progress
start: 2026-07-08
end: 2026-07-21
goal: "Tenant: Register — public self-signup creates pending_kyc tenant + tenant_admin login."
roadmap_sprint: 6
platform: Tenant
depends_on: [SPRINT-003]
release_target: v0.7.0
---

# SPRINT-006 — Tenant: Register

## Goal

Ship a **public tenant registration** flow: a prospect submits company + admin credentials, the backend creates a `pending_kyc` tenant with a default brand stub and **tenant_admin** user, returns JWT tokens, and a new **tenant portal** surface at `/tenant` captures signups. Platform admins can list pending tenants ahead of Sprint 7 KYC.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0025 | 3 | todo | devops | Postgres `tenant_registrations`, `brands` stub, `tenants.status` + `pending_kyc` |
| TASK-0026 | 5 | todo | dev | Registration store + `POST /api/public/tenant/register` + validation |
| TASK-0027 | 3 | todo | dev | Auth integration — JWT for new tenant_admin; pending tenant resolution |
| TASK-0028 | 2 | todo | dev | Platform admin `GET /api/platform/tenants` list + status filter |
| TASK-0029 | 3 | todo | dev | Tenant web app — `/tenant/register` form + confirmation |

**Committed:** 16 points · **Completed:** 0 points · **Velocity target:** 16

## Scope boundary

**In**
- Public registration API (no Bearer required; basic rate-limit stub)
- `tenant_registrations` row per signup (audit + Sprint 7 KYC input)
- Default `brands` row per tenant (`name` = company name)
- `apps/tenant-web` static build served at `/tenant/`
- Platform tenant list for `platform_admin`
- `sprint-tech-specs` design pack **before** TASK-0026 implementation

**Out** (→ backlog / later sprints)
- KYC approve/reject, document upload (Sprint 7)
- Email verification / invitations
- Package purchase on signup (Sprint 9)
- Tenant admin KM/settings/embed (Sprints 14–17)
- Customer registration (Sprint 19)
- HeyGen / LiveAvatar lip-sync (deferred)

## Feature

- [FEAT-0006 — Tenant self-registration](../01-features/FEAT-0006-tenant-register.md)

## Design pack (`sprint-tech-specs`)

| Artifact | Path | Status |
| --- | --- | --- |
| Tenant register deep spec | [11-tenant-register-spec.md](../02-design/11-tenant-register-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §18–21 | `approved` |
| ER diagram | [03-er-diagram.md](../02-design/03-er-diagram.md) | `approved` |
| API spec | [04-api-spec.md](../02-design/04-api-spec.md) § Tenant registration | `approved` |
| UX/UI ASCII | [05-ux-ui.md](../02-design/05-ux-ui.md) § T1–T3, P11 | `approved` |

> Design pack approved — DEV may start **TASK-0025** (schema) then **TASK-0026**.

## Verification

```bash
make build && make test
make infra-init && make restart
open http://localhost:8091/tenant/register
# Register acme → login as tenant_admin → GET /api/auth/me
curl -H "Authorization: Bearer $PLATFORM_TOKEN" http://localhost:8091/api/platform/tenants?status=pending_kyc
```

- Manual: `docs/sdlc/06-manual-tests/SPRINT-006-manual.md` (Tester, at VERIFY)

## Risks

| Risk | Mitigation |
| --- | --- |
| Slug squatting / abuse | Slug format validation; rate-limit stub; KYC gate in Sprint 7 |
| `tenants.status` migration | `ALTER` check constraint idempotently in `ensureSchema` |
| Third Svelte app build time | Reuse platform-admin patterns; `make tenant-web` target |
| Pending tenant using product | Document `pending_kyc` limitations until Sprint 7; no entitlement auto-assign |

## Definition of done

- Design pack approved · code reviewed · ACs verified by Tester · `make build` · tag **v0.7.0** at sprint close