---
id: READINESS-RELEASE
status: active
updated: 2026-07-07
current_sprint: SPRINT-004
release_target: v0.5.0
---

# Release Readiness Checklist

Use this checklist before **demo**, **sprint sign-off**, or **`release-cut`** (git tag).

## A. Code & build

- [ ] Branch is clean or PR merged to `main`
- [ ] `go test ./...` passes
- [ ] `make build` succeeds (Svelte portal + Go binary)
- [ ] No known P0/P1 defects open for this sprint

## B. Infrastructure

- [ ] Shared containers running: `poc-gml-postgres`, `poc-gml-redis`, `poc-gml-minio`
- [ ] Sprint 2+: `poc-gml-clickhouse` running
- [ ] `make infra-init` applied (schema + MinIO bucket)
- [ ] `make infra-check` — all required services report healthy
- [ ] Monti compose up (`monti-nats`, `monti-livekit`) or accepted as optional degraded

## C. Configuration

- [ ] `infra/.env.dev` exists (from `.env.dev.example`)
- [ ] `GEMINI_API_KEY` set and valid
- [ ] Sprint 2+: `CLICKHOUSE_URL` and `CLICKHOUSE_DB=monti_jarvis`
- [ ] `DEMO_TENANT_ID=demo` for single-tenant demos

## D. Runtime smoke (5 min)

```bash
make start
make status
curl -fsS http://localhost:8091/healthz
curl -fsS http://localhost:8091/api/infra
curl -fsS http://localhost:8091/api/workforce
```

- [ ] `/healthz` → `"ok": true`, sprint flag matches active sprint
- [ ] `/api/infra` → postgres, redis, minio `ok`; clickhouse `ok` (Sprint 2+)
- [ ] `/api/workforce` → four agents

## E. Sprint-specific data

### SPRINT-001 (v0.2.0)

- [ ] Portal loads at http://localhost:8091
- [ ] Agent selection + text chat + voice smoke (see [SPRINT-001 manual](../06-manual-tests/SPRINT-001-manual.md))

### SPRINT-002 (v0.3.0 target)

- [ ] `make km-seed` or `POST /api/km/seed` succeeds
- [ ] `GET /api/km/agents/max` shows documents and chunks
- [ ] Billing RAG curl returns `sources[]` (see [KM_SETUP](../../KM_SETUP.md))
- [ ] Citation chips visible in portal after grounded question

## F. Documentation

- [ ] Sprint doc status accurate (`docs/sdlc/03-sprints/SPRINT-NNN.md`)
- [ ] Task statuses match implementation (`docs/sdlc/04-tasks/`)
- [ ] Manual test checklist completed ([`06-manual-tests/`](../06-manual-tests/))
- [ ] Test matrix scenarios for this sprint marked pass ([`05-test-scenarios/TEST-MATRIX.md`](../05-test-scenarios/TEST-MATRIX.md))
- [ ] `AGENTS.md` current sprint line updated
- [ ] API spec matches shipped routes ([`02-design/04-api-spec.md`](../02-design/04-api-spec.md))

## G. Sign-off

| Role | Name | Date | Notes |
| --- | --- | --- | --- |
| Dev | | | Implementation + unit tests |
| Tester | | | Manual UAT green |
| PM | | | ACs accepted |
| DevOps | | | Infra + deploy verified |

## H. Release-cut (PM + DevOps)

After sections A–G are green:

```bash
# Example for v0.3.0 — run release-cut skill for semver + VERSION_HISTORY
git tag v0.3.0
git push origin v0.3.0
```

- [ ] Tag pushed to `origin`
- [ ] Sprint marked `completed` in `03-sprints/`
- [ ] `_velocity.json` updated
- [ ] ROADMAP current sprint pointer advanced

## Quick demo script (stakeholder, ~10 min)

1. `make up && make km-seed`
2. Open portal → select **Max** → **Billing** tab
3. Ask *"When are invoices due?"* → show citation chips
4. Start voice call with **Ava** → one general question
5. Show `/legacy/` for visual continuity
6. `make down` when finished