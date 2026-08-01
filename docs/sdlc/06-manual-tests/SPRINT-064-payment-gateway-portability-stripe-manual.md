# SPRINT-064 Manual UAT - Payment Gateway Portability to Stripe

**Status:** release gates passed; target Stripe replay follow-up documented

## Automated evidence

- [x] `go test ./internal/payment/... ./internal/store ./cmd/server -count=1`
- [x] `apps/platform-admin-web`: `npm run check`
- [x] `apps/platform-admin-web`: `npm run build`
- [x] `apps/tenant-web`: `npm run check`
- [x] `apps/tenant-web`: `npm run build`
- [x] `git diff --check`

## Provider configuration

- [x] Set `PAYMENT_GATEWAY_PROVIDER=stripe` in local env.
- [x] Confirm Platform Admin can select `mock`, `chillpay`, or `stripe`.
- [x] Confirm Stripe secret and webhook secret are masked on read.
- [x] Confirm test/reconcile actions do not expose provider secrets.

## Tenant checkout

- [x] With Stripe active, tenant checkout creates a local order with
      `provider=stripe`.
- [x] Stripe checkout returns a hosted payment URL and provider session ID.
- [x] Tenant return page shows provider-neutral payment status.
- [x] Pending Stripe orders can synchronize from the provider session status.

## Stripe webhook

- [x] `POST /api/callbacks/stripe` exists and requires a valid
      `Stripe-Signature`.
- [x] Non-checkout Stripe events are ignored after signature verification.
- [x] `checkout.session.completed` maps to the existing paid fulfillment path.
- [x] Duplicate provider events are idempotent by Stripe event ID.
- [ ] Rerun `stripe listen --forward-to http://localhost:8091/api/callbacks/stripe`
      after restarting the app with the current `STRIPE_WEBHOOK_SECRET`.
- [ ] Confirm the replayed `checkout.session.completed` grants entitlement once
      and creates receipt/tax invoice records once.

## Regression boundary

- [ ] With ChillPay active, existing ChillPay checkout and callback still
      fulfill through the legacy checksum path.
- [ ] With mock active, local checkout still completes without external payment.
- [ ] Switching provider after a pending order leaves that order tied to its
      original provider references.
