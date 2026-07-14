---
id: FEAT-0023
title: Authenticated Workforce Selection and Customer Quota Enforcement
status: in_progress
sprint: SPRINT-021
owner: PM
updated: 2026-07-13
---

# Feature: Authenticated Workforce Selection and Customer Quota Enforcement

## Problem

SPRINT-020 lets customers authenticate with tenant-scoped OTP sessions, but the customer portal still allows conversation entry before consistently enforcing tenant policy, customer identity, and per-customer time limits. Tenants need a controlled customer flow where OTP can be required before AI workforce selection and all chat/voice usage is attributed to the signed-in customer.

## Scope

**In**

- Tenant policy to require customer OTP before AI workforce selection.
- Customer portal flow that gates avatar/workforce selection behind a valid customer session when required.
- Customer-aware workforce picker that filters unavailable/disabled/unassigned avatars.
- Per-customer call duration and daily usage enforcement using existing tenant/package/tier/group quota foundations.
- Clear customer-facing quota/time-limit states before and during calls.

**Out**

- New identity providers beyond SPRINT-020 email OTP.
- Tickets, conversation history, satisfaction surveys, or agent handoff.
- New billing package model beyond enforcing existing entitlement/quota settings.
- Production-scale load testing beyond the sprint UAT gate.

## Acceptance criteria

1. Tenant can configure whether customer OTP is required before workforce selection.
2. When required, anonymous customers cannot select an AI workforce or start chat/voice until OTP sign-in succeeds.
3. Signed-in customer context is displayed and preserved across workforce selection, chat, and voice call creation.
4. The workforce picker only shows avatars available to the tenant and allowed for the selected/customer context.
5. Per-customer daily call time and per-call duration limits are enforced with safe API errors.
6. Customer UI shows remaining/blocked quota states without leaking tenant internals.
7. Quota/rate-limit counters include tenant, customer, tier/group, and selected avatar context.
8. Manual UAT covers optional-auth and required-auth tenants plus exhausted quota behavior.

## Dependencies

- [FEAT-0005 - Avatar Catalog](FEAT-0005-avatar-catalog.md)
- [FEAT-0013 - Quota & Rate Limit](FEAT-0013-quota-rate-limit.md)
- [FEAT-0016 - Tenant Settings, Locale, Limits](FEAT-0016-tenant-settings-locale-limits.md)
- [FEAT-0022 - Customer Authentication and Domain Enforcement](FEAT-0022-customer-auth.md)

## Design

- [DES-0024 — Authenticated Workforce Selection and Customer Quota Specification](../02-design/24-authenticated-workforce-selection-spec.md)
- [API contract — Authenticated Workforce Selection & Customer Quota](../02-design/04-api-spec.md#authenticated-workforce-selection--customer-quota-sprint-21)
- [Workflow §64–65](../02-design/02-workflow.md#64-required-auth-workforce-selection-sprint-21)
- [UX § Sprint 21](../02-design/05-ux-ui.md#sprint-21--authenticated-workforce-selection-c14t14)
- Sprint plan: [SPRINT-021](../03-sprints/SPRINT-021.md)
