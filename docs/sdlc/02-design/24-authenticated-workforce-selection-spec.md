---
id: DES-0024
title: Authenticated Workforce Selection and Customer Quota Specification
status: approved
updated: 2026-07-13
sprint: SPRINT-021
owner: SA
---

# Authenticated Workforce Selection and Customer Quota — Design Spec

**Sprint:** SPRINT-021 · **Release target:** v2.2.0  
**Feature:** [FEAT-0023](../01-features/FEAT-0023-authenticated-workforce-selection.md)  
**Depends on:** [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md), [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md), [23-customer-auth-spec.md](23-customer-auth-spec.md)

## 1. Goals

- Let tenants require customer OTP sign-in before AI workforce selection.
- Keep optional-auth tenants compatible with the existing public no-auth customer portal.
- Return only active tenant-assigned avatars in the customer workforce picker.
- Enforce customer-aware daily call/chat time and per-call duration limits.
- Show customer-facing quota and blocked states without exposing package internals.

## 2. Non-goals

- Password, OAuth, social login, or new identity providers.
- Tickets, satisfaction surveys, human handoff, or conversation history.
- New package SKUs or billing rules.
- Broad production load testing beyond manual UAT evidence.

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `CUSTOMER_WORKFORCE_AUTH_REQUIRED_DEFAULT` | `false` | Default for tenants without explicit setting |
| `CUSTOMER_DAILY_CALL_SECONDS_DEFAULT` | `0` | `0` means no additional customer daily cap beyond package/tier rules |
| `CUSTOMER_MAX_CALL_SECONDS_DEFAULT` | `0` | `0` means no additional per-call cap beyond tenant/package rules |

## 4. Data model

SPRINT-021 extends `tenant_customer_auth_settings`:

| Column | Type | Notes |
| --- | --- | --- |
| `require_auth_for_workforce` | boolean | Blocks workforce selection and call/chat start until customer session is valid |
| `customer_daily_call_seconds` | int | Optional per-customer daily cap |
| `customer_max_call_seconds` | int | Optional max seconds for one voice call |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

Optional usage ledger if existing call/counter records are not enough:

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `cuevt_{ulid}` |
| `tenant_id` | text FK | Tenant isolation |
| `customer_id` | text FK | Customer charged for usage |
| `session_id` | text FK | Customer session when present |
| `avatar_id` | text FK | Selected AI workforce |
| `usage_type` | text | `chat` or `voice` |
| `reserved_seconds` | int | Reservation before voice starts |
| `consumed_seconds` | int | Actual committed usage |
| `status` | text | `reserved`, `committed`, `released`, `denied` |
| `deny_reason` | text | Safe quota denial code |
| `usage_date` | date | Tenant-local usage day |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

## 5. Redis

| Key | Purpose |
| --- | --- |
| `monti_jarvis:quota:{tenant}:customer:{customer}:day:{yyyymmdd}` | Daily customer usage seconds |
| `monti_jarvis:quota:{tenant}:customer:{customer}:avatar:{avatar}:day:{yyyymmdd}` | Optional per-avatar usage attribution |
| `monti_jarvis:rate:{tenant}:customer:{customer}:chat` | Customer chat rate limiter |
| `monti_jarvis:rate:{tenant}:customer:{customer}:call` | Customer call start limiter |

## 6. API summary

See [04-api-spec.md](04-api-spec.md) § Authenticated Workforce Selection & Customer Quota.

| Method | Path | Role |
| --- | --- | --- |
| `GET` | `/api/customer/portal-policy` | public |
| `GET` | `/api/customer/workforce` | public or `customer`, depending on tenant policy |
| `GET` | `/api/customer/quota` | public or `customer`, depending on tenant policy |
| `PUT` | `/api/tenant/customer-auth/settings` | `tenant_admin` |
| `POST` | `/api/chat` | public or `customer`, depending on tenant policy |
| `POST` | `/api/calls` | public or `customer`, depending on tenant policy |

## 7. RBAC

| Action | public | customer | tenant_admin | platform_admin |
| --- | --- | --- | --- | --- |
| Read safe portal policy | yes | yes | no | no |
| Select workforce when auth optional | yes | yes | no | no |
| Select workforce when auth required | no | yes | no | no |
| Start chat/voice when auth required | no | yes | no | no |
| Edit workforce auth/quota settings | no | no | yes | platform support only if existing policy allows |

## 8. Verification

```bash
make test && make build
curl "http://localhost:8091/api/customer/portal-policy?tenant_id=libra-tech-co-ltd"
curl -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  "http://localhost:8091/api/customer/workforce?tenant_id=libra-tech-co-ltd"
curl -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  "http://localhost:8091/api/customer/quota?tenant_id=libra-tech-co-ltd"
```

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | | ☐ |
| Dev | | | ☐ |
