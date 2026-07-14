---
id: SPRINT-023
status: completed
start: 2026-07-14
end: 2026-07-14
closed: 2026-07-14
updated: 2026-07-14
design_pack: approved
release_target: v2.4.0
release: v2.4.0
goal: "Tenant/Customer: let the AI offer a human follow-up ticket, then give tenant teams a tenant-scoped queue and lifecycle to resolve it."
roadmap_sprint: 23
platform: Tenant / Customer
depends_on: [SPRINT-003, SPRINT-020, SPRINT-022]
---

# SPRINT-023 - Tenant/Customer: Tickets and Human Escalation

## Goal

When a customer needs human help, the AI asks for explicit confirmation and creates a tenant-scoped ticket that operators can triage and close.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S20-S22) | 16, 16, 16 -> **avg 16** |
| **Commitment** | **16** |
| **Completed** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0107](../04-tasks/TASK-0107.md) | 3 | completed | devops | Ticket and ticket-event schema with tenant-safe audit contract |
| [TASK-0108](../04-tasks/TASK-0108.md) | 5 | completed | dev | Customer creation plus tenant queue/detail/lifecycle APIs |
| [TASK-0109](../04-tasks/TASK-0109.md) | 3 | completed | dev | AI escalation offer, explicit confirmation, and ticket event publication |
| [TASK-0110](../04-tasks/TASK-0110.md) | 4 | completed | dev | Tenant ticket queue, detail timeline, filters, and lifecycle controls |
| [TASK-0111](../04-tasks/TASK-0111.md) | 1 | completed | tester | Ticket creation, isolation, lifecycle, and escalation UAT |

**Committed:** 16 points · **Completed:** 16 points · **Completion:** 100%

## Scope boundary

**In**

- Customer confirmation before a ticket is created from an AI escalation offer.
- Tenant-scoped ticket and ticket-event records linked to a conversation record, call, customer, and avatar where available.
- Tenant queue filters, ticket detail, priority, assignee, status, and internal notes.
- Lifecycle events published with tenant context for later notification/worker integration.
- Safe anonymous-customer handling using existing customer-auth policy and contact capture rules.

**Out**

- Real-time transfer or live human agent takeover.
- SLA timers, email/SMS notifications, external CRM connectors, and platform-wide ticket dashboards.
- Customer-facing ticket history or two-way ticket messaging after creation.
- Cross-tenant platform support queue and ClickHouse ticket analytics.

## Feature

- [FEAT-0025 - Tenant Tickets and Human Escalation](../01-features/FEAT-0025-tickets-human-escalation.md)

## Design

| Artifact | Path | Status |
| --- | --- | --- |
| Deep spec | [26-tickets-human-escalation-spec.md](../02-design/26-tickets-human-escalation-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §68-70 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 23 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Tickets & Human Escalation | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § Sprint 23 | `approved` |

## Verification

```bash
/usr/local/go/bin/go test ./...
/usr/local/go/bin/go build ./cmd/server
cd apps/customer-web && npm run build
cd ../tenant-web && npm run build
# Customer receives a structured ticket offer and must confirm before creation.
# Tenant admin sees only the current tenant's tickets and can update lifecycle fields.
# Duplicate confirmation is idempotent and does not create a second open ticket.
# Cross-tenant ticket ids return 404 without metadata leakage.
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-023-manual.md
```

## Risks

| Risk | Mitigation |
| --- | --- |
| AI creates tickets without customer intent | Require a structured offer and explicit confirmation flag on the create API |
| Ticket data leaks across tenants | Resolve tenant only from the authenticated context and scope every query/update |
| Repeated voice confirmations create duplicates | Require an idempotency key and return the existing ticket for duplicate requests |
| Operators treat the MVP as a live handoff | Label the queue as follow-up work; defer transfer, SLA, and notifications |

## Links

- Depends: [SPRINT-003](SPRINT-003.md), [SPRINT-020](SPRINT-020.md), [SPRINT-022](SPRINT-022.md)
- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Sprint 23
- Target: **v2.4.0**
