---
id: DES-0014
title: Tenant Buy Package Specification
status: approved
updated: 2026-07-09
sprint: SPRINT-009
owner: SA
---

# Tenant Buy Package — Design Spec

**Sprint:** SPRINT-009 · **Release target:** v1.0.0  
**Feature:** [FEAT-0009](../01-features/FEAT-0009-buy-package.md)  
**Depends on:** [08-packages-spec.md](08-packages-spec.md), [13-payment-gateway-spec.md](13-payment-gateway-spec.md), [11-tenant-register-spec.md](11-tenant-register-spec.md)

## 1. Goals

- **Active tenants** (`tenant_admin`, `tenants.status = active`) browse purchasable packages and start checkout.
- **ChillPay** `InitPayment` returns `PaymentUrl`; tenant pays on hosted page (SPRINT-008 gateway config).
- **Callback fulfillment** on `PaymentStatus=0`: order `paid`, prior entitlement revoked, new package assigned, Redis cache invalidated.
- **Mock path** for CI and local dev without ChillPay credentials.
- **Combined UAT** with SPRINT-008 gateway (§0 in manual test doc).

## 2. Non-goals (Sprint 9)

- Invoices, receipts, tax invoices (Sprints 10–12).
- Subscription renewals, proration, coupons.
- Platform operator purchase on behalf of tenant.
- Auto-grant Starter on signup/KYC without payment.
- `payment_intents` separate table (single `payment_orders` suffices).
- NATS `payment.completed` events (stretch).

## 3. Environment

Uses SPRINT-008 vars. Additional:

| Variable | Purpose |
| --- | --- |
| `CHILLPAY_RETURN_URL` | Must point to `/tenant/billing/return` (browser return after ChillPay) |
| `PAYMENT_MOCK_AUTO_FULFILL` | `true` — mock checkout completes without extra click (CI) |

## 4. Data model (Postgres `callcenter`)

### `payment_orders` (Sprint 9)

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `ord_` + uuid |
| `tenant_id` | text FK → `tenants` | |
| `package_id` | text FK → `packages` | Snapshot target package |
| `order_no` | text UK | ChillPay `OrderNo`; max **20** alphanumeric only (`MJ` + 2-char tenant fingerprint + 16 hex). No `_`/`-` (ChillPay code 1006). |
| `amount_cents` | int | From `packages.price_cents` at checkout time |
| `currency` | text | Gateway `currency` (e.g. `764`) |
| `status` | text | `pending` \| `paid` \| `failed` \| `cancelled` |
| `provider` | text | `chillpay` \| `mock` |
| `transaction_id` | text | ChillPay `TransactionId` after init/callback |
| `payment_url` | text | Redirect URL from `InitPayment` or mock page |
| `paid_at` | timestamptz | Set on success callback |
| audit cols | | standard |

**Indexes:** unique `order_no`; `(tenant_id, status)`; `(tenant_id, created_at DESC)`.

### Relationships

```text
tenants 1──* payment_orders *──1 packages
payment_orders.fulfillment → tenant_entitlements (logical; no FK on entitlement)
```

## 5. Order lifecycle

| Status | Meaning | Next |
| --- | --- | --- |
| `pending` | Checkout created; awaiting ChillPay callback | `paid`, `failed`, or timeout cancel (manual) |
| `paid` | Callback `PaymentStatus=0`; entitlement assigned | terminal |
| `failed` | Callback `PaymentStatus=2` | terminal |
| `cancelled` | Operator/abandon (future) | terminal |

## 6. API summary

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `GET` | `/api/tenant/packages` | `tenant_admin` + `active` tenant | Purchasable catalog |
| `POST` | `/api/tenant/checkout` | same | `{package_id}` → order + `payment_url` |
| `GET` | `/api/tenant/orders/{id}` | same | Poll order status (own tenant only) |
| `POST` | `/api/callbacks/chillpay` | public | **Sprint 9:** fulfill on success |
| `POST` | `/api/dev/mock-pay/{order_id}` | dev only | Mock fulfillment trigger |

