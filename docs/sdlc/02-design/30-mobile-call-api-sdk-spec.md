---
id: DES-0030
title: Mobile Call API and SDK Specification
status: approved
updated: 2026-07-15
sprint: SPRINT-027
owner: SA
---

# Mobile Call API and SDK - Design Spec

Sprint: SPRINT-027 · Release target: v2.8.0  
Feature: FEAT-0029 - Mobile Call API and SDK  
Depends on: 23-customer-auth-spec.md and 24-authenticated-workforce-selection-spec.md

## 1. Goals

- Provide a versioned mobile facade over the existing tenant-safe call session and voice relay behavior.
- Keep tenant, customer, avatar, quota, rate-limit, and scope policy decisions on the Go server.
- Give mobile clients one typed lifecycle for create, connect, observe, reconnect, end, and rate.
- Support customer email OTP login with optional non-sensitive iOS and Android push notifications.
- Return tenant bootstrap data in one mobile response: assigned avatars, default avatar, resolved locale, timezone, and server-authoritative call limits.
- Hide Gemini, LiveKit, MinIO, Redis, Postgres topology, credentials, and raw provider errors.
- Preserve the existing web /api/calls and /ws/voice behavior while the mobile contract is introduced.

## 2. Non-goals

- New tenant or customer authentication issuance. Sprint 27 reuses customer access and refresh tokens from Sprint 20.
- Direct provider SDK access from iOS, Android, React Native, or Flutter applications.
- Cross-tenant administration, audit history, platform monitoring, billing, or analytics reporting.
- Audio transcoding or noise-removal policy beyond the existing voice relay contract.
- Shipping native wrappers for every platform before the SDK target decision is approved.

## 3. Environment and compatibility

| Variable / header | Contract |
| --- | --- |
| AUTH_DISABLED | Local development may retain the existing public no-auth flow; production tenant policies still enforce their configured customer-auth mode. |
| MOBILE_CALL_API_ENABLED | Feature gate for /api/mobile/v1 and /ws/mobile/v1; default false until the release is enabled. |
| MOBILE_WS_MAX_FRAME_BYTES | Maximum decoded audio payload per message; default 32768. Oversized frames return audio_frame_too_large. |
| MOBILE_PUSH_ENABLED | Optional APNs/FCM notification delivery; default false. Email OTP remains the required authentication channel. |
| MOBILE_PUSH_PROVIDER | Server-side provider selection: apns, fcm, or auto. Credentials are never shipped to clients. |
| MOBILE_PUSH_TOKEN_TTL | Maximum retention for transient active-call notification metadata; default 15m after call end. |
| Idempotency-Key | Required on mobile create; recommended on end and rating. Keys are scoped to tenant, caller, route, and a bounded TTL. |
| X-Monti-SDK-Version | Client SDK version for support telemetry; never used for authorization. |

The mobile API uses the same deployment base URL and CORS policy as the existing API. The server must not return provider hostnames, LiveKit room tokens, Gemini keys, MinIO paths, or raw infrastructure errors.

## 4. Data model

Sprint 27 adds no new Postgres table, migration, ClickHouse table, MinIO object, or new long-lived session entity. OTP challenges and customer sessions remain the Sprint 20 source of truth. Push metadata is transient and bounded to authentication or active-call lifetimes.

| Existing contract | Mobile use |
| --- | --- |
| call_sessions | Source of truth for call_id, tenant, room, avatar, customer, status, start, and end timestamps. Existing audit columns remain required. |
| call_turns | Transcript persistence and record retrieval after the call. |
| customer_sessions | Customer access/refresh token lifecycle and tenant binding. |
| ai_avatars + tenant_avatar_assignments | Validate the requested avatar is active and assigned to the resolved tenant. |
| conversation_ratings | Existing one-row-per-tenant/call 1-5 rating contract. |

Transient mobile coordination keys are:

