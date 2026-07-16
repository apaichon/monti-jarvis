---
id: UAT-027
title: Sprint 27 mobile call API and SDK
sprint: SPRINT-027
status: signed_off
release: v2.8.0
updated: 2026-07-16
---

# Sprint 27 UAT

## Preconditions

- Server runs with `MOBILE_CALL_API_ENABLED=true`.
- Customer authentication from Sprint 20 is available for authenticated mobile callers.
- Tenant has at least one active assigned AI employee.
- Redis quota and rate-limit settings use the Monti Jarvis DB/prefix defaults.

## Scenarios

| ID | Scenario | Expected result |
| --- | --- | --- |
| UAT-027-01 | Call `GET /api/mobile/v1/bootstrap` with a valid customer token | Response includes tenant policy, locale/timezone, assigned active avatars, default avatar, and authoritative limit values. |
| UAT-027-02 | Create a mobile call with an assigned avatar | Server creates a tenant/customer-scoped call session without trusting a body tenant selector. |
| UAT-027-03 | Create a mobile call with an unassigned avatar or another tenant's session | API returns a safe authorization or not-found error without exposing tenant internals. |
| UAT-027-04 | Repeat create/end requests with the same idempotency key | Duplicate requests replay the safe result and do not create duplicate active sessions or close side effects. |
| UAT-027-05 | Connect to `/ws/mobile/v1/calls/{call_id}` as the owning customer | WebSocket accepts the session and emits only versioned Monti event envelopes. |
| UAT-027-06 | Connect with an expired, invalid, oversized, or cross-tenant request | WebSocket rejects or closes with bounded provider-independent errors. |
| UAT-027-07 | Fetch transcript, end the call, and submit a 1-5 star rating | Lifecycle endpoints return mobile-safe data and persist the normal conversation/rating side effects. |
| UAT-027-08 | Use the TypeScript SDK against injectable fetch, storage, and WebSocket primitives | SDK exposes OTP, bootstrap, token refresh, call lifecycle, reconnect, transcript, end, and rating operations without provider credentials. |
| UAT-027-09 | Existing web chat, call, archive, quota, statistics, and `/healthz` flows still run | Mobile facade does not change existing web call behavior. |

## Automated Checks

- `/usr/local/go/bin/go test ./...`
- `/usr/local/go/bin/go vet ./...`
- `/usr/local/go/bin/go build ./cmd/server`
- `tsc -p packages/monti-mobile-sdk/tsconfig.json`
- `git diff --check`

## Release Close Note

Automated API contract, tenant isolation, mobile lifecycle, idempotency, redaction, Go, server-build, SDK TypeScript, and diff validation passed on 2026-07-16. Device-level iOS/Android smoke coverage and APNs/FCM provider delivery remain follow-up work; the shipped API accepts safe notification metadata and reports provider delivery as `not_configured` until credentials/adapters are configured.

## Sign-off

| Tester | Date | Result | Defects |
| --- | --- | --- | --- |
| Codex release verification | 2026-07-16 | Pass for v2.8.0 release | Device smoke and push provider delivery deferred |
