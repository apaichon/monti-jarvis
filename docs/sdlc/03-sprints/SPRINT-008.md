---
id: SPRINT-008
status: in_progress
start: 2026-07-09
end: 2026-07-23
updated: 2026-07-09
release_target: v0.9.0
goal: "Platform Admin: Payment Gateway — configure ChillPay (or mock), test connection, and receive MD5-verified callbacks for Sprint 9 checkout."
roadmap_sprint: 8
platform: Platform Admin
depends_on: [SPRINT-003]
---

# SPRINT-008 — Platform Admin: Payment Gateway

## Goal

Let **platform admins** connect **ChillPay** (Thailand payment gateway) or local **mock**, validate credentials, and accept **MD5-verified payment callbacks** — without tenant checkout yet. Unblocks Sprint 9 package purchase (`InitPayment` → `PaymentUrl` redirect).

## Context from prior sprints

- **Sprint 4:** `packages`, `tenant_entitlements`, platform admin portal at `/admin`
- **Sprint 6–7:** Tenant register + KYC; approved tenants are `active`
- **No payment tables or provider code** exist today
- **ChillPay pattern** (harvest-core reference): form-urlencoded POST, MD5 `CheckSum`, callback form fields, `PaymentStatus` 0/1/2
- Pattern: optional external deps (Resend) — no-op / mock when unset; infra block on `/api/infra`

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0035 | 3 | completed | devops | Payment gateway schema — `payment_gateway_configs` + `payment_callback_events` |
| TASK-0036 | 5 | completed | dev | Provider abstraction — ChillPay + mock adapters; platform config GET/PUT/test APIs |
| TASK-0037 | 3 | completed | dev | ChillPay callback receiver — MD5 verify + idempotent event log |
| TASK-0038 | 3 | completed | dev | Platform admin UI — `/admin/settings/payment` configure + test connection |
| TASK-0039 | 2 | completed | dev | Infra status, env docs, E2E smoke |

**Committed:** 16 points · **Target velocity:** 16 (avg from Sprints 1–7)

## Scope boundary

**In**
- Single platform-wide gateway config (not per-tenant)
- Providers: `chillpay` (test/live mode) and `mock` (local dev)
- ChillPay config: `merchant_code`, `api_key`, `md5_key`, `base_url`, `route_no`, `currency` (default `THB`), `callback_url`, `return_url`
- Secret handling: prefer `CHILLPAY_API_KEY` / `CHILLPAY_MD5_KEY` env; mask keys in API responses
- `ChillPayClient`: `InitPayment`, `InquiryPaymentStatus`, `VerifyCallback` (checksum per ChillPay V2 manual)
- Platform APIs: get/update config, test connection
- Callback `POST /api/callbacks/chillpay` logs events only (checkout fulfillment → Sprint 9)
- Platform admin settings screen with feedback dialog pattern from Sprint 6–7
- `sprint-tech-specs` design pack **before** TASK-0036 implementation

**Out** (→ backlog / later sprints)
- Tenant buy-package checkout + `InitPayment` redirect (Sprint 9)
- Billing, invoices, receipts, tax invoices (Sprints 10–12)
- Auto-assign Starter on KYC approve (Sprint 9)
- Stripe / Omise adapters (interface extensibility only)
- NATS payment events
- PCI — card data never touches Monti server (ChillPay hosted page)

## Feature

- [FEAT-0008 — Platform Payment Gateway](../01-features/FEAT-0008-payment-gateway.md)

## Design pack (`sprint-tech-specs`)

| Artifact | Path | Status |
| --- | --- | --- |
| Payment gateway deep spec | [13-payment-gateway-spec.md](../02-design/13-payment-gateway-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §25–27 | `approved` |
| ER diagram | [03-er-diagram.md](../02-design/03-er-diagram.md) | `approved` |
| API spec | [04-api-spec.md](../02-design/04-api-spec.md) § Payment Gateway | `approved` |
| UX/UI ASCII | [05-ux-ui.md](../02-design/05-ux-ui.md) § P13 | `approved` |

> Design pack approved — DEV may start **TASK-0035** (schema) then **TASK-0036** (APIs).

## Verification

```bash
make build && make test
make infra-init && make restart
# Platform admin: configure mock provider → test connection
open http://localhost:8091/admin/settings/payment
# Simulate ChillPay callback (form POST with valid CheckSum or dev bypass)
curl -X POST http://localhost:8091/api/callbacks/chillpay \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "TransactionId=123&Amount=10000&OrderNo=ord-1&PaymentStatus=0&CheckSum=..."
```

- Manual: deferred — **combined with** `SPRINT-009-manual.md` §0 Gateway (Tester, at Sprint 9 VERIFY)
- E2E: `e2e/tests/platform-payment-gateway.spec.ts` (regression in Sprint 9)

## Risks

| Risk | Mitigation |
| --- | --- |
| Secrets in DB | Mask `api_key` / `md5_key` in API; prefer env overrides |
| Callback replay | Unique index on `(provider, transaction_id)` |
| ChillPay unavailable in CI | Default `mock` provider; ChillPay integration tests skip without keys |
| Checksum field order | Port exact concat order from harvest-core `VerifyCallback` / `InitPayment` |
| Sprint 9 coupling | Callback handler logs only; document `PaymentStatus` values Sprint 9 will consume |

## Definition of done

- Design pack approved · code reviewed · ACs verified · `make build` · tag **v0.9.0** at sprint close