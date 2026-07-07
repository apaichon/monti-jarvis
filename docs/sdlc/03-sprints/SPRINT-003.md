---
id: SPRINT-003
status: completed
start: 2026-07-07
end: 2026-07-07
closed: 2026-07-07
updated: 2026-07-07
release: v0.4.0
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
| TASK-0010 | 3 | completed | devops | Postgres tenants, users, roles schema + dev seed |
| TASK-0011 | 5 | completed | dev | JWT auth service + `/api/auth/*` endpoints |
| TASK-0012 | 5 | completed | dev | Auth middleware + RBAC on protected routes |
| TASK-0013 | 3 | completed | dev | Tenant resolution in KM/calls; `AUTH_DISABLED` dev bypass |

**Committed:** 16 points · **Completed:** 16 points · **Velocity:** 16

## Shipped (v0.4.0)

- JWT login, refresh, logout, `GET /api/auth/me` with HS256 access + refresh rotation
- Roles: `platform_admin`, `tenant_admin`, `customer` (stub)
- RBAC on KM writes (`tenant_admin`+) and `POST /api/km/seed` (`platform_admin`)
- `AUTH_DISABLED=true` default — Sprint 1–2 customer demo unchanged
- Redis read-through cache, write-behind refresh persist, `jti` denylist
- NATS JetStream `MONTI_AUTH` + ClickHouse `auth_events` ingestion
- Auditable columns on all Postgres/ClickHouse tables + `make db-migrate`
- Auth-aware `make km-seed` · `LOCAL-DEV.md` auth section

## Feature

- [FEAT-0003 — Auth and RBAC](../01-features/FEAT-0003-auth-rbac.md)
- Design: [auth-spec](../02-design/auth-spec.md) · [auth-cache-events-spec](../02-design/auth-cache-events-spec.md)

## Verification

```bash
make test && make build
make infra-init && make restart
curl -X POST http://localhost:8091/api/auth/login \
  -H 'content-type: application/json' \
  -d '{"email":"admin@demo.local","password":"demo-admin"}'
make km-seed   # with AUTH_DISABLED=false uses platform login
```

- Manual: [SPRINT-003 UAT](../06-manual-tests/SPRINT-003-manual.md)

## Definition of done

- [x] Code shipped · ACs met · auth in `infra-init` · `AUTH_DISABLED` documented · tag v0.4.0