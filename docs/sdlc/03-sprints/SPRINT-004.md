---
id: SPRINT-004
status: completed
start: 2026-07-07
end: 2026-07-21
closed: 2026-07-07
updated: 2026-07-07
release: v0.5.0
goal: "Platform Admin: Portal + Packages — login/profile UI and commercial catalog with entitlements."
roadmap_sprint: 4
platform: Platform Admin
depends_on: [SPRINT-003]
release_target: v0.5.0
---

# SPRINT-004 — Platform Admin: Portal + Packages

## Goal

Ship the **platform admin Svelte portal** (`/admin`) with **login, logout, profile**, and **package + entitlement management**, backed by JSONB rules APIs and Redis entitlement cache.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0014 | 3 | completed | devops | Postgres packages schema + dev seeds + provider catalog stub |
| TASK-0015 | 5 | completed | dev | Package catalog service + platform CRUD API |
| TASK-0016 | 4 | completed | dev | Tenant entitlement assign/revoke + read APIs |
| TASK-0017 | 3 | completed | dev | Entitlement resolver + Redis read-through cache |
| TASK-0018 | 5 | completed | dev | Platform admin portal — login, logout, profile, shell |
| TASK-0019 | 5 | completed | dev | Platform admin portal — packages + entitlement UI |

**Committed:** 25 points · **Completed:** 25 points · **Velocity:** 16 (stretch)

## Shipped (v0.5.0)

- `apps/platform-admin-web` at `/admin` — login, profile, packages list/create/edit, tenant entitlement UI
- Postgres: `package_rule_schemas`, `packages`, `package_limits`, `tenant_entitlements` + seeds (`rules-v1`, Starter/Pro/Enterprise, demo → Starter)
- Platform APIs: rule-schemas, packages CRUD, tenant entitlement assign/revoke, `GET /api/entitlements/me`
- `internal/packages` JSONB rules validation against schema `fields`
- `internal/entitlements` resolver + Redis cache (`monti_jarvis:entitlement:{tenant_id}`)
- `make platform-admin-web` · CORS PUT/DELETE · `/api/infra` `entitlement_cache` status
- Fix: audit-column DDL commas so packages schema bootstraps on startup

## Scope boundary

**In**
- `apps/platform-admin-web` at `/admin` (SvelteKit + Tailwind)
- Auth screens: login, logout, profile (`/api/auth/*`)
- Package catalog UI + tenant entitlement assign/revoke (demo tenant min.)
- Backend: `package_rule_schemas`, packages APIs, entitlements resolver, Redis cache
- Provider catalog stub seeds
- Design: `08-packages-spec`, `09-platform-admin-portal-spec`, UX/workflow/API updates

**Out** (→ backlog / later sprints)
- Tenant admin portal (Sprint 15+)
- Payment, checkout, billing (Sprints 8–12)
- Quota/rate-limit enforcement on live paths (Sprint 13)
- Login rate limit (auth-cache Phase F)
- NATS entitlement events
- Password reset / MFA

## Feature

- [FEAT-0004 — Packages and Platform Admin Portal](../01-features/FEAT-0004-packages-entitlements.md)

## Design pack (`sprint-tech-specs`)

| Artifact | Path | Status |
| --- | --- | --- |
| Packages | [08-packages-spec.md](../02-design/08-packages-spec.md) | `shipped` |
| Platform portal | [09-platform-admin-portal-spec.md](../02-design/09-platform-admin-portal-spec.md) | `shipped` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §9–13 | `approved` |
| ER diagram | [03-er-diagram.md](../02-design/03-er-diagram.md) | `approved` |
| API spec | [04-api-spec.md](../02-design/04-api-spec.md) | `approved` |
| UX/UI ASCII | [05-ux-ui.md](../02-design/05-ux-ui.md) § P0–P6 screens | `approved` |
| Auth (prior) | [06-auth-spec.md](../02-design/06-auth-spec.md) | `shipped` |

## Verification

```bash
make build && make test
make infra-init && make restart
# AUTH_DISABLED=false in infra/.env.dev
open http://localhost:8091/admin/login
# platform@monti.local / monti-platform → Packages → Profile → Logout
```

- Manual: [SPRINT-004 UAT](../06-manual-tests/SPRINT-004-manual.md)

## Definition of done

- [x] Code shipped · ACs met · portal + API UAT · `make build` · tag v0.5.0