---
id: SPRINT-009
status: in_progress
start: 2026-07-09
end: 2026-07-23
updated: 2026-07-09
release_target: v1.0.0
goal: "Tenant: Buy Package — ChillPay checkout, callback fulfillment, and entitlement assignment; combined E2E verify with SPRINT-008 gateway."
roadmap_sprint: 9
platform: Tenant
depends_on: [SPRINT-004, SPRINT-006, SPRINT-008]
---

# SPRINT-009 — Tenant: Buy Package

## Goal

Let **active tenants** browse the package catalog, **checkout via ChillPay** (`InitPayment` → `PaymentUrl`), and receive **entitlement on successful callback** — completing the SPRINT-008 + SPRINT-009 end-to-end commerce path.

## Context

| Sprint | Shipped capability |
| --- | --- |
| 4 | `packages`, `tenant_entitlements`, platform assign API |
| 6–7 | Tenant register, KYC, `active` tenant |
| 8 | ChillPay config, `InitPayment` client, callback receiver + event log *(code: `d65a122`)* |

**Combined verify (PM request):** Manual UAT and E2E must exercise **SPRINT-008 gateway + SPRINT-009 checkout** in one flow — not isolated per sprint.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0040 | 3 | todo | devops | `payment_orders` schema + order lifecycle states |
| TASK-0041 | 5 | todo | dev | Tenant catalog + checkout APIs — `InitPayment`, mock path |
| TASK-0042 | 5 | todo | dev | Callback fulfillment — order match, entitlement assign, cache invalidation |
| TASK-0043 | 3 | todo | dev | Tenant billing UI — `/tenant/billing`, return page, feedback dialogs |
| TASK-0044 | 2 | todo | tester | Combined E2E + `SPRINT-009-manual.md` (includes SPRINT-008 gateway steps) |

**Committed:** 18 points · **Stretch:** +2 vs avg 16 (commerce E2E)

## Scope boundary

**In**
- `payment_orders` linked to ChillPay `OrderNo` / `TransactionId`
- Tenant checkout for `tenant_admin` on `active` tenants only
- ChillPay `InitPayment` using SPRINT-008 gateway config (env + DB)
- Callback handler upgrade: `PaymentStatus=0` → `paid` + entitlement
- Mock provider: dev/CI checkout without real ChillPay
- Tenant UI at `/tenant/billing` + `/tenant/billing/return`
- `GET /api/entitlements/me` reflects purchase after callback
- `sprint-tech-specs` design pack before TASK-0041

**Out** (→ backlog)
- Billing cycles, invoices, receipts (Sprints 10–12)
- Starter auto-grant on signup/KYC (must purchase)
- Upgrade proration / plan changes mid-cycle
- Platform operator checkout on behalf of tenant
- NATS `payment.completed` events (stretch)

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