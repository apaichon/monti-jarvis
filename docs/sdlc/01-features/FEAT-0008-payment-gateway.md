# Feature: Platform Payment Gateway   (FEAT-0008)
**Sprint:** SPRINT-008   **Owner:** DEV   **Status:** active   **Release target:** v0.9.0

## Problem

Sprints 4–7 delivered packages, tenant onboarding, and KYC approval, but there is **no payment provider** configured. Platform operators cannot connect **ChillPay** (or a dev mock) for tenant commerce. Sprint 9 (Buy Package) needs a **validated gateway config** and callback receiver before checkout can ship.

## Scope

In:
- Postgres `payment_gateway_configs` — platform-level ChillPay settings (`merchant_code`, `api_key`, `md5_key`, `base_url`, `route_no`, `currency`, `callback_url`, `return_url`, `mode`, `status`)
- Postgres `payment_callback_events` — idempotent callback audit log (`transaction_id`, `order_no`, `payment_status`, payload hash)
- `internal/payment/chillpay` — `ChillPayClient` (ported from harvest-core pattern):
  - `InitPayment` — form POST + MD5 `CheckSum` (params 1–20 + `md5_key`)
  - `InquiryPaymentStatus` — `POST …/api/v2/PaymentStatus/` with MD5 checksum
  - `VerifyCallback` — MD5 over callback form fields + `md5_key`
- `internal/payment` — `Provider` interface; **ChillPay** adapter; **mock** provider for local dev without keys
- Platform APIs (`platform_admin` only):
  - `GET /api/platform/payment-gateway` — current config + connection status (secrets masked)
  - `PUT /api/platform/payment-gateway` — set provider + ChillPay fields; env overrides (`CHILLPAY_API_KEY`, `CHILLPAY_MD5_KEY`)
  - `POST /api/platform/payment-gateway/test` — validate credentials (`InquiryPaymentStatus` or checksum self-test; mock ping)
- Callback: `POST /api/callbacks/chillpay` — parse `application/x-www-form-urlencoded` body, `VerifyCallback`, persist event, `200` ack (no checkout side-effects yet)
- Platform admin UI `/admin/settings/payment` — configure ChillPay, test connection, view last callback status
- `GET /api/infra` adds `payment_gateway` block (configured, provider, mode)
- Design pack via `sprint-tech-specs` (`13-payment-gateway-spec.md`)

Out:
- Tenant checkout / package purchase UI (Sprint 9) — `InitPayment` + redirect to `PaymentUrl` ships in Sprint 9
- Invoices, receipts, tax invoices (Sprints 10–12)
- Subscription billing cycles and usage metering (Sprint 10+)
- Multiple concurrent providers per tenant (single platform default only)
- Card data collection in Monti UI (ChillPay hosted payment page in Sprint 9)
- Stripe / Omise / 2C2P adapters (future; interface ready)
- Refunds, disputes, payout reporting

## Acceptance criteria

1. `ensureSchema` creates `payment_gateway_configs` (singleton row) and `payment_callback_events` idempotently with audit columns.
2. `GET /api/platform/payment-gateway` returns provider, mode (`test`|`live`), masked `api_key` tail, `md5_key_set`, `merchant_code`, `base_url`, `route_no`, `currency`, `callback_url`, `return_url`, `connection_status`; `403` for non-platform-admin.
3. `PUT /api/platform/payment-gateway` accepts `provider` (`chillpay`|`mock`), ChillPay fields, mode; env `CHILLPAY_API_KEY` / `CHILLPAY_MD5_KEY` override DB secrets when set; returns updated config.
4. `POST /api/platform/payment-gateway/test` returns `200 {ok:true}` when ChillPay config valid (mock always OK); `502` with message when keys/base URL invalid.
5. `POST /api/callbacks/chillpay` verifies MD5 `CheckSum` via `VerifyCallback`, inserts event once (idempotent on `transaction_id`), returns `200`; rejects bad checksum with `400`.
6. `/admin/settings/payment` — form for merchant code, API key, MD5 key, base URL, route no, currency, return URL; **Test connection** button; feedback dialog; nav link in platform shell.
7. `/api/infra` includes `payment_gateway: {configured, provider, mode}`.
8. `go test ./...`; manual UAT in `docs/sdlc/06-manual-tests/SPRINT-008-manual.md` (Tester, at VERIFY).

## ChillPay reference (integration contract)

| Operation | Endpoint | Checksum |
| --- | --- | --- |
| Init payment | `base_url` (form POST) | MD5(`MerchantCode`+`OrderNo`+…+`CustName`+`MD5Key`) — 20 fields |
| Payment status | `{base}/api/v2/PaymentStatus/` | MD5(`MerchantCode`+`TransactionId`+`ApiKey`+`MD5Key`) |
| Callback verify | `POST /api/callbacks/chillpay` | MD5(`TransactionId`+`Amount`+…+`CustomerName`+`MD5Key`) |

Callback `PaymentStatus`: `"0"` success · `"1"` pending · `"2"` failed.

## Test notes

- API: RBAC, secret masking, idempotent callback insert, `VerifyCallback` unit tests with fixture forms
- Dev without ChillPay: `mock` provider passes test endpoint; `PAYMENT_CALLBACK_DEV_BYPASS=true` for local callback curl
- Browser UAT: platform admin configures mock → test OK → simulate callback form POST
- E2E: `e2e/tests/platform-payment-gateway.spec.ts` (config page smoke)

## Links

- Sprint: [SPRINT-008](../03-sprints/SPRINT-008.md)
- Depends on: [FEAT-0003](FEAT-0003-auth-rbac.md) (platform_admin RBAC)
- Prior commerce: [FEAT-0004](FEAT-0004-packages-entitlements.md)
- Design: [13-payment-gateway-spec.md](../02-design/13-payment-gateway-spec.md) · [04-api-spec.md](../02-design/04-api-spec.md) § Payment Gateway · [05-ux-ui.md](../02-design/05-ux-ui.md) § P13 · [02-workflow.md](../02-design/02-workflow.md) §25–27
- Roadmap: Sprint 8 · Phase C
- Next: [FEAT-0009 Buy Package](FEAT-0009-buy-package.md) — combined E2E verify