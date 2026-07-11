---
id: MANUAL-SPRINT-009
sprint: SPRINT-009
release_target: v1.0.0
updated: 2026-07-09
---

# SPRINT-009 Manual UAT — Buy Package + Gateway verify (v1.0.0)

Combined verification for SPRINT-008 (payment gateway) and SPRINT-009 (tenant checkout).

## Prerequisites

- [ ] `make infra-init` · Postgres, Redis up
- [ ] `infra/.env.dev` has `AUTH_DISABLED=false`, `JWT_SECRET` (≥32 bytes)
- [ ] `make build && make restart`

## §0 — Payment gateway (SPRINT-008)

```bash
PLATFORM=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)
```

- [ ] `GET /api/platform/payment-gateway` with Bearer → 200, `configured` reflects provider
- [ ] `PUT /api/platform/payment-gateway` mock provider → 200
- [ ] `POST /api/platform/payment-gateway/test` → `ok: true`
- [ ] `GET /api/infra` → `payment_gateway.provider` set
- [ ] Platform UI `/admin/settings/payment` saves mock config

## §1 — Tenant checkout (SPRINT-009, mock)

Configure mock gateway:

```bash
curl -s -X PUT -H "Authorization: Bearer $PLATFORM" -H 'Content-Type: application/json' \
  -d '{"provider":"mock","mode":"test","merchant_code":"MOCK","base_url":"http://localhost","route_no":1,"currency":"764","return_url":"http://localhost:8091/tenant/billing/return"}' \
  http://localhost:8091/api/platform/payment-gateway | jq .
```

Login active tenant (`demo` seed):

```bash
TENANT=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@demo.local","password":"demo-admin"}' | jq -r .access_token)
```

- [ ] `GET /api/tenant/packages` → active catalog + `current_entitlement`
- [ ] `POST /api/tenant/checkout` `{package_id: pkg-pro}` → `payment_url` with `mock-pay`
- [ ] Tenant UI `/tenant/billing` → Buy Pro → mock pay page → Complete → return shows paid
- [ ] `POST /api/dev/mock-pay/{order_id}` → order `paid`
- [ ] `GET /api/entitlements/me` → Pro package
- [ ] `pending_kyc` tenant → `GET /api/tenant/packages` → 403

## §2 — ChillPay sandbox (optional)

Requires `CHILLPAY_*` in `infra/.env.dev` and ngrok callback URL.

- [ ] Platform gateway shows ChillPay `merchant_code` from env seed
- [ ] `POST /api/tenant/checkout` returns real `PaymentUrl`
- [ ] Pay in ChillPay sandbox → callback `POST /api/callbacks/chillpay` → order `paid` + entitlement assigned
- [ ] `payment_callback_events` row logged; replay callback is idempotent (no duplicate entitlement)

## Regression

```bash
cd e2e && npx playwright test tenant-checkout.spec.ts platform-payment-gateway.spec.ts
```

- [ ] E2E tenant checkout (mock) green
- [ ] E2E platform payment gateway green
- [ ] `GET /healthz` → `sprint: SPRINT-009`