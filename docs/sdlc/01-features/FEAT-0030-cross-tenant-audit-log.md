---
id: FEAT-0030
title: "Cross-Tenant Audit Log"
status: completed
owner: product
created: 2026-07-16
updated: 2026-07-16
sprint: SPRINT-028
---

# FEAT-0030: Cross-Tenant Audit Log

## Purpose

Give platform administrators a reliable, searchable history of security-sensitive and operational changes across tenants without putting ClickHouse availability on the request path or retaining local audit files indefinitely.

## Scope

- Emit a structured, immutable audit event with tenant, actor, request, resource, outcome, and timestamp context.
- Append events to backend-local JSONL spool files named `audit_log_YYYYMMDD-HH-MM-SS.jsonl`.
- Transfer closed files to ClickHouse every five seconds by default, with a configurable interval and explicit ClickHouse sink mode.
- Retry failed or interrupted transfers, deduplicate logical events by deterministic event ID, and retain failed files.
- Delete only acknowledged local files after the configured one-hour retention period.
- Provide platform-admin audit search with tenant, actor, action, resource, outcome, and date filters.
- Expose bounded delivery health information for operators and tests.

## Out of scope

- Tenant/customer-facing audit history.
- Raw request and response bodies, credentials, tokens, OTP values, audio, transcripts, and unrestricted personal data.
- SIEM, webhook, legal-hold, compliance certification, or tamper-proof archival integrations.
- Replacing existing Postgres audit columns.

## Acceptance criteria

1. A covered authenticated or system mutation creates one structured audit event with the resolved tenant and actor context.
2. Spool mode writes closed JSONL files using the required timestamp naming convention and does not block the request on ClickHouse.
3. The background worker transfers closed files on the configured interval, retries transient failures, and retains files when acknowledgement is missing.
4. A local file is deleted only after complete batch acknowledgement and after the one-hour retention threshold.
5. Retry after an uncertain ClickHouse response does not produce duplicate logical events in platform queries.
6. Platform administrators can search and paginate audit events without tenant scope leakage; non-platform roles cannot access the cross-tenant API.
7. Audit payloads are allowlisted and redacted; credentials, tokens, OTPs, audio, transcripts, and raw request bodies never enter ClickHouse or spool files.

## Design links

- [DES-0031 - Cross-Tenant Audit Log Specification](../02-design/31-cross-tenant-audit-log-spec.md)
- [04-api-spec.md](../02-design/04-api-spec.md) Sprint 28 section
- [02-workflow.md](../02-design/02-workflow.md) Sprint 28 sections
- [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 28 storage contract
- [05-ux-ui.md](../02-design/05-ux-ui.md) Sprint 28 platform surface
