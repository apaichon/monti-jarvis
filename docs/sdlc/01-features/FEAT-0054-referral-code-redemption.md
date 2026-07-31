---
id: FEAT-0054
title: "Referral code redemption adds bonus quota"
status: completed
release: v2.35.0
roadmap_sprint: 62
priority: M+
depends_on: [SPRINT-013, SPRINT-045, SPRINT-046, SPRINT-051]
design: DES-0058
design_spec: ../02-design/58-referral-code-redemption-spec.md
updated: 2026-07-31
---

# FEAT-0054: Referral Code Redemption Adds Bonus Quota

## Purpose

Let a tenant admin enter and apply a referral code to receive eligible bonus
quota, with validation, abuse controls, expiry, and ledger tracking.

## Problem

S46 shipped referral attribution and bonus-quota rewards for the referrer path.
Operators also need manual redemption: enter a code, validate, and grant bonus
quota without support editing entitlements.

## Scope

### In

- Tenant portal referral-code input, validation, apply
- Server validation: existence, campaign/status, expiry, eligibility, one-use /
  per-period limits, self-referral, duplicate, abuse rules
- Bonus grant via existing bonus ledger (not base package mutation)
- Dimensions aligned to S45/S46 (call minutes, mobile minutes, KM, storage,
  active avatars where configured)
- Tenant usage shows base + referral bonus + expiry
- Platform admin inspect / revoke / reverse
- Idempotent apply (retries never double-grant)

### Out

- Cash payouts / affiliate settlement
- Cross-tenant quota pooling beyond the grant
- Manual SQL entitlement edits
- Unlimited concurrent-call increases without capacity policy
- Applying codes to invoices/tax/historical usage

## Acceptance criteria

1. Tenant can apply a valid code and see bonus quota in usage views.
2. Duplicate retry is idempotent (no second grant).
3. Expired, self, and ineligible codes fail with safe errors.
4. Grants are tenant-isolated and audited.
5. Platform admin can reverse/revoke a grant and remaining bonus updates.

## Dependencies

- design: [DES-0058](../02-design/58-referral-code-redemption-spec.md)
- prior: SPRINT-046 referral + bonus ledger
