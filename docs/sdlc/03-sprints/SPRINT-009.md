---
id: SPRINT-009
status: completed
start: 2026-07-09
end: 2026-07-11
updated: 2026-07-11
release_target: v1.3.0
release: v1.3.0
goal: "Tenant: Buy Package — method select (card/PromptPay), ChillPay checkout, callback fulfillment, entitlement, MVP receipt+tax invoice on paid; combined E2E with SPRINT-008."
roadmap_sprint: 9
platform: Tenant
depends_on: [SPRINT-004, SPRINT-006, SPRINT-008]
---

# SPRINT-009 — Tenant: Buy Package

## Goal

Let **active tenants** browse packages, choose **Credit Card** or **QR PromptPay**, pay via **ChillPay** (or mock), land on **payment status**, receive **entitlement**, and get **MVP receipt + tax invoice** on success — closing SPRINT-008+009 commerce E2E.

## Context

| Sprint | Shipped capability |
| --- | --- |
| 4 | `packages`, `tenant_entitlements`, platform assign API |
| 6–7 | Tenant register, KYC, `active` tenant |
| 8 | ChillPay config, `InitPayment` client, callback receiver + event log |

**Combined verify:** Manual UAT and E2E exercise **gateway + checkout + status + docs** in one flow.

**Commerce chain (re-scoped 2026-07-11):** See [15-commerce-chain-plan.md](../02-design/15-commerce-chain-plan.md). S10–S12 build **platform billing ledger** and **document ops/compliance** on top of S9 MVP documents — not greenfield issue.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0040 | 3 | completed | devops | `payment_orders` (+ `payment_method`) schema + lifecycle |
| TASK-0041 | 5 | completed | dev | Catalog + checkout APIs — method → ChannelCode, InitPayment, mock |
| TASK-0042 | 5 | completed | dev | Callback fulfillment — paid/failed + entitlement + MVP `payment_documents` |
| TASK-0043 | 3 | completed | dev | Tenant UI — method modal, return status, receipt/tax links |
| TASK-0044 | 2 | completed | tester | Combined E2E path in `e2e/tests/tenant-checkout.spec.ts` |

**Committed:** 18 points · **Stretch:** +2 vs avg 16 (commerce E2E)

## Scope boundary

**In**
- `payment_orders` linked to ChillPay `OrderNo` / `TransactionId`; `payment_method` (`credit_card` \| `qr_promptpay`)
- Tenant checkout for `tenant_admin` on `active` tenants only
- ChillPay `InitPayment` with `ChannelCode` + return URL including `order_id`
- Callback: `PaymentStatus=0` → `paid` + entitlement; `2` → `failed`; idempotent
- Mock provider: success/fail simulate for local/CI
- Tenant UI: `/tenant/billing` (method select), `/return` (status), `/mock-pay`
- **MVP documents:** auto-issue `receipt` + `tax_invoice` on paid; HTML view + JSON API
- `GET /api/entitlements/me` reflects purchase after callback

**Out** (→ S10–S12 / later)
- Platform billing ledger, usage metering, subscription cycles (**S10**)
- Platform admin receipt queue, renumber, void/reissue, PDF branding (**S11**)
- Tenant tax-invoice compliance (buyer tax ID, RD fields, e-tax export) (**S12**)
- Starter auto-grant on signup/KYC · proration · NATS payment events

## Feature

- [FEAT-0009 — Tenant Buy Package](../01-features/FEAT-0009-buy-package.md)

## Design pack (`sprint-tech-specs`)

| Artifact | Path | Status |
| --- | --- | --- |
| Buy package deep spec | [14-buy-package-spec.md](../02-design/14-buy-package-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §28–31 | `approved` |
| ER diagram | [03-er-diagram.md](../02-design/03-er-diagram.md) | `approved` |
| API spec | [04-api-spec.md](../02-design/04-api-spec.md) § Tenant Checkout | `approved` |
| UX/UI ASCII | [05-ux-ui.md](../02-design/05-ux-ui.md) § T4–T6 | `approved` |

> Design pack approved — DEV may start **TASK-0040** (schema) then **TASK-0041** (checkout APIs).

## Verification (combined SPRINT-008 + SPRINT-009)

```bash
make build && make test
make infra-init && make restart

# 1) SPRINT-008 — gateway ready (ChillPay sandbox from infra/.env.dev)
open http://localhost:8091/admin/settings/payment
curl -s http://localhost:8091/api/infra | jq .payment_gateway

# 2) SPRINT-009 — tenant buys package
# Active tenant login → /tenant/billing → Buy Pro → ChillPay sandbox
# ngrok callback → entitlement updated

curl -H "Authorization: Bearer $TENANT_TOKEN" http://localhost:8091/api/entitlements/me | jq .
```

- Manual: `docs/sdlc/06-manual-tests/SPRINT-009-manual.md` — **§0 Gateway (Sprint 8)** + **§1 Checkout (Sprint 9)**
- E2E: `e2e/tests/tenant-checkout.spec.ts` (mock) + optional ChillPay sandbox tag

## Risks

| Risk | Mitigation |
| --- | --- |
| OrderNo collision | Prefix `ord_{tenant_id}_{uuid}` |
| Callback before redirect | Order stays `pending`; poll `GET /api/tenant/orders/{id}` |
| Double entitlement | Revoke prior active in same transaction as assign |
| ChillPay/ngrok flake | Mock path for CI; sandbox manual in UAT doc |
| Sprint 8 not tagged yet | Combined VERIFY gate; tag **v1.0.0** at Sprint 9 close (includes gateway) |

## Definition of done

- Design pack approved · SPRINT-008+009 E2E verified · `make build` · tag **v1.0.0** at sprint close