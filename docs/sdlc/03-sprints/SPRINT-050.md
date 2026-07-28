---
id: SPRINT-050
status: completed
start: 2026-07-28
end: 2026-07-28
closed: 2026-07-28
updated: 2026-07-28
design_pack: shipped
release: v2.23.0
release_target: v2.23.0
roadmap_sprint: 50
feature: FEAT-0042
platform: Platform Admin / Finance
depends_on: [SPRINT-004, SPRINT-009, SPRINT-010, SPRINT-011, SPRINT-012, SPRINT-013]
goal: "Platform admin can grant a promotional quota package to a tenant and must set the active plan and issue a tax invoice."
velocity_basis: "Last 3 closed: S42=12, S45=13, S46=17 → avg 14.0; commit 12"
---

# SPRINT-050 — Admin Promotional Package Grant (Active Plan + Tax Invoice)

## Goal

Make promotional package grants available on the **platform admin web**: when an
admin provides a promotion package to a tenant, the system **must set the active
plan** and **issue a tax invoice** for that tenant in one grant workflow.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed slices (S42, S45, S46) | 12, 13, 17 → **avg 14.0** |
| **Committed** | **12** |
| **Completed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0182](../04-tasks/TASK-0182.md) Promotion grant API + store (active plan + tax invoice) | 5 | completed | dev | Atomic platform-admin grant; plan + tax invoice |
| [TASK-0183](../04-tasks/TASK-0183.md) Admin web promotion grant UX | 4 | completed | dev | Tenant UI grant form with plan + invoice confirmation |
| [TASK-0184](../04-tasks/TASK-0184.md) Audit, isolation, verification, UAT | 3 | completed | tester/dev | Unit tests + manual checklist |
| **Total** | **12** | **12/12** | | **Shipped v2.23.0** |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0042](../01-features/FEAT-0042-admin-promotion-package-grant.md) | shipped |
| Deep spec | **`shipped`** | [DES-0046](../02-design/46-admin-promotion-package-grant-spec.md) |
| Workflow | **`shipped`** | [02-workflow.md](../02-design/02-workflow.md) §116–117 |
| ER | **`shipped`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 50 |
| API | **`shipped`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 50 |
| UX | **`shipped`** | [05-ux-ui.md](../02-design/05-ux-ui.md) A50 |

## Scope boundary

### In

- Platform-admin promotional grant for a tenant from admin web
- Select active catalog package → set as **active plan/entitlement**
- **Issue tax invoice** for the tenant on the same grant (atomic TX)
- Grant reason, optional validity, optional amount, idempotency key
- Audit ledger `promotion_grants` + reuse `payment_orders` / `payment_documents`
- Tenant isolation for entitlement and invoice reads

### Out

- Shared Cloud / Dedicated VM commercial redesign (**SPRINT-051**)
- Billing scheduler, proration, dunning
- Referral bonus ledger changes (S46)
- Public product-web grants or tenant self-serve promo checkout
- Changing ChillPay paid checkout fulfillment
- Automatic tax-invoice email delivery

## Relationship to existing entitlement assign

| Path | Active plan | Tax invoice |
| --- | --- | --- |
| `POST …/entitlement` | yes | no |
| Tenant paid checkout | yes | yes (best-effort after pay) |
| **`POST …/promotion-grants` (S50)** | **yes (required)** | **yes (required, same TX)** |

## Shipped summary (v2.23.0)

- `POST/GET /api/platform/tenants/{id}/promotion-grants` and grant detail route
- Atomic store grant: paid promotion order + active entitlement + tax invoice + audit
- Admin UI promotion card on `/admin/tenants/[id]/entitlement`
- Migration/bootstrap: `promotion_grants` (`033_promotion_grants.sql` + ensure)
- Manual UAT checklist: [SPRINT-050-promotion-grant-manual.md](../06-manual-tests/SPRINT-050-promotion-grant-manual.md)

## Verification target

```bash
go test ./internal/store ./cmd/server -count=1
cd apps/platform-admin-web && npm run check
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Half-applied grant (plan without invoice) | Single Postgres TX |
| Duplicate tax invoices on retry | Idempotency key + unique issued doc per order type |
| Confusion with plain entitlement assign | Distinct API + UI “Promotion grant” |

## Notes

- Former roadmap commercial ops track is **SPRINT-051**.
- Release **v2.23.0** (v2.22.0 was Sprint 41 security hardening).
