---
id: DES-0023
title: Customer Authentication and Domain Enforcement Specification
status: shipped
updated: 2026-07-13
sprint: SPRINT-020
owner: SA
---

# Customer Authentication and Domain Enforcement — Design Spec

**Sprint:** SPRINT-020 · **Release target:** v2.1.0
**Feature:** [FEAT-0022](../01-features/FEAT-0022-customer-auth.md)
**Depends on:** [06-auth-spec.md](06-auth-spec.md), [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md), [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md), [22-customer-account-import-spec.md](22-customer-account-import-spec.md)

## 1. Goals

1. Let tenants enable customer authentication after SPRINT-019 customer identities exist.
2. Bind customer email OTP verification and sessions to a tenant-scoped `customers` row.
3. Enforce SPRINT-019 domain `allow`/`deny` rules before sending an OTP and again before session issuance.
4. Preserve the public no-auth conversation path unless tenant settings explicitly require auth.
5. Pass customer, tier, group, and locale context into chat/voice quota and RAG resolution.
6. Produce UAT evidence for multi-user quota/rate-limit tenant isolation.

## 2. Non-goals

- Full customer OAuth/social login.
- Password-based customer login, password reset, or long-lived shared secrets.
- Magic links, invitation workflow, or customer email preference management.
- Customer-facing billing, tickets, KYC, discounts, or conversation history.
- Opening production customer traffic without signed readiness evidence.

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `CUSTOMER_AUTH_OTP_LENGTH` | `6` | Numeric OTP code length |
| `CUSTOMER_AUTH_OTP_TTL_MINUTES` | `10` | OTP challenge lifetime |
| `CUSTOMER_AUTH_OTP_RESEND_SECONDS` | `60` | Minimum resend interval for same tenant/email |
| `CUSTOMER_AUTH_OTP_MAX_ATTEMPTS` | `5` | Verification attempts before challenge is locked |
| `CUSTOMER_AUTH_SESSION_TTL_MINUTES` | `60` | Customer access/session cache TTL |
| `CUSTOMER_AUTH_REFRESH_TTL_MINUTES` | `43200` | Refresh token lifetime, 30 days |
| `CUSTOMER_AUTH_RATE_LIMIT_PER_MINUTE` | `5` | OTP request/verify attempt cap per tenant/email hash |

Email delivery uses the existing project mailer/resend configuration. SPRINT-020 does not introduce a new mail provider.

## 4. Data model — Postgres `monti_jarvis.callcenter`

### `tenant_customer_auth_settings`

| Column | Type | Notes |
| --- | --- | --- |
| `tenant_id` | text PK/FK | Tenant owner |
| `enabled` | boolean | Enables customer auth endpoints |
| `mode` | text | `disabled`, `optional`, `required` |
| `domain_enforcement` | text | `off`, `allowlist`, `denylist`, `allowlist_and_denylist` |
| `allow_public_no_auth` | boolean | Keeps current no-auth portal path available |
| `session_ttl_minutes` | int | Bounded access/session TTL |
| `refresh_ttl_minutes` | int | Bounded refresh TTL |
| audit columns | timestamptz/text | `created_at`, `updated_at`, `created_by`, `updated_by` |

### `customer_auth_identities`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `caid_{ulid}` |
| `tenant_id` | text FK | Tenant isolation |
| `customer_id` | text FK | SPRINT-019 customer |
| `email_normalized` | text | Lower-case customer auth identity |
| `status` | text | `active` or `locked` |
| `verified_at` | timestamptz | First successful OTP verification |
| `last_login_at` | timestamptz | Successful OTP login timestamp |
| audit columns | timestamptz/text | Standard audit contract |

Unique constraints: `(tenant_id, email_normalized)` and `(tenant_id, customer_id)` for active customer auth identity.

### `customer_otp_challenges`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `cotp_{ulid}`; returned as `challenge_id` |
| `tenant_id` | text FK | Tenant isolation |
| `customer_id` | text FK | Nullable until self-claim creates customer |
| `email_normalized` | text | Lower-case email receiving OTP |
| `purpose` | text | `login`, `register`, or `claim` |
| `otp_hash` | text | Hash of OTP code; never returned or logged |
| `status` | text | `pending`, `verified`, `expired`, `locked` |
| `attempt_count` | int | Incremented on verify attempts |
| `expires_at` | timestamptz | OTP expiry |
| `sent_at` | timestamptz | Last email send timestamp |
| `verified_at` | timestamptz | Successful verification timestamp |
| `ip_hash` | text | Hashed remote address |
| audit columns | timestamptz/text | Standard audit contract |

### `customer_sessions`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `csess_{ulid}` |
| `tenant_id` | text FK | Tenant isolation |
| `customer_id` | text FK | SPRINT-019 customer |
| `auth_identity_id` | text FK | OTP identity that issued the session |
| `refresh_token_hash` | text | Opaque refresh token hash |
| `status` | text | `active`, `revoked`, `expired` |
| `expires_at` | timestamptz | Access/session expiry |
| `refresh_expires_at` | timestamptz | Refresh expiry |
| `revoked_at` | timestamptz | Logout/security revocation |
| `user_agent` | text | Optional display/debug |
| `ip_hash` | text | Hashed remote address for abuse checks |
| audit columns | timestamptz/text | Standard audit contract |

### `customer_auth_events`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `caevt_{ulid}` |
| `tenant_id` | text FK | Tenant isolation |
| `customer_id` | text FK | Nullable for failed pre-identity attempts |
| `event_type` | text | `otp_requested`, `otp_sent`, `otp_verified`, `login`, `refresh`, `logout`, `domain_denied` |
| `result` | text | `success` or `failure` |
| `reason` | text | Error code without secrets |
| `ip_hash` | text | Hashed remote address |
| `metadata` | jsonb | Bounded, no passwords/tokens |
| audit columns | timestamptz/text | Standard audit contract |

