---
id: FEAT-0022
title: Customer Authentication and Domain Enforcement
status: shipped
sprint: SPRINT-020
owner: PM
updated: 2026-07-13
---

# Feature: Customer Authentication and Domain Enforcement

## Problem

SPRINT-019 created tenant-scoped customer identities, but customers still cannot authenticate, claim an imported identity, or receive tier/group benefits through a session. Tenants also need a controlled way to enforce domain rules before opening authenticated customer traffic.

## Scope

**In**

- Customer OTP identity/session tables linked to SPRINT-019 `customers`.
- Email OTP request/verify, logout, session refresh, and current-customer profile APIs.
- Tenant-controlled customer-auth settings and domain-rule enforcement.
- Customer portal login/account UX that preserves the existing public no-auth conversation path.
- Quota/rate-limit isolation verification under authenticated multi-user traffic.

**Out**

- Full OAuth/social login for customers.
- Password login, password reset, magic links, invitation workflow, or self-service billing.
- Customer tickets, conversation history, discounts, or KYC.
- Opening production customer traffic before the readiness gate is signed off.

## Acceptance criteria

1. Active tenant admin can enable/disable customer authentication and choose domain-rule enforcement mode.
2. Customer OTP identities bind to a tenant-scoped SPRINT-019 customer without cross-tenant leakage.
3. Customer OTP request sends a short code by email and returns safe challenge metadata.
4. Customer OTP verification issues a customer-scoped session/JWT with role `customer`; logout and refresh behave predictably.
5. Domain `allow`/`deny` rules apply before OTP delivery and session issuance according to tenant settings.
6. Customer portal can show OTP login/account state while preserving the no-auth conversation entry.
7. Authenticated chat/voice traffic applies tenant, tier, group, quota, and rate-limit context correctly.
8. Tenant A customer sessions cannot read or consume Tenant B data/quota.
9. Manual UAT records multi-user quota/rate-limit isolation before customer production traffic is allowed.

## Dependencies

- [FEAT-0003 - Auth and RBAC](FEAT-0003-auth-rbac.md)
- [FEAT-0013 - Quota & Rate Limit](FEAT-0013-quota-rate-limit.md)
- [FEAT-0016 - Tenant Settings, Locale, Limits](FEAT-0016-tenant-settings-locale-limits.md)
- [FEAT-0021 - Customer Account Import and Domain Integration](FEAT-0021-customer-account-import.md)

## Design

- [DES-0023 — Customer Authentication and Domain Enforcement Specification](../02-design/23-customer-auth-spec.md)
- [API contract — Customer Authentication & Domain Enforcement](../02-design/04-api-spec.md#customer-authentication--domain-enforcement-sprint-20)
- Existing customer import/domain design: [22-customer-account-import-spec.md](../02-design/22-customer-account-import-spec.md)
