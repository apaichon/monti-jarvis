---
id: SPRINT-016
status: completed
start: 2026-07-12
end: 2026-07-12
closed: 2026-07-12
updated: 2026-07-12
design_pack: shipped
release_target: v1.7.0
release: v1.7.0
goal: "Tenant: Settings, Locale, and operational call limits under package quotas — self-service usage view + caps."
roadmap_sprint: 16
platform: Tenant
depends_on: [SPRINT-013, SPRINT-015]
---

# SPRINT-016 — Tenant: Settings, Locale, Limits

## Goal

Let **active tenants** configure **locale/timezone**, see **package usage vs limits**, and set **operational call-time caps** (per call / per day) that sit under S13 package ceilings — without platform admin help.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S13–S15) | 16, 16, 16 → **avg 16** |
| Trailing average | **16** |
| **Commitment** | **16** |
| **Completed** | **16** |

## Context

| Sprint | Capability used |
| --- | --- |
| 13 | Redis quotas, monthly call minutes, package rules-v1 |
| 6–7 | Tenant admin + active status |
| 15 | Tenant portal shell (`/tenant/km`, `/tenant/embed`) |

**Gap:** No tenant-facing settings; no daily/per-call voice caps; usage UI is platform-only.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0072](../04-tasks/TASK-0072.md) | 3 | done | devops | `tenant_settings` + `tenant_call_limits` schema, Redis keys |
| [TASK-0073](../04-tasks/TASK-0073.md) | 5 | done | dev | Settings/usage/call-limits APIs + enforce voice caps |
| [TASK-0074](../04-tasks/TASK-0074.md) | 4 | done | dev | Tenant UI `/tenant/settings` |
| [TASK-0075](../04-tasks/TASK-0075.md) | 3 | done | dev | Locale preference wiring (portal + AI reply hint) + tier/group labels scaffold |
| [TASK-0076](../04-tasks/TASK-0076.md) | 1 | done | tester | Manual UAT checklist |

**Committed:** 16 points

## Scope boundary

**In**
- Tenant settings: locale (`th`|`en`), timezone, optional brand display name override
- Read-only usage snapshot for own tenant
- Call limits: `max_minutes_per_call`, `max_call_minutes_per_day` (nullable / 0 = unset)
- Enforcement on voice open/close (Redis daily + session duration)
- Scaffold fields: `user_tier_label`, `user_group_label` (free text for ops notes — full model S18)
- Design DES-0019 + UAT

**Out**
- Full customer tier product (S18)
- Preview sandbox (S17)
- Overage billing
- Per-caller identity quotas (S19+)
- Complete UI translation of entire tenant portal

## Feature

- [FEAT-0016 — Tenant Settings, Locale, Limits](../01-features/FEAT-0016-tenant-settings-locale-limits.md)

## Design pack (`sprint-tech-specs`)

| Artifact | Path | Status |
| --- | --- | --- |
| Deep spec | [19-tenant-settings-limits-spec.md](../02-design/19-tenant-settings-limits-spec.md) **DES-0019** | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §46–48 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 16 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Tenant settings | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § T9 | `approved` |

> **Gate:** Design pack approved — implement TASK-0072 → 0076.

## Verification

```bash
make build && make test
# Tenant admin → /tenant/settings
# Set max_minutes_per_call=1 → start voice → exceeds → denied/ended
# GET /api/tenant/usage shows package meters
```

- **Manual UAT:** [SPRINT-016-manual.md](../06-manual-tests/SPRINT-016-manual.md) (TASK-0076 — create at VERIFY)

## Risks

| Risk | Mitigation |
| --- | --- |
| Double-count minutes (S13 monthly + S16 daily) | Clear Redis key namespaces; document both |
| Cap > package confuses users | UI shows package ceiling; clamp on save |
| Voice already open when daily hits 0 | Enforce on open; end on per-call cap only |
| Locale unused by AI | Store preference; pass as system hint on chat/voice |

## Shipped summary (v1.7.0)

| Area | Outcome |
| --- | --- |
| Schema | `tenant_settings`, `tenant_call_limits` (lazy create) |
| APIs | `GET/PUT /api/tenant/settings`, `GET /api/tenant/usage`, `GET/PUT /api/tenant/call-limits` |
| Redis | `call_daily:{tenant}:{YYYYMMDD}` (tenant timezone day) under S13 monthly keys |
| Voice | Daily + per-call caps; mono-language prompt + caption merge polish |
| UI | `/tenant/settings` — workspace, usage meters, call limits, tier/group scaffold |
| UAT | [SPRINT-016-manual.md](../06-manual-tests/SPRINT-016-manual.md) |

## Production launch gate (do not skip)

> **Before launching production to end customers** — specifically after **tenant customer-user auth** is integrated (SPRINT-019–020: customer accounts for tenants) — **must verify that rate limit and quota management work end-to-end** for real multi-user load.

| Gate | Why |
| --- | --- |
| **S13 package quotas** | Monthly minutes, concurrent, KM, avatars, voice/RAG flags |
| **S13 rate limits** | Chat / voice / KM per-minute buckets under multi-customer concurrent use |
| **S16 operational caps** | Daily + per-call minutes under package ceiling |
| **Tenant isolation** | Customer of tenant A must not consume or see quota of tenant B |
| **Fail-open / fail-closed** | Confirm production env flags (`QUOTA_*`, `RATE_LIMIT_*`) match intended safety |

Checklist owner: **DevOps + Tester** at pre-prod for S19–20 / customer go-live. Track against [FEAT-0013](../01-features/FEAT-0013-quota-rate-limit.md) + this sprint’s call limits.

## Links

- Depends: [SPRINT-013](SPRINT-013.md), [SPRINT-015](SPRINT-015.md)  
- Next: SPRINT-017 Test and Preview  
- Release: **v1.7.0**
