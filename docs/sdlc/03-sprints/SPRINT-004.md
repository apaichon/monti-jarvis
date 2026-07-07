---
id: SPRINT-004
status: in_progress
start: 2026-07-07
end: 2026-07-21
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
| TASK-0014 | 3 | todo | devops | Postgres packages schema + dev seeds + provider catalog stub |
| TASK-0015 | 5 | todo | dev | Package catalog service + platform CRUD API |
| TASK-0016 | 4 | todo | dev | Tenant entitlement assign/revoke + read APIs |
| TASK-0017 | 3 | todo | dev | Entitlement resolver + Redis read-through cache |
| TASK-0018 | 5 | todo | dev | Platform admin portal — login, logout, profile, shell |
| TASK-0019 | 5 | todo | dev | Platform admin portal — packages + entitlement UI |

**Committed:** 25 points · **Completed:** 0 points · **Velocity target:** 16 (stretch — portal adds 10 pts)

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
| Packages | [08-packages-spec.md](../02-design/08-packages-spec.md) | `approved` |
| Platform portal | [09-platform-admin-portal-spec.md](../02-design/09-platform-admin-portal-spec.md) | `approved` |
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

- Manual: `docs/sdlc/06-manual-tests/SPRINT-004-manual.md` (Tester)

## Risks

| Risk | Mitigation |
| --- | --- |
| Sprint points above velocity (25 vs 16) | Portal split TASK-0018/0019; API tasks parallelizable |
| Package schema churn | `rules` jsonb + `package_rule_schemas` versions |
| Token in sessionStorage (XSS) | Dev-only; document; httpOnly cookies in hardening sprint |
| Entitlement cache stale | Redis `DEL` on every write path |

## Definition of done

- Code reviewed · ACs verified by Tester · portal + API UAT · `make build` · tag v0.5.0 at sprint close