---
id: SPRINT-003
status: in_progress
start: 2026-07-07
end: 2026-07-21
updated: 2026-07-07
goal: "Backend: Auth — JWT sessions, tenant model, and RBAC skeleton for protected APIs."
roadmap_sprint: 3
platform: Backend
depends_on: [SPRINT-002]
release_target: v0.4.0
---

# SPRINT-003 — Backend: Auth

## Goal

Introduce **JWT authentication** and a **three-role RBAC skeleton** (platform / tenant / customer) so KM admin and future tenant APIs are tenant-safe, while the public customer portal keeps working in **auth-disabled dev mode**.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0010 | 3 | todo | devops | Postgres tenants, users, roles schema + dev seed |
| TASK-0011 | 5 | todo | dev | JWT auth service + `/api/auth/*` endpoints |
| TASK-0012 | 5 | todo | dev | Auth middleware + RBAC on protected routes |
| TASK-0013 | 3 | todo | dev | Tenant resolution in KM/calls; `AUTH_DISABLED` dev bypass |

**Committed:** 16 points

## Phase: design review (current)

**Do not implement** until design sign-off.

| Doc | Status |
| --- | --- |
| [auth-spec.md](../02-design/auth-spec.md) | `review_pending` — **primary review** |
| [api-spec.md](../02-design/api-spec.md) | `review_pending` |
| [er-diagram.md](../02-design/er-diagram.md) | `review_pending` |
| [architecture.md](../02-design/architecture.md) | `review_pending` |
| [workflow.md](../02-design/workflow.md) | `review_pending` |
| [ux-ui.md](../02-design/ux-ui.md) | `review_pending` |

Open questions: auth-spec §13 (km/seed role, public calls, refresh storage, platform tenant override).

**Build starts after:** approver checkboxes in auth-spec §Approver + tasks moved from `todo` → `in_progress`.

## Feature

- [FEAT-0003 — Auth and RBAC](../01-features/FEAT-0003-auth-rbac.md)
- Design updates: [api-spec](../02-design/api-spec.md) · [er-diagram](../02-design/er-diagram.md) · [architecture](../02-design/architecture.md)

## Carry-over from Sprint 2

- Customer portal stays **no-login** for inbound demo when `AUTH_DISABLED=true`.
- `demo` tenant and sample KB remain the default dev dataset.
- ClickHouse RAG and voice preload optimizations ship as-is; auth wraps APIs, not retrieval logic.
- Continue `net/http` monolith (`cmd/server`); Fiber migration out of scope.

## Scope boundary

**In scope:**

- Postgres tables: `tenants`, `users`, `user_roles`, `refresh_tokens`
- Password login (bcrypt) for platform_admin and tenant_admin dev users
- JWT access (short TTL) + refresh (longer TTL, rotatable)
- Middleware: `Authorization: Bearer <access>`
- Roles: `platform_admin`, `tenant_admin`, `customer` (customer role stub only)
- Protect: `POST /api/km/agents/*/documents`, `/reset`, `POST /api/km/seed`
- Env: `JWT_SECRET`, `AUTH_DISABLED` (default `true` for local demo)

**Out of scope:**

- Customer register/login UI (Sprint 19–20)
- Tenant registration wizard (Sprint 6)
- OAuth, MFA, password reset email
- Packages, quotas, rate limits (Sprint 4, 13)
- Encrypting JWT with asymmetric keys (HS256 dev only)
- Admin Svelte apps for user management

## Stack touchpoints

```text
internal/auth        JWT, login, middleware, RBAC
internal/store       users, tenants, refresh tokens
cmd/server           /api/auth/*, wrap KM handlers
infra/.env.example   JWT_SECRET, AUTH_DISABLED
scripts/infra-init   auth schema + seed users
docs/sdlc/02-design  api-spec, er-diagram updates
```

## Verification

```bash
make infra-init   # applies auth schema + seed
make test && make build && make start

# Auth disabled (default) — regression
curl -fsS http://localhost:8091/api/chat -H 'content-type: application/json' \
  -d '{"agent_id":"ava","message":"hello","history":[]}'
make km-seed

# Auth enabled
AUTH_DISABLED=false make restart
curl -X POST http://localhost:8091/api/auth/login \
  -H 'content-type: application/json' \
  -d '{"email":"admin@demo.local","password":"demo-admin"}'
# Use access_token on KM upload — expect 201; without token — 401
```

- `go test ./internal/auth/...`
- Manual: [SPRINT-003 UAT](../06-manual-tests/SPRINT-003-manual.md) (create at VERIFY)
- Readiness: [RELEASE-READINESS](../08-readiness/RELEASE-READINESS.md)

## Definition of done

- Code reviewed · ACs verified · auth migration in `infra-init` · docs/api-spec updated · `AUTH_DISABLED` documented in LOCAL-DEV

## Risks

- Breaking local demo if `AUTH_DISABLED` default flips to false — keep default `true`.
- Header `X-Tenant-Id` must not override JWT tenant when auth enabled.
- Refresh token storage: Postgres table vs Redis — start Postgres for auditability.