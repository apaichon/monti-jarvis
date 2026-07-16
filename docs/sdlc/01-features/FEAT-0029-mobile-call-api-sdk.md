---
id: FEAT-0029
title: "Mobile Call API and SDK"
status: shipped
owner: product
created: 2026-07-15
updated: 2026-07-16
sprint: SPRINT-027
---

# FEAT-0029: Mobile Call API and SDK

## Purpose

Give tenant integrators a stable mobile contract for starting and operating inbound AI voice calls without coupling their applications to the web embed surface or provider credentials.

## Scope

- Versioned mobile call-session creation, status, end-call, transcript, and rating contracts.
- Mobile bootstrap response with tenant policy, active assigned avatars, default avatar, locale, timezone, and time limits.
- Customer email OTP login with optional non-sensitive APNs/FCM notification handoff for iOS and Android.
- Reuse customer Bearer authentication, tenant resolution, avatar assignment, quota, rate-limit, and tenant-isolation policy.
- Mobile-safe WebSocket handshake and a bounded PCM16 audio/event protocol.
- Typed SDK lifecycle surface with token refresh, reconnect, transcript callbacks, and explicit end-call control.
- Reference integration, compatibility guidance, provider-error redaction, and contract tests.

## Out of scope

- A new customer identity or tenant entitlement model.
- Direct Gemini, LiveKit, MinIO, or other provider credentials in a mobile client.
- Native production applications for every mobile platform.
- Replacing the existing web embed routes or /api/calls contract.
- Platform audit logs, system monitoring, billing, or analytics dashboards.

## Acceptance criteria

1. An authenticated mobile client can create a call session for an assigned avatar without sending a tenant selector in the request body.
2. The mobile API returns only versioned Monti session data and a bounded client-facing error contract; provider URLs, credentials, raw errors, and internal tenant metadata are never returned.
3. A mobile client can connect to the call WebSocket, receive readiness, transcript, audio, turn, status, and error events, and reconnect within the documented lifecycle rules.
4. Create and end operations are idempotent; duplicate requests do not create duplicate active sessions or duplicate close side effects.
5. Customer auth-required tenants reject missing or invalid customer sessions, while local AUTH_DISABLED behavior remains compatible with the existing public development flow.
6. The SDK exposes token refresh, call state, transcript callbacks, connection recovery, explicit end-call, and 1-5 star rating operations.
7. Cross-tenant call access, unassigned avatars, quota/rate-limit violations, expired sessions, and oversized audio frames return stable safe errors.
8. Bootstrap returns only active tenant-assigned avatars, resolves locale using customer then tenant settings with en fallback, and exposes Asia/Bangkok as the default timezone when unset.
9. The call contract exposes the server-authoritative per-call cap, daily remaining seconds, 10-second warning threshold, and tenant-local reset boundary.

## Design links

- Sprint: [SPRINT-027](../03-sprints/SPRINT-027.md)
- DES-0030 - Mobile Call API and SDK Specification: [30-mobile-call-api-sdk-spec.md](../02-design/30-mobile-call-api-sdk-spec.md)
- API specification: [04-api-spec.md](../02-design/04-api-spec.md) - Mobile Call API and SDK
