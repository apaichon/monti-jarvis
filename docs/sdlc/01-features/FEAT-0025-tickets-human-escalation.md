---
id: FEAT-0025
title: Tenant Tickets and Human Escalation
status: completed
sprint: SPRINT-023
owner: PM
updated: 2026-07-14
---

# Feature: Tenant Tickets and Human Escalation

## Problem

Some customer conversations need follow-up from a human. The AI must be able to offer that path without silently creating work, while tenant teams need a small, auditable queue to triage and resolve requests.

## Scope

**In**

- Structured AI escalation offer during an active customer chat or voice conversation.
- Explicit customer confirmation before ticket creation.
- Tenant-scoped tickets linked to conversation records, calls, customers, and avatars when available.
- Tenant queue filters, detail timeline, priority, assignee, status, and internal notes.
- Idempotent ticket creation and tenant-context lifecycle events.

**Out**

- Real-time human transfer, supervisor listen mode, or live agent takeover.
- SLA timers, email/SMS notifications, external CRM integration, and platform-wide queues.
- Customer-facing ticket history or two-way ticket messaging after creation.
- Cross-tenant ticket analytics in ClickHouse.

## Acceptance criteria

1. The customer receives a structured human-follow-up offer and the create API rejects requests without explicit confirmation.
2. A confirmed request creates one tenant-scoped ticket linked to the source conversation/call when available.
3. Repeated submission with the same idempotency key returns the existing ticket and does not duplicate work.
4. Tenant admins can list and filter only their own tickets, inspect the linked conversation summary, and see the event timeline.
5. Tenant admins can update status, priority, assignee, and internal notes using validated lifecycle transitions.
6. Ticket creation and updates publish tenant-context events without putting transcript or contact secrets in event subjects.
7. Anonymous and authenticated customer paths follow existing tenant customer-auth policy and never expose another tenant's ticket.
8. Manual UAT covers customer confirmation, tenant isolation, duplicate submission, lifecycle transitions, and the no-live-handoff boundary.

## Dependencies

- [FEAT-0001 - Workforce QA](FEAT-0001-workforce-qa.md)
- [FEAT-0003 - Auth and RBAC](FEAT-0003-auth-rbac.md)
- [FEAT-0022 - Customer Authentication and Domain Enforcement](FEAT-0022-customer-auth.md)
- [FEAT-0024 - Conversation Records and Knowledge Gap Review](FEAT-0024-conversation-records-knowledge-gaps.md)

## Design

- [DES-0026 - Tickets and Human Escalation Specification](../02-design/26-tickets-human-escalation-spec.md)
- [API contract - Tickets & Human Escalation](../02-design/04-api-spec.md#tickets--human-escalation-sprint-23)
- [Workflow §68-70](../02-design/02-workflow.md#68-customer-confirms-human-follow-up-ticket-sprint-23)
- [UX § Sprint 23](../02-design/05-ux-ui.md#sprint-23---tickets-and-human-escalation-t16)
- Sprint plan: [SPRINT-023](../03-sprints/SPRINT-023.md)