Full contract: [04-api-spec.md](04-api-spec.md) § Tenant Checkout.

## 7. Checkout flow (ChillPay)

1. Validate tenant `active`, package `active`, gateway configured.
2. Insert `payment_orders` `pending` with unique `order_no`.
3. Build `ChillPayRequestInfo`:
   - `OrderNo` = `order_no`
   - `CustomerID` = `tenant_id`
   - `Amount` = `amount_cents / 100` (THB baht; client sends satang internally)
   - `Description` = package name
   - `CustEmail` = tenant admin email (from registration)
4. `ChillPayClient.InitPayment` → store `transaction_id`, `payment_url` on order.
5. Return `{order_id, order_no, payment_url, status: pending}`.

**Amount:** ChillPay expects satang in form field `Amount`; `price_cents` maps 1:1 to satang for THB.

## 8. Callback fulfillment (Sprint 9 upgrade)

After SPRINT-008 verify + event log:

1. Lookup `payment_orders` by `order_no` = form `OrderNo`.
2. `PaymentStatus`:
   - `0` → transaction: update order `paid`, revoke active entitlement, assign package entitlement, invalidate Redis `monti_jarvis:entitlement:{tenant_id}`.
   - `2` → order `failed`.
   - `1` → no order status change (still `pending`).
3. Idempotent: if order already `paid`, skip entitlement mutation; return `200`.

## 9. Mock provider

When `payment_gateway_configs.provider = mock`:

- `InitPayment` skipped; `payment_url` = `{APP_PUBLIC_URL}/tenant/billing/mock-pay?order_id={id}`.
- Mock page calls `POST /api/dev/mock-pay/{order_id}` (Bearer required, same tenant) OR auto-fulfills when `PAYMENT_MOCK_AUTO_FULFILL=true`.
- Fulfillment reuses callback logic (synthetic `PaymentStatus=0`).

## 10. RBAC

| Route | `tenant_admin` | `active` tenant | `pending_kyc` |
| --- | --- | --- | --- |
| `/api/tenant/packages` | ✅ | required | `403` |
| `/api/tenant/checkout` | ✅ | required | `403` |
| `/api/tenant/orders/{id}` | ✅ own tenant | required | `403` |

## 11. Verification (combined SPRINT-008 + 009)

```bash
make build && make test && make restart

# §0 — Gateway (Sprint 8)
PLATFORM=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)
curl -s -H "Authorization: Bearer $PLATFORM" http://localhost:8091/api/platform/payment-gateway | jq .
curl -s http://localhost:8091/api/infra | jq .payment_gateway

# §1 — Checkout (Sprint 9) — mock
curl -s -X PUT -H "Authorization: Bearer $PLATFORM" -H 'Content-Type: application/json' \
  -d '{"provider":"mock","mode":"test","merchant_code":"MOCK","base_url":"http://localhost","route_no":1,"currency":"764","return_url":"http://localhost:8091/tenant/billing/return"}' \
  http://localhost:8091/api/platform/payment-gateway

TENANT=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@demo.local","password":"demo-admin"}' | jq -r .access_token)

curl -s -H "Authorization: Bearer $TENANT" http://localhost:8091/api/tenant/packages | jq .
curl -s -X POST -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"package_id":"pkg-pro"}' http://localhost:8091/api/tenant/checkout | jq .

# After mock pay / callback:
curl -s -H "Authorization: Bearer $TENANT" http://localhost:8091/api/entitlements/me | jq .
```

## 12. UI

- **T4** billing catalog: [05-ux-ui.md](05-ux-ui.md) § Screen T4
- **T5** return page: § Screen T5
- **T6** mock pay: § Screen T6

## Links

- Workflow: [02-workflow.md](02-workflow.md) §28–31
- ER: [03-er-diagram.md](03-er-diagram.md)
- Sprint: [SPRINT-009](../03-sprints/SPRINT-009.md)
- Tasks: TASK-0040–TASK-0044