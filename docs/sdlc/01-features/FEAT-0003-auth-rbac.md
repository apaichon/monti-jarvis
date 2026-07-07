# Feature: Auth and RBAC Skeleton   (FEAT-0003)
**Sprint:** SPRINT-003   **Owner:** DEV

## Problem

Sprints 1–2 run in **no-auth** mode with `DEMO_TENANT_ID=demo` and a trusted `X-Tenant-Id` header. Multi-tenant SaaS (packages, billing, tenant KM admin in later sprints) requires **authenticated actors** and **role-based access** before any commerce or admin UI ships.

## Scope

In:
- Postgres `tenants`, `users`, `user_roles` (platform / tenant / customer skeleton)
- JWT access + refresh tokens (HS256 dev; env-configured secret)
- Auth API: login, refresh, logout, `GET /api/auth/me`
- Go middleware: resolve `tenant_id`, `user_id`, `role` from Bearer token
- Protect KM admin routes (`/api/km/*` write/seed) for `tenant_admin` and `platform_admin`
- `AUTH_DISABLED=true` dev bypass preserving current no-auth demo behavior
- Audit-friendly auth events logged to server (login success/failure)

Out:
- Customer self-registration and KYC (Sprint 19–20)
- OAuth / SAML / social login
- Full platform-admin UI (API only this sprint)
- Packages, quotas, billing (Sprint 4+)
- Migrate from `net/http` to Fiber (deferred)
- Per-route NATS authz on subjects (enforce in Go handlers only)

## Acceptance criteria

1. Operator can create a dev tenant and tenant-admin user via seed/migration; login returns access + refresh JWT.
2. Protected route rejects missing/invalid token with `401`; wrong role returns `403`.
3. When `AUTH_DISABLED=false`, KM ingest/reset/seed requires `tenant_admin` or `platform_admin`; reads use `tenant_id` from token not header spoofing.
4. When `AUTH_DISABLED=true` (default dev), Sprint 1–2 customer portal and public chat/voice work unchanged with `demo` tenant.
5. Refresh token rotation invalidates old refresh on logout.
6. `GET /api/auth/me` returns user id, email, role, tenant_id for valid access token.

## Test notes

- Unit: JWT issue/parse, password verify, middleware role checks
- Integration: login → call protected KM upload → 201; no token → 401
- Manual: `curl` login flow in sprint verification; portal smoke with `AUTH_DISABLED=true`
- Regression: `go test ./...`, public `/api/chat` and `/ws/voice` with auth disabled

## Dependencies

- Blueprint §11.3, §12.1, multi-tenant isolation rules
- Sprint 2: KM APIs, `DEMO_TENANT_ID`, Postgres `callcenter` schema
- Design updates: `02-design/api-spec.md`, `er-diagram.md`, `architecture.md`