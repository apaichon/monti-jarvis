---
id: FEAT-0056
title: "Payment gateway portability from ChillPay to Stripe"
status: completed
release: v2.36.0
roadmap_sprint: 64
priority: P+
depends_on: [SPRINT-008, SPRINT-009, SPRINT-010, SPRINT-012, SPRINT-013, SPRINT-045, SPRINT-048, SPRINT-050, SPRINT-051]
design: DES-0059
design_spec: ../02-design/59-payment-gateway-portability-stripe-spec.md
updated: 2026-08-01
---

# FEAT-0056: Payment Gateway Portability from ChillPay to Stripe

## Purpose

Let platform operators switch tenant package checkout from ChillPay to Stripe
without breaking existing ChillPay orders, receipts, tax invoices, or entitlement
fulfillment.

## Problem

Monti's self-serve package checkout is operational through ChillPay and the local
mock provider, but the implementation still treats ChillPay as the only real
payment provider. Moving to Stripe needs a provider-neutral contract for checkout,
webhooks, payment status, reconciliation, and finance documents so the active
provider can be changed safely.

## Scope

### In

- Add `stripe` as a supported platform payment provider alongside `mock` and
  `chillpay`.
- Preserve current ChillPay config, callback, browser return, order, receipt,
  tax-invoice, subscription, and entitlement behavior.
- Create Stripe Checkout Sessions for Shared Cloud package checkout.
- Receive and verify Stripe signed webhooks, then reuse the existing idempotent
  paid/failed fulfillment path.
- Store Stripe provider references on payment orders and callback events.
- Platform Admin payment settings can configure, test, and switch the active
  provider with masked secret/status metadata.
- Add a reconciliation action to compare local order state with provider state
  and report mismatches for operator follow-up.

### Out

- Native Stripe subscription billing or Stripe-hosted recurring subscription
  management.
- Refund, dispute, payout, or chargeback automation.
- Changing package pricing, tax calculation, invoice numbering, or entitlement
  rules.
- Allowing tenants to bring their own payment gateway.
- Running more than one provider for a single checkout attempt.

## Acceptance criteria

1. Platform Admin can select `mock`, `chillpay`, or `stripe` as the active
   payment provider and see safe configuration/test status.
2. Existing ChillPay checkout and callback fulfillment continue to pass when
   `chillpay` is active.
3. When `stripe` is active, tenant checkout creates a Stripe Checkout Session and
   returns a hosted `payment_url`.
4. A signed Stripe `checkout.session.completed` webhook marks the matching order
   paid exactly once, grants the entitlement once, and issues receipt/tax
   documents through the existing path.
5. Duplicate, delayed, failed, or replayed webhooks are recorded but do not
   double-grant quota or duplicate finance documents.
6. Switching the active provider affects only new orders; existing orders keep
   their original provider references and remain queryable/reconcilable.

## Dependencies

- design: [DES-0059](../02-design/59-payment-gateway-portability-stripe-spec.md)
- prior: [DES-0013](../02-design/13-payment-gateway-spec.md),
  [DES-0014](../02-design/14-buy-package-spec.md),
  [DES-0048](../02-design/48-commercial-plans-billing-operations-spec.md)

## Release

Shipped in v2.36.0 as SPRINT-064. The release adds Stripe as an active
checkout provider option while preserving mock and ChillPay behavior.
