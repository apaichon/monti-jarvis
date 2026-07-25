---
id: SPRINT-044
status: backlog
hold_reason: security_review_required
start: 2026-07-25
end: 2026-07-31
updated: 2026-07-25
design_pack: approved
release_target: v2.18.0
roadmap_sprint: 44
feature: FEAT-0038
platform: Customer / Tenant
depends_on: [SPRINT-001, SPRINT-020, SPRINT-021, SPRINT-043]
---

# SPRINT-044 — Customer Generative AI (ON HOLD)

> **ON HOLD:** S44 implementation has been removed pending a security review.
> No customer or tenant CLI/generative execution path is enabled.

## Goal

Give authenticated customers a tenant-scoped generative workspace that can run
bounded provider jobs and return useful artifacts — HTML, image, canvas, link,
report, or document — while keeping provider credentials server-side and
preserving the existing inbound call-center experience.

The first release must deliver one usable provider path end to end. The other
providers may expose bounded adapters and explicit `not_configured` or
`unsupported` states until their runtime contract is available.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded full slices (S37, S39, S42) | 14, 14, 12 → **avg 13.3** |
| **Committed** | **14** |

The commitment is intentionally close to the recent full-slice velocity. It
excludes unrestricted shell execution, model training, and a complete
third-party provider marketplace.

## Commitment

No unassigned `proposed` or `approved` task files were available when this
sprint was planned. The roadmap item is therefore decomposed into work
packages; TASK tickets are created or refined during the technical-spec pass.

| Work package | Points | Owner | Outcome |
| --- | ---: | --- | --- |
| [TASK-0158](../04-tasks/TASK-0158.md) Provider adapter and job contract | 3 | dev | **backlog / held** · Requires security review before any provider execution |
| [TASK-0159](../04-tasks/TASK-0159.md) Credential vault | 2 | dev | **backlog / held** · Requires security review before accepting customer/provider credentials |
| [TASK-0160](../04-tasks/TASK-0160.md) Job execution and artifact persistence | 3 | dev | **backlog / held** · No job or artifact runtime is enabled |
| [TASK-0161](../04-tasks/TASK-0161.md) Artifact outputs | 3 | dev | **backlog / held** · No customer download/preview runtime is enabled |
| [TASK-0162](../04-tasks/TASK-0162.md) Customer workspace UI | 2 | dev | **backlog / held** · Workspace implementation removed |
| [TASK-0163](../04-tasks/TASK-0163.md) Safety, quota, audit, and verification | 1 | tester/dev | **backlog / held** · Security review is a prerequisite |

**Held:** 14 points remain unshipped. The implementation and customer-facing
runtime were removed; no S44 task is complete for release purposes.

## Scope boundary

### In

- Customer-authenticated provider connection using API key or supported
  subscription/login flow.
- Server-side provider adapters with bounded execution, timeout, cancellation,
  and explicit failure states.
- Tenant/customer-scoped generation jobs and artifact metadata.
- MinIO storage under tenant/customer prefixes; Postgres remains the metadata
  authority.
- Output types: HTML template, image, canvas/JSON, signed/public link, report,
  and document export where the provider/runtime supports it.
- Usage events for billing, AI cost, quota, and audit without raw provider
  bodies or credentials.

### Out

- Arbitrary shell execution on a customer device or server.
- Training/fine-tuning custom models.
- Unrestricted third-party skill marketplace.
- Replacing Gemini voice/chat for inbound calls.
- Free unlimited generation, overage billing, or automatic package upgrades.

## Dependencies and design gates

1. Reuse S20 customer authentication and S21 tenant/customer quota context;
   anonymous customer access cannot create generation jobs unless an explicit
   existing policy allows it.
2. Reuse S43 tenant provider configuration, prompts, tools, and skills without
   exposing tenant Gemini or provider secrets to the browser.
3. Approve the provider adapter, credential vault, artifact retention, and
   tenant-isolation contracts before adding new storage or execution paths.
4. Define the S45 quota dimensions and usage-event contract before generation
   usage is surfaced in billing or customer quota cards.

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | approved | [FEAT-0038](../01-features/FEAT-0038-customer-generative-ai.md) |
| Deep spec | approved | [DES-0040](../02-design/40-customer-generative-workspace-spec.md) |
| Workflow | approved | [02-workflow.md](../02-design/02-workflow.md) §97–99 |
| ER | approved | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 44 |
| API | approved | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 44 |
| UX | approved | [05-ux-ui.md](../02-design/05-ux-ui.md) T23 |

## Verification target

```bash
make test
make build
cd apps/customer-web && npm run check && npm run build
git diff --check
# one configured provider creates and completes a tenant-scoped job
# credentials are encrypted/masked and never returned in plaintext
# HTML/image/report/doc artifacts persist and remain tenant-isolated
# Codex and Antigravity return bounded not_configured/unsupported states
# duplicate/retried jobs do not duplicate logical usage or artifacts
# quota, audit, and provider failures return safe user-facing states
```

**UAT checklist:** [SPRINT-044-manual.md](../06-manual-tests/SPRINT-044-manual.md)

## Risks

| Risk | Mitigation |
| --- | --- |
| Provider CLI/API hangs or runs beyond budget | Server-side timeout, cancellation, bounded payload/output, and job state machine |
| API keys or subscription tokens leak | Encrypt at rest, metadata-only responses, redacted logs, and secret-absence tests |
| Generated HTML or links execute unsafe content | Sandboxed preview, signed links, content policy hooks, and safe download behavior |
| Artifacts cross tenant/customer boundaries | Tenant-derived authorization on every job/artifact read and object prefix |
| Provider cost exceeds package quota | Preflight quota check, post-commit usage event, idempotent metering, and explicit unavailable state |

## Release

Target **v2.18.0** — first customer generative workspace slice, subject to
provider credential and artifact-retention design approval.
