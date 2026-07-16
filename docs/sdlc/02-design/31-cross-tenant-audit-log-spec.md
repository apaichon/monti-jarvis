---
id: DES-0031
title: Cross-Tenant Audit Log Specification
status: approved
updated: 2026-07-16
sprint: SPRINT-028
owner: SA
---

# Cross-Tenant Audit Log - Design Spec

Sprint: SPRINT-028 · Release target: v2.9.0
Feature: FEAT-0030 - Cross-Tenant Audit Log
Depends on: 06-auth-spec.md, 29-tenant-system-performance-spec.md

## 1. Goals

- Capture a consistent audit event for security-sensitive and operational mutations across tenant and platform surfaces.
- Preserve tenant and actor context from the request boundary through asynchronous delivery.
- Keep the request path independent from ClickHouse availability by default.
- Make backend-local audit files recoverable, replayable, and safe to delete only after confirmed delivery.
- Give platform administrators a bounded cross-tenant search surface with explicit filters and redacted metadata.

## 2. Non-goals

- Tenant/customer audit views, SIEM export, webhooks, legal holds, or compliance certification.
- Full request/response logging or storage of credentials, tokens, OTPs, audio, transcripts, or raw personal data.
- Replacing Postgres row-level `created_by` and `updated_by` columns.
- Synchronous ClickHouse writes in the default request path.

## 3. Environment variables

| Variable | Default | Contract |
| --- | --- | --- |
| `AUDIT_LOG_MODE` | `spool` | `spool` appends to local files and transfers closed files; `clickhouse` uses the ClickHouse sink as the active destination. Reject unknown values at startup. |
| `AUDIT_LOG_DIR` | `./var/audit` | Instance-local directory for active files, closed files, and transfer acknowledgement markers. |
| `AUDIT_LOG_FLUSH_INTERVAL` | `5s` | Go worker interval for scanning and transferring closed files. Must be positive and bounded. |
| `AUDIT_LOG_RETENTION` | `1h` | Minimum age before an acknowledged local file may be deleted. |
| `AUDIT_LOG_BATCH_SIZE` | `500` | Maximum events per ClickHouse insert. |
| `AUDIT_LOG_QUEUE_SIZE` | `10000` | Maximum in-memory event queue before the writer records a drop/error metric. |
| `AUDIT_LOG_RETRY_BACKOFF` | `1s` | Initial retry delay with bounded exponential backoff. |

ClickHouse connection variables remain `CLICKHOUSE_URL`, `CLICKHOUSE_DB`, `CLICKHOUSE_USER`, and `CLICKHOUSE_PASSWORD`. `AUDIT_LOG_MODE=clickhouse` requires a configured ClickHouse client; `spool` remains able to capture files while ClickHouse is disabled or unavailable.

## 4. Event contract

The event is immutable and JSON-serializable. Metadata is an allowlisted, bounded object; it is not a copy of the request body.

```json
{
  "event_id": "evt_01J2...",
  "occurred_at": "2026-07-16T09:10:11.123Z",
  "tenant_id": "tenant_01",
  "actor_id": "user_01",
  "actor_type": "platform_admin|tenant_admin|customer|system|anonymous",
  "action": "tenant.avatar.assignment.updated",
  "resource_type": "tenant_avatar_assignment",
  "resource_id": "assignment_01",
  "request_id": "req_01",
  "source": "platform_api",
  "outcome": "success|denied|failure",
  "metadata": {"avatar_id": "ava", "changed_fields": ["status"]}
}
```

Rules:

- `event_id` is generated once at emission and remains stable across all retries.
- `occurred_at` is UTC. The filename timestamp is also UTC and uses `YYYYMMDD-HH-MM-SS`.
- `tenant_id` is the resolved scope. Platform-only events may use an empty tenant ID and must identify the platform source.
- `actor_id` comes from `internal/auditctx`; missing request actors resolve to the explicit `system` actor.
- `request_id` is propagated from the HTTP request or generated for a background operation.
- `outcome=denied` is emitted for authorization failures only when the route has enough context to avoid leaking another tenant's existence.
- Metadata is limited by key allowlist and byte size. Redaction occurs before enqueueing.

Initial action families include authentication/session changes, tenant and avatar assignment changes, package/entitlement changes, customer import/auth changes, billing mutations, ticket lifecycle changes, KM changes, and platform configuration changes. The implementation must centralize event construction so adding a covered mutation does not bypass redaction.

## 5. Local spool state machine

Each instance owns `AUDIT_LOG_DIR`. It must not assume that the directory is shared by multiple server instances.

