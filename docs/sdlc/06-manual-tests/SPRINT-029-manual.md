---
id: UAT-029
title: Sprint 29 Platform System Performance Monitoring UAT
status: deferred_uat
updated: 2026-07-17
sprint: SPRINT-029
release: v2.10.0
---

# Sprint 29 Manual UAT

Use a platform-admin session against the platform admin web app. Run the API checks with a separate tenant-admin and anonymous session to verify access control.

| ID | Scenario | Expected result | Status |
| --- | --- | --- | --- |
| S29-U01 | Open `/admin/monitoring` as platform admin | Loading state resolves to summary, dependency matrix, audit delivery, and tenant table | pending |
| S29-U02 | Apply an exact `tenant_id` filter | Only the requested tenant is returned; `total`, `limit`, and `offset` are coherent | pending |
| S29-U03 | Filter by `operational`, `degraded`, and `unavailable` | Results use derived tenant status and preserve stable table layout | pending |
| S29-U04 | Make ClickHouse analytics unavailable or stale | Snapshot remains safe; analytics is marked unavailable/stale and raw provider details are absent | pending |
| S29-U05 | Create an audit delivery backlog | Audit state shows bounded status/counts without local directories, file names, or event payloads | pending |
| S29-U06 | Use tenant-admin and anonymous sessions | API returns safe `403`/`401`; no tenant metadata is rendered | pending |
| S29-U07 | Trigger a slow dependency probe | Request completes within the configured monitoring timeout and renders degraded/unavailable state | pending |
| S29-U08 | Expire the platform session while the page is open | Local auth state is cleared and the app returns to login without an infinite retry loop | pending |
| S29-U09 | Open at a viewport below 700px | Navigation collapses, filters stack, and tenant rows remain readable without horizontal scrolling | pending |

See [32-platform-system-performance-spec.md](../02-design/32-platform-system-performance-spec.md), [SPRINT-029](../03-sprints/SPRINT-029.md), and [05-ux-ui.md](../02-design/05-ux-ui.md) A21.

## Release close note

Automated Go tests, server build, platform-admin type checks/build, authenticated API smoke, authorization/error checks, and `git diff --check` passed. Browser, responsive-layout, and dependency-failure UAT scenarios remain `pending` and are deferred to the next tester run; this does not claim those manual scenarios passed.
