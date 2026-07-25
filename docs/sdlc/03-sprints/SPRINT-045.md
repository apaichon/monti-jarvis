---
id: SPRINT-045
status: completed
start: 2026-08-01
end: 2026-08-07
updated: 2026-07-25
design_pack: approved
release_target: v2.18.1
release: v2.18.1
release_scope: partial_6_of_13_points_manual_uat_deferred
closed: 2026-07-25
roadmap_sprint: 45
feature: FEAT-0039
platform: Platform / Tenant / Mobile
depends_on: [SPRINT-013, SPRINT-016, SPRINT-025, SPRINT-027, SPRINT-030, SPRINT-031, SPRINT-043]
---

# SPRINT-045 — AiaaS Mass-Market Packages and Usage Reconciliation

> **CLOSED AS PARTIAL:** v2.18.1 ships the verified 6-point slice plus the
> follow-up ledger/projection implementation. Manual UAT and the remaining
> acceptance evidence are explicitly deferred to Sprint 46; 7 points carry
> over and are not counted as complete.

## Goal

Offer four simple monthly AiaaS packages and make package limits, web/mobile
usage, storage, KM, avatars, concurrency, billing, and statistics agree on the
same tenant-scoped dimension contract.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed slices (S37, S39, S42) | 14, 14, 12 → **avg 13.3** |
| **Proposed capacity** | **13** |

## Proposed work packages

| Work package | Points | Owner | Outcome |
| --- | ---: | --- | --- |
| Package initialization and entitlement snapshots | 3 | dev | Idempotent four-tier defaults, platform-admin catalog CRUD, package-change history — **TASK-0164 completed** |
| Dimensioned quota enforcement | 3 | dev | Storage/mobile dimensions and stable quota response contract — **TASK-0165 completed** |
| Idempotent usage ledger and reconciliation | 3 | devops/dev | Replay-safe events, source watermarks, mismatch/correction states — **TASK-0166 carry-over** |
| Statistics and billing projections | 2 | dev | Tenant/platform current-vs-historical usage consistency — **TASK-0167 carry-over** |
| Mobile enforcement and verification | 2 | dev/tester | Mobile quota metadata, lifecycle release, two-tenant/load UAT — **TASK-0168 carry-over** |

**Delivered:** 6 points across package initialization and dimensioned quota
enforcement. **Carry-over:** 7 points across TASK-0166, TASK-0167, and
TASK-0168. Manual UAT is deferred by release decision and tracked in the
Sprint 45 manual runbook.

## Partial release record

TASK-0164 and TASK-0165 are completed and verified. The release includes the
four idempotent AiaaS catalog defaults, rules-v2 storage/mobile dimensions,
separate mobile/web counters, mobile bootstrap metadata, safe unavailable
source reporting, and storage projection reads. Docker-backed two-tenant/load
UAT and idempotent usage-ledger/reconciliation work remain open under the
three remaining tasks below.

## Carry-over tasks to Sprint 46

| Task | Points | Status | Outcome |
| --- | ---: | --- | --- |
| [TASK-0166](../04-tasks/TASK-0166.md) | 3 | in_progress | Idempotent usage ledger and bounded reconciliation runs |
| [TASK-0167](../04-tasks/TASK-0167.md) | 2 | in_progress | Tenant/platform current-vs-historical statistics and billing projections |
| [TASK-0168](../04-tasks/TASK-0168.md) | 2 | in_progress | Mobile lifecycle verification, two-tenant isolation, and load UAT |

## Scope boundary

### In

- Catalog-driven AiaaS package tiers and entitlement snapshots.
- Dimensioned quota enforcement for web, embed, mobile, KM, avatars, storage,
  and concurrent calls.
- Idempotent usage events and bounded reconciliation projections.
- Consistent tenant/platform billing and statistics contracts.
- Mobile allowance metadata and failure/retry/disconnect verification.

### Out

- Customer generative workspace or any Claude/Codex/CLI execution.
- Overage billing, automatic upgrades, quota pooling, and enterprise pricing.
- Replacement of Gemini inbound chat/voice.

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0039](../01-features/FEAT-0039-aiaas-packages-usage-reconciliation.md) | `in_progress` |
| Deep spec | [DES-0042](../02-design/42-aiaas-packages-usage-reconciliation-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §100–104 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 45 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 45 | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) T24/A24/M2 | `approved` |

Implementation is gated on PM approval of commercial values, the dimension
contract, reconciliation authority, and all API/ER changes.

## Verification target

```bash
make test
make build
go test ./internal/quota ./internal/metering ./cmd/server -count=1
# all four tiers return catalog-driven dimensions
# web/embed/mobile use the same tenant entitlement snapshot
# duplicate usage events do not double count
# failed starts, disconnects, timeouts, deletes, and retries reconcile safely
# two-tenant concurrent-load tests preserve isolation
git diff --check
```

## Close note

Automated tests, build, migration, and transactional duplicate-safety smoke
checks passed. The local Docker stack was healthy during migration. The
step-by-step manual UAT remains deferred and must be executed before broad
customer production traffic or declaring the carry-over tasks complete.
