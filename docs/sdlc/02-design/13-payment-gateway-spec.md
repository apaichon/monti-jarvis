---
id: DES-0013
title: Platform Payment Gateway (ChillPay) Specification
status: approved
updated: 2026-07-09
sprint: SPRINT-008
owner: SA
---

# Platform Payment Gateway — ChillPay Design Spec

**Sprint:** SPRINT-008 · **Release target:** v0.9.0  
**Feature:** [FEAT-0008](../01-features/FEAT-0008-payment-gateway.md)  
**Depends on:** [06-auth-spec.md](06-auth-spec.md) (platform_admin RBAC), [08-packages-spec.md](08-packages-spec.md) (catalog for Sprint 9 checkout)

## 1. Goals

- Platform operators **configure** a single platform-wide ChillPay gateway (or **mock** for local dev).
- **Test connection** validates merchant credentials before Sprint 9 checkout.
- **Receive payment callbacks** from ChillPay with MD5 `CheckSum` verification; persist idempotent audit log.
- Platform admin UI at `/admin/settings/payment` with standardized feedback dialogs.

## 2. Non-goals (Sprint 8)

- Tenant checkout / `InitPayment` redirect to `PaymentUrl` (Sprint 9).
- Entitlement assignment on successful payment (Sprint 9).
- Billing, invoices, receipts, tax invoices (Sprints 10–12).
- Stripe / Omise / 2C2P adapters (interface extensibility only).
- Refunds, disputes, payout reconciliation.
- NATS payment events (optional stretch — log only this sprint).

## 3. Environment

| Variable | Purpose |
| --- | --- |
| `CHILLPAY_MERCHANT_CODE` | Override DB `merchant_code` |
| `CHILLPAY_API_KEY` | Override DB `api_key` (masked in API responses) |
| `CHILLPAY_MD5_KEY` | Override DB `md5_key` (never returned in API) |
| `CHILLPAY_BASE_URL` | ChillPay payment init endpoint (form POST) |
| `CHILLPAY_ROUTE_NO` | Route number (integer, default `1`) |
| `CHILLPAY_CURRENCY` | ISO currency code (default `THB`) |
| `CHILLPAY_RETURN_URL` | Browser return after payment (Sprint 9 checkout) |
| `APP_PUBLIC_URL` | Derive `callback_url` = `{APP_PUBLIC_URL}/api/callbacks/chillpay` |
| `PAYMENT_CALLBACK_DEV_BYPASS` | `true` — skip checksum verify on callback (local only) |

Env overrides take precedence over DB-stored secrets when set (same pattern as Resend).

## 4. Data model (Postgres `callcenter`)

### `payment_gateway_configs` (singleton)

One active row (`id = 'default'`). Platform-wide — not per-tenant.

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | Always `default` |
| `provider` | text | `chillpay` \| `mock` |
| `mode` | text | `test` \| `live` |
| `status` | text | `inactive` \| `active` |
| `merchant_code` | text | ChillPay merchant code |
| `api_key` | text | Stored encrypted-at-rest optional; masked in GET |
| `md5_key` | text | MD5 secret; never returned; `md5_key_set` bool in API |
| `base_url` | text | Payment init URL |
| `route_no` | int | Default `1` |
| `currency` | text | Default `THB` |
| `callback_url` | text | Server-derived; read-only in UI |
| `return_url` | text | Browser redirect URL |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

### `payment_callback_events`

Idempotent callback audit log. Sprint 9 will consume `payment_status = 0` (success).

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `pce_` + uuid |
| `provider` | text | `chillpay` |
| `transaction_id` | text | ChillPay `TransactionId` |
| `order_no` | text | ChillPay `OrderNo` |
| `payment_status` | text | `0` success · `1` pending · `2` failed |
| `amount` | text | Satang string from callback |
| `customer_id` | text | ChillPay `CustomerId` |
| `payload_hash` | text | SHA-256 of raw form body |
| `received_at` | timestamptz | Server receive time |
| audit cols | | standard |

**Unique:** `(provider, transaction_id)`

## 5. ChillPay client (`internal/payment/chillpay`)

Ported from harvest-core integration pattern.

### 5.1 InitPayment (Sprint 9 — documented here, not called in Sprint 8)

`POST` `application/x-www-form-urlencoded` to `base_url`.

**CheckSum (param 21):** MD5 hex of concatenation (no separators):

```text
MerchantCode + OrderNo + CustomerId + Amount + PhoneNumber + Description +
ChannelCode + Currency + LangCode + RouteNo + IPAddress + ApiKey + TokenFlag +
CreditToken + CreditMonth + ShopID + ProductImageUrl + CustEmail + CardType +
CustName + MD5SecretKey
```

`CallbackUrl` and `ReturnUrl` are sent but **not** included in checksum.

