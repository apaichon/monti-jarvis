---
id: FEAT-0044
title: "Shared Cloud and Dedicated VM commercial plans and billing operations"
status: shipped
roadmap_sprint: 51
priority: P
depends_on: [SPRINT-009, SPRINT-010, SPRINT-012, SPRINT-013, SPRINT-025, SPRINT-031, SPRINT-045, SPRINT-048, SPRINT-050]
design: DES-0048
design_spec: ../02-design/48-commercial-plans-billing-operations-spec.md
updated: 2026-07-29
release: v2.25.0
---

# FEAT-0044: Shared Cloud and Dedicated VM Commercial Operations

## Purpose

Provide one tenant-safe commercial authority for Shared Cloud subscriptions and
Dedicated VM quotations, including catalog calculations, quota visibility,
renewal scheduling, receipts, and tax invoices.

## Acceptance criteria

1. Packages expose explicit `deployment_mode` and `purchase_mode` values.
2. Shared Cloud plans use server-authoritative monthly or annual checkout.
3. Dedicated VM plans collect company/capacity information and never enter the
   tenant payment gateway.
4. Platform admins can move quotes through review, capacity, quotation,
   acceptance, provisioning, and activation.
5. Current plan returns package, subscription period, next bill, quota
   dimensions, freshness, compact utilization, and document links.
6. Billing and sidebar consume the same current-plan contract.
7. Renewal cycles are unique by subscription and period, use bounded retries,
   and settle only after receipt and tax invoice issuance.
8. Historical catalog, calculation, subscription, and document snapshots do
   not change when a package is edited.

## Delivery state

Shipped in **v2.25.0**. The release includes Shared Cloud checkout, Dedicated
quotation operations, current-plan quota/next-bill contracts, replay-safe
billing cycles, receipt/tax-invoice issuance, and a visible platform admin quote
request monitor. Production scheduler activation remains an environment
enablement step because `BILLING_SCHEDULER_ENABLED` defaults to `false`.
