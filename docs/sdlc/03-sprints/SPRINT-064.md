---
id: SPRINT-064
status: completed
start: 2026-08-01
end: 2026-08-01
updated: 2026-08-01
design_pack: approved
roadmap_sprint: 64
feature: FEAT-0056
platform: Platform Admin / Tenant / Finance
depends_on: [SPRINT-008, SPRINT-009, SPRINT-010, SPRINT-012, SPRINT-013, SPRINT-045, SPRINT-048, SPRINT-050, SPRINT-051]
goal: "Switch tenant package checkout between ChillPay and Stripe with safe provider config, signed webhooks, idempotent fulfillment, and reconciliation."
velocity_basis: "Last 3 closed: S61=12, S62=12, S63=3 -> avg 9; S63 was a narrow patch, so commit a bounded 12-point feature slice with refunds/subscriptions out of scope."
release_target: v2.36.0
release: v2.36.0
---

# SPRINT-064 - Payment Gateway Portability: ChillPay to Stripe

## Goal

Platform operators can switch the active tenant checkout provider from ChillPay
to Stripe while keeping existing ChillPay orders, entitlement fulfillment,
receipts, tax invoices, and billing views intact.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S61-S63) | 12, 12, 3 -> **avg 9** |
| Normal feature baseline (S60-S62) | **12** |
| **Committed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0230](../04-tasks/TASK-0230.md) Payment provider config model and Stripe settings | 3 | completed | dev | Config, masking, provider switch |
| [TASK-0231](../04-tasks/TASK-0231.md) Stripe checkout adapter and tenant checkout routing | 4 | completed | dev | Stripe Checkout Session path |
| [TASK-0232](../04-tasks/TASK-0232.md) Stripe webhook fulfillment and reconciliation | 3 | completed | dev | Signed webhooks + idempotent fulfill |
| [TASK-0233](../04-tasks/TASK-0233.md) Payment gateway admin UX and regression verification | 2 | completed | tester/dev | Admin UX, release gates, manual checklist |
| **Total** | **12** | **12/12** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0056](../01-features/FEAT-0056-payment-gateway-portability-stripe.md) | completed |
| Deep spec | **`shipped`** | [DES-0059](../02-design/59-payment-gateway-portability-stripe-spec.md) |
| Workflow | **`approved`** | [02-workflow.md](../02-design/02-workflow.md) §141-143 |
| ER | **`approved`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 64 |
| API | **`approved`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 64 |
| UX | **`approved`** | [05-ux-ui.md](../02-design/05-ux-ui.md) A64 / T64 |

## Scope boundary

### In

- Platform payment provider switch among `mock`, `chillpay`, and `stripe`
- Stripe Checkout Sessions for Shared Cloud package checkout
- Signed Stripe webhooks and provider event idempotency
- Existing fulfillment reuse: order paid/failed, entitlement, receipt, tax invoice
- Platform Admin payment settings, test connection, webhook status, reconciliation
- ChillPay regression safety

### Out

- Stripe Billing subscriptions
- Refunds, disputes, payouts, or chargebacks
- Pricing/tax/invoice numbering changes
- Tenant-owned custom gateways
- Multiple providers on one checkout attempt

## Risks

| Risk | Mitigation |
| --- | --- |
| Double fulfillment from webhook retries | Provider event uniqueness + existing paid-order idempotency |
| ChillPay regression | Keep existing adapter paths and run ChillPay tests/UAT |
| Secret leakage | Mask keys, never return plaintext, scrub logs/audit events |
| Stripe amount drift | Server amount and package snapshot remain authoritative |
| Provider switch confusion | Existing orders keep original provider and references |

## Verification

```bash
go test ./internal/payment ./internal/store ./cmd/server -count=1
cd apps/platform-admin-web && npm run check && npm run build
cd ../tenant-web && npm run check && npm run build
```

Manual: [SPRINT-064-payment-gateway-portability-stripe-manual.md](../06-manual-tests/SPRINT-064-payment-gateway-portability-stripe-manual.md)

## Review - PASS

The provider switch is now resolved through a provider-neutral gateway with
Stripe-specific checkout, signed webhook parsing, idempotent callback event
records, provider references on orders, and admin reconciliation. Tenant
checkout keeps existing mock and ChillPay behavior while routing new orders to
Stripe when the active provider or `PAYMENT_GATEWAY_PROVIDER` override is
`stripe`.

## Release

- Version: v2.36.0
- Closed: 2026-08-01
- Type: minor feature release
- Tag: `v2.36.0`
- Automated verification: Go payment/server tests, Platform Admin check/build,
  Tenant Web check/build, `git diff --check`.
- Manual verification: local Stripe checkout and webhook configuration path
  documented; rerun signed Stripe CLI replay after restarting with the current
  `STRIPE_WEBHOOK_SECRET`.
