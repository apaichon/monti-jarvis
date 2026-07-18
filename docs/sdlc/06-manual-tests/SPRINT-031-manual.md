---
id: UAT-031
title: Sprint 31 Platform Billing, Quota, and AI Usage UAT
status: deferred_uat
updated: 2026-07-17
sprint: SPRINT-031
release: v2.12.0
---

# Sprint 31 Manual UAT

Use a platform-admin session with two active tenants and controlled Postgres, Redis, and ClickHouse fixtures. Keep a separate tenant-admin and anonymous session for authorization checks.

| ID | Scenario | Expected result | Status |
| --- | --- | --- | --- |
| S31-U01 | Open `/admin/billing/usage` as platform admin | Paid value, reporting minutes, quota snapshot, AI coverage, freshness, and tenant rows load | pending |
| S31-U02 | Apply inclusive date boundaries containing paid and unpaid orders plus call facts | Only paid orders and facts inside the inclusive range are included; timezone is visible | pending |
| S31-U03 | Deliver the same AI event twice | One logical usage fact remains; aggregate cost is not doubled | pending |
| S31-U04 | Use observed, estimated, unavailable, and missing-rate AI fixtures | States and costs remain separated; unavailable data is not shown as zero/exact total | pending |
| S31-U05 | Compare reporting minutes with Redis monthly/daily quota counters | Historical activity and current enforcement are shown as separate dimensions; divergence is labeled | pending |
| S31-U06 | Create paid-order/entitlement mismatch | Orders and package state remain read-only and reconciliation shows a warning | pending |
| S31-U07 | Page active tenants and filter by an exact tenant id | Limit/offset are bounded and only allowlisted aggregate fields are returned | pending |
| S31-U08 | Make ClickHouse or AI usage projection unavailable | Required activity failure shows a safe retry state; optional AI failure remains partial/unavailable without raw infrastructure details | pending |
| S31-U09 | Use tenant-admin and anonymous sessions | API returns safe `403`/`401`; no platform aggregate is rendered | pending |
| S31-U10 | Expire the platform session while the page is open | Session is cleared and the existing login flow is used without an infinite retry loop | pending |
| S31-U11 | Open below 700px width and exercise loading, empty, stale, retry, and pagination states | Filters stack, metrics remain readable, and tenant usage remains usable on mobile | pending |
| S31-U12 | Repeat call, voice, archive, quota, payment callback, audit, and tenant-statistics smoke checks | Existing customer and enforcement paths remain compatible | pending |

See [DES-0034](../02-design/34-platform-billing-quota-ai-cost-spec.md), [SPRINT-031](../03-sprints/SPRINT-031.md), and [05-ux-ui.md](../02-design/05-ux-ui.md) A23.

## Verification note

Automated Go tests, server build, platform-admin type checks/build, and `git diff --check` pass. Browser, responsive-layout, and dependency-failure scenarios remain pending for tester execution; this document does not claim those manual scenarios passed.
