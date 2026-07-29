---
id: SPRINT-053
status: completed
closed: 2026-07-29
start: 2026-07-28
end: 2026-08-07
updated: 2026-07-29
design_pack: shipped
release_target: v2.26.0
release: v2.26.0
roadmap_sprint: 53
feature: FEAT-0046
platform: Tenant / Customer
depends_on: [SPRINT-016, SPRINT-019, SPRINT-020, SPRINT-021, SPRINT-052]
goal: "Tenant can enable conversation email+OTP auto-register of customers; UI shows app version matching VERSION/tag."
worktree: .worktrees/SPRINT-053
branch: feature/sprint-053-conversation-auto-register-version
velocity_basis: "Last 3 closed: S46=17, S50=12, S52=12 → avg 13.7; commit 12"
---

# SPRINT-053 — Conversation Auto-Register (Email OTP) + App Version on UI

## Goal

Ship a tenant setting that auto-registers customers when they enter email and
complete OTP during conversation, and display the release **app version**
(same as git tag / `VERSION`) on primary UIs.

## Worktree

| Item | Value |
| --- | --- |
| Path | `.worktrees/SPRINT-053` |
| Branch | `feature/sprint-053-conversation-auto-register-version` |
| Base | `main` @ v2.24.0 |

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S46, S50, S52) | 17, 12, 12 → **avg 13.7** |
| **Committed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0188](../04-tasks/TASK-0188.md) Tenant setting + store for conversation auto-register OTP | 4 | completed | dev | Setting get/put; schema; defaults off |
| [TASK-0189](../04-tasks/TASK-0189.md) Conversation email OTP auto-register API + customer UX | 5 | completed | dev | Email→OTP→create/bind in conversation |
| [TASK-0198](../04-tasks/TASK-0198.md) App version API + UI (tenant, admin, optional customer) | 3 | completed | dev | VERSION/tag visible on shells |
| **Total** | **12** | | | |

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0046](../01-features/FEAT-0046-conversation-auto-register-app-version.md) | planned |
| Deep spec | **`approved`** | [DES-0050](../02-design/50-conversation-auto-register-app-version-spec.md) |
| Workflow | **`approved`** | [02-workflow.md](../02-design/02-workflow.md) §120–122 |
| ER | **`approved`** | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 53 |
| API | **`approved`** | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 53 |
| UX | **`approved`** | [05-ux-ui.md](../02-design/05-ux-ui.md) T53 / C53 / A53 |

**Implementation gate:** design approved. Open `in_progress` in this worktree when coding starts.

## Scope boundary

### In

- `auto_register_on_conversation_otp` tenant customer-auth setting (default false)
- Conversation email OTP → auto-create/reuse customer → session bind
- Domain + rate-limit reuse from S20
- App version from `VERSION` on `/healthz` or `/api/version` + UI footers

### Out

- SMS OTP, social login
- Forcing email on all anonymous demos when setting off
- S51 commercial redesign

## Verification target

```bash
go test ./internal/store ./cmd/server -count=1
cd apps/tenant-web && npm run check
cd apps/customer-web && npm run check
cd apps/platform-admin-web && npm run check
# Manual: settings toggle; conversation OTP auto-register; version label == VERSION
```

Manual checklist: `docs/sdlc/06-manual-tests/SPRINT-053-auto-register-version-manual.md` (at VERIFY)

## Risks

| Risk | Mitigation |
| --- | --- |
| Double account create on retry | Idempotent customer upsert by tenant+email |
| Open registration spam | Default off; domain policy; OTP rate limit |
| Version drift vs static assets | Single embed of VERSION at build/serve |
| Clash with require-workforce-auth | Spec co-existence rules in DES-0050 |


## Shipped summary (v2.26.0)

- Tenant setting `auto_register_on_conversation_otp` (default off)
- OTP verify can auto-create customer when enabled
- `GET /api/version` + UI version surfaces (tenant/admin)
- Migration `036_conversation_auto_register.sql`
- Design: DES-0050 · FEAT-0046 · TASK-0188, 0189, 0198

**Closed:** 2026-07-29
