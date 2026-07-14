---
id: SPRINT-024
status: completed
start: 2026-07-14
end: 2026-07-14
updated: 2026-07-14
closed: 2026-07-14
design_pack: approved
release_target: v2.5.0
release: v2.5.0
goal: "Tenant: collect customer satisfaction after AI conversations and provide tenant-scoped statistics for service quality review."
roadmap_sprint: 24
platform: Tenant
depends_on: [SPRINT-022, SPRINT-023]
---

# SPRINT-024 - Tenant: Customer Satisfaction Review and Statistics

## Goal

After a chat or voice conversation ends, invite the customer to submit a 1-5 star satisfaction review and give tenant users reliable, date-filtered statistics to understand AI service quality.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S19, S20, S23) | 16, 16, 16 -> **avg 16** |
| **Commitment** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0112](../04-tasks/TASK-0112.md) | 3 | completed | devops | Tenant-scoped satisfaction review schema and API |
| [TASK-0113](../04-tasks/TASK-0113.md) | 5 | completed | dev | Chat/voice star review prompt and follow-up flow |
| [TASK-0114](../04-tasks/TASK-0114.md) | 5 | completed | dev | Tenant statistics API and dashboard with date filters |
| [TASK-0115](../04-tasks/TASK-0115.md) | 3 | completed | tester | Satisfaction review and statistics UAT |

**Committed:** 16 points · **Completed:** 16 points

## Scope boundary

**In**

- One idempotent 1-5 review linked to each completed call/conversation.
- AI voice and chat prompt after conversation completion.
- Star icon review control and follow-up prompt when the customer skips rating.
- Tenant statistics with default today filter, date range, rating distribution, average, completion rate, avatar, and channel breakdowns.
- Tenant isolation and UAT evidence.

**Out**

- Public reviews or cross-tenant platform analytics.
- Sentiment analysis, free-form comments, CSAT/NPS variants, and billing changes.
- Reopening or extending a completed call to collect a review.

## Feature

- [FEAT-0026 - Customer Satisfaction Review and Tenant Statistics](../01-features/FEAT-0026-customer-satisfaction-statistics.md)

## Design

| Artifact | Path | Status |
| --- | --- | --- |
| Deep spec | [27-customer-satisfaction-statistics-spec.md](../02-design/27-customer-satisfaction-statistics-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §71-72 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 24 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Customer Satisfaction | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § Sprint 24 | `approved` |

Implementation must preserve the existing conversation archive, call-close, ticket, and tenant-scope contracts until the pack is approved.

## Verification

```bash
/usr/local/go/bin/go test ./...
/usr/local/go/bin/go build ./cmd/server
cd apps/customer-web && npm run build
cd ../tenant-web && npm run build
# Completed chat and voice conversations invite a 1-5 star review.
# Unrated calls expose a follow-up review prompt without reopening the call.
# Tenant statistics default to today and support start/end date filters.
# Cross-tenant review/statistics access returns 404/403 without metadata leakage.
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-024-manual.md
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Rating prompt interrupts call close | Ask after completion and keep archive/session close independent of review submission |
| Duplicate or delayed voice events create multiple reviews | Enforce one review per call/conversation and idempotent submission |
| Tenant statistics expose customer context | Aggregate only and enforce tenant scope on every query |
| Empty or sparse data misleads operators | Show review count, completion rate, empty states, and explicit date range |

## Links

- Depends: [SPRINT-022](SPRINT-022.md), [SPRINT-023](SPRINT-023.md)
- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Sprint 24
- Target: **v2.5.0**
