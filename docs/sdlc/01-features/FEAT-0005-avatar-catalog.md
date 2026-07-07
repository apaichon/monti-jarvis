# Feature: Platform Avatar Catalog + Tenant Assignment   (FEAT-0005)
**Sprint:** SPRINT-005   **Owner:** DEV   **Status:** in_progress

## Problem

AI avatars (Ava, Max, Luna, Neo) are hardcoded in `internal/workforce/workforce.go`. Platform operators cannot manage the catalog or control which avatars each tenant may use. Package rules already define `max_ai_employees`, but there is no data model to enforce assignment.

## Scope

In:
- Postgres `ai_avatars` (platform catalog) + `tenant_avatar_assignments`
- Dev seeds: migrate four prototype avatars; assign subset to tenant `demo`
- Platform-admin CRUD API + tenant assign/revoke/list APIs
- `GET /api/workforce` resolves active avatars for tenant (DB-first, static fallback)
- Enforce `max_ai_employees` from active entitlement on tenant assign
- Platform admin portal: avatars list/create/edit + tenant assignment UI (`/admin`)
- Design pack via `sprint-tech-specs` (workflow, ER, API, UX, `10-avatars-spec.md`)

Out:
- Full `ai_employee_versions`, languages, tools, guardrails (blueprint §16.3 — Sprint 21)
- MinIO image upload pipeline (URL field only; upload sprint later)
- Customer portal UI redesign (continues using `/api/workforce`)
- Tenant admin self-service avatar picker (Sprint 15+)
- Quota enforcement on live calls (Sprint 13)

## Acceptance criteria

1. `ensureSchema` creates `ai_avatars` and `tenant_avatar_assignments` with audit columns; seeds four avatars matching current workforce metadata.
2. `platform_admin` CRUD avatars via `/api/platform/avatars*`; `tenant_admin` receives `403` on platform routes.
3. `platform_admin` assign/revoke/list tenant avatars; assign blocked when count exceeds entitlement `max_ai_employees`.
4. `GET /api/workforce` returns DB-assigned active avatars for resolved tenant; falls back to static catalog when none assigned.
5. Portal: `/admin/avatars` list/create/edit; `/admin/tenants/{id}/avatars` assign UI.
6. `go test ./...`; customer portal `/` unchanged with `AUTH_DISABLED=true`.

## Test notes

- API: CRUD, RBAC, assignment limit, workforce resolver
- Browser UAT: login → avatars → assign demo → verify customer portal agent list

## Links

- Sprint: [SPRINT-005](../03-sprints/SPRINT-005.md)
- Roadmap: Sprint 5 · Phase B
- Depends on: [FEAT-0003](FEAT-0003-auth-rbac.md), [FEAT-0004](FEAT-0004-packages-entitlements.md)