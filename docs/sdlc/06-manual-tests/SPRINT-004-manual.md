---
id: MANUAL-SPRINT-004
sprint: SPRINT-004
release_target: v0.5.0
updated: 2026-07-07
---

# SPRINT-004 Manual UAT — Platform Admin Portal + Packages (v0.5.0)

## Prerequisites

- [ ] `make infra-init` · Postgres, Redis up
- [ ] `infra/.env.dev` has `AUTH_DISABLED=false`, `JWT_SECRET` (≥32 bytes)
- [ ] `make build && make restart`

## S1 — Schema and seeds

- [ ] Server log has no `postgres schema` warnings on startup
- [ ] `psql` — tables exist: `callcenter.packages`, `package_limits`, `tenant_entitlements`
- [ ] `GET /api/infra` → `entitlement_cache: ok` (when Redis up)

## S2 — Platform API (curl)

Login platform admin:

```bash
TOKEN=$(curl -sS -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)
```

- [X] `GET /api/platform/packages` with Bearer → 200, starter/pro/enterprise
- [X] `GET /api/platform/rule-schemas` → `rules-v1`
- [X] `GET /api/platform/tenants/demo/entitlement` → active Starter
- [X] Login `admin@demo.local` / `demo-admin` → `GET /api/platform/packages` → 403
- [X] Tenant admin `GET /api/entitlements/me` → 200 with demo entitlement

## S3 — Portal UI (`/admin`)

Open `http://localhost:8091/admin/login`

- [X] Login `platform@monti.local` / `monti-platform` → redirects to `/admin/packages`
- [X] Packages table shows starter, pro, enterprise
- [X] Profile → email, role `platform_admin`, user id
- [X] Assign demo → `/admin/tenants/demo/entitlement` shows current package
- [X] Logout → returns to login; packages page requires re-auth
- [X] `admin@demo.local` login → rejected (tenant admin not allowed on portal)

## S4 — Customer portal regression

Set `AUTH_DISABLED=true`, restart.

- [X] `http://localhost:8091/` loads customer portal without login
- [X] `POST /api/chat` works without Bearer

## Sign-off

| Role | Date | Result |
| --- | --- | --- |
| Dev | 2026-07-07 | Pass (v0.5.0) |
| Tester | 2026-07-10| Pass (v0.6.0)|