---
id: FEAT-0038
title: "Customer generative workspace and tenant-scoped artifacts"
status: backlog
hold_reason: security_review_required
roadmap_sprint: 44
priority: K
depends_on: [SPRINT-001, SPRINT-020, SPRINT-021, SPRINT-043]
updated: 2026-07-25
---

# FEAT-0038: Customer Generative AI (ON HOLD)

> **ON HOLD:** The S44 implementation was removed. Customer and tenant CLI or
> generative execution is disabled until security review and re-approval.

## Purpose

Give an authenticated customer a bounded workspace for connecting a supported
provider, submitting a generation job, and previewing a tenant/customer-scoped
artifact without exposing provider credentials to the browser.

## Acceptance criteria

1. An authenticated customer can save, rotate, mask, expire, and revoke a
   Claude API key; ciphertext is stored server-side and plaintext is never
   returned or written to audit metadata.
2. The API exposes one unified provider/job contract. Claude completes a job;
   Codex, Antigravity, and Grok remain explicit `not_configured` or
   `unsupported` capability states until their bounded adapters are available.
3. Jobs are tenant/customer scoped, bounded by prompt/output size and timeout,
   idempotent for duplicate requests, rate limited, and persisted in Postgres.
4. HTML, image, canvas, link, report, and document outputs have explicit types;
   artifacts are stored under a tenant/customer MinIO prefix and can only be
   read by the owning authenticated customer.
5. The customer workspace shows provider state, credential state, progress,
   history, safe failure codes, and sandboxed HTML/SVG previews.

## Scope

In scope: Claude API adapter, bounded provider capability contract, encrypted
customer credentials, generation jobs, MinIO artifacts, metering/audit hooks,
customer workspace UI, rate limiting, and isolation verification.

Out of scope: arbitrary shell execution, model training, provider marketplace,
subscription-login runtime, unrestricted public generation, and changes to the
inbound Gemini voice/chat path.

## Design links

- [DES-0040 — Customer Generative Workspace](../02-design/40-customer-generative-workspace-spec.md)
- [API contract](../02-design/04-api-spec.md) — Sprint 44
- [Workflow](../02-design/02-workflow.md) — §97–99
- [UX/UI](../02-design/05-ux-ui.md) — T23