```text
active .jsonl
    | rotate on flush/shutdown/size boundary
    v
closed .jsonl
    | worker claims file with an exclusive lock
    v
transferring
    | ClickHouse accepts every batch
    v
ack marker (.uploaded with count/checksum/time)
    | file age >= AUDIT_LOG_RETENTION
    v
delete .jsonl and marker

closed/transferring --failure--> retained .jsonl --retry--> transferring
```

File contract:

- Name: `audit_log_YYYYMMDD-HH-MM-SS.jsonl`.
- One complete JSON event per line, newline terminated, UTF-8.
- The active file is never uploaded while open. Rotation closes and flushes it before the worker can claim it.
- A collision in the same second reuses the active file rather than creating an ambiguous second timestamped file; implementations may add an internal instance suffix only when required for multi-process safety.
- The acknowledgement marker records event count, byte count, checksum, ClickHouse batch result, and acknowledgement time. It is written atomically after the final batch succeeds.
- On startup, stale locks are recovered only when the owning process is no longer active. Files without a valid marker remain eligible for transfer.

The worker is a single goroutine per server instance with a ticker configured by `AUDIT_LOG_FLUSH_INTERVAL`. It drains the bounded in-memory queue into the active file, scans closed files oldest-first, sends bounded batches, and emits delivery metrics. Shutdown stops new intake, drains the queue, closes the active file, performs a bounded final transfer, and leaves unacknowledged files on disk.

## 6. ClickHouse model

Database: `monti_jarvis`. Table: `audit_events`.

```sql
CREATE TABLE IF NOT EXISTS audit_events (
  event_id String,
  occurred_at DateTime64(3, 'UTC'),
  tenant_id String,
  actor_id String,
  actor_type LowCardinality(String),
  action LowCardinality(String),
  resource_type LowCardinality(String),
  resource_id String,
  request_id String,
  source LowCardinality(String),
  outcome LowCardinality(String),
  metadata_json String,
  ingested_at DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = ReplacingMergeTree(ingested_at)
ORDER BY (tenant_id, occurred_at, event_id);
```

`event_id` is the logical identity. Retries may insert the same event again after an uncertain network response; platform queries use `FINAL` or an equivalent deterministic deduplication query so one logical event is displayed. The local file is deleted only when the complete batch request receives a successful ClickHouse response. A timeout or partial response is not an acknowledgement.

ClickHouse stores redacted audit metadata only. It does not receive local paths, credentials, raw request bodies, or audio/transcript content.

## 7. Delivery modes and failure behavior

| Mode | Request path | Worker | ClickHouse unavailable |
| --- | --- | --- | --- |
| `spool` | Enqueue event and return after bounded local writer acceptance | Transfers closed files every interval | Keep files and retry; expose delivery health; do not delete |
| `clickhouse` | Enqueue to the same bounded writer; active sink is ClickHouse | Flushes batches to ClickHouse; may use a short local failure buffer | Preserve the local failure buffer/files and retry; never silently discard |

The implementation must not make a customer or platform request wait for an unbounded ClickHouse call. Queue overflow is observable and returns the route's existing safe success/failure policy; security-critical routes must not claim audit success when the event was not accepted by the local writer.

## 8. Platform API and RBAC

Only `platform_admin` may query cross-tenant audit events. Tenant administrators do not receive this API in Sprint 28.

- `GET /api/platform/audit-logs` supports `tenant_id`, `actor_id`, `action`, `resource_type`, `outcome`, `start_date`, `end_date`, `limit`, and `cursor`.
- `GET /api/platform/audit-logs/health` returns sink mode, queue depth, last successful transfer, oldest pending file age, failed file count, and ClickHouse availability without returning local paths or raw errors.
- Limits are bounded; default page size is 50 and maximum is 200. Dates are interpreted as UTC unless an explicit platform timezone contract is added later.
- Metadata is returned only after server-side redaction and truncation. No endpoint exposes spool file contents or ClickHouse credentials.

Denied roles receive the existing unauthorized/forbidden envelope. Invalid filters return `validation_error`; unavailable analytics returns `analytics_unavailable` with retry guidance and no infrastructure details.

## 9. Verification

```bash
go test ./...
go build ./cmd/server
git diff --check
# Event actor and tenant context are correct for authenticated, anonymous, and system paths.
# Event emission is non-blocking and queue overflow is observable.
# File names, newline-delimited JSON, rotation, restart recovery, and graceful drain pass tests.
# Failed ClickHouse inserts retain files and retry on the configured interval.
# Successful insert acknowledgement writes the marker; deletion requires the retention threshold.
# Duplicate delivery of one event_id returns one logical event from platform queries.
# Platform admin filters and pagination work; tenant/admin/customer roles cannot query cross-tenant data.
# Secret/token/body/audio/transcript redaction tests pass.
```

See [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), [04-api-spec.md](04-api-spec.md), and [05-ux-ui.md](05-ux-ui.md).
