---
id: MANUAL-SPRINT-003
sprint: SPRINT-003
release_target: v0.4.0
updated: 2026-07-07
---

# SPRINT-003 Manual UAT — Backend Auth (v0.4.0)

## Prerequisites

- [ ] `make infra-init` · Postgres, Redis, ClickHouse, NATS up
- [ ] `infra/.env.dev` has `AUTH_DISABLED=false`, `JWT_SECRET` (≥32 bytes), auth cache flags
- [ ] `make restart`

## S1 — Auth disabled regression

Set `AUTH_DISABLED=true`, restart.

- [ ] `curl http://localhost:8091/healthz` → `auth_disabled: true`
- [ ] `curl -X POST http://localhost:8091/api/chat -d '{"agent_id":"ava","message":"hi","history":[]}'` → 200
- [ ] `make km-seed` → 200 without login

## S2 — Login and RBAC

Set `AUTH_DISABLED=false`, restart.

- [ ] Login tenant admin → 200 with `access_token`, `refresh_token`
- [ ] `GET /api/auth/me` with Bearer → profile `tenant_admin`, `tenant_id: demo`
- [ ] `POST /api/km/agents/ava/documents` without token → 401
- [ ] Same with Bearer → 201 (or 502 if no Gemini for embed)
- [ ] `POST /api/km/seed` as tenant admin → 403
- [ ] Login `platform@monti.local` / `monti-platform` → `make km-seed` → 200

## S3 — Cache and events

- [ ] `GET /api/infra` → `auth_cache: ok`, `auth_events: ok`, `auth_write_behind_lag: 0`
- [ ] After login, ClickHouse `auth_events` has `auth.user.logged_in` row
- [ ] Logout with refresh + Bearer → subsequent `/me` with old access → 401 (jti denylist)

## Sign-off

| Role | Date | Result |
| --- | --- | --- |
| Dev | 2026-07-07 | Pass (v0.4.0) |
| Tester | | |