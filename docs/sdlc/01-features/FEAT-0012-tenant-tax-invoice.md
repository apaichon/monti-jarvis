# Feature: Tenant Tax Invoice Compliance (FEAT-0012)

**Sprint:** SPRINT-012 · **Status:** shipped · **Release:** v1.3.0

## Problem

Tax invoices lack formal buyer tax ID / branch; tenants need a document vault.

## Scope

In: tax profile CRUD, document list/view, optional reissue of tax invoices on save.  
Out: RD online submission.

## Acceptance criteria

1. `GET/PUT /api/tenant/tax-profile` with company, tax_id, branch, address.
2. `GET /api/tenant/billing/documents` lists own docs; HTML preview.
3. `refresh_invoices=true` reissues active tax invoices with new buyer fields.
4. UI: `/tenant/billing/documents`, `/tenant/billing/tax`.

## Links

- Sprint: [SPRINT-012](../03-sprints/SPRINT-012.md)
