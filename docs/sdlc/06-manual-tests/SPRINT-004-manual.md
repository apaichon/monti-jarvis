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

- [ ] `GET /api/platform/packages` with Bearer → 200, starter/pro/enterprise
- [ ] `GET /api/platform/rule-schemas` → `rules-v1`
- [ ] `GET /api/platform/tenants/demo/entitlement` → active Starter
- [ ] Login `admin@demo.local` / `demo-admin` → `GET /api/platform/packages` → 403
- [ ] Tenant admin `GET /api/entitlements/me` → 200 with demo entitlement

## S3 — Portal UI (`/admin`)

Open `http://localhost:8091/admin/login`

- [ ] Login `platform@monti.local` / `monti-platform` → redirects to `/admin/packages`
- [ ] Packages table shows starter, pro, enterprise
- [ ] Profile → email, role `platform_admin`, user id
- [ ] Assign demo → `/admin/tenants/demo/entitlement` shows current package
- [ ] Logout → returns to login; packages page requires re-auth
- [ ] `admin@demo.local` login → rejected (tenant admin not allowed on portal)

## S4 — Customer portal regression

Set `AUTH_DISABLED=true`, restart.

- [ ] `http://localhost:8091/` loads customer portal without login
- [ ] `POST /api/chat` works without Bearer

## Sign-off

| Role | Date | Result |
| --- | --- | --- |
| Dev | 2026-07-07 | Pass (v0.5.0) |
| Tester | | |