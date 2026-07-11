---
id: DES-0015
title: Phase C Commerce Chain Plan (Sprints 8–12)
status: approved
updated: 2026-07-11
sprint: SPRINT-009
owner: PM
---

# Phase C — Commerce chain plan (S8–S12)

**Decision (2026-07-11):** Re-scope S9–S12 so each sprint owns a clear layer of one continuous flow. MVP receipt + tax invoice ship with Buy Package; later sprints add platform ops and compliance — not a second greenfield issue path.

## 1. End-to-end flow

```text
[Platform] Gateway config (S8)
        │
        ▼
[Tenant] Browse packages → select payment method (S9)
        │  credit_card | qr_promptpay
        ▼
[ChillPay] InitPayment (ChannelCode) → hosted pay (S9)
        │
        ├─ async callback PaymentStatus=0|1|2  (S8 receive → S9 fulfill)
        └─ browser return → /tenant/billing/return (S9)
                │
                ├─ paid   → entitlement + MVP receipt + tax invoice (S9)
                └─ failed → status only; retry buy
        │
        ▼
[Platform] Billing ledger / history / filters (S10)
        │
        ▼
[Platform] Receipt ops — list, void, reissue, branding (S11)
        │
        ▼
[Tenant]  Tax invoice compliance — tax ID, address, RD export (S12)
```

## 2. What already exists in code (as of 2026-07-11)

| Layer | Status | Key surfaces |
| --- | --- | --- |
| **S8 Gateway** | Code shipped | `payment_gateway_configs`, ChillPay client, `/admin/settings/payment`, `POST /api/callbacks/chillpay` |
| **S9 Checkout** | Code largely shipped; VERIFY open | `payment_orders`, method select, checkout API, mock pay, return status poll |
| **S9 MVP docs** | Code shipped (ahead of old roadmap) | `payment_documents` (`receipt` \| `tax_invoice`), HTML + JSON on paid |
| **S10 Billing** | **Shipped** | `/admin/billing`, `GET /api/platform/billing/orders` |
| **S11 Receipt ops** | **Shipped** | void/reissue, seller branding, `/admin/billing/receipts` |
| **S12 Tax compliance** | **Shipped** | tax profile, document vault, refresh invoices |

## 3. Re-scoped sprint goals

| Sprint | Platform | Goal | Depends | Release note |
| ---: | --- | --- | --- | --- |
| **8** | Platform Admin | **Payment Gateway** — ChillPay/mock config, test connection, MD5 callback log | 3 | Code shipped; **VERIFY with S9** |
| **9** | Tenant | **Buy Package** — method → ChillPay → status → entitlement → **MVP receipt + tax invoice** | 4, 6, 8 | **v1.0.0** (includes S8) |
| **10** | Platform Admin | **Billing** — cross-tenant payment ledger, filters, package/order history, optional usage stub | 9 | Platform commercial ops |
| **11** | Platform Admin | **Receipt** — admin document list, void/reissue, seller branding, PDF-ready export | 10 | Hardens S9 receipt |
| **12** | Tenant | **Tax Invoice** — buyer tax ID / address on KYC or profile, tenant document vault, RD-oriented fields | 10, 11 | Hardens S9 tax invoice |

## 4. Ownership of shared objects

| Object | Create | Read (tenant) | Read (platform) | Mutate later |
| --- | --- | --- | --- | --- |
| `payment_gateway_configs` | S8 | — | S8 | S8 |
| `payment_callback_events` | S8/S9 | — | S8/S10 | — |
| `payment_orders` | S9 | own | S10 ledger | S9 status; S10 notes? |
| `tenant_entitlements` | S4 assign / S9 buy | me | S4 | S9 replace |
| `payment_documents` | S9 auto on paid | own + HTML | S11 list/ops | S11 void/reissue; S12 tax fields |

## 5. Proposed commitment sketches (not opened until prior closes)

### SPRINT-010 — Platform Admin: Billing (~16 pts)

| Draft task | Pts | Outcome |
| --- | ---: | --- |
| Billing list API + schema views | 5 | `GET /api/platform/billing/orders` filter by tenant/status/date |
| Platform admin UI `/admin/billing` | 5 | Table, detail drawer, link to tenant + package |
| Usage / period stub (optional) | 3 | `billing_periods` or read-model from paid orders only |
| E2E + manual | 3 | Platform sees paid mock order from S9 path |

**Out of S10:** PDF brand kit, void document, RD tax schema.

### SPRINT-011 — Platform Admin: Receipt (~14–16 pts)

| Draft task | Pts | Outcome |
| --- | ---: | --- |
| Platform document APIs | 4 | list/get/void/reissue receipts |
| Seller profile / branding | 3 | company name, tax ID, logo on docs |
| Admin UI receipt console | 5 | search, preview, void |
| PDF or print CSS polish | 2 | printable consistency |
| Tests | 2 | void + reissue idempotent |

**In:** Operates on `payment_documents` from S9.  
**Out:** Full e-Tax (ETDA) submission.

### SPRINT-012 — Tenant: Tax Invoice (~14–16 pts)

| Draft task | Pts | Outcome |
| --- | ---: | --- |
| Buyer tax profile fields | 4 | tax ID, branch, address on tenant/KYC |
| Document regeneration with tax fields | 4 | re-issue tax invoice when profile complete |
| Tenant document vault UI | 4 | `/tenant/billing/documents` list + download |
| E2E + manual | 2–4 | paid order → tax invoice shows buyer tax ID |

## 6. VERIFY order (combined gates)

```text
S8 code ──┐
          ├── S9 VERIFY (mock E2E + optional ChillPay sandbox) ── tag v1.0.0
S9 code ──┘
                │
                ▼
         open S10 → S11 → S12 (each with own VERIFY)
```

Do **not** open S10 until S9 TASK-0044 combined UAT passes.

## 7. Risks across the chain

| Risk | Mitigation |
| --- | --- |
| Double-counting “receipt” in S9 and S11 | S9 = issue MVP; S11 = ops only |
| ChillPay ChannelCode for PromptPay merchant-specific | Config override later; default `promptpay` / `creditcard` |
| VAT 7% assumption wrong for some packages | S10/S12 make rate configurable per currency |
| Doc status drift (tasks still `todo` while code exists) | km-sync + doc-audit at S9 close |

## Links

- S8: [SPRINT-008](../03-sprints/SPRINT-008.md) · [FEAT-0008](../01-features/FEAT-0008-payment-gateway.md) · [13-payment-gateway-spec.md](13-payment-gateway-spec.md)
- S9: [SPRINT-009](../03-sprints/SPRINT-009.md) · [FEAT-0009](../01-features/FEAT-0009-buy-package.md) · [14-buy-package-spec.md](14-buy-package-spec.md)
- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Phase C
