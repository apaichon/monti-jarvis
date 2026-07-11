---
id: SPRINT-013
status: completed
start: 2026-07-11
end: 2026-07-11
closed: 2026-07-11
updated: 2026-07-11
release_target: v1.4.0
release: v1.4.0
goal: "Platform Admin: Quota & Rate Limit — Redis counters + entitlement enforcement on chat/voice/KM/avatars; platform usage view."
roadmap_sprint: 13
platform: Platform Admin
depends_on: [SPRINT-003, SPRINT-004]
---

# SPRINT-013 — Platform Admin: Quota, Rate Limit

## Goal

Enforce **package entitlements** as live **quotas** (Redis counters) and **API rate limits** on hot paths, and give **platform admins** a read-only **usage vs limits** view. Unblocks Sprint 16 tenant self-service limits.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S10–S12) | 16, 14, 14 → **avg 14.7** |
| Trailing average (`_velocity.json`) | **16** |
| **Commitment target** | **16** |

No unassigned backlog tasks (`proposed`/`approved` without sprint). New commitment below.

## Context from prior sprints

| Sprint | Shipped capability |
| --- | --- |
| 3 | JWT + RBAC; Redis available |
| 4 | `package_limits.rules` jsonb, `tenant_entitlements`, `internal/entitlements` Redis cache |
| 5 | Avatar assign (count toward `max_ai_employees`) |
| 8–12 | Commerce; purchase assigns entitlement |

**Gap today:** limits are stored and shown, but chat / voice / KM do not check them.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0057](../04-tasks/TASK-0057.md) | 3 | completed | devops | Quota Redis key design, env, `/api/infra` status, optional usage audit table |
| [TASK-0058](../04-tasks/TASK-0058.md) | 5 | completed | dev | `internal/quota` — resolve limits, check/incr/release counters |
| [TASK-0059](../04-tasks/TASK-0059.md) | 4 | completed | dev | Enforce on chat, voice, KM ingest, avatar assign; 429/403 codes |
| [TASK-0060](../04-tasks/TASK-0060.md) | 3 | completed | dev | Platform API + `/admin` usage panel (tenant usage vs limits) |
| [TASK-0061](../04-tasks/TASK-0061.md) | 1 | completed | tester | Manual UAT checklist drafted; full browser UAT deferred |

**Committed:** 16 points · **Completed:** 16 · **Velocity:** 16 · **Release:** v1.4.0

## Shipped (v1.4.0)

- `internal/quota` — Redis counters, rate limits, Snapshot, structured errors
- Env: `QUOTA_*`, `RATE_LIMIT_*`; `/api/infra` quota + rate_limit status
- Enforcement: chat, voice WS, KM ingest, avatar assign
- Platform `GET /api/platform/tenants/{id}/usage` + `/admin/tenants/{id}/usage`
- Design pack DES-0016 + workflow §32–36; manual checklist (partial UAT deferred)

**Deferred UAT:** concurrent voice multi-session, live rate-limit burst under Gemini load, package-flag browser toggles — re-run [SPRINT-013-manual.md](../06-manual-tests/SPRINT-013-manual.md) when convenient.

## Scope boundary

**In**
- Redis counters under `monti_jarvis:quota:*` and rate-limit keys `monti_jarvis:rl:*`
- Enforcement for rules-v1 keys: `max_ai_employees`, `max_monthly_call_minutes`, `max_km_documents`, `max_concurrent_calls`, `voice_enabled`, `rag_enabled`
- Per-tenant (or demo-tenant) rate limit on `POST /api/chat`, voice WS open, KM writes
- Platform `GET /api/platform/tenants/{id}/usage` + admin UI panel
- Fail mode when Redis unavailable: env `QUOTA_FAIL_OPEN` (default true for local demo)
- Design pack before TASK-0058 implementation

**Out** (→ later)
- Tenant portal limit/settings UI (**SPRINT-016**)
- ClickHouse usage warehouse, billing overage (**S25+**)
- Login rate limit (auth Phase F)
- NATS usage events
- Per-customer (end-user) quotas (S19+)

## Feature

- [FEAT-0013 — Quota & Rate Limit](../01-features/FEAT-0013-quota-rate-limit.md)

## Design pack (`sprint-tech-specs`)

| Artifact | Path | Status |
| --- | --- | --- |
| Quota deep spec | [16-quota-rate-limit-spec.md](../02-design/16-quota-rate-limit-spec.md) | **`shipped`** |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §32–36 | **`shipped`** |
| ER diagram | [03-er-diagram.md](../02-design/03-er-diagram.md) Redis + no DDL | **`shipped`** |
| API spec | [04-api-spec.md](../02-design/04-api-spec.md) § Quota | **`shipped`** |
| UX/UI ASCII | [05-ux-ui.md](../02-design/05-ux-ui.md) § P14 | **`shipped`** |

## Verification

```bash
make build && make test
make infra-init && make restart
# Assign Starter, exhaust KM docs, assert 429
# Open concurrent voice slots = max_concurrent_calls, next fails
# Platform admin: open tenant usage panel
open http://localhost:8091/admin/tenants/demo/usage
```

- **Manual UAT:** [SPRINT-013-manual.md](../06-manual-tests/SPRINT-013-manual.md) (TASK-0061)
- Unit: `go test ./internal/quota/... ./cmd/server/ -count=1 -run Quota`

## Risks

| Risk | Mitigation |
| --- | --- |
| Demo/`AUTH_DISABLED` breaks customer UX | Document fallback: resolve `demo` entitlement or fail-open when no tenant |
| Clock skew on monthly windows | Use UTC calendar month keys `YYYYMM` |
| Double-count concurrent on crash | TTL on concurrent keys + release on WS close |
| Redis down blocks all traffic | `QUOTA_FAIL_OPEN=true` default for dev |

## Links

- Depends: [SPRINT-003](SPRINT-003.md), [SPRINT-004](SPRINT-004.md)
- Next: [ROADMAP](../00-roadmap/ROADMAP.md) SPRINT-014 Embed to Web · SPRINT-016 tenant limits UI
