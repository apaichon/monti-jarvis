---
id: SPRINT-066
status: skipped
start: 2026-08-01
end: 2026-08-02
updated: 2026-08-02
design_pack: skipped
roadmap_sprint: 66
feature: external-backup-tools
platform: Infra / Platform Admin / DevOps
depends_on: [SPRINT-002, SPRINT-022, SPRINT-025, SPRINT-028, SPRINT-029, SPRINT-036, SPRINT-041, SPRINT-049]
goal: "Evaluate backup/restore for Monti Postgres, ClickHouse, and MinIO."
velocity_basis: "Skipped by product decision; no shipped points."
release_target: skipped
release: none
---

# SPRINT-066 - Backup/Restore Evaluation (Skipped)

## Decision

Sprint 66 is closed as skipped. No v2.38.0 release will be cut, no tag will be
created, and no backup/restore implementation will be merged to main.

## Rationale

The in-app backup/restore slice created too much operational surface for the
current release. Monti will use dedicated external tooling for Postgres,
ClickHouse, and MinIO backup/restore, then revisit app-level visibility only
after the external runbooks are selected.

## Cleanup

- Removed Platform Admin backup/restore UI and API routes from main.
- Removed backup metadata tables, migration, store helpers, and runner code.
- Removed backup-related env variables from examples.
- Removed Sprint 66 design/API/ER/UX additions from shared design docs.

## Outcome

| Item | Result |
| --- | --- |
| Release | skipped; no version bump |
| Merge | not merged to main |
| Tag | none |
| Velocity | 0 shipped points |
| Follow-up | Use external backup tools for Postgres, ClickHouse, and MinIO later |
