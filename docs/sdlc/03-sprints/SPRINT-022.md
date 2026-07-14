---
id: SPRINT-022
status: completed
start: 2026-07-14
end: 2026-07-15
updated: 2026-07-14
design_pack: approved
release_target: v2.3.0
goal: "Platform/Tenant: persist conversation records to MinIO with configurable archive protection and surface knowledge gaps for tenant review."
roadmap_sprint: 22
platform: Platform / Tenant
depends_on: [SPRINT-001, SPRINT-003, SPRINT-015, SPRINT-021]
---

# SPRINT-022 - Platform/Tenant: Conversation Records and Knowledge Gaps

## Goal

Persist tenant conversation artifacts for operations review and convert low-confidence or knowledge-miss turns into tenant-reviewable knowledge-gap records.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S18-S20) | 16, 16, 16 -> **avg 16** |
| **Commitment** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0102](../04-tasks/TASK-0102.md) | 4 | completed | devops | Conversation archive schema and MinIO object contract |
| [TASK-0103](../04-tasks/TASK-0103.md) | 4 | completed | dev | Chat/voice transcript and artifact archive writer |
| [TASK-0104](../04-tasks/TASK-0104.md) | 4 | completed | dev | Knowledge-gap detection, API, and lifecycle state |
| [TASK-0105](../04-tasks/TASK-0105.md) | 3 | completed | dev | Tenant conversation record and knowledge-gap UI |
| [TASK-0106](../04-tasks/TASK-0106.md) | 1 | completed | tester | Archive/gap UAT and cross-tenant access checks |

**Committed:** 16 points

## Scope boundary

**In**

- Tenant-scoped conversation record metadata.
- MinIO archive writes under `calls/` with deterministic tenant/call paths.
- Configurable archive protection/encryption mode where supported by deployment.
- Knowledge-gap candidate creation from RAG misses, low confidence, or fallback answers.
- Tenant review UI/API for records and gap lifecycle.

**Out**

- Ticket routing and human support workflow.
- Satisfaction surveys and tenant statistics dashboards.
- Full retention policy automation and cold storage.

## Feature

- [FEAT-0024 - Conversation Records and Knowledge Gap Review](../01-features/FEAT-0024-conversation-records-knowledge-gaps.md)

## Design pack

| Artifact | Path | Status |
| --- | --- | --- |
| Deep spec | [25-conversation-records-knowledge-gaps-spec.md](../02-design/25-conversation-records-knowledge-gaps-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §66–67 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 22 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Conversation Records & Knowledge Gaps | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § Sprint 22 | `approved` |

> **Closed:** 2026-07-14 as release v2.3.0 scope. Frontend checks/builds passed; Go validation was blocked by a local incomplete Go toolchain (`go: no such tool "vet"`).

## Verification

```bash
make test && make build
# Chat and voice create MinIO archive objects.
# Tenant admin can list conversation records and inspect safe metadata.
# RAG miss creates knowledge-gap candidate and tenant can resolve/snooze it.
# Cross-tenant record/object access returns 404/403.
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-022-manual.md
```

## Risks

| Risk | Mitigation |
| --- | --- |
| PII leakage in object paths or logs | Use tenant/call IDs only; never include customer email in paths |
| Archive writes slow down live calls | Write asynchronously where safe and record retryable state |
| Encryption behavior varies by MinIO/deployment | Document supported local mode and expose a clear setting |
| Knowledge-gap noise overwhelms tenants | Start with conservative detection and lifecycle filters |

## Links

- Depends: [SPRINT-001](SPRINT-001.md), [SPRINT-003](SPRINT-003.md), [SPRINT-015](SPRINT-015.md), [SPRINT-021](SPRINT-021.md)
- Target: **v2.3.0**
