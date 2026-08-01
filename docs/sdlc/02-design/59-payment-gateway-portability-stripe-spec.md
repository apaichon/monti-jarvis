---
id: DES-0059
title: Payment Gateway Portability and Stripe Specification
status: shipped
updated: 2026-08-01
sprint: SPRINT-064
owner: SA
feature: FEAT-0056
release_target: v2.36.0
release: v2.36.0
---

# DES-0059 - Payment Gateway Portability and Stripe

**Sprint:** SPRINT-064 · **Release:** v2.36.0  
**Feature:** [FEAT-0056](../01-features/FEAT-0056-payment-gateway-portability-stripe.md)  
**Tasks:** TASK-0230-TASK-0233  
**Prior:** [DES-0013](13-payment-gateway-spec.md), [DES-0014](14-buy-package-spec.md), [DES-0048](48-commercial-plans-billing-operations-spec.md)

## 1. Goals

1. Add Stripe as a real hosted-checkout provider beside existing ChillPay and
   mock providers.
2. Let Platform Admin switch the active provider safely for new checkout
   attempts.
3. Keep all order, subscription, entitlement, receipt, tax-invoice, and billing
   views provider-neutral.
4. Verify Stripe webhooks and process them idempotently.
5. Preserve ChillPay behavior and data compatibility.

## 2. Non-goals

- Stripe Billing subscriptions or provider-managed recurring plans.
- Refund/dispute/payout automation.
- Package price, tax, invoice numbering, or entitlement-rule changes.
- Tenant-owned payment-provider settings.
- More than one provider per checkout attempt.

## 3. Environment

| Variable | Purpose |
| --- | --- |
| `PAYMENT_GATEWAY_PROVIDER` | Optional active provider override: `mock`, `chillpay`, or `stripe` |
| `STRIPE_SECRET_KEY` | Optional env override for Stripe secret key |
| `STRIPE_PUBLISHABLE_KEY` | Optional env override for display/public metadata |
| `STRIPE_WEBHOOK_SECRET` | Optional env override for signed webhook verification |
| `STRIPE_API_BASE_URL` | Optional test stub override; default Stripe API |
| `STRIPE_SUCCESS_URL` | Optional override; default `/tenant/billing/return` |
| `STRIPE_CANCEL_URL` | Optional override; default `/tenant/billing` |
| existing `CHILLPAY_*` vars | Continue to override ChillPay config |
| `PAYMENT_CALLBACK_DEV_BYPASS` | Local only; never bypass Stripe signature in production |

## 4. Data Model

Migration: `scripts/migrations/037_payment_gateway_stripe.sql` or equivalent
idempotent `ensureSchema` delta.

### 4.1 `payment_gateway_configs` extension

Existing singleton row remains the active-provider authority.

| Column | Type | Notes |
| --- | --- | --- |
| `provider` | text | add `stripe`; valid values `mock`, `chillpay`, `stripe` |
| `stripe_publishable_key` | text | Stripe publishable key or env override |
| `stripe_secret_key` | text | Stripe secret key; never returned by API |
| `stripe_webhook_secret` | text | Stripe webhook signing secret; never returned by API |
| `stripe_api_base_url` | text | Stripe API base URL; test stubs can override |
| `stripe_success_url` | text | hosted checkout success redirect |
| `stripe_cancel_url` | text | hosted checkout cancel redirect |
| `last_test_status` | text | `unknown`, `ok`, `failed` |
| `last_tested_at` | timestamptz nullable | last provider test |
| `last_test_error` | text | safe operator-facing error class/message |
| `last_webhook_status` | text | `unknown`, `ok`, `failed` |
| `last_webhook_at` | timestamptz nullable | latest verified provider webhook |
| audit cols | existing | `created_at`, `updated_at`, `created_by`, `updated_by` |

Legacy ChillPay columns (`merchant_code`, `api_key`, `md5_key`, `base_url`,
`route_no`, `currency`, `callback_url`, `return_url`) remain readable and
writable for backward compatibility. Sprint 64 keeps the existing payment
gateway storage pattern and masks Stripe secrets from API responses; encrypted
payment-provider secret storage is a future hardening item.

### 4.2 `payment_orders` extension

| Column | Type | Notes |
| --- | --- | --- |
| `provider_session_id` | text | Stripe Checkout Session ID |
| `provider_payment_id` | text | Stripe PaymentIntent ID or ChillPay transaction ID |
| `provider_status` | text | raw normalized provider status |
| `checkout_expires_at` | timestamptz nullable | Stripe session expiry |
| `last_provider_sync_at` | timestamptz nullable | reconciliation timestamp |

`provider`, `order_no`, `status`, `amount_cents`, `currency`, and document
relationships remain the local authority.

