---
id: SPRINT-011
status: completed
start: 2026-07-11
end: 2026-07-11
updated: 2026-07-11
release_target: v1.3.0
release: v1.3.0
goal: "Platform Admin: Receipt ops — list, void, reissue, seller branding, printable HTML."
roadmap_sprint: 11
platform: Platform Admin
depends_on: [SPRINT-010]
---

# SPRINT-011 — Platform Admin: Receipt ops

## Goal

Operate on `payment_documents`: **list / void / reissue**, configure **seller branding**, print-ready HTML.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0049 | 4 | completed | dev | Document list + void + reissue APIs |
| TASK-0050 | 3 | completed | dev | Seller branding GET/PUT |
| TASK-0051 | 5 | completed | dev | UI `/admin/billing/receipts` + seller |
| TASK-0052 | 2 | completed | tester | Void/reissue smoke |

**Committed:** 14 points

## Scope

**In:** status `issued|voided`, reissue creates new active doc, seller block on new issues.  
**Out:** e-Tax government submission, PDF binary store.

## Links

- [15-commerce-chain-plan.md](../02-design/15-commerce-chain-plan.md)
- [FEAT-0011](../01-features/FEAT-0011-platform-receipts.md)
