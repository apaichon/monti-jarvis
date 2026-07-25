---
id: SPRINT-046
status: in_progress
start: 2026-08-08
end: 2026-08-14
updated: 2026-07-25
design_pack: pending
release_target: v2.19.0
roadmap_sprint: 46
platform: Platform / Tenant / Growth
feature: Tenant referral affiliate program and bonus quota
depends_on: [SPRINT-009, SPRINT-010, SPRINT-013, SPRINT-031, SPRINT-045]
carry_over_from: SPRINT-045
---

# SPRINT-046: Tenant Referral Affiliate Program and Bonus Quota

## Goal

Establish a tenant-scoped, auditable referral foundation. A referral can be
captured once at signup and qualified only after the referred tenant is active,
KYC-approved, and has a paid, non-voided package order. Bonus quota grants and
the full affiliate UX remain later slices; this sprint does not mutate base
entitlements.

## Commitment

| Task | Scope | Points | Status |
| --- | --- | ---: | --- |
| TASK-0166 | Carry over S45 usage/reconciliation hardening | 3 | in_progress |
| TASK-0167 | Carry over S45 dashboard/reporting verification | 2 | in_progress |
| TASK-0168 | Carry over S45 manual UAT | 2 | in_progress |
| TASK-0169 | Referral attribution and qualification foundation | 3 | completed |
| **Total** | | **10** | |

## Acceptance focus

1. Each tenant can retrieve one tenant-scoped referral code.
2. Signup attribution is immutable and idempotent; self, duplicate, circular,
   disabled, and cross-tenant misuse is rejected.
3. Referral states include `clicked`, `attributed`, `pending`, `qualified`,
   `rejected`, and `reversed`, with auditable status changes.
4. Qualification requires active tenant status, approved KYC, and a paid order
   without a voided payment document.
5. No referral operation changes the purchased package, historical usage, or
   quota limits. Bonus ledger work is explicitly out of this slice.

## Verification

```bash
go test ./...
make test
make build
git diff --check
```
