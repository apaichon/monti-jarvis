---
id: SPRINT-060
status: completed
start: 2026-07-30
end: 2026-07-31
closed: 2026-07-31
updated: 2026-07-31
design_pack: approved
roadmap_sprint: 60
feature: FEAT-0052
platform: Tenant / AI Operations / Security
depends_on: [SPRINT-041, SPRINT-043, SPRINT-052]
goal: "Require validated tenant Gemini keys for production AI; no env GEMINI_API_KEY fallback; AI Settings test connection."
velocity_basis: "Last 3 closed: S56=12, S57=12, S58=12 → avg 12; commit 12"
release_target: v2.33.0
release: v2.35.0
git_tag: v2.35.0
---

# SPRINT-060 — Tenant-Owned Gemini Key Enforcement

## Goal

Production tenant chat/voice uses only a validated tenant Gemini key. Tenant
admins manage and test keys in AI Settings; env platform key is never used for
tenant production traffic.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S56–S58) | 12, 12, 12 → **avg 12** |
| **Committed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0216](../04-tasks/TASK-0216.md) Gemini key validation API + metadata | 4 | completed | dev | Test/connect status endpoints |
| [TASK-0217](../04-tasks/TASK-0217.md) Production runtime fail-closed (no env fallback) | 3 | completed | dev | Resolver enforces tenant key |
| [TASK-0218](../04-tasks/TASK-0218.md) Tenant AI Settings key UX | 3 | completed | dev | Enter/test/replace/delete UI |
| [TASK-0219](../04-tasks/TASK-0219.md) Audit, readiness, verification | 2 | completed | tester/dev | Audit + UAT + posture signal |
| **Total** | **12** | **12/12** | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0052](../01-features/FEAT-0052-tenant-owned-gemini-key-enforcement.md) | design_approved |
| Deep spec | **`approved`** | [DES-0056](../02-design/56-tenant-owned-gemini-key-enforcement-spec.md) |
| Workflow | **`approved`** | [02-workflow.md](../02-design/02-workflow.md) §132–134 |
| ER | **`approved`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 60 |
| API | **`approved`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 60 |
| UX | **`approved`** | [05-ux-ui.md](../02-design/05-ux-ui.md) T60 |

## Scope boundary

### In
- Key test/validation, metadata, production fail-closed, AI Settings UX, audit

### Out
- Platform shared pool, multi-provider, billing redesign

## Worktree

```text
.worktrees/SPRINT-060-062
branch: docs/sprint-060-062-plan
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Breaking local dev | Explicit `ALLOW_PLATFORM_GEMINI_FALLBACK=1` for non-prod only |
| Gemini test cost | Bounded lightweight request; rate-limit test endpoint |
| Secret leakage | Never return ciphertext/plaintext; audit scrub |

## Verification

```bash
go test ./internal/store ./internal/gemini ./cmd/server -count=1
cd apps/tenant-web && npm run check
```

Manual UAT: [SPRINT-060-062-manual.md](../06-manual-tests/SPRINT-060-062-manual.md)

## Review - PASS

Validated-key runtime enforcement, production fail-closed behavior, bounded
connection testing, safe status metadata, audit middleware coverage, and
readiness posture were reviewed on 2026-07-31. Automated gates passed; live
Gemini credential UAT remains listed in the manual checklist.

## Shipped summary

Released in the combined Sprint 60–62 train as **v2.35.0** on 2026-07-31
with 12/12 points completed. Production tenant AI now requires a validated
tenant-owned Gemini key, fails closed without one, and exposes bounded key
testing, audit, rate-limit, and readiness behavior.
