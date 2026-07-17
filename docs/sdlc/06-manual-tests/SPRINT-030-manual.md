---
id: UAT-030
title: Sprint 30 Platform Call Center Statistics by Tenant UAT
status: deferred_uat
updated: 2026-07-17
sprint: SPRINT-030
release: v2.11.0
---

# Sprint 30 Manual UAT

Use a platform-admin session against the platform admin web app. Use controlled ClickHouse/Postgres fixtures and separate tenant-admin/anonymous sessions for authorization checks.

| ID | Scenario | Expected result | Status |
| --- | --- | --- | --- |
| S30-U01 | Open `/admin/call-center` as platform admin | Today is selected; KPIs, breakdowns, freshness, and tenant rows load | pending |
| S30-U02 | Apply an explicit start/end date range | Range is inclusive and response timezone is visible | pending |
| S30-U03 | Use Today after changing dates | Both dates reset to today and offset resets to zero | pending |
| S30-U04 | Compare aggregate totals with the sum of tenant rows | Controlled facts reconcile for completed count, duration, and range minutes | pending |
| S30-U05 | Page through tenant rows and use an exact tenant filter | Limit/offset are bounded and only the requested tenant row is returned | pending |
| S30-U06 | Use a valid range with no completed facts | Zero totals and an explicit empty state are shown, not an analytics error | pending |
| S30-U07 | Make ClickHouse unavailable or stale | Stale values remain labeled; unavailable returns a safe retry state and never zeroes activity | pending |
| S30-U08 | Make ratings or package enrichment unavailable | Activity remains visible and only the affected enrichment is marked unavailable | pending |
| S30-U09 | Use tenant-admin and anonymous sessions | API returns safe `403`/`401`; no aggregate or tenant metadata is rendered | pending |
| S30-U10 | Open below 700px width | Filters stack, KPI cards remain readable, and tenant rows do not require horizontal scrolling | pending |
| S30-U11 | Expire the platform session while open | Local auth is cleared and the app returns to login without an infinite retry loop | pending |
| S30-U12 | Repeat the Sprint 29 monitoring/audit smoke checks | Existing monitoring, audit, tenant statistics, calls, and quota behavior remains compatible | pending |

See [33-platform-call-center-statistics-spec.md](../02-design/33-platform-call-center-statistics-spec.md), [SPRINT-030](../03-sprints/SPRINT-030.md), and [05-ux-ui.md](../02-design/05-ux-ui.md) A22.

## Release close note

Automated Go tests, server build, platform-admin type checks/build, aggregate contract tests, authorization/error checks, and `git diff --check` passed. Browser, responsive-layout, and dependency-failure UAT scenarios remain `pending` and are deferred to the next tester run; this does not claim those manual scenarios passed.
