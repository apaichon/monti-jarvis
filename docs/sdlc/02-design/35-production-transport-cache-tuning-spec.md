---
id: DES-0035
title: Production Transport and Cache Tuning Specification
status: review_pending
updated: 2026-07-18
sprint: SPRINT-032
owner: SA
depends_on: [SPRINT-031, DES-0034]
---

# Production Transport and Cache Tuning - Design Spec

**Sprint:** SPRINT-032 · **Roadmap:** gRPC switch mode, Cache on Prod
**Status:** Design track only; implementation requires separately approved tasks.
**Depends on:** [DES-0034](34-platform-billing-quota-ai-cost-spec.md), [SPRINT-031](../03-sprints/SPRINT-031.md)

## 1. Goals

- Define a reversible internal gRPC switch mode without changing browser-facing HTTP or WebSocket contracts.
- Define a production cache profile that preserves Redis DB 4 isolation, `monti_jarvis:` namespacing, TTL ownership, and fail-open behavior where safe.
- Make transport and cache rollout observable, bounded, and reversible before implementation begins.
- Keep payment, entitlement, quota-authority, customer-call, and platform-admin authorization semantics unchanged.

## 2. Non-goals

- Implementing a gRPC server, replacing the Fiber/HTTP server, or exposing gRPC to browsers in Sprint 32.
- Adding a new public endpoint, changing the billing usage response, or moving ClickHouse/Postgres authorities into a cache.
- Caching payment transitions, entitlement mutations, authorization decisions, raw prompts, transcripts, audio, or provider credentials.
- Changing Redis DB allocation: Monti Jarvis remains on DB 4 with the `monti_jarvis:` prefix.

## 3. Current-state constraints

- The customer and platform surfaces use HTTP/JSON and the voice surface uses WebSocket; these contracts remain stable.
- Existing Redis consumers include auth, entitlement, quota, call-active state, OAuth state, and rate limits. New cache keys must not overlap those namespaces.
- There is no current gRPC service boundary in the repository. Any switch must begin behind an internal adapter and preserve the existing HTTP path as rollback fallback.

## 4. Proposed transport switch

The switch is internal-only and opt-in per service boundary. Proposed deployment configuration, not implemented in Sprint 32:

| Setting | Values | Default / rule |
| --- | --- | --- |
| `GRPC_MODE` | `disabled`, `shadow`, `preferred`, `required` | `disabled`; production rollout starts `shadow`. |
| `GRPC_FALLBACK_ENABLED` | boolean | `true` for `shadow`/`preferred`; `required` may disable only after a separate approval. |
| `GRPC_TIMEOUT` | duration | Must be bounded and less than the caller deadline. |
| `GRPC_TARGET` | internal service target | Never accepted from browser input; deployment-only. |

Mode behavior:

1. `disabled`: use the existing HTTP/internal path.
2. `shadow`: execute only a bounded comparison probe; the HTTP result remains authoritative and only normalized metrics are recorded.
3. `preferred`: use gRPC when healthy; fall back to the HTTP path on timeout, transport error, or incompatible response.
4. `required`: fail safely when gRPC is unavailable; enable only after shadow and preferred evidence meet the release gate.

The adapter must propagate request id, tenant scope, deadline, and normalized error code. It must never copy customer content into transport metrics or logs.

## 5. Proposed production cache profile

Use cache-aside behavior for read-heavy, non-authoritative data only:

| Domain | Key shape | Default TTL | Failure behavior |
| --- | --- | ---: | --- |
| Package/entitlement display | `monti_jarvis:cache:entitlement:{tenant_id}` | 15m | Bypass cache and read Postgres; never mutate entitlement from cache. |
| Platform aggregate display metadata | `monti_jarvis:cache:platform:{scope}:{hash}` | 30s | Bypass cache; stale display must be labeled if ever enabled. |
| AI rate catalog | `monti_jarvis:cache:ai-rate:{provider}:{model}:{modality}:{version}` | 1h | Keep historical fact pricing; missing rate remains unavailable. |

Rules:

- `CACHE_PROD_ENABLED` is an explicit deployment flag; default is `false` until metrics and reset tests are complete.
- Cache keys use a bounded, hashed query component where user-controlled filters are included; no prompt, transcript, email, phone, or payment payload enters a key.
- Cache writes are best-effort and never block customer calls, quota enforcement, payment callbacks, or platform authorization.
- Invalidation is explicit on package/entitlement changes; TTL expiry is a fallback, not the source of truth.
- Quota counters, rate-limit buckets, auth sessions, and OAuth state retain their existing key contracts and are not folded into the generic cache namespace.

## 6. Observability and rollout gates

Required normalized metrics/log fields:

- `transport_mode`, `transport_selected`, `transport_fallback`, `transport_latency_ms`
- `cache_domain`, `cache_hit`, `cache_miss`, `cache_bypass`, `cache_write_error`
- `tenant_scope_hash`, `request_id`, `error_code`, and bounded dependency latency

Do not emit raw targets, Redis keys, SQL, credentials, customer identifiers, or provider payloads. Rollout requires:

1. Unit tests for mode selection, deadline propagation, fallback, and incompatible response handling.
2. Cache key namespace/TTL/reset tests with isolated Redis DB 4 fixtures.
3. Shadow comparison with no customer-visible result change.
4. Preferred-mode soak evidence and a documented rollback to `GRPC_MODE=disabled`, `CACHE_PROD_ENABLED=false`.

## 7. API and RBAC impact

There is no public API or RBAC change in this design track. Existing `/api/infra` may report normalized transport/cache health only after a separate implementation task; it must not expose targets, keys, or raw dependency errors. Existing platform-admin and tenant/customer authorization boundaries remain authoritative.

## 8. Verification curl block

Until implementation tasks are approved, verification uses existing endpoints only:

```bash
curl -fsS http://localhost:8091/healthz
curl -fsS http://localhost:8091/api/infra
curl -fsS \
  -H "Authorization: Bearer $PLATFORM_ADMIN_TOKEN" \
  "http://localhost:8091/api/platform/billing/usage?start_date=2026-07-18&end_date=2026-07-18"
```

The Sprint 32 implementation commitment remains TASK-0144/TASK-0145. gRPC and production-cache implementation must receive separate task IDs and approval before code changes are made.
