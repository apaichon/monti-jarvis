---
id: FEAT-0046
title: "Conversation auto-register customer via email OTP + app version on UI"
status: shipped
release: v2.26.0
roadmap_sprint: 53
priority: D+
depends_on: [SPRINT-016, SPRINT-019, SPRINT-020, SPRINT-021, SPRINT-052]
design: DES-0050
design_spec: ../02-design/50-conversation-auto-register-app-version-spec.md
updated: 2026-07-29
---

# FEAT-0046: Conversation Auto-Register (Email OTP) + App Version on UI

## Purpose

1. Let tenants enable **auto-register** when a customer enters **email** during
   conversation and completes **OTP** — create/reuse the customer and bind the
   session without a separate signup path.
2. Show the **same app version as the git tag / `VERSION` file** on primary UIs.

## Problem

- Friction-light tenants want identity capture mid-chat without forcing
  pre-conversation OTP when workforce auth is optional.
- Operators cannot tell which release is running from the UI alone.

## Scope

### In

- Tenant setting: `auto_register_on_conversation_otp` (default **off**)
- Conversation UI: email → request OTP → verify → auto-create/reuse customer →
  bind session
- Reuse S20 OTP challenges, domain policy, rate limits, `customers` directory
- App version from `VERSION` / tag exposed via API and shown on tenant +
  platform-admin shells (customer portal optional footer)

### Out

- SMS OTP, social login
- Mandatory email on every anonymous demo when setting is off
- Changing S51 commercial catalog
- Marketing email blasts

## Acceptance criteria

1. Tenant admin can enable/disable auto-register in Settings; default off.
2. When on: customer can enter email in conversation, receive OTP, verify;
   missing customer is created; existing customer is reused; session binds.
3. When off: conversation does not offer this auto-register path.
4. Domain allowlist and OTP rate limits still apply; tenant isolation holds.
5. Require-auth-for-workforce coexists without double registration.
6. UI shows app version matching deployed `VERSION` / tag (e.g. `v2.24.0`).
7. Version display leaks no secrets or env names.

## Test notes

- Setting on/off; domain reject; rate limit; two-tenant isolation
- Verify session used for quota/attribution after OTP
- Health/version endpoint matches `VERSION` file

## Dependencies

- `internal/store` customer auth + customers; `cmd/server`; tenant settings UI;
  customer-web conversation shell
- Design: [DES-0050](../02-design/50-conversation-auto-register-app-version-spec.md)
