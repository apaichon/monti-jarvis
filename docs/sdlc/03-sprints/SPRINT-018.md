---
id: SPRINT-018
status: completed
start: 2026-07-12
end: 2026-07-12
closed: 2026-07-12
updated: 2026-07-12
design_pack: shipped
release_target: v1.9.0
release: v1.9.0
goal: "Tenant: Customer Tier catalog and groups — define VIP/standard rules before customer identity."
roadmap_sprint: 18
platform: Tenant
depends_on: [SPRINT-016]
---

# SPRINT-018 — Tenant: Customer Tier

## Goal

Let **active tenants** manage a **customer tier catalog** (and optional groups) with default agent, locale hint, and optional call-cap overrides — configuration ready for S19–20 customer accounts.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S15–S17) | 16, 16, 16 → **avg 16** |
| Trailing average | **16** |
| **Commitment** | **16** |
| **Completed** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0082](../04-tasks/TASK-0082.md) | 3 | completed | devops | `customer_tiers` + `customer_groups` schema |
| [TASK-0083](../04-tasks/TASK-0083.md) | 5 | completed | dev | Tenant tiers/groups REST APIs + validation |
| [TASK-0084](../04-tasks/TASK-0084.md) | 4 | completed | dev | Tenant UI `/tenant/tiers` |
| [TASK-0085](../04-tasks/TASK-0085.md) | 3 | completed | dev | Apply tier overrides on preview (+ settings link) |
| [TASK-0086](../04-tasks/TASK-0086.md) | 1 | completed | tester | Manual UAT checklist |

**Committed:** 16 · **Completed:** 16

## Shipped summary (v1.9.0)

| Area | Outcome |
| --- | --- |
| Schema | `customer_tiers`, `customer_groups` |
| APIs | Full CRUD `/api/tenant/tiers` and `/api/tenant/groups` |
| UI | `/tenant/tiers` + nav **Tiers** |
| Preview | Optional `tier_id` for locale + call-cap overrides |
| Settings | Link to structured Tiers admin |

## Feature

- [FEAT-0020 — Customer Tier](../01-features/FEAT-0020-customer-tier.md)

## Design pack

| Artifact | Path | Status |
| --- | --- | --- |
| Deep spec | [21-customer-tier-spec.md](../02-design/21-customer-tier-spec.md) | `shipped` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §52–54 | `shipped` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 18 | `shipped` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Customer tiers | `shipped` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § T11 | `shipped` |

## Verification

```bash
make build && make test
open http://localhost:8091/tenant/tiers
# UAT: docs/sdlc/06-manual-tests/SPRINT-018-manual.md
```

## Production launch gate (carry forward)

Before customer production (post S19–20 auth): re-verify rate limit + quota **with tier overrides**.

## Links

- Depends: [SPRINT-016](SPRINT-016.md)  
- Next: SPRINT-019 Customer Account Import / Integration  
- Release: **v1.9.0**