**Response:** JSON `{Status, Code, Message, TransactionId, PaymentUrl, ...}` — `Status = 0` success.

### 5.2 InquiryPaymentStatus (test connection)

`POST` to `{base_url trimmed}/api/v2/PaymentStatus/`

**Form fields:** `MerchantCode`, `TransactionId`, `ApiKey`, `CheckSum`

**CheckSum:** MD5(`MerchantCode` + `TransactionId` + `ApiKey` + `MD5SecretKey`)

Sprint 8 **test** endpoint: when no known `TransactionId`, validate config by checksum self-test + HTTP reachability; optional probe with `TransactionId=0` expecting ChillPay error (proves credentials accepted).

### 5.3 VerifyCallback

**Callback form fields** (`ChillPayCallbackForm`):

| Field | Notes |
| --- | --- |
| `OrderNo` | Merchant order reference |
| `Amount` | Satang string |
| `TransactionId` | Provider transaction id |
| `CustomerId` | End-user id |
| `CustomerName` | |
| `BankCode` | |
| `PaymentDate` | |
| `PaymentStatus` | `0` success · `1` pending · `2` failed |
| `PaymentDescription` | |
| `BankRefCode` | |
| `Currency` | |
| `CreditCardToken` | |
| `CurrentDate` | |
| `CurrentTime` | |
| `CheckSum` | MD5 verify field |

**Verify CheckSum:** MD5 hex of:

```text
TransactionId + Amount + OrderNo + CustomerId + BankCode + PaymentDate +
PaymentStatus + BankRefCode + CurrentDate + CurrentTime + PaymentDescription +
CreditCardToken + Currency + CustomerName + MD5SecretKey
```

Compare case-insensitive to `form.CheckSum`.

## 6. API summary

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `GET` | `/api/platform/payment-gateway` | `platform_admin` | Current config + connection status |
| `PUT` | `/api/platform/payment-gateway` | `platform_admin` | Upsert gateway config |
| `POST` | `/api/platform/payment-gateway/test` | `platform_admin` | Validate credentials |
| `POST` | `/api/callbacks/chillpay` | public | ChillPay payment callback (form POST) |

Full contract: [04-api-spec.md](04-api-spec.md) § Payment Gateway.

## 7. RBAC

| Route | `platform_admin` | `tenant_admin` | Public |
| --- | --- | --- | --- |
| `/api/platform/payment-gateway*` | ✅ | `403` | `401` |
| `/api/callbacks/chillpay` | N/A | N/A | ✅ (checksum gate) |

## 8. Callback handler (Sprint 8)

1. Parse `application/x-www-form-urlencoded` into `ChillPayCallbackForm`.
2. Load active gateway config; `503` if not configured.
3. `VerifyCallback` unless `PAYMENT_CALLBACK_DEV_BYPASS=true`.
4. `INSERT` into `payment_callback_events` ON CONFLICT `(provider, transaction_id) DO NOTHING`.
5. Return `200` empty body (ChillPay expects ack).
6. **No** entitlement or order fulfillment — Sprint 9.

## 9. Mock provider

When `provider = mock`:

- `Ping` always succeeds.
- Callback bypass allowed without checksum.
- UI shows “Mock — no ChillPay credentials required”.

## 10. Verification

```bash
make build && make test
make infra-init && make restart

PLATFORM=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)

# Configure mock
curl -s -X PUT -H "Authorization: Bearer $PLATFORM" -H "Content-Type: application/json" \
  -d '{"provider":"mock","mode":"test","merchant_code":"MOCK","base_url":"http://localhost","route_no":1,"currency":"THB","return_url":"http://localhost:8091/tenant/billing/return"}' \
  http://localhost:8091/api/platform/payment-gateway | jq .

curl -s -X POST -H "Authorization: Bearer $PLATFORM" \
  http://localhost:8091/api/platform/payment-gateway/test | jq .

# Simulate callback (dev bypass)
PAYMENT_CALLBACK_DEV_BYPASS=true curl -s -X POST \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "TransactionId=12345&Amount=10000&OrderNo=ord-test-1&PaymentStatus=0&OrderNo=ord-test-1&CustomerId=cust-1" \
  http://localhost:8091/api/callbacks/chillpay

curl -s http://localhost:8091/api/infra | jq .payment_gateway
```

## 11. UI

- **P13** settings screen: [05-ux-ui.md](05-ux-ui.md) § Screen P13
- **Flows P-E, P-F** — configure → test; callback ops
- Feedback via `FeedbackDialog`

## Links

- Workflow: [02-workflow.md](02-workflow.md) §25–27
- ER: [03-er-diagram.md](03-er-diagram.md)
- Sprint: [SPRINT-008](../03-sprints/SPRINT-008.md)
- Tasks: TASK-0035–TASK-0039