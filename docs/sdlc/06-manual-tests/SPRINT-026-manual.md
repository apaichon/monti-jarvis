---
id: UAT-026
title: Sprint 26 tenant system performance monitoring
sprint: SPRINT-026
status: deferred_uat
release: v2.7.0
---

# Sprint 26 UAT

## Preconditions

- Tenant admin can sign in to the tenant console.
- The server is running with the tenant-web application available.
- The tenant has a valid active tenant-admin session.
- Postgres, Redis, MinIO, ClickHouse, NATS, LiveKit, and Gemini may be enabled or disabled independently for state coverage.

## Scenarios

| ID | Scenario | Expected result |
| --- | --- | --- |
| UAT-026-01 | Open `/tenant/monitoring` as an active tenant admin | The page loads a normalized system snapshot for the current tenant and does not require a tenant ID query parameter. |
| UAT-026-02 | Open `GET /api/tenant/system-performance` without a tenant session | The API returns `401` with the standard safe error shape and no dependency metadata. |
| UAT-026-03 | Probe a configured healthy dependency | The component shows `operational`, a non-negative latency, and a checked timestamp. |
| UAT-026-04 | Disable an optional dependency | The component shows `disabled`; no raw configuration, URL, credential, or provider detail is shown. |
| UAT-026-05 | Make a configured dependency unavailable or exceed the probe timeout | The component shows `unavailable`; the page shows the degraded/unavailable guidance and retry action. |
| UAT-026-06 | Return stale or unavailable ClickHouse analytics | Analytics shows `stale` or `unavailable`, while the overall state is `degraded` unless a dependency is unavailable. |
| UAT-026-07 | Attempt to use a tenant-admin session from another tenant | The response contains only that session's tenant-scoped snapshot; no other tenant identifiers or records are exposed. |
| UAT-026-08 | Use the retry action after a transient failure | The page requests a fresh snapshot and replaces the prior feedback without a full navigation. |
| UAT-026-09 | Inspect desktop and mobile widths | Status, latency, timestamps, analytics freshness, and retry controls remain readable without overlap or horizontal breakage. |
| UAT-026-10 | Run existing customer call, archive, quota, statistics, and `/healthz` flows | Existing flows continue to work; monitoring is read-only and does not change their response or persistence behavior. |

## Evidence

- Capture the API response for UAT-026-01 with component statuses redacted to safe categories.
- Capture the `401` response for UAT-026-02.
- Capture screenshots of operational, degraded/unavailable, stale, retry, and mobile states.
- Record the existing regression commands and `/healthz` response used for UAT-026-10.

## Automated checks

- `/usr/local/go/bin/go test ./...`
- `/usr/local/go/bin/go build ./cmd/server`
- `npm run check` in `apps/tenant-web`
- `npm run build` in `apps/tenant-web`

## Release close note

Automated API authorization, redaction, timeout, Go, server-build, tenant-web check/build, and `/healthz` smoke validation passed on 2026-07-15. Manual browser and responsive-layout evidence is deferred to the next tester run.
