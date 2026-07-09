---
id: SPRINT-006
status: completed
start: 2026-07-08
end: 2026-07-21
closed: 2026-07-09
updated: 2026-07-09
release: v0.7.0
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
| TASK-0025 | 3 | completed | devops | Postgres `tenant_registrations`, `brands` stub, `tenants.status` + `pending_kyc` |
| TASK-0026 | 5 | completed | dev | Registration store + `POST /api/public/tenant/register` + validation |
| TASK-0027 | 3 | completed | dev | Auth integration — JWT for new tenant_admin; pending tenant resolution |
| TASK-0028 | 2 | completed | dev | Platform admin `GET /api/platform/tenants` list + status filter |
| TASK-0029 | 3 | completed | dev | Tenant web app — `/tenant/register` form + confirmation |

**Committed:** 16 points · **Completed:** 16 points · **Velocity:** 16

## Shipped (v0.7.0)

- Postgres: `tenant_registrations`, `brands` stub, `tenants.status` includes `pending_kyc`
- Public API: `POST /api/public/tenant/register`, email verify, OAuth (Google/GitHub) signup
- Auth: JWT for new `tenant_admin`; email-verification gate on login; `pending_kyc` KM write block
- Tenant KYC backoffice: contact, photo, document upload + submit for platform review
- `apps/tenant-web` at `/tenant/` — register, login, OAuth, check-email, verify, backoffice
- Platform admin: `GET /api/platform/tenants` + `/admin/tenants` list UI
- Resend email integration for verification links
- **UI:** shadcn-style feedback dialog for success/error across tenant + platform admin portals
- **Avatars:** static portrait images only — **no HeyGen / LiveAvatar** in this sprint
- Dev: `scripts/dev-hosts.sh` + `make dev-hosts` for `monti-jarvis-dev.local`

## Scope boundary

**In**
- Public registration API (no Bearer required; basic rate-limit stub)
- `tenant_registrations` row per signup (audit + Sprint 7 KYC input)
- Default `brands` row per tenant (`name` = company name)
- `apps/tenant-web` static build served at `/tenant/`
- Platform tenant list for `platform_admin`
- `sprint-tech-specs` design pack **before** TASK-0026 implementation
- Static avatar portraits (MinIO / asset URLs) — voice via existing Gemini relay

**Out** (→ backlog / later sprints)
- KYC approve/reject by platform admin (Sprint 7)
- HeyGen / LiveAvatar / animated lip-sync avatars (not used; deferred indefinitely)
- Package purchase on signup (Sprint 9)
- Tenant admin KM/settings/embed (Sprints 14–17)
- Customer registration (Sprint 19)

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

## Verification

```bash
make build && make test
make infra-init && make restart
open http://monti-jarvis-dev.local:8091/tenant/register
# Register acme → verify email → login as tenant_admin → KYC backoffice
curl -H "Authorization: Bearer $PLATFORM_TOKEN" http://localhost:8091/api/platform/tenants?status=pending_kyc
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Slug squatting / abuse | Slug format validation; rate-limit stub; KYC gate in Sprint 7 |
| `tenants.status` migration | `ALTER` check constraint idempotently in `ensureSchema` |
| Third Svelte app build time | Reuse platform-admin patterns; `make tenant-web` target |
| Pending tenant using product | Document `pending_kyc` limitations until Sprint 7; no entitlement auto-assign |

## Definition of done

- Design pack approved · code reviewed · ACs verified · `make build` · tag **v0.7.0** at sprint close ✅