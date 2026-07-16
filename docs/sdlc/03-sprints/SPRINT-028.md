---
id: SPRINT-028
status: completed
start: 2026-07-16
end: 2026-07-17
updated: 2026-07-16
closed: 2026-07-16
design_pack: approved
release_target: v2.9.0
release: v2.9.0
goal: "Platform: capture tenant-scoped audit events, spool them safely on the backend, and deliver them to ClickHouse with retryable retention."
roadmap_sprint: 28
platform: Platform
depends_on: [SPRINT-003, SPRINT-026]
---

# SPRINT-028 - Platform: Cross-Tenant Audit Log

## Goal

Give platform operators a reliable, cross-tenant audit trail for security-sensitive and operational changes. Events must be captured with actor and tenant context, remain durable on the backend while ClickHouse is unavailable, and be removed locally only after a confirmed ClickHouse insert and the configured retention period.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S25, S26, S27) | 16, 16, 16 -> **avg 16** |
| **Proposed commitment** | **16** |

## Proposed commitment

The roadmap work is committed at 16 points and decomposed into four implementation tasks.

| Work package | Points | Owner | Outcome |
| --- | ---: | --- | --- |
| Audit event contract and emission | 4 | dev | A stable event envelope with tenant, actor, action, resource, request, outcome, and timestamp context; instrument security-sensitive platform and tenant mutations without recording secrets or raw customer content |
| Backend local spool and lifecycle worker | 4 | devops | Append-only JSONL spool with `audit_log_YYYYMMDD-HH-MM-SS.jsonl` naming, atomic rotation, configurable flush interval defaulting to 5 seconds, retry/backoff, and graceful shutdown drain |
| ClickHouse sink, acknowledgement, and retention | 4 | dev | ClickHouse audit table and batch insert path with deterministic event IDs, retry-safe delivery, explicit sink mode, and deletion only for files older than one hour whose transfer has been confirmed |
| Platform audit query, verification, and operations | 4 | tester/dev | Platform-scoped API/UI filters, operational health signals, tenant isolation tests, failure/replay tests, and UAT evidence for local recovery and ClickHouse delivery |

## Implementation tasks

| Task | Scope | Points | Status |
| --- | --- | ---: | --- |
| [TASK-0128](../04-tasks/TASK-0128.md) | Audit event contract and HTTP emission | 4 | `completed` |
| [TASK-0129](../04-tasks/TASK-0129.md) | Local audit spool and transfer worker | 4 | `completed` |
| [TASK-0130](../04-tasks/TASK-0130.md) | ClickHouse audit sink and query projection | 4 | `completed` |
| [TASK-0131](../04-tasks/TASK-0131.md) | Platform audit API, UI, and verification | 4 | `completed` |

**Committed:** 16 points · **Implementation completed:** 16 points · **Manual UAT:** pending

## Design

The Sprint 28 technical design pack is approved and implementation is complete. Manual browser and infrastructure UAT is deferred to the next tester run.

| Artifact | Scope | Status |
| --- | --- | --- |
| Feature | [FEAT-0030 - Cross-Tenant Audit Log](../01-features/FEAT-0030-cross-tenant-audit-log.md) | `approved` |
| Deep spec | [31-cross-tenant-audit-log-spec.md](../02-design/31-cross-tenant-audit-log-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §§80–82 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 28 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Platform Audit Log | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) A20 | `approved` |

## Scope boundary

**In**

- A structured immutable audit event containing at least `event_id`, `occurred_at`, `tenant_id`, `actor_id`, `actor_type`, `action`, `resource_type`, `resource_id`, `request_id`, `source`, `outcome`, and bounded metadata.
- Existing `internal/auditctx` actor propagation reused at request and background-job boundaries.
- Backend-local append-only JSONL files named `audit_log_YYYYMMDD-HH-MM-SS.jsonl`; event timestamps use RFC3339 UTC and filenames use UTC for deterministic operations.
- Default `AUDIT_LOG_MODE=spool`, which writes locally and transfers closed files to ClickHouse from a Go routine every `AUDIT_LOG_FLUSH_INTERVAL` (default `5s`).
- `AUDIT_LOG_MODE=clickhouse` for deployments that select ClickHouse as the active sink, with the same event schema and retry/error handling contract.
- Configurable local spool directory, flush interval, batch size, retry backoff, and retention period. The retention default is one hour.
- ClickHouse batch insertion with deterministic event IDs and a retry-safe deduplication strategy for uncertain network outcomes.
- File transfer state or an equivalent durable acknowledgement marker so a local file is deleted only after ClickHouse accepts the complete batch and the file is older than the retention period.
- Platform-only audit query access with tenant, actor, action, resource, outcome, and date filters, bounded pagination, and redacted metadata.
- Metrics/logging for queue depth, transfer failures, oldest pending file, last successful transfer, and files retained because acknowledgement is missing.

