---
id: SPRINT-012
status: completed
start: 2026-07-11
end: 2026-07-11
updated: 2026-07-11
release_target: v1.3.0
release: v1.3.0
goal: "Tenant: Tax invoice compliance — buyer tax profile, document vault, reissue on profile update."
roadmap_sprint: 12
platform: Tenant
depends_on: [SPRINT-010, SPRINT-011]
---

# SPRINT-012 — Tenant: Tax Invoice compliance

## Goal

Tenants maintain **buyer tax profile**, browse **document vault**, and optionally **refresh tax invoices** when profile changes.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0053 | 4 | completed | dev | `tenant_tax_profiles` + GET/PUT tax-profile |
| TASK-0054 | 4 | completed | dev | Document vault APIs + reissue on refresh |
| TASK-0055 | 4 | completed | dev | UI `/tenant/billing/documents` + `/tax` |
| TASK-0056 | 2 | completed | tester | Tax ID appears on reissued invoice |

**Committed:** 14 points

## Scope

**In:** tax_id, branch, company, address; vault list/view; refresh tax invoices.  
**Out:** Full RD/e-Tax submission APIs.

## Links

- [15-commerce-chain-plan.md](../02-design/15-commerce-chain-plan.md)
- [FEAT-0012](../01-features/FEAT-0012-tenant-tax-invoice.md)
