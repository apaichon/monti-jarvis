---
id: SPRINT-017
status: completed
start: 2026-07-12
end: 2026-07-12
closed: 2026-07-12
updated: 2026-07-12
design_pack: shipped
release_target: v1.8.0
release: v1.8.0
goal: "Tenant: Test and Preview sandbox — embed-like desk with package-charged chat/voice before go-live."
roadmap_sprint: 17
platform: Tenant
depends_on: [SPRINT-015, SPRINT-016]
---

# SPRINT-017 — Tenant: Test and Preview

## Goal

Give **active tenants** a first-class **Preview** desk (embed-like avatar UI) to validate agents, KM, and locale **before go-live**, with the **same package metering** as production and sessions logged as `source=preview`.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S14–S16) | 16, 16, 16 → **avg 16** |
| Trailing average | **16** |
| **Commitment** | **16** |
| **Completed** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0077](../04-tasks/TASK-0077.md) | 3 | done | devops | Preview session flag + schema/keys |
| [TASK-0078](../04-tasks/TASK-0078.md) | 5 | done | dev | Preview chat/voice APIs (package quotas apply) |
| [TASK-0079](../04-tasks/TASK-0079.md) | 4 | done | dev | Tenant UI `/tenant/preview` (embed-like) |
| [TASK-0080](../04-tasks/TASK-0080.md) | 3 | done | dev | Scenario checklist + embed link + lang/voice UX |
| [TASK-0081](../04-tasks/TASK-0081.md) | 1 | done | tester | Manual UAT checklist |

**Committed:** 16 points · **Completed:** 16

## Shipped summary (v1.8.0)

| Area | Outcome |
| --- | --- |
| APIs | `GET /api/tenant/preview/scenarios`, `POST /api/tenant/preview/chat`, `GET /ws/tenant/preview/voice` |
| Quota | Same package path as production (rate, concurrent, monthly minutes, S16 daily/per-call) |
| Logging | `call_sessions.source=preview` |
| UI | Embed-like panel: avatar portrait, agent/topic/lang, chat, voice, scenarios |
| Voice UX | Status loading steps, agent greets first, multi-lang (`auto`/`en`/`th`), localhost mic handoff |
| Live desk | Connecting status + greeting-first on customer `/` and embed |

## Feature

- [FEAT-0019 — Tenant Test and Preview](../01-features/FEAT-0019-tenant-test-preview.md)

## Design pack

| Artifact | Path | Status |
| --- | --- | --- |
| Deep spec | [20-tenant-test-preview-spec.md](../02-design/20-tenant-test-preview-spec.md) | `shipped` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §49–51 | `shipped` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 17 | `shipped` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Tenant preview | `shipped` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § T10 | `shipped` |

## Verification

```bash
make build && make test
open http://localhost:8091/tenant/preview
# UAT: docs/sdlc/06-manual-tests/SPRINT-017-manual.md
```

## Production launch gate (carry forward)

Before **customer production** after tenant **customer-user auth** (S19–20): verify **rate limit + package quota** under multi-user load.

## Links

- Depends: [SPRINT-015](SPRINT-015.md), [SPRINT-016](SPRINT-016.md)  
- Next: SPRINT-018 Customer Tier  
- Release: **v1.8.0**
