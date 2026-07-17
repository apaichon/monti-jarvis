---
id: SPRINT-029
status: completed
start: 2026-07-17
end: 2026-07-18
updated: 2026-07-17
closed: 2026-07-17
design_pack: approved
release_target: v2.10.0
release: v2.10.0
goal: "Platform: expose a tenant-safe cross-tenant view of service health, dependency latency, and analytics freshness."
roadmap_sprint: 29
feature: FEAT-0031
platform: Platform
depends_on: [SPRINT-026, SPRINT-028]
---

# SPRINT-029 - Platform: System Performance Monitoring

## Goal

Give platform operators a bounded, cross-tenant view of Monti service health and analytics freshness. The platform surface must normalize dependency state without exposing provider details, credentials, URLs, customer data, or tenant-owned monitoring records.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S26, S27, S28) | 16, 16, 16 -> **avg 16** |
| **Committed** | **16** |
| **Completed** | **16** |

## Proposed commitment

No unassigned proposed/approved task files were available. The roadmap work is proposed as four implementation packages totaling 16 points:

| Work package | Points | Owner | Outcome |
| --- | ---: | --- | --- |
| [TASK-0132](../04-tasks/TASK-0132.md) Cross-tenant health snapshot | 4 | devops | Bounded, concurrent probes for configured Postgres, Redis, MinIO, ClickHouse, NATS, LiveKit, Gemini, and audit delivery state with normalized statuses |
| [TASK-0133](../04-tasks/TASK-0133.md) Platform performance API | 4 | dev | Platform-admin authorization, tenant filters, bounded pagination, analytics freshness, timeout behavior, and redacted error contracts |
| [TASK-0134](../04-tasks/TASK-0134.md) Platform monitoring dashboard | 6 | dev | `/admin/monitoring` view with overall health, dependency matrix, tenant drill-down, freshness state, degraded/unavailable states, and retry behavior |
| [TASK-0135](../04-tasks/TASK-0135.md) Verification and operations | 2 | tester | Unit/API isolation tests, response redaction checks, performance bounds, UAT checklist, and release readiness evidence |

**Committed:** 16 points · **Implementation completed:** 16 points · **Manual UAT:** deferred

## Scope boundary

**In**

- Read-only platform-admin monitoring of configured service dependencies across tenants.
- Stable normalized statuses: `operational`, `degraded`, `unavailable`, `disabled`, and `stale` where applicable.
- Bounded probe timeout and concurrency; monitoring remains outside customer call, voice relay, quota, archive, and chat critical paths.
- Tenant-scoped analytics freshness and safe aggregate health summaries.
- Audit spool/ClickHouse delivery health from Sprint 28 without exposing local paths or raw infrastructure errors.
- Tenant isolation, platform RBAC, response redaction, retry states, and responsive platform-admin UI.

**Out**

- Alert delivery, paging, email, SMS, webhooks, or SIEM integrations.
- Persistent time-series storage, SLO history, billing, quota, or AI infrastructure cost dashboards.
- Customer-facing health details or tenant-admin monitoring changes.
- Raw provider errors, dependency URLs, credentials, customer identifiers, transcripts, audio paths, or audit spool contents.

## Dependencies and design gates

1. Reuse the tenant monitoring probe contract where safe, while defining platform-only aggregation and tenant visibility rules.
2. Define the API response shape, filter/pagination contract, timeout budgets, and safe error mapping.
3. Define dashboard loading, partial failure, stale analytics, empty, unauthorized, and retry states.
4. Define whether snapshots are request-time only or cached briefly, with explicit freshness semantics.
5. Define UAT coverage for cross-tenant isolation and no infrastructure-detail leakage.

## Verification

```bash
go test ./...
go build ./cmd/server
cd apps/platform-admin-web && npm run check && npm run build
git diff --check
# Platform admins can view normalized health across tenants without tenant leakage.
# Dependency timeouts remain bounded and do not block customer operations.
# Disabled, degraded, unavailable, and stale states are distinct and actionable.
# Audit delivery health is visible without exposing local paths, raw errors, or credentials.
# Non-platform roles receive the standard unauthorized/forbidden response.
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-029-manual.md
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Cross-tenant probes amplify load on shared dependencies | Bound concurrency, use short timeouts, and avoid polling faster than the operator refresh interval |
| Platform views leak tenant or infrastructure metadata | Centralize allowlisted response fields and test authorization/redaction with multiple tenants |
| Stale analytics appears healthy | Keep dependency health and analytics freshness separate, with explicit timestamps and stale state |
| Monitoring becomes part of call critical paths | Use read-only request paths and never invoke platform monitoring from call, archive, or quota operations |

## Feature

- [FEAT-0031 - Platform System Performance Monitoring](../01-features/FEAT-0031-platform-system-performance-monitoring.md) - `completed`

## Design

| Artifact | Status | Scope |
| --- | --- | --- |
| [32-platform-system-performance-spec.md](../02-design/32-platform-system-performance-spec.md) | `shipped` | Platform snapshot, status derivation, RBAC, redaction, and verification |
| [02-workflow.md](../02-design/02-workflow.md) §83 | `shipped` | Platform administrator monitoring request and bounded probe flow |
| [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 29 | `shipped` | Ephemeral read model and existing-store contract; no migration |
| [04-api-spec.md](../02-design/04-api-spec.md) Platform System Performance | `shipped` | Platform-admin endpoint, filters, response, and errors |
| [05-ux-ui.md](../02-design/05-ux-ui.md) A21 | `shipped` | Platform monitoring screen, states, flows, and component mapping |

Implementation and automated release verification are complete. Manual browser and infrastructure-failure UAT is deferred to the next tester run.

## Links

- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Sprint 29
- Depends: [SPRINT-026](SPRINT-026.md), [SPRINT-028](SPRINT-028.md)
- Related tenant monitoring design: [29-tenant-system-performance-spec.md](../02-design/29-tenant-system-performance-spec.md)
- Feature and design pack: [FEAT-0031](../01-features/FEAT-0031-platform-system-performance-monitoring.md), [DES-0032](../02-design/32-platform-system-performance-spec.md)
- Next planned: Sprint 30 platform overall call-center statistics

## Release close

Automated Go tests, server build, platform-admin type checks/build, authenticated API smoke, authorization/error checks, and `git diff --check` passed. Manual browser, responsive-layout, and dependency-failure UAT remains deferred and is tracked in [SPRINT-029-manual.md](../06-manual-tests/SPRINT-029-manual.md).

**Closed:** 2026-07-17 as release `v2.10.0`.
