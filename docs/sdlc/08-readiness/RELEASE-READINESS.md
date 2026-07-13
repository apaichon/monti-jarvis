---
id: READINESS-RELEASE
status: completed
updated: 2026-07-13
current_sprint: SPRINT-019
release_target: v2.0.0
release: v2.0.0
---

# Release Readiness Checklist

Use this checklist before **demo**, **sprint sign-off**, or **`release-cut`** (git tag).

## Production customer launch gate (post S16 / before open traffic)

> After **tenant customer-user authentication** is integrated (SPRINT-019–020), and **before production launch to end customers**, sign off that **rate limit and quota management work** under multi-user load.

- [ ] S13 package quotas enforce (monthly minutes, concurrent, KM, avatars, voice/RAG flags)
- [ ] S13 rate limits enforce under concurrent chat/voice/KM
- [ ] S16 daily + per-call operational caps enforce under package ceiling
- [ ] Tenant isolation: customer of A cannot consume B’s quota
- [ ] Production `QUOTA_*` / `RATE_LIMIT_*` env flags reviewed (fail-open vs fail-closed)
- [ ] Load or soak test notes attached (or UAT multi-session evidence)

**Owner:** DevOps + Tester · **Blocked by:** customer auth (SPRINT-020) if not yet shipped · **Recorded in:** [SPRINT-016](../03-sprints/SPRINT-016.md) shipped notes

This gate is not a blocker for the SPRINT-019 data/import release because SPRINT-019 does not open customer-authenticated production traffic.

## A. Code & build

- [x] Branch is clean or PR merged to `main`
- [x] `go test ./...` passes
- [x] `make build` succeeds (Svelte portal + Go binary)
- [x] No known P0/P1 defects open for this sprint

## B. Infrastructure

- [x] Shared containers running: Postgres, Redis, MinIO
- [x] Sprint 2+: ClickHouse running
- [x] `make infra-init` applied (schema + MinIO bucket)
- [x] `make infra-check` — all required services report healthy
- [x] Monti compose up (`monti-nats`, `monti-livekit`)

## C. Configuration

- [x] `infra/.env.dev` exists (from `.env.dev.example`)
- [x] `GEMINI_API_KEY` configured for local runtime
- [x] Sprint 2+: `CLICKHOUSE_URL` and `CLICKHOUSE_DB=monti_jarvis`
- [x] `DEMO_TENANT_ID=demo` for single-tenant demos

## D. Runtime smoke (5 min)

```bash
make start
make status
curl -fsS http://localhost:8091/healthz
curl -fsS http://localhost:8091/api/infra
curl -fsS http://localhost:8091/api/workforce
```

- [x] `/healthz` → `"ok": true`, sprint flag matches SPRINT-019
- [x] `/api/infra` → postgres, redis, minio `ok`; clickhouse `ok` (Sprint 2+)
- [x] `/api/workforce` → available tenant/demo agents

## E. Sprint-specific data

### SPRINT-001 (v0.2.0)

- [ ] Portal loads at http://localhost:8091
- [ ] Agent selection + text chat + voice smoke (see [SPRINT-001 manual](../06-manual-tests/SPRINT-001-manual.md))

### SPRINT-002 (v0.3.0 target)

- [ ] `make km-seed` or `POST /api/km/seed` succeeds
- [ ] `GET /api/km/agents/max` shows documents and chunks
- [ ] Billing RAG curl returns `sources[]` (see [KM_SETUP](../../KM_SETUP.md))
- [ ] Citation chips visible in portal after grounded question

### SPRINT-019 (v2.0.0)

- [x] Tenant customer CRUD and deactivate pass.
- [x] CSV dry-run validates without writes.
- [x] CSV commit creates/updates valid rows and rejects invalid rows.
- [x] Repeat import is idempotent for `(source, external_id)`.
- [x] Domain defaults and explicit tier/group precedence pass.
- [x] Cross-tenant customer, import, and rule ids return 404.
- [x] Customer auth/token endpoints remain unavailable until SPRINT-020.

## F. Documentation

- [x] Sprint doc status accurate (`docs/sdlc/03-sprints/SPRINT-NNN.md`)
- [x] Task statuses match implementation (`docs/sdlc/04-tasks/`)
- [x] Manual test checklist completed ([`06-manual-tests/`](../06-manual-tests/))
- [x] Test matrix scenarios for this sprint marked pass ([`05-test-scenarios/TEST-MATRIX.md`](../05-test-scenarios/TEST-MATRIX.md))
- [x] `AGENTS.md` current sprint line updated
- [x] API spec matches shipped routes ([`02-design/04-api-spec.md`](../02-design/04-api-spec.md))

## G. Sign-off

| Role | Name | Date | Notes |
| --- | --- | --- | --- |
| Dev | Codex release verification | 2026-07-13 | Implementation + unit tests |
| Tester | Codex release verification | 2026-07-13 | Manual UAT green |
| PM | User-authorized release close | 2026-07-13 | ACs accepted |
| DevOps | Codex release verification | 2026-07-13 | Infra + deploy verified |

## H. Release-cut (PM + DevOps)

After sections A–G are green:

```bash
# SPRINT-019 release
git tag -a v2.0.0 -m "v2.0.0 - SPRINT-019 customer account import and integration"
git push origin v2.0.0
```

- [x] Tag pushed to `origin`
- [x] Sprint marked `completed` in `03-sprints/`
- [x] `_velocity.json` updated
- [x] ROADMAP current sprint pointer advanced

## Quick demo script (stakeholder, ~10 min)

1. `make up && make km-seed`
2. Open portal → select **Max** → **Billing** tab
3. Ask *"When are invoices due?"* → show citation chips
4. Start voice call with **Ava** → one general question
5. Show `/legacy/` for visual continuity
6. `make down` when finished
