---
id: SPRINT-031
status: in_progress
start: 2026-07-17
end: 2026-07-18
updated: 2026-07-17
design_pack: approved
release_target: v2.12.0
goal: "Platform: provide read-only billing, quota, and AI infrastructure cost usage with source reconciliation and explicit measurement coverage."
roadmap_sprint: 31
feature: FEAT-0033
platform: Platform
depends_on: [SPRINT-010, SPRINT-013, SPRINT-025, SPRINT-030]
---

# SPRINT-031 - Platform Billing, Quota, and AI Infrastructure Cost Usage

## Goal

Give platform administrators a safe, read-only operational view of paid package value, current quota enforcement, historical usage, and AI infrastructure cost coverage by tenant and date range.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S28, S29, S30) | 16, 16, 16 -> **avg 16** |
| **Proposed commitment** | **16** |

## Commitment proposal

| Work package | Points | Owner | Outcome |
| --- | ---: | --- | --- |
| [TASK-0140](../04-tasks/TASK-0140.md) Usage metering and reconciliation contract | 5 | devops | Idempotent ClickHouse AI usage projection, versioned rates, and reporting-source contract |
| [TASK-0141](../04-tasks/TASK-0141.md) Platform billing/quota/AI usage API | 4 | dev | Bounded platform-admin aggregate API with RBAC, redaction, and partial-failure states |
| [TASK-0142](../04-tasks/TASK-0142.md) AI instrumentation and usage dashboard | 5 | dev | Gemini usage capture plus responsive platform billing usage surface |
| [TASK-0143](../04-tasks/TASK-0143.md) Reconciliation and UAT | 2 | tester | Payment, quota, AI coverage, failure-state, and responsive-layout verification |

**Status:** In progress; implementation and automated verification are complete. Manual UAT remains pending per UAT-031.

## Scope boundary

**In**

- Read-only platform view of paid orders, active packages, quota enforcement snapshots, historical call minutes, and AI usage/cost coverage.
- Provider-neutral AI meter with `observed`, `estimated`, and `unavailable` states.
- Idempotent ClickHouse AI usage projection and versioned rate metadata.
- Bounded aggregate API, tenant pagination, platform-admin RBAC, redaction, and explicit partial-source failure behavior.
- Responsive platform billing usage UI and reconciliation/UAT fixtures.

**Out**

- Charging, invoice generation, refunds, tax documents, auto-upgrades, entitlement mutation, or customer-facing cost display.
- Replacing Redis quota enforcement or treating reporting facts as a second quota authority.
- Raw prompts, responses, transcripts, audio, customer identifiers, provider payloads, exports, alerts, or scheduled reports.

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0033 - Platform Billing, Quota, and AI Cost Usage](../01-features/FEAT-0033-platform-billing-quota-ai-cost-usage.md) | `in_progress` |
| Deep spec | [34-platform-billing-quota-ai-cost-spec.md](../02-design/34-platform-billing-quota-ai-cost-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) Sprint 31 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 31 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Platform Billing Usage | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) Sprint 31 | `approved` |

Implementation is gated on approval of the deep spec and API contract. AI cost totals must not be exposed as exact until the measurement state and rate version are present.

## Verification target

```bash
make test
make build
cd apps/platform-admin-web && npm run check && npm run build
git diff --check
```

See [34-platform-billing-quota-ai-cost-spec.md](../02-design/34-platform-billing-quota-ai-cost-spec.md) for API, data, workflow, UX, privacy, and reconciliation acceptance criteria.

## Risks

| Risk | Mitigation |
| --- | --- |
| Gemini provider usage is incomplete for voice or older responses | Preserve observed/estimated/unavailable states and show coverage explicitly. |
| Billing and entitlement state diverge | Report reconciliation warnings without mutating either authority. |
| Redis counters differ from historical range usage | Label current enforcement and historical reporting as separate periods. |
| Meter retries duplicate cost | Deterministic event ids, replacing projection rows, and duplicate-delivery tests. |
