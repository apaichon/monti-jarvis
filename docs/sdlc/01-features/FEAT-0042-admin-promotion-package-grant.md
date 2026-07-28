---
id: FEAT-0042
title: "Admin promotional package grant with active plan and tax invoice"
status: shipped
roadmap_sprint: 50
priority: P
depends_on: [SPRINT-004, SPRINT-009, SPRINT-010, SPRINT-011, SPRINT-012, SPRINT-013]
design: DES-0046
design_spec: ../02-design/46-admin-promotion-package-grant-spec.md
release: v2.23.0
updated: 2026-07-28
---

# FEAT-0042: Admin Promotional Package Grant (Active Plan + Tax Invoice)

## Purpose

Let platform administrators grant a **promotional quota package** to a tenant
from the platform admin web. Every successful grant **must** set the selected
package as the tenant's **active plan** and **issue a tax invoice** for that
tenant.

## Problem

Operators can already manage the package catalog and assign entitlements, but a
plain entitlement assign does not create finance documents. Sales-approved
promotions, complimentary trials, and operational goodwill grants need the same
commercial trail as paid packages:

- an active plan (entitlement snapshot) the tenant can use immediately
- a tax invoice for finance/compliance
- an audit of who granted which package and why

Without a first-class grant path, operators either skip invoices (finance gap)
or hand-edit records outside the product.

## Scope

### In

- Platform-admin-only promotional grant flow on admin web (tenant context)
- Select an **active catalog package** as the promotion plan
- Set that package as the tenant's **sole active entitlement/plan**
- Issue a **tax invoice** for the tenant linked to the promotion grant
- Capture grant reason / notes, optional validity window, acting admin identity
- Atomic success or rollback (no active plan without invoice; no orphan invoice)
- Tenant isolation and audit-friendly responses for platform receipt/tax search
- Reuse existing package catalog, entitlement, and payment-document authorities

### Out

- Tenant self-serve promotional checkout
- Changing Shared Cloud / Dedicated VM catalog redesign (S51)
- Billing scheduler, dunning, proration calculator (S51)
- Referral bonus-quota ledger changes (S46 remains as-is)
- Public product-web entitlement grants
- Hard-coded tax rates outside finance configuration
- Automatic email delivery of tax invoices (download/history is enough)

## Acceptance criteria

1. A `platform_admin` can open a tenant in admin web and grant a promotional
   package chosen from the **active** package catalog.
2. On successful grant, the tenant has **exactly one active plan/entitlement**
   for the selected package with a rules snapshot (previous active entitlement
   is superseded with an auditable reason such as `promotion_grant`).
3. On successful grant, the system **issues a tax invoice** for that tenant
   with package line context, promotion source, immutable document number, and
   seller/buyer tax fields as configured.
4. Grant is atomic: if tax invoice issuance fails, the tenant's active plan is
   not left half-applied; if plan activation fails, no tax invoice is issued.
5. Non-`platform_admin` callers cannot grant promotions; tenant A cannot read
   tenant B's grant outcome or tax invoice.
6. Existing paid checkout fulfillment and plain entitlement assign (if still
   exposed) remain available, but the **promotion grant** path is the supported
   admin path that guarantees plan + tax invoice together.
7. Platform receipt/tax search can find the issued tax invoice; tenant billing
   history (where already exposed) can show the document for that tenant only.

## Test notes

- Functional: grant with no prior entitlement; grant superseding an existing plan
- Failure: force invoice failure and assert plan not left active incorrectly
- AuthZ: tenant_admin / customer cannot call grant API
- Isolation: two tenants; each only sees own invoice/entitlement
- Finance: document type `tax_invoice`, status issued, unique number
- UI: admin confirmation shows active plan name + tax invoice number

## Dependencies

- packages: `internal/packages`, `internal/entitlements`, `internal/store`,
  payment documents / billing ops, `cmd/server` platform handlers
- admin UI: `apps/platform-admin-web` tenant entitlement / promotion surface
- design: [DES-0046](../02-design/46-admin-promotion-package-grant-spec.md),
  [08-packages-spec.md](../02-design/08-packages-spec.md), S9–S12 tax invoice paths
- API: [04-api-spec.md](../02-design/04-api-spec.md) § Sprint 50
- UX: [05-ux-ui.md](../02-design/05-ux-ui.md) A50
- blueprint: `docs/monti_multi_tenant_ai_call_center_blueprint.md`
