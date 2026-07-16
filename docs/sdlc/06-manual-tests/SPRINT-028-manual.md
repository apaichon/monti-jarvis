---
id: UAT-028
title: Sprint 28 cross-tenant audit log
sprint: SPRINT-028
status: deferred_uat
release: v2.9.0
updated: 2026-07-16
---

# Sprint 28 UAT

## Preconditions

- Server is running with ClickHouse configured and `AUDIT_LOG_MODE=spool`.
- Platform-admin credentials are available; a tenant-admin credential is available for denial checks.
- The platform-admin web build is served from the configured `PLATFORM_ADMIN_WEB_DIR`.
- `AUDIT_LOG_DIR` is writable by the backend process.

## Scenarios

| ID | Scenario | Expected result |
| --- | --- | --- |
| UAT-028-01 | Perform a protected platform or tenant mutation | One JSONL event contains tenant, actor, action, outcome, request ID, and bounded metadata; no request body or secret is present. |
| UAT-028-02 | Inspect the local spool after rotation | File name matches `audit_log_YYYYMMDD-HH-MM-SS.jsonl`; active files are not transferred while open. |
| UAT-028-03 | Stop or disable ClickHouse, then perform a mutation | The local file remains, transfer health becomes degraded/unavailable, and no file is deleted. |
| UAT-028-04 | Restore ClickHouse and wait for the configured interval | Closed files transfer oldest-first and receive an acknowledgement marker. |
| UAT-028-05 | Age an acknowledged file beyond `AUDIT_LOG_RETENTION` | The JSONL file and marker are removed; an unacknowledged file remains. |
| UAT-028-06 | Open `/admin/audit-logs` as a platform admin | Events, delivery status, filters, metadata expansion, and pagination render correctly. |
| UAT-028-07 | Request audit endpoints as tenant admin or anonymous | API returns unauthorized/forbidden and no cross-tenant event data. |
| UAT-028-08 | Retry one event after an uncertain ClickHouse response | The platform list shows one logical event after deduplication. |
| UAT-028-09 | Inspect events containing quotes or newline-like metadata | JSONL and ClickHouse JSONEachRow remain valid; values are redacted/bounded. |

## Automated checks

- `/usr/local/go/bin/go test -vet=off ./...`
- `/usr/local/go/bin/go build ./cmd/server`
- `cd apps/platform-admin-web && npm run check`
- `cd apps/platform-admin-web && npm run build`
- `git diff --check`

## Sign-off

| Tester | Date | Result | Defects |
| --- | --- | --- | --- |
| Pending manual verification |  |  |  |

## Release close note

Automated audit event, spool, ClickHouse sink, platform API/UI build, tenant logo response, Go, and diff validation passed on 2026-07-16. Manual browser, ClickHouse outage/recovery, and retention evidence is deferred to the next tester run.
