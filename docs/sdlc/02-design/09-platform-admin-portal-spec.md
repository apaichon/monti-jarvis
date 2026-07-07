---
id: DES-0009
title: Platform Admin Portal Specification
status: approved
updated: 2026-07-08
sprint: SPRINT-004
owner: SA
---

# Platform Admin Portal — Design Spec

**Sprint:** SPRINT-004 · **Release target:** v0.5.0  
**Feature:** [FEAT-0004](../01-features/FEAT-0004-packages-entitlements.md)  
**Depends on:** [06-auth-spec.md](06-auth-spec.md), [08-packages-spec.md](08-packages-spec.md)

## 1. Goals

Deliver a **SvelteKit platform admin portal** at `/admin` so operators can log in, view profile, and manage commercial packages + tenant entitlements without curl.

## 2. Non-goals (Sprint 4)

- Tenant admin portal (Sprint 15+)
- Customer login UI (Sprints 19–20)
- Full tenant list / KYC / billing screens (Sprints 6–12)
- Password reset, MFA, OAuth
- Mobile-native app (responsive web only)

## 3. Deployment

| Item | Value |
| --- | --- |
| App path | `apps/platform-admin-web/` |
| Build output | `apps/platform-admin-web/build` |
| Served at | `http://localhost:8091/admin/` |
| Go handler | `internal/platformweb` (mirror `customerweb` SPA fallback) |
| Makefile | `make platform-admin-web`, included in `make build` |

Customer portal `/` unchanged.

## 4. Auth UX

| Screen | Route | API |
| --- | --- | --- |
| Login | `/admin/login` | `POST /api/auth/login` |
| Logout | header action | `POST /api/auth/logout` + clear tokens |
| Profile | `/admin/profile` | `GET /api/auth/me` |

**Token storage:** `sessionStorage` — `access_token`, `refresh_token` (dev; httpOnly cookie deferred).

**Route guard:** All `/admin/*` except `/admin/login` require `platform_admin` role from `/api/auth/me`; `tenant_admin` sees “wrong portal” message; unauthenticated → redirect login.

**401 handling:** Clear storage, redirect `/admin/login?next=…`.

## 5. Packages UX

| Screen | Route | API |
| --- | --- | --- |
| Package list | `/admin/packages` | `GET /api/platform/packages` |
| Create package | `/admin/packages/new` | `GET /api/platform/rule-schemas`, `POST /api/platform/packages` |
| Edit package | `/admin/packages/[id]` | `GET/PUT /api/platform/packages/{id}` |
| Archive | list/detail action | `DELETE /api/platform/packages/{id}` |
| Tenant entitlement | `/admin/tenants/[id]/entitlement` | `GET/POST/DELETE /api/platform/tenants/{id}/entitlement` |

**Rules form:** Load active schema from `rule-schemas`; render fields from `fields` jsonb (int/bool inputs); submit as `rules` + `rules_schema_id`.

**Dev shortcut:** Link from packages list to assign `demo` tenant (seed tenant id from env or hardcode `demo`).

## 6. UX/UI screen specs (canonical)

Full ASCII wireframes, zone→API maps, states, and component paths live in **[05-ux-ui.md](05-ux-ui.md) § Platform Admin Portal**:

| Screen | Doc section |
| --- | --- |
| Login | **P0** — `/admin/login` |
| App shell | **P1** — shared nav + logout |
| Profile | **P2** — `/admin/profile` |
| Packages list | **P3** — `/admin/packages` |
| Package create | **P4** — `/admin/packages/new` |
| Package edit | **P5** — `/admin/packages/[id]` |
| Tenant entitlement | **P6** — `/admin/tenants/[id]/entitlement` |

## 7. Stack

- SvelteKit 2 + Svelte 5 + Tailwind (match `customer-web` toolchain)
- Shared design tokens from [05-ux-ui.md](05-ux-ui.md) (dark admin variant)
- `src/lib/api/auth.ts`, `packages.ts` — fetch with Bearer from sessionStorage

## 8. Verification

```bash
make platform-admin-web && make restart
# AUTH_DISABLED=false
open http://localhost:8091/admin/login
# platform@monti.local / monti-platform → packages list → edit Starter → assign demo
```

## 9. Related artifacts

| Artifact | Path |
| --- | --- |
| UX wireframes | [05-ux-ui.md](05-ux-ui.md) § Platform Admin Portal |
| Workflows | [02-workflow.md](02-workflow.md) §12–13 |
| API | [04-api-spec.md](04-api-spec.md) |
| Packages domain | [08-packages-spec.md](08-packages-spec.md) |