### 4.3 `payment_callback_events` extension

| Column | Type | Notes |
| --- | --- | --- |
| `provider_event_id` | text | Stripe event ID; blank for legacy ChillPay |
| `event_type` | text | e.g. `checkout.session.completed` |
| `signature_verified` | boolean | true only after provider verification |
| `processing_status` | text | `received`, `processed`, `ignored`, `failed` |
| `processed_at` | timestamptz nullable | |
| `error_code` | text | safe failure code |

Unique indexes:

- Existing `(provider, transaction_id)` for compatibility.
- New partial unique `(provider, provider_event_id)` where
  `provider_event_id <> ''`.

## 5. Provider Contract

```go
type Provider interface {
  Name() string
  Ping(ctx context.Context, cfg ResolvedConfig) error
  CreateCheckout(ctx context.Context, input CheckoutInput) (CheckoutResult, error)
  VerifyWebhook(r *http.Request, cfg ResolvedConfig) (ProviderEvent, error)
  Status(ctx context.Context, providerRef string) (ProviderStatus, error)
}
```

Provider outputs normalize to:

| Local status | Stripe source | ChillPay source |
| --- | --- | --- |
| `pending` | open Checkout Session | `PaymentStatus=1` |
| `paid` | `checkout.session.completed` with paid PaymentIntent | `PaymentStatus=0` |
| `failed` | failed/expired/canceled session or payment | `PaymentStatus=2` |

## 6. Stripe Checkout Rules

1. `POST /api/tenant/checkout` validates active tenant, active Shared Cloud
   package, billing interval, and server-calculated amount.
2. Create local `payment_orders` first with `provider=stripe`.
3. Create Stripe Checkout Session with:
   - `mode=payment`
   - `client_reference_id=order_no`
   - metadata: `order_id`, `order_no`, `tenant_id`, `package_id`,
     `billing_interval`, `payment_method`
   - server-calculated amount/currency only
   - success URL to `/tenant/billing/return?order_id=...&provider=stripe`
   - cancel URL to `/tenant/billing?checkout_cancelled=1`
4. Store `session.id`, `payment_intent`, provider URL, expiry.
5. Return the existing checkout response shape with `provider=stripe`.

## 7. Webhook Fulfillment

1. `POST /api/callbacks/stripe` reads raw body and verifies
   `Stripe-Signature` against configured webhook secret.
2. Insert `payment_callback_events` by Stripe event ID. Duplicate event returns
   `200` after confirming the previous record.
3. For `checkout.session.completed`, locate order by metadata `order_id` or
   `client_reference_id`.
4. If local amount/currency/order provider mismatch, record failed event and do
   not fulfill.
5. For paid events, call the existing fulfillment path once:
   `payment_orders paid -> tenant_entitlements active -> payment_documents`.
6. For expired/failed events, mark the order failed when safe and leave no
   entitlement change.

## 8. API Summary

| Method | Path | Role | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/platform/payment-gateway` | platform_admin | active provider config, masked metadata, webhook/test status |
| `PUT` | `/api/platform/payment-gateway` | platform_admin | upsert provider-specific config and switch active provider |
| `POST` | `/api/platform/payment-gateway/test` | platform_admin | test selected provider |
| `POST` | `/api/platform/payment-gateway/reconcile` | platform_admin | compare local/provider order state |
| `POST` | `/api/tenant/checkout` | tenant_admin | existing checkout endpoint; routes by active provider |
| `POST` | `/api/callbacks/stripe` | public signed | Stripe webhook receiver |
| `POST` | `/api/callbacks/chillpay` | public checksum | unchanged ChillPay callback receiver |

Full contract: [04-api-spec.md](04-api-spec.md) Sprint 64.

## 9. RBAC and Security

- Platform gateway config routes require `platform_admin`.
- Tenant checkout requires active `tenant_admin` and own tenant context.
- Provider webhooks are public but protected by provider signatures/checksums.
- Secret values are write-only and masked on reads.
- Logs and audit events may include provider, mode, order ID, provider event ID,
  and safe status, never card data, API keys, webhook secrets, or raw provider
  payloads.

## 10. Verification

```bash
# Unit/API regression
go test ./internal/payment ./internal/store ./cmd/server -count=1

# Admin + tenant web checks
cd apps/platform-admin-web && npm run check && npm run build
cd ../tenant-web && npm run check && npm run build

# Manual sandbox
# 1. Configure Stripe test keys.
# 2. Create tenant checkout.
# 3. Send signed checkout.session.completed webhook.
# 4. Verify payment_orders paid, entitlement active, receipt/tax invoice issued.
# 5. Replay webhook; verify no duplicate grant/documents.
```

## See also

Workflow §141-143 · ER Sprint 64 · API Sprint 64 · UX A64/T64
