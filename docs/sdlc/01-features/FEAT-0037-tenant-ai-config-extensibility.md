---
id: FEAT-0037
title: "Embed auth, grouped configuration, and tenant AI extensibility"
status: planned
roadmap_sprint: 43
priority: D+
depends_on: [SPRINT-014, SPRINT-015, SPRINT-016, SPRINT-039]
updated: 2026-07-19
---

# FEAT-0037: Tenant AI Configuration and Embed Auth

## Purpose

Give tenants controlled configuration of their embed access and AI runtime while
preserving the existing public embed default, platform secret boundaries, and
tenant isolation.

## Acceptance criteria

1. An embed configured with `auth=true` blocks chat and voice until the customer
   completes the configured authentication flow; `auth=false` preserves the
   current public embed behavior.
2. Core infrastructure settings are separated from grouped operational and
   product configuration, with documented local and production loading behavior.
3. A tenant can save an encrypted Gemini API key; subsequent tenant-scoped AI
   calls use it when configured, and no API returns the plaintext key.
4. A tenant can manage a bounded custom system prompt that is applied to its
   chat and voice orchestration within safety and length limits.
5. A tenant can register and enable allowlisted call tools scoped to its tenant.
6. A tenant can create and assign skills composed of prompt and tool bundles,
   with tenant isolation and audit-safe CRUD behavior.

## Scope

In scope: embed auth mode, environment/config groups, encrypted tenant Gemini
key storage and resolution, tenant prompts, allowlisted tools, and tenant
skills. Out of scope: a third-party skill marketplace, multi-provider LLM
switching, and replacing the platform Gemini default for every tenant.

## Dependencies

- Sprint 14 public embed resolve and iframe surface.
- Sprint 15 tenant KM/scope boundaries.
- Sprint 16 tenant settings and limits shell.
- Sprint 39 tenant branding and embed configuration surface.

## Design links

- [DES-0039 — Tenant AI Configuration and Embed Auth](../02-design/39-tenant-ai-config-extensibility-spec.md)
- [API contract](../02-design/04-api-spec.md) — Sprint 43
- [Workflow](../02-design/02-workflow.md) — §93–96
- [UX/UI](../02-design/05-ux-ui.md) — T22