## 5. Redis / NATS / ClickHouse

| Key / subject | Purpose |
| --- | --- |
| `monti_jarvis:customer_session:{session_id}` | Hot session lookup with access/session TTL |
| `monti_jarvis:rate:customer_auth:{tenant}:{email_hash}` | OTP request/verify abuse limit |
| `monti_jarvis:customer_otp:{challenge_id}` | Optional hot OTP challenge cache with OTP TTL |
| Existing S13 quota/rate keys | Authenticated chat/voice attribution must include tenant and customer context |

No new NATS subject is required. No ClickHouse or MinIO table/object is added in SPRINT-020.

## 6. Domain policy

| Mode | Behavior |
| --- | --- |
| `off` | Ignore domain rules for auth |
| `allowlist` | Require active `allow` rule for email domain |
| `denylist` | Reject active `deny` rule for email domain |
| `allowlist_and_denylist` | Deny wins; otherwise require active allow |

Domain matching uses the normalized email domain and tenant-owned active `customer_domain_rules`.

## 7. API summary

See [04-api-spec.md](04-api-spec.md#customer-authentication--domain-enforcement-sprint-20).

| Method | Path | Role |
| --- | --- | --- |
| `GET`, `PUT` | `/api/tenant/customer-auth/settings` | active `tenant_admin` |
| `POST` | `/api/customer/auth/request-otp` | public tenant context |
| `POST` | `/api/customer/auth/verify-otp` | public tenant context |
| `POST` | `/api/customer/auth/refresh` | refresh token |
| `POST` | `/api/customer/auth/logout` | `customer` |
| `GET` | `/api/customer/me` | `customer` |

### OTP request and email response contract

`POST /api/customer/auth/request-otp` validates tenant settings and domain policy, creates a pending OTP challenge, sends the OTP to the customer email, and returns only safe challenge metadata.

Request:

```json
{
  "email": "jane@example.com",
  "purpose": "login",
  "display_name": "Jane Doe",
  "external_id": "crm-42"
}
```

Response `202`:

```json
{
  "challenge_id": "cotp_01",
  "status": "otp_sent",
  "delivery": {
    "channel": "email",
    "to": "j***@example.com",
    "expires_in": 600,
    "resend_after": 60
  },
  "customer_hint": {
    "matched_existing_customer": true,
    "requires_profile_completion": false
  }
}
```

The email body contains a short numeric OTP and expiry. It must not include access tokens, refresh tokens, tenant ids that are not already visible to the user, or customer metadata beyond the greeting.

`POST /api/customer/auth/verify-otp` verifies the challenge and code, creates or activates the customer auth identity, issues the customer session, and returns the session response.

Request:

```json
{
  "challenge_id": "cotp_01",
  "otp": "123456"
}
```

Response `200`:

```json
{
  "status": "authenticated",
  "access_token": "<jwt>",
  "refresh_token": "<opaque>",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_expires_in": 2592000,
  "customer": {
    "id": "cust_01",
    "tenant_id": "demo",
    "display_name": "Jane Doe",
    "email": "jane@example.com",
    "tier_id": "tier_vip",
    "group_ids": ["grp_retail"],
    "locale": "th",
    "role": "customer"
  }
}
```

## 8. RBAC

| Action | `platform_admin` | active `tenant_admin` | `customer` | public |
| --- | ---: | ---: | ---: | ---: |
| Configure tenant customer auth | no | yes | no | no |
| Request customer OTP | no | no | no | yes, when enabled |
| Verify OTP / refresh customer session | no | no | no | yes, when enabled |
| Read current customer profile | no | no | yes | no |
| Revoke own customer session | no | no | yes | no |
| Access another tenant's customer/session | no | no | no | no |

## 9. Chat/voice integration

Customer Bearer tokens are optional on public chat/voice paths while tenant mode is `optional`. When present, the server loads customer, tier, group, and locale context before quota, rate-limit, RAG, and prompt construction. When tenant mode is `required`, unauthenticated customer-specific access returns `401 unauthorized`; the public no-auth portal remains available only if `allow_public_no_auth=true`.

## 10. Verification

```bash
make test && make build && make infra-check

curl -sS -X PUT -H "Authorization: Bearer $TENANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled":true,"mode":"optional","domain_enforcement":"allowlist_and_denylist","allow_public_no_auth":true}' \
  http://localhost:8091/api/tenant/customer-auth/settings | jq .

curl -sS -X POST -H "Content-Type: application/json" \
  -d '{"email":"jane@example.com","purpose":"login"}' \
  http://localhost:8091/api/customer/auth/request-otp | jq .

curl -sS -X POST -H "Content-Type: application/json" \
  -d '{"challenge_id":"cotp_01","otp":"123456"}' \
  http://localhost:8091/api/customer/auth/verify-otp | jq .
```

Manual UAT must use at least two tenants and multiple customers to verify cross-tenant isolation and quota/rate-limit attribution.

## 11. Related artifacts

| Artifact | Link |
| --- | --- |
| Sprint | [SPRINT-020](../03-sprints/SPRINT-020.md) |
| Feature | [FEAT-0022](../01-features/FEAT-0022-customer-auth.md) |
| Workflow | [02-workflow.md](02-workflow.md) §59–63 |
| ER | [03-er-diagram.md](03-er-diagram.md) § Sprint 20 |
| API | [04-api-spec.md](04-api-spec.md) § Customer Authentication & Domain Enforcement |
| UX | [05-ux-ui.md](05-ux-ui.md) § T13 |

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | | ☐ |
| Dev | | | ☐ |
| Tester | | | ☐ |
