---
id: SPRINT-010
status: completed
start: 2026-07-11
end: 2026-07-11
updated: 2026-07-11
release_target: v1.3.0
release: v1.3.0
goal: "Platform Admin: Billing ledger — cross-tenant payment orders, filters, order detail with documents."
roadmap_sprint: 10
platform: Platform Admin
depends_on: [SPRINT-009]
---

# SPRINT-010 — Platform Admin: Billing ledger

## Goal

Platform operators browse **all payment orders**, filter by tenant/status, and open order detail (docs link).

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0045 | 5 | completed | dev | `GET /api/platform/billing/orders` + detail |
| TASK-0046 | 5 | completed | dev | Platform UI `/admin/billing` |
| TASK-0047 | 3 | completed | dev | Order detail + nav |
| TASK-0048 | 3 | completed | tester | Smoke via mock checkout path |

**Committed:** 16 points

## Scope

**In:** payment order ledger (read-model from `payment_orders`), filters, package/tenant names.  
**Out:** usage metering deep, void/reissue (S11), tax profile (S12).

## Verification

```bash
make restart
# platform login → /admin/billing
curl -H "Authorization: Bearer $PLATFORM" http://localhost:8091/api/platform/billing/orders | jq .
```

## Links

- [15-commerce-chain-plan.md](../02-design/15-commerce-chain-plan.md)
- [FEAT-0010](../01-features/FEAT-0010-platform-billing.md)
