---
id: DES-0060
title: Queued Concurrent-Call Admission Specification
status: review_pending
updated: 2026-08-01
sprint: SPRINT-065
owner: SA
feature: FEAT-0057
release_target: v2.37.0
---

# DES-0060 - Queued Concurrent-Call Admission

**Sprint:** SPRINT-065 · **Release target:** v2.37.0  
**Feature:** [FEAT-0057](../01-features/FEAT-0057-queued-concurrent-call-admission.md)  
**Prior:** [DES-0016](16-quota-rate-limit-spec.md), [DES-0019](19-tenant-settings-limits-spec.md), [DES-0024](24-authenticated-workforce-selection-spec.md), [DES-0042](42-aiaas-packages-usage-reconciliation-spec.md), [DES-0048](48-commercial-plans-billing-operations-spec.md), [DES-0052](52-caller-desk-branding-audio-devices-spec.md)

## 1. Goals

1. Queue a caller when a tenant's active voice calls equal the package
   `max_concurrent_calls` limit.
2. Admit the first eligible queued caller when another call releases capacity.
3. Keep active calls at or below the package/bonus concurrent-call limit across
   browser retries, disconnects, server restarts, and multiple app instances.
4. Show useful wait state to callers and operational capacity to tenant/admin
   users.
5. Show total calls and busy/live status on the customer call screen, not only
   in tenant/admin support views.

## 2. Non-goals

- Raising or bypassing tenant concurrent-call limits.
- Priority queues, paid fast lanes, callback scheduling, or human-agent routing.
- Replacing S16 daily/per-call caps, monthly-minute quota, mobile quota, or
  rate limits.
- Durable Postgres queue history in the S65 MVP.

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `CALL_QUEUE_ENABLED` | `true` | Enable queue-on-concurrent-full for `/ws/voice`. |
| `CALL_QUEUE_MAX_WAIT` | `120s` | Maximum time a caller may wait before timeout. |
| `CALL_QUEUE_MAX_PER_TENANT` | `50` | Bounded queued entries per tenant/channel. |
| `CALL_QUEUE_PROMOTION_LOCK_TTL` | `10s` | Redis lock TTL for multi-instance promotion. |
| `CALL_QUEUE_POSITION_REFRESH` | `2s` | Status frame cadence while waiting. |

## 4. Data model

S65 does not add a Postgres table. Active calls remain the existing Redis
concurrent lease; monthly usage remains the S45 usage ledger.

### Redis DB 4

| Key | Type | TTL | Purpose |
| --- | --- | --- | --- |
| `monti_jarvis:quota:{tenant}:concurrent` | int | `QUOTA_CONCURRENT_TTL` | Existing active-call lease count. |
| `monti_jarvis:callq:{tenant}:voice` | sorted set | none | FIFO queue ordered by enqueue time. |
| `monti_jarvis:callq:{tenant}:entry:{admission_id}` | hash | `CALL_QUEUE_MAX_WAIT` + grace | Caller metadata, status, agent, call id, and expiry. |
| `monti_jarvis:callq:{tenant}:client:{client_key}` | string | entry TTL | Idempotent retry mapping to admission id. |
| `monti_jarvis:callq:{tenant}:promote_lock` | string | lock TTL | Multi-instance promotion lock. |
| `monti_jarvis:callq:{tenant}:recent_timeouts` | counter/list | 24h | Tenant/admin visibility. |

### Capacity summary

| Field | Meaning |
| --- | --- |
| `active_calls` | Current reserved concurrent voice slots for the tenant. |
| `queued_callers` | Current waiting callers in the tenant voice queue. |
| `total_calls` | `active_calls + queued_callers`; must render on the call screen. |
| `max_concurrent_calls` | Tenant package/bonus concurrent-call limit. |
| `busy_status` | `available`, `busy`, `queued`, `admitted`, `live`, `timeout`, or `cancelled`. |

### Queue entry states

| State | Meaning |
| --- | --- |
| `queued` | Caller is waiting and receives position frames. |
| `admitted` | Slot reserved; handler may start Gemini relay. |
| `cancelled` | Browser closed or caller cancelled before admission. |
| `timed_out` | Wait exceeded `CALL_QUEUE_MAX_WAIT`. |
| `expired` | Cleanup removed stale entry after handler/server loss. |

## 5. Admission algorithm

1. Resolve tenant from request context, query, or embed key as today.
2. Run existing rate, `voice_enabled`, monthly, mobile, S16 daily, and S16/S18
   per-call checks. Fail these immediately; do not queue.
3. If active concurrent count is below limit, atomically reserve a slot and
   return `admitted` with a capacity summary for the call screen.
4. If active count is at limit, enqueue with idempotency key and return/send
   queue metadata including total calls and `busy_status=busy`.
5. On release, cancel, timeout, or stale cleanup, acquire the promotion lock,
   remove ineligible entries, re-check active count, reserve one slot, and mark
   the first eligible entry `admitted`.

## 6. API summary

See [04-api-spec.md](04-api-spec.md) Sprint 65.

| Method | Path | Role | Purpose |
| --- | --- | --- | --- |
| `GET` | `/ws/voice` | public/customer optional | Voice WebSocket now supports queued admission frames. |
| `GET` | `/api/tenant/concurrent-call-queue/status` | `tenant_admin` | Tenant active/queued/limit snapshot. |
| `GET` | `/api/platform/tenants/{tenant_id}/usage` | `platform_admin` | Adds queue snapshot for support. |

## 7. RBAC

| Action | Public/customer | `tenant_admin` | `platform_admin` |
| --- | --- | --- | --- |
| Join voice queue for selected tenant | yes, current S54/S21 rules | n/a | n/a |
| Cancel own queued admission | yes, by closing WS or cancel frame | n/a | n/a |
| View tenant queue snapshot | no | own tenant | any tenant |
| Force remove queue entry | no | no | deferred |

## 8. Verification

```bash
go test ./internal/quota ./cmd/server -count=1
cd apps/customer-web && npm run check
cd apps/tenant-web && npm run check
```

Manual UAT:

1. Tenant A package limit 1: caller A starts immediately.
2. Tenant A caller B starts while A is active: B sees queue position 1.
3. Caller A ends: caller B is admitted automatically; active count remains 1.
4. Tenant B caller starts during Tenant A queue: Tenant B starts independently.
5. Customer call screen shows busy/live state and total calls in queued and
   active-call states.
6. Queued caller cancel, browser close, and timeout remove queue entries.

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | | [ ] |
| Dev | | | [ ] |
