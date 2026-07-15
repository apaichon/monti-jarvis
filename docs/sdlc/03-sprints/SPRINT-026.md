---
id: SPRINT-026
status: completed
start: 2026-07-15
end: 2026-07-16
updated: 2026-07-15
closed: 2026-07-15
design_pack: approved
release_target: v2.7.0
release: v2.7.0
goal: "Tenant: expose safe system performance and dependency health signals for day-to-day operations."
roadmap_sprint: 26
platform: Tenant
depends_on: [SPRINT-025]
---

# SPRINT-026 - Tenant: System Performance Monitoring

## Goal

Give tenant administrators a tenant-safe view of service availability, dependency latency, and analytics freshness without exposing internal infrastructure errors or cross-tenant data.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S23, S24, S25) | 16, 16, 16 -> **avg 16** |
| **Commitment** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0120](../04-tasks/TASK-0120.md) | 4 | completed | devops | Normalized dependency probes and bounded performance snapshot |
| [TASK-0121](../04-tasks/TASK-0121.md) | 4 | completed | dev | Tenant-safe system performance API |
| [TASK-0122](../04-tasks/TASK-0122.md) | 6 | completed | dev | Tenant monitoring dashboard with degraded and stale states |
| [TASK-0123](../04-tasks/TASK-0123.md) | 2 | completed | tester | Monitoring API, tenant isolation, and UAT evidence |

**Committed:** 16 points · **Completed:** 16 points

**Closed:** 2026-07-15 as release `v2.7.0`. Automated validation and the reproducible UAT checklist passed; manual browser evidence remains deferred for a follow-up tester run.

## Scope boundary

**In**

- Read-only, tenant-admin monitoring of the Monti service path used by the active tenant.
- Normalized status and bounded latency for configured Postgres, Redis, MinIO, ClickHouse, NATS, LiveKit, and Gemini dependencies.
- Call-center analytics freshness and a clear distinction between healthy, degraded, unavailable, disabled, and stale data.
- Tenant dashboard loading, empty, degraded, unavailable, unauthorized, and retry states.
- Tenant isolation, timeout bounds, response redaction, and manual UAT evidence.

**Out**

- Platform-wide or cross-tenant monitoring; planned for Sprint 29.
- Audit logs; planned for Sprint 28.
- Billing, quota, AI infrastructure cost, paging, email, SMS, or webhook alerts.
- Customer-facing health details or raw provider error messages.
- Persistent time-series storage or a replacement for `/healthz` and `/api/infra`.

## Feature

- [FEAT-0028 - Tenant System Performance Monitoring](../01-features/FEAT-0028-tenant-system-performance-monitoring.md)

## Design

The Sprint 26 technical design pack is approved for implementation.

| Artifact | Planned scope | Status |
| --- | --- | --- |
| Deep spec | [29-tenant-system-performance-spec.md](../02-design/29-tenant-system-performance-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) Sprint 26 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 26 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Tenant System Performance | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) Sprint 26 | `approved` |

## Verification

```bash
make test
make build
cd apps/tenant-web && npm run check && npm run build
# Tenant admin receives only normalized component status and bounded latency.
# Dependency timeout and analytics outage render degraded/unavailable states.
# Tenant A cannot read Tenant B monitoring data or provider error details.
# Existing customer calls, archive writes, quota checks, and /healthz remain compatible.
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-026-manual.md
```

Release validation passed on 2026-07-15 with the full Go test suite, server build, tenant-web checks/build, and the running `/healthz` smoke test. Manual browser UAT is explicitly deferred and does not block this release cut.

## Risks

| Risk | Mitigation |
| --- | --- |
| Health probes add load to shared dependencies | Use short timeouts, bounded concurrency, and no polling faster than the UI refresh interval |
| Provider errors leak credentials or topology | Map errors to stable public codes and keep raw details in server logs only |
| A dependency outage blocks tenant operations | Monitoring is read-only and must not sit on call, quota, archive, or chat request paths |
| Stale ClickHouse data looks healthy | Return freshness metadata and render stale separately from healthy |

## Links

- Depends: [SPRINT-025](SPRINT-025.md)
- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Sprint 26
- Target: **v2.7.0**
- Next planned: Sprint 27 Mobile Call API and SDK
