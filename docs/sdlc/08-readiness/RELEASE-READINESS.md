---
id: READINESS-RELEASE
status: completed
updated: 2026-07-16
current_sprint: SPRINT-028
release_target: v2.9.0
release: v2.9.0
---

# Release Readiness Checklist

Use this checklist before **demo**, **sprint sign-off**, or **`release-cut`** (git tag).

## Production customer launch gate (post S16 / before open traffic)

> After **tenant customer-user authentication** is integrated (SPRINT-019–020), and **before production launch to end customers**, sign off that **rate limit and quota management work** under multi-user load.

- [x] S13 package quotas enforce (monthly minutes, concurrent, KM, avatars, voice/RAG flags)
- [x] S13 rate limits enforce under tenant-scoped chat/voice/KM keys
- [x] S16 daily + per-call operational caps remain enforced under package ceiling
- [x] Tenant isolation: customer-auth tenant routing verified on Libra Tech tenant; cross-tenant UAT documented
- [x] Production `QUOTA_*` / `RATE_LIMIT_*` env flags reviewed for local release defaults
- [x] Load or soak test notes attached as SPRINT-020 manual checklist for target-environment re-run

**Owner:** DevOps + Tester · **Status:** local release gate accepted for v2.1.0; re-run target-environment multi-session test before broad customer traffic · **Recorded in:** [SPRINT-020 manual checklist](../06-manual-tests/SPRINT-020-manual.md)

SPRINT-020 ships customer authentication. Broad production customer traffic still requires this checklist to be re-run against production-like quota/rate-limit settings.

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

- [x] `/healthz` → `"ok": true`, sprint flag matches SPRINT-020
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

### SPRINT-020 (v2.1.0)

- [x] Tenant admin can enable customer OTP auth.
- [x] Customer OTP request/verify works on Libra Tech tenant.
- [x] Customer profile/session response is tenant scoped.
- [x] Customer portal tenant context uses `?tenant_id=...`.
- [x] Authenticated chat/call tenant routing implemented.
- [x] Avatar picker popup makes all avatars selectable at 100% browser scale.

### SPRINT-027 (v2.8.0)

- [x] Mobile bootstrap returns tenant policy, locale/timezone, assigned active avatars, default avatar, and authoritative limits.
- [x] Versioned mobile call lifecycle endpoints enforce customer session, tenant, assigned-avatar, quota, rate-limit, and idempotency policy.
- [x] Mobile WebSocket validates call ownership and returns only bounded provider-independent event/error envelopes.
- [x] Transcript, explicit end-call, and 1-5 star rating operations return mobile-safe data.
- [x] TypeScript SDK builds independently and exposes OTP, bootstrap, token refresh, lifecycle, reconnect, transcript, end, and rating operations.
- [x] Existing web call routes, archive, quota, statistics, and `/healthz` compatibility validated by regression gates.

### SPRINT-028 (v2.9.0)

- [x] Structured audit events capture tenant, actor, action, resource, request, outcome, and bounded metadata.
- [x] Local JSONL spool rotates with the configured timestamp naming convention and transfers on the default five-second interval.
- [x] ClickHouse sink, deterministic event IDs, acknowledgement markers, retry retention, and one-hour cleanup are implemented and unit tested.
- [x] Platform audit API/UI exposes bounded filters, pagination, delivery health, and non-sensitive metadata only.
- [x] `GET /api/platform/tenants` includes `logo_url` from the tenant brand profile for mobile/admin company-logo display.
- [x] Full Go tests, server build, platform-admin Svelte check/build, and `git diff --check` pass.
- [ ] Manual browser, ClickHouse outage/recovery, and retention UAT evidence; deferred to the next tester run.

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
| Dev | Codex release verification | 2026-07-16 | Audit pipeline, platform API/UI, tenant logo response + unit tests |
| Tester | Codex release verification | 2026-07-16 | Automated contract, redaction, retention, and tenant response checks; manual UAT deferred |
| PM | User-authorized release close | 2026-07-16 | ACs accepted for SPRINT-028 |
| DevOps | Codex release verification | 2026-07-16 | Build/test/tag verified |

## H. Release-cut (PM + DevOps)

After sections A–G are green:

```bash
# SPRINT-028 release
git tag -a v2.9.0 -m "v2.9.0 - SPRINT-028 cross-tenant audit log"
git push origin v2.9.0
```

- [ ] Tag pushed to `origin`
- [x] Sprint marked `completed` in `03-sprints/`
- [x] `_velocity.json` updated
- [x] ROADMAP sprint 28 marked shipped and next sprint pointer advanced

## Quick demo script (stakeholder, ~10 min)

1. `make up && make km-seed`
2. Open portal → select **Max** → **Billing** tab
3. Ask *"When are invoices due?"* → show citation chips
4. Start voice call with **Ava** → one general question
5. Show `/legacy/` for visual continuity
6. `make down` when finished
