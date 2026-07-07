# Feature: Platform Admin Portal + Packages   (FEAT-0004)
**Sprint:** SPRINT-004   **Owner:** DEV   **Status:** shipped v0.5.0

## Problem

Sprint 3 delivered JWT auth and RBAC, but operators have no UI and tenants share implicit limits. Platform SaaS needs a **portal** for login/profile and a **package catalog** with per-tenant entitlements before commerce (Sprint 9) and quotas (Sprint 13).

## Scope

In:
- **Platform admin portal** `apps/platform-admin-web` at `/admin` — login, logout, profile, packages UI
- Postgres `package_rule_schemas`, `packages`, `package_limits` (`rules` jsonb), `tenant_entitlements`
- Platform-admin CRUD API + entitlement assign/revoke
- `GET /api/entitlements/me` for `tenant_admin` (API; no tenant UI this sprint)
- `internal/entitlements` resolver + Redis cache
- Dev seeds: `rules-v1`, Starter/Pro/Enterprise, `demo` → Starter
- Design: [08-packages-spec.md](../02-design/08-packages-spec.md), [09-platform-admin-portal-spec.md](../02-design/09-platform-admin-portal-spec.md), linked `02`–`05` artifacts

Out:
- Tenant admin Svelte portal
- Payment, checkout, invoicing (Sprints 8–12)
- Quota enforcement on call/KM paths (Sprint 13)
- Self-service package purchase (Sprint 9)
- Login rate limit (deferred)
- NATS entitlement events

## Acceptance criteria

1. `make infra-init` creates package tables + `rules-v1` seed + three packages; `demo` → `pkg-starter`.
2. `platform_admin` CRUD packages via API; `tenant_admin` receives `403` on platform routes.
3. `platform_admin` assign/revoke tenant entitlement via API and portal UI.
4. Portal: login at `/admin/login`, logout clears session, profile shows `/api/auth/me` fields.
5. Portal: list/create/edit/archive packages; rules form driven by `rule-schemas`.
6. `entitlements.Resolve` + Redis cache invalidation on assign/revoke.
7. `go test ./...`; customer portal `/` unchanged with `AUTH_DISABLED=true`.

## Test notes

- Browser UAT: login → packages → assign demo → profile → logout
- API integration tests for CRUD, RBAC, cache
- Regression: customer portal smoke

## Dependencies

- FEAT-0003 auth/RBAC
- Blueprint §5.3 Platform Admin Portal
- `internal/platformweb`, `apps/platform-admin-web`, `internal/packages`, `internal/entitlements`