~~~text
monti_jarvis:mobile:idempotency:{tenant_id}:{caller_subject}:{route}:{key}
monti_jarvis:mobile:notify:{call_id}
~~~

The idempotency key stores a bounded response digest and expires after the configured idempotency window. The notify key stores only an encrypted or opaque APNs/FCM token reference and expires at the call end plus MOBILE_PUSH_TOKEN_TTL. Active-call state continues to use the existing monti_jarvis:call:active:{call_id} key. No customer audio, transcript, email, OTP, or raw push token is stored in Redis.

### Mobile bootstrap contract

`GET /api/mobile/v1/bootstrap` is the first mobile request after the app knows the tenant context. Authenticated callers derive tenant context from the customer token; public development/optional-auth callers use the existing trusted portal context. The mobile client must not use a request-body tenant_id to switch tenants.

Response 200:

~~~json
{
  "tenant": { "display_name": "Libra Tech Co.,Ltd", "slug": "libra-tech-co-ltd" },
  "auth": {
    "enabled": true,
    "mode": "required",
    "require_auth_for_workforce": true,
    "allow_public_no_auth": false,
    "otp": { "channel": "email", "ttl_seconds": 600, "resend_after_seconds": 60 }
  },
  "locale": {
    "default": "th",
    "customer": "th",
    "ai_reply": "th",
    "supported": ["en", "th"],
    "timezone": "Asia/Bangkok"
  },
  "avatars": [
    {
      "id": "ava",
      "name": "Ava",
      "role": "General Support",
      "trait": "Warm & Patient",
      "image_url": "/api/assets/avatars/ava/portrait.png",
      "status": "active",
      "is_default": true
    }
  ],
  "default_avatar_id": "ava",
  "limits": {
    "max_call_seconds": 300,
    "daily_limit_seconds": 1800,
    "daily_remaining_seconds": 1200,
    "warning_at_seconds": 10,
    "reset_at": "2026-07-16T00:00:00+07:00",
    "state": "quota_available"
  }
}
~~~

Bootstrap rules:

- avatars contains only active avatars assigned to the resolved tenant, ordered by the tenant assignment order or stable avatar id. Archived, disabled, and unassigned avatars are excluded.
- default_avatar_id uses this precedence: valid request selection, tenant or embed default_agent_id when still assigned, then the first active assigned avatar. If no assignment exists, retain the existing static workforce fallback only for the documented demo/development compatibility path.
- locale resolution is customer.locale, then tenant ai_reply_locale, then tenant locale, then en. Supported locales are en and th. The default tenant timezone is Asia/Bangkok when no tenant setting exists.
- max_call_seconds is the lowest non-zero cap from customer policy, tenant call limits, tier/package policy, or null when no cap applies. daily_remaining_seconds is the server-authoritative minimum remaining balance across applicable daily limits.
- reset_at is the next tenant-local quota-day boundary. The SDK must use server values for countdown decisions; device clock is display-only.
- anonymous callers receive auth and policy metadata, but customer_id and customer-specific remaining quota are omitted or returned as null.

### Customer OTP and mobile notification flow

Customer login remains passwordless OTP. Email is the authentication channel of record.

1. The app calls POST /api/customer/auth/request-otp with email, tenant context, locale, and optional mobile notification metadata.
2. The server validates customer-auth mode and domain policy, creates a six-digit challenge, and sends the OTP by email when the email is valid and delivery is configured.
3. If platform is ios or android and a push_token is supplied, APNs or FCM sends a non-sensitive notification such as Check your email for the Monti sign-in code. The OTP is never included in APNs/FCM payloads.
4. The app calls POST /api/customer/auth/verify-otp with challenge_id and otp, then stores the returned access and refresh tokens in the platform secure storage.
5. The SDK refreshes access tokens through POST /api/customer/auth/refresh and retries one interrupted status/WebSocket request. It never retries OTP creation automatically after an auth or domain failure.

Request additions:

~~~json
{
  "email": "jane@example.com",
  "locale": "th",
  "notification": {
    "platform": "ios",
    "push_token": "<apns-or-fcm-token>",
    "app_version": "1.0.0"
  }
}
~~~

notification is optional. Valid platform values are ios and android. Email remains required for the current S20 OTP contract; a missing email returns email_required rather than silently using push-only authentication.

OTP response notification state:

~~~json
{
  "challenge_id": "cotp_01",
  "status": "otp_sent",
  "delivery": { "channel": "email", "to": "j***@example.com" },
  "notifications": { "push": "queued|disabled|rejected" },
  "expires_in": 600,
  "resend_after": 60
}
~~~

Push notification types are otp_email_sent, call_ending_soon, call_ended, and review_available. A push failure never exposes an OTP and does not turn a successful email delivery into an auth failure. APNs and FCM credentials remain server secrets; the SDK only supplies a device token.

During an active call, the SDK may provide the same push token in the create-call notification object. The server may retain only an encrypted or opaque token reference in the bounded active-call Redis state for the call lifetime plus 15 minutes. No push token, email, OTP, or audio is written to a conversation record.

## 5. Mobile API summary

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| POST | /api/mobile/v1/calls | customer Bearer when required by tenant policy | Create or replay a mobile call session. |
| GET | /api/mobile/v1/calls/{call_id} | same caller and tenant context | Read safe lifecycle status. |
| POST | /api/mobile/v1/calls/{call_id}/end | same caller and tenant context | Idempotently request end-call. |
| POST | /api/mobile/v1/calls/{call_id}/rating | same caller and tenant context | Save or update a 1-5 star rating. |
| GET | /ws/mobile/v1/calls/{call_id} | Bearer during upgrade | Exchange bounded audio/control frames and receive events. |
| GET | /api/mobile/v1/bootstrap | public or customer context | Return mobile auth policy, assigned avatars, locale, timezone, and applicable limits. |
| POST | /api/customer/auth/request-otp | public tenant context | Send email OTP and optionally queue a non-sensitive APNs/FCM notification. |
| POST | /api/customer/auth/verify-otp | public tenant context | Verify OTP and issue customer access/refresh tokens. |
| POST | /api/customer/auth/refresh | refresh token | Rotate customer tokens for mobile session recovery. |

Tenant context is derived from the customer token or the existing trusted host/embed development context. tenant_id is never accepted as an authoritative body field. Every call id is checked against both the resolved tenant and caller policy before data is returned.

## 6. WebSocket protocol

The SDK opens GET /ws/mobile/v1/calls/{call_id} with the customer Bearer token and X-Monti-SDK-Version. The server validates the call before accepting media. v1 audio is mono PCM16 at 16 kHz, carried as base64 JSON chunks no larger than the decoded frame limit.

Client messages:

~~~json
{ "type": "start_audio", "sample_rate": 16000, "channels": 1, "encoding": "pcm_s16le" }
{ "type": "audio", "data_base64": "<pcm16 chunk>" }
{ "type": "text", "text": "optional typed input" }
{ "type": "end" }
~~~

Server events:

~~~json
{ "type": "ready", "call_id": "call_01", "avatar_id": "ava", "encoding": "pcm_s16le" }
{ "type": "audio", "data_base64": "<pcm16 chunk>" }
{ "type": "transcript", "role": "caller|assistant", "text": "...", "final": true }
{ "type": "call_status", "status": "active|ending|ended|failed" }
{ "type": "turn_complete" }
{ "type": "error", "code": "rate_limited", "retryable": true }
~~~

The server may close with a normal close code after call_status=ended. The SDK must treat a clean close after an ended status as success, reconnect only for retryable transport errors, and never replay audio after end.

## 7. SDK contract

The target platform is selected during implementation planning. The minimum typed surface is platform-neutral:

~~~typescript
type CallState = "idle" | "creating" | "connecting" | "active" | "ending" | "ended" | "failed";

