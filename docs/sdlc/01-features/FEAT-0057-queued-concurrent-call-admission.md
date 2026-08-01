---
id: FEAT-0057
title: "Queued concurrent-call admission"
status: completed
release: v2.37.0
roadmap_sprint: 65
priority: A+
depends_on: [SPRINT-013, SPRINT-016, SPRINT-021, SPRINT-045, SPRINT-051, SPRINT-056]
design: DES-0060
design_spec: ../02-design/60-queued-concurrent-call-admission-spec.md
updated: 2026-08-01
---

# FEAT-0057: Queued Concurrent-Call Admission

## Purpose

Let callers wait when a tenant's package concurrent-call limit is full, then
start automatically when capacity opens.

## Problem

Sprint 13 protects tenant capacity by rejecting `/ws/voice` when active calls
equal `max_concurrent_calls`. That is correct for quota safety, but it creates a
poor caller experience during short traffic spikes because callers must retry
manually even if another customer finishes seconds later.

## Scope

### In

- Tenant-scoped waiting queue for public/customer voice calls.
- Admission controller that checks existing rate, feature, monthly, daily, and
  per-call rules before reserving a concurrent slot or enqueueing the caller.
- Redis-backed FIFO queue, active-slot lease, queue-entry TTL, and promotion
  lock using DB 4 prefix `monti_jarvis:`.
- Promotion on slot release, browser disconnect, failed start, and timeout.
- WebSocket queue status frames that let customer-web show position and
  auto-start when admitted.
- Tenant/admin visibility for active calls, queued callers, package limit, and
  recent queue timeouts.
- Audit/metrics events for queued, admitted, promoted, cancelled, timed out, and
  rejected states.

### Out

- Allowing active calls above tenant package/bonus concurrent-call limit.
- Paid priority queue, callback scheduling, or human-agent handoff.
- Cross-tenant capacity pooling.
- Replacing mobile, monthly-minute, daily, per-call, or rate-limit logic.
- Durable queue analytics beyond bounded operational counters.

## Acceptance criteria

1. If a tenant is below its package concurrent-call limit, the caller starts
   immediately and consumes one concurrent slot.
2. If a tenant is at the concurrent-call limit, the next caller is queued and
   sees live position/wait metadata instead of immediate `quota_exceeded`.
3. When an active call releases a slot, the first eligible queued caller is
   admitted and active calls never exceed the tenant limit.
4. If a queued caller cancels, closes the browser, or times out, the queue entry
   is removed and remaining positions are updated.
5. Two tenants have independent queues and cannot consume each other's
   concurrent-call capacity.

## Design

- Deep spec: [DES-0060](../02-design/60-queued-concurrent-call-admission-spec.md)
- Workflow: [02-workflow.md](../02-design/02-workflow.md), section 144
- API contract: [04-api-spec.md](../02-design/04-api-spec.md), Sprint 65
- UX mapping: [05-ux-ui.md](../02-design/05-ux-ui.md), C65/T65

## Release

Shipped in v2.37.0 as SPRINT-065. The release adds bounded Redis-backed
waiting for over-limit voice calls, automatic promotion on slot release,
customer queue status, tenant capacity visibility, and simplified conversation
controls.
