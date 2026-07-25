---
id: UAT-SPRINT-045
status: deferred
updated: 2026-07-25
sprint: SPRINT-045
owner: Tester
---

# SPRINT-045 — Manual Test Checklist

The verified v2.18.0 slice is covered by automated tests and builds. Docker-
backed two-tenant/load verification remains deferred with the usage-ledger and
reconciliation follow-up.

## Automated evidence

- [x] Four AiaaS seed definitions and rules-v2 validation pass.
- [x] Separate web/mobile minute counters pass focused tests.
- [x] Snapshot rows expose source/freshness and null unavailable usage.
- [x] Mobile SDK build, `make test`, `make build`, and `git diff --check` pass.

## Deferred integration UAT

- [ ] Run Postgres, Redis DB 4, MinIO, and ClickHouse fixtures.
- [ ] Verify two-tenant isolation across all four package tiers.
- [ ] Verify mobile disconnect, timeout, retry, failed-start, and concurrent
  load behavior.
- [ ] Verify usage-ledger replay and reconciliation correction states.

The deferred checks must be completed before broad customer production traffic.