**Out**

- Audit history exposed to tenant customers or anonymous callers.
- Full request/response bodies, access tokens, passwords, OTPs, provider credentials, audio, transcripts, or unrestricted personal data in audit payloads.
- Compliance certification, legal hold, tamper-proof archival, or external SIEM/webhook delivery.
- Replacing the existing Postgres `created_by`/`updated_by` audit columns; this sprint adds an event trail alongside them.
- Platform-wide performance, billing, quota, or AI-cost dashboards planned for later Phase G sprints.

## Design gates

The technical-spec pack must resolve these decisions before implementation:

1. Define the event envelope, action/resource taxonomy, actor types, redaction rules, and the minimum required events.
2. Define the local spool state machine: active versus closed files, rotation boundaries, transfer markers, crash recovery, and graceful shutdown behavior.
3. Define `AUDIT_LOG_MODE` semantics, configuration validation, default values, bounded queue behavior, and what happens when ClickHouse is disabled or unavailable.
4. Define ClickHouse table engine/order/deduplication strategy, batch acknowledgement semantics, replay behavior after uncertain responses, and query consistency expectations.
5. Define platform authorization, filters, pagination, retention observability, and manual UAT coverage.

## Verification

```bash
go test ./...
go build ./cmd/server
git diff --check
# An authenticated mutation produces one structured event with the correct tenant and actor.
# Anonymous/system events use an explicit system actor and never inherit a previous request actor.
# Spool mode creates correctly named JSONL files and transfers closed files on the configured interval.
# ClickHouse outage retains local files, retries without loss, and does not delete unacknowledged data.
# A successful batch acknowledgement followed by the retention threshold permits deletion only for that file.
# Retry after an uncertain insert does not create duplicate logical events in audit queries.
# Platform users can filter audit events without crossing tenant authorization boundaries.
# Audit payloads contain no credentials, tokens, OTPs, audio, transcripts, or unbounded request bodies.
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-028-manual.md
```

Release validation passed on 2026-07-16 with the full Go test suite, server build, platform-admin Svelte check/build, API contract coverage for tenant `logo_url`, and `git diff --check`. Manual browser and ClickHouse outage/recovery UAT remains deferred.

**Closed:** 2026-07-16 as release `v2.9.0`.

## Risks

| Risk | Mitigation |
| --- | --- |
| ClickHouse outage grows the local spool until disk pressure | Bound queue/batch sizes, expose oldest-pending metrics, preserve files on failure, and document an operator alert threshold |
| Network timeout occurs after ClickHouse accepted a batch | Use deterministic event IDs and a ClickHouse deduplication/query strategy before deleting local files |
| Audit logging slows customer or admin requests | Use a bounded asynchronous writer and make failure behavior explicit; never perform unbounded ClickHouse work in the request path |
| Sensitive request data leaks into the audit stream | Allowlisted fields only, bounded metadata, central redaction tests, and no raw bodies or credentials |
| Multiple server instances write to one local directory | Treat the spool directory as instance-local and include an instance/source field; do not assume shared filesystem coordination |

## Links

- Depends: [SPRINT-003](SPRINT-003.md), [SPRINT-026](SPRINT-026.md)
- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Sprint 28
- Existing actor context: `internal/auditctx`
- Existing persistence audit columns: `internal/store/audit.go` and `scripts/migrations/001_audit_columns_postgres.sql`
- Feature: [FEAT-0030](../01-features/FEAT-0030-cross-tenant-audit-log.md)
- Deep spec: [31-cross-tenant-audit-log-spec.md](../02-design/31-cross-tenant-audit-log-spec.md)
- Target: **v2.9.0**
- Next planned: Sprint 29 platform system performance monitoring
