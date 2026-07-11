# Feature: Platform Quota & Rate Limit   (FEAT-0013)
**Sprint:** SPRINT-013   **Owner:** DEV   **Status:** shipped v1.4.0

## Problem

Packages ship numeric limits (`max_monthly_call_minutes`, `max_km_documents`, `max_concurrent_calls`, `max_ai_employees`) and feature flags, but nothing **enforces** them on live chat, voice, KM ingest, or avatar assignment. Without Redis counters and middleware, a Starter tenant can consume unbounded resources. Platform operators also cannot see **usage vs entitlement** for support and capacity planning.

## Scope

**In:**
- `internal/quota` — resolve limits from `internal/entitlements`, Redis usage counters, check/increment/release
- API **rate limit** (per-tenant or per-IP sliding window) on hot paths: chat, voice WS open, KM write
- **Quota enforcement** hooks:
  - concurrent calls / active voice sessions → `max_concurrent_calls`
  - monthly voice minutes (approx. on session end) → `max_monthly_call_minutes`
  - KM document count → `max_km_documents`
  - AI employees assigned → `max_ai_employees`
  - feature flags `voice_enabled` / `rag_enabled`
- Platform APIs: tenant usage snapshot; optional global rate-limit config
- Platform admin UI: usage vs limits for a tenant (read-only)
- `/api/infra` quota/rate-limit status; env docs
- Design pack: [16-quota-rate-limit-spec.md](../02-design/16-quota-rate-limit-spec.md)

**Out:**
- Tenant self-service limit UI (→ SPRINT-016)
- ClickHouse usage metering warehouse / cost dashboards (→ SPRINT-025+)
- Login rate limit Phase F (auth-cache)
- Billing overage charges / auto-upgrade
- NATS usage events (optional later)

## Acceptance criteria

1. With Starter entitlement, exceeding `max_km_documents` returns **HTTP 429** (or 403 for feature flag) with stable error code; document is not ingested.
2. Concurrent voice sessions above `max_concurrent_calls` are rejected until a slot is released.
3. Chat/voice respect `voice_enabled` / `rag_enabled` from effective entitlement.
4. Redis keys use prefix `monti_jarvis:` and documented shapes for monthly + concurrent counters.
5. `platform_admin` can `GET` usage vs limits for any tenant; portal shows a usage panel.
6. Rate limit excess returns **429** with `Retry-After` when configured; disabled when Redis down (fail-open or fail-closed per env — documented).
7. `go test ./internal/quota/...` and manual UAT checklist green; customer no-auth demo path still works when `AUTH_DISABLED=true` with demo entitlement or permissive fallback documented.

## Test notes

- Functional: assign Starter → push KM docs to limit → assert 429 → raise package or revoke docs → succeed
- Concurrent: open N voice sessions = limit, (N+1) fails
- Rate limit: burst chat requests from one tenant
- Languages: Thai + English error messages optional; English codes required

## Dependencies

- packages: `internal/entitlements`, `internal/store`, Redis 8
- FEAT-0003 Auth/RBAC, FEAT-0004 Packages
- blueprint: Redis quotas/rate limits; Phase B Sprint 13

## Design links

| Artifact | Path |
| --- | --- |
| Deep spec | [16-quota-rate-limit-spec.md](../02-design/16-quota-rate-limit-spec.md) |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Quota |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §32–36 |
| ER / Redis | [03-er-diagram.md](../02-design/03-er-diagram.md) |
| UX P14 | [05-ux-ui.md](../02-design/05-ux-ui.md) § P14 |
| Sprint | [SPRINT-013](../03-sprints/SPRINT-013.md) |
