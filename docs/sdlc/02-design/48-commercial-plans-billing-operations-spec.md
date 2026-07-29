---
id: DES-0048
title: Shared Cloud and Dedicated VM Commercial Operations Specification
status: approved
updated: 2026-07-28
sprint: SPRINT-051
owner: SA
feature: FEAT-0044
---

# DES-0048 — Shared Cloud and Dedicated VM Commercial Operations

## Goals and invariants

- Persist `shared_cloud|dedicated_vm` deployment mode and
  `self_serve|quote` purchase mode.
- Treat the server catalog version as the authority for price, annual discount,
  tax, currency, and quota rules; ignore browser-supplied amounts.
- Allow payment checkout only for `self_serve`.
- A Dedicated quote submission creates no payment order, document,
  subscription, or entitlement.
- Activate Dedicated service only after controlled quote and provisioning
  transitions.
- Use one current-plan response for billing and sidebar.
- Treat unavailable usage as unknown, never as zero.
- Make renewal replay safe by unique subscription/period cycle.

## Data model

Migration: `scripts/migrations/035_commercial_operations.sql`.

| Entity | Authority |
| --- | --- |
| `packages` | Explicit deployment and purchase modes |
| `package_versions` | Immutable price, tax, currency, and rules snapshot |
| `dedicated_quote_requests` | Tenant company/capacity request and operator lifecycle |
| `tenant_subscriptions` | Billing anchor, period, state, immutable calculation |
| `billing_cycles` | One renewal action per subscription and period |
| Existing `payment_orders` | Payment state |
| Existing `payment_documents` | Receipt and tax-invoice authority |
| Existing `tenant_entitlements` | Runtime package/quota snapshot |

One live subscription is allowed per tenant. A cycle is unique on
`(subscription_id, period_key)`.

## Calculation

```text
monthly base = package_version.monthly_price_cents
annual base  = monthly base × 12
discount     = annual base × annual_discount_bps / 10_000
taxable      = max(base - discount, 0)
tax          = round(taxable × tax_rate_bps / 10_000)
amount due   = taxable + tax
```

The default catalog policy is a 20% annual discount and 7% tax. Dedicated
calculations are indicative and return `quote_required=true`.

## Workflows

```text
Shared:
catalog → calculate → payment order + pending subscription
        → provider success → documents → entitlement + active subscription

Dedicated:
catalog → company/capacity request
        → review → capacity confirmed → quoted → accepted → provisioning
        → operator activation → entitlement + active invoice subscription

Renewal:
due subscription → SKIP LOCKED claim → unique cycle + one order
                 → paid → receipt + tax invoice → settle period once
                 └→ bounded retry → grace → suspension
```

Calendar anchors clamp to the final valid day of the target month. For example,
31 January renews on 28/29 February.

## API

| Method | Route | Role |
| --- | --- | --- |
| `POST` | `/api/tenant/commercial/calculate` | tenant admin |
| `POST/GET` | `/api/tenant/commercial/quotes` | own tenant |
| `GET` | `/api/tenant/commercial/current-plan` | own tenant |
| `GET` | `/api/platform/commercial/quotes` | platform admin |
| `PATCH` | `/api/platform/commercial/quotes/{id}` | platform admin |
| `POST` | `/api/platform/billing/cycles/{id}/retry` | platform admin |
| `POST` | `/api/tenant/checkout` | Shared Cloud only |

Dedicated checkout returns `409 PACKAGE_REQUIRES_QUOTE`.

## Scheduler controls

| Variable | Default |
| --- | --- |
| `BILLING_SCHEDULER_ENABLED` | `false` |
| `BILLING_SCHEDULER_POLL_INTERVAL` | `1m` |
| `BILLING_GRACE_PERIOD` | `72h` |
| `BILLING_RETRY_DELAYS` | `1h,6h,24h` |

Production must enable the scheduler only after migration and UAT. Database
claims use `FOR UPDATE SKIP LOCKED`; retries never create a second cycle/order.

## UX contract

- Shared cards show **Buy** and monthly/annual controls.
- Dedicated cards show **Request quotation** and company/contact/capacity form.
- Current Plan shows package/state, billing interval, current period, next
  charge, dimensioned quota usage/freshness, compact utilization, and documents.
- The sidebar reads the same store and does not contain hard-coded plan or
  utilization values.
- Compact utilization is the highest reliable finite quota ratio.

## Verification record

- 263 Go tests passed across 36 packages.
- Tenant Svelte check: 0 errors, 0 warnings.
- Platform Svelte check: 0 errors; 29 baseline warnings in untouched pages.
- Shared annual checkout, authoritative amount, payment, entitlement, receipt,
  and tax invoice passed against Postgres.
- Dedicated quote lifecycle passed and created zero payment orders.
- Repeated scheduler ticks produced one cycle and one order; payment settled
  the cycle and advanced the subscription once.

## Approval

The product-owner instruction to build Sprint 51 authorizes implementation of
this design. Release remains subject to the linked manual UAT checklist.
