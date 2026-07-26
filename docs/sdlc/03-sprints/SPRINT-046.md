---
id: SPRINT-046
status: completed
start: 2026-08-08
end: 2026-08-14
updated: 2026-07-25
closed: 2026-07-25
design_pack: pending
release_target: v2.20.0
release: pending
release_scope: complete_17_of_17_points_manual_uat_ready
roadmap_sprint: 46
platform: Platform / Tenant / Growth
feature: Tenant referral affiliate program and bonus quota
depends_on: [SPRINT-009, SPRINT-010, SPRINT-013, SPRINT-031, SPRINT-045]
carry_over_from: SPRINT-045
---

# SPRINT-046: Tenant Referral Affiliate Program and Bonus Quota

> **REBUILT COMPLETE:** the S45 carry-over, referral foundation, bonus ledger,
> quota enforcement, reporting, and tenant referral UX are implemented. The
> release cut is pending; the refreshed manual UAT is ready for local sign-off.

## Goal

Establish a tenant-scoped, auditable referral and bonus-quota program. A
referral can be captured once at signup and qualified only after the referred
tenant is active, KYC-approved, and has a paid, non-voided package order. A
qualified referral grants configurable, expiring quota through a separate
bonus layer without mutating base entitlements.

## Commitment

| Task | Scope | Points | Status |
| --- | --- | ---: | --- |
| TASK-0166 | Carry over S45 usage/reconciliation hardening | 3 | completed |
| TASK-0167 | Carry over S45 dashboard/reporting verification | 2 | completed |
| TASK-0168 | Carry over S45 manual UAT | 2 | completed |
| TASK-0169 | Referral attribution and qualification foundation | 3 | completed |
| TASK-0175 | Referral bonus quota ledger, enforcement, and affiliate UX | 7 | completed |
| **Total** | | **17** | **17 completed** |

## Acceptance focus

1. Each tenant can retrieve one tenant-scoped referral code.
2. Signup attribution is immutable and idempotent; self, duplicate, circular,
   disabled, and cross-tenant misuse is rejected.
3. Referral states include `clicked`, `attributed`, `pending`, `qualified`,
   `rejected`, and `reversed`, with auditable status changes.
4. Qualification requires active tenant status, approved KYC, and a paid order
   without a voided payment document.
5. Qualification grants configurable bonus quota through an append-only ledger;
   purchased package limits and historical usage remain unchanged.
6. Tenant and platform usage show base, bonus, and total values, and quota
   enforcement uses valid bonus remaining across supported dimensions.

## Verification

Manual UAT: [SPRINT-046-referrals-manual.md](../06-manual-tests/SPRINT-046-referrals-manual.md)

```bash
go test ./...
make test
make build
git diff --check
```

## Close note

Automated verification, migration, and web builds passed for the rebuilt
scope. The refreshed step-by-step manual UAT is ready for a local fixture run;
it must be attached before the v2.20.0 release cut.
