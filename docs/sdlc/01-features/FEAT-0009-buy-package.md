# Feature: Tenant Buy Package   (FEAT-0009)
**Sprint:** SPRINT-009   **Owner:** DEV   **Status:** active   **Release target:** v1.0.0

## Problem

Sprints 4–8 delivered packages, tenant onboarding, KYC, and ChillPay gateway config, but **tenants cannot self-purchase** a package. Operators must manually assign entitlements. Sprint 9 closes the commerce loop: **checkout → ChillPay → callback → entitlement**.

## Scope

In:
- Postgres `payment_orders` — tenant checkout orders (`order_no`, `package_id`, `amount`, ChillPay `transaction_id`, `status`)
- Tenant APIs (`tenant_admin`, `active` tenant only):
  - `GET /api/tenant/packages` — purchasable catalog (active packages + prices)
  - `POST /api/tenant/checkout` — `{package_id}` → create order, `InitPayment`, return `{order_id, payment_url}`
  - `GET /api/tenant/orders/{id}` — order status poll
- **Callback fulfillment** — extend `POST /api/callbacks/chillpay`: on `PaymentStatus=0`, mark order `paid`, assign/replace `tenant_entitlements`, invalidate Redis cache
- **Mock checkout** — when gateway `provider=mock`, skip ChillPay; simulate paid via dev endpoint or instant fulfillment for CI
- Tenant UI `/tenant/billing` — package cards, Buy → redirect to `payment_url`; `/tenant/billing/return` — post-payment status
- Combined **E2E + manual UAT** proving SPRINT-008 gateway + SPRINT-009 checkout (user request)
- Design pack via `sprint-tech-specs` (`14-buy-package-spec.md`)

Out:
- Invoices, receipts, tax invoices (Sprints 10–12)
- Subscription renewals, proration, coupons (Sprint 10+)
- Platform-admin purchase on behalf of tenant (manual assign stays in Sprint 4)
- Partial payments, installments
- Auto-assign Starter on KYC approve without payment (explicit buy required)
- Quota enforcement on live paths (Sprint 13)

## Acceptance criteria

1. `ensureSchema` creates `payment_orders` with unique `order_no`, FK to `tenants` + `packages`, audit columns.
2. `GET /api/tenant/packages` returns active packages with `price_cents`, `billing_period`; `403` for non-`tenant_admin` or non-`active` tenant.
3. `POST /api/tenant/checkout` creates `pending` order, calls ChillPay `InitPayment` (or mock), returns `payment_url`; `409` if package archived or tenant not `active`.
4. `POST /api/callbacks/chillpay` with `PaymentStatus=0` marks matching order `paid`, assigns entitlement (revokes prior active), invalidates entitlement cache; idempotent on replay.
5. `GET /api/tenant/orders/{id}` reflects `pending` → `paid`/`failed`; tenant can only read own orders.
6. `/tenant/billing` shows packages + current entitlement; Buy redirects to ChillPay; return page shows outcome.
7. **Combined verify:** ChillPay sandbox (SPRINT-008 env) + tenant checkout (SPRINT-009) → callback → `GET /api/entitlements/me` shows new package.
8. `go test ./...`; manual `docs/sdlc/06-manual-tests/SPRINT-009-manual.md` includes SPRINT-008 gateway checks.

## Test notes

- E2E mock path: platform mock gateway → tenant checkout → mock pay → entitlement updated
- E2E/manual ChillPay sandbox: real `InitPayment` + ngrok callback (from `infra/.env.dev`)
- Regression: platform admin assign entitlement still works
- Callback replay does not double-assign entitlement

## Links

- Sprint: [SPRINT-009](../03-sprints/SPRINT-009.md)
- Depends on: [FEAT-0004](FEAT-0004-packages-entitlements.md), [FEAT-0006](FEAT-0006-tenant-register.md), [FEAT-0008](FEAT-0008-payment-gateway.md)
- Prior: [13-payment-gateway-spec.md](../02-design/13-payment-gateway-spec.md)
- Design: [14-buy-package-spec.md](../02-design/14-buy-package-spec.md) · [04-api-spec.md](../02-design/04-api-spec.md) § Tenant Checkout · [05-ux-ui.md](../02-design/05-ux-ui.md) § T4–T6 · [02-workflow.md](../02-design/02-workflow.md) §28–31
- Roadmap: Sprint 9 · Phase C
- Next: Sprint 10 Billing