interface MontiCallClient {
  getBootstrap(): Promise<MobileBootstrap>;
  createCall(input: { avatarId?: string; locale?: string; pushToken?: string }): Promise<CallHandle>;
  refreshToken(): Promise<void>;
  requestOTP(input: OTPRequest): Promise<OTPChallenge>;
  verifyOTP(input: { challengeId: string; otp: string }): Promise<CustomerSession>;
}

interface CallHandle {
  readonly id: string;
  readonly state: CallState;
  on(event: "transcript" | "status" | "error", listener: (event: unknown) => void): () => void;
  connect(): Promise<void>;
  sendAudio(pcm16: ArrayBuffer): void;
  sendText(text: string): void;
  end(reason?: "customer" | "timeout" | "network"): Promise<void>;
  rate(score: 1 | 2 | 3 | 4 | 5): Promise<void>;
}
~~~

The SDK owns frame splitting, reconnect backoff, access-token refresh, state transitions, and safe error mapping. It does not own tenant policy, provider credentials, raw WebSocket URLs, or database identifiers.

## 8. RBAC and policy matrix

| Action | Platform admin | Tenant admin | Authenticated customer | Anonymous |
| --- | --- | --- | --- | --- |
| Create mobile call in active tenant | no mobile caller route | no impersonation | yes, subject to tenant policy | only when existing public no-auth policy allows |
| Read call status | no | no cross-customer access | own call only | own anonymous call only |
| Connect voice WebSocket | no | no | own active call only | own public call only when allowed |
| End call | no | no | own call only | own public call only when allowed |
| Submit rating | no | no | own call only | existing public rating policy |
| Read mobile bootstrap | no | policy-scoped active tenant | own resolved tenant context | policy-scoped public context |
| Request or verify customer OTP | no | no impersonation | public pre-auth flow | public pre-auth flow |

Denied calls return stable 401, 403, or 404 responses without revealing whether another tenant or call exists. Quota, rate-limit, avatar assignment, and customer-auth checks occur before session creation and again at WebSocket upgrade.

## 9. Verification

~~~bash
curl -i -X POST http://localhost:8091/api/mobile/v1/calls \
  -H 'Authorization: Bearer <customer-access-token>' \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: mobile-demo-001' \
  -H 'X-Monti-SDK-Version: 0.1.0' \
  --data '{"avatar_id":"ava","locale":"th"}'

curl -i http://localhost:8091/api/mobile/v1/calls/call_01 \
  -H 'Authorization: Bearer <customer-access-token>'

curl -i -X POST http://localhost:8091/api/mobile/v1/calls/call_01/end \
  -H 'Authorization: Bearer <customer-access-token>' \
  -H 'Idempotency-Key: mobile-end-001'
~~~

Expected checks:

- Repeating create with the same idempotency key returns the same call id and does not reserve a second active call.
- An unassigned avatar, expired customer token, quota exhaustion, cross-tenant call id, or oversized frame returns the documented safe code.
- WebSocket events contain only the versioned Monti event envelope and never provider credentials or raw infrastructure errors.
- End-call is idempotent, closes the mobile session, and leaves the existing archive/rating path compatible.
- Existing /api/calls and /ws/voice customer-web flows remain unchanged.

- Bootstrap returns only active tenant-assigned avatars and the resolved default locale/timezone.
- A required-auth tenant rejects bootstrap call creation without a verified customer token.
- A valid email receives the OTP by email; APNs/FCM receives only a non-sensitive notification when configured.
- Missing email, invalid domain, expired OTP, invalid OTP, disabled push, and unavailable mail provider map to stable error codes.
- A mobile call receives the minimum applicable per-call/daily limit and server reset boundary; the SDK warns at 10 seconds but the server remains authoritative.

## Sibling artifacts

See 02-workflow.md, 03-er-diagram.md, 04-api-spec.md, and 05-ux-ui.md.
