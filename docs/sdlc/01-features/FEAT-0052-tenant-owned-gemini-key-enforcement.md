---
id: FEAT-0052
title: "Tenant-owned Gemini key enforcement (no production env fallback)"
status: design_approved
release: pending
roadmap_sprint: 60
priority: D+
depends_on: [SPRINT-041, SPRINT-043, SPRINT-052]
design: DES-0056
design_spec: ../02-design/56-tenant-owned-gemini-key-enforcement-spec.md
updated: 2026-07-30
---

# FEAT-0052: Tenant-Owned Gemini Key Enforcement

## Purpose

Require each tenant to own a validated Gemini API key for production AI
runtime. Stop using the shared platform `GEMINI_API_KEY` env fallback for
tenant chat/voice traffic in production.

## Problem

Sprint 43 stores encrypted tenant Gemini keys, but runtime can still fall back
to a platform env key when the tenant has none. That weakens cost attribution
and commercial isolation. Tenants need AI Settings to enter, test, rotate, and
remove their own key without exposing plaintext to the browser.

## Scope

### In

- AI Settings UI: enter/replace/delete key, masked last4, validation status,
  last-tested timestamp, test-connection action
- Server-side Gemini connectivity test (bounded, non-secret errors)
- Encrypted-at-rest key storage (reuse S43 secretbox); metadata only on APIs
- Production fail-closed when no validated tenant key
- Dev/test env fallback only when explicitly allowed and labeled non-production
- Audit events for create/replace/delete/test (no plaintext)
- Readiness/posture signal: env fallback disabled for production tenant runtime

### Out

- Returning Gemini keys to browser or platform-admin read APIs
- Platform-funded shared Gemini pool for production tenant traffic
- Multi-provider keys, billing plan redesign, KM translation

## Acceptance criteria

1. Tenant admin can enter a Gemini key and run **Test connection**.
2. Valid key is marked usable with masked metadata and `last_validated_at`.
3. Invalid key shows a safe failure and is not used for runtime.
4. In production mode, tenant chat/voice never uses env `GEMINI_API_KEY` when
   tenant key is missing or invalid.
5. Key create/replace/delete/test are audited without secret values.

## Dependencies

- packages: `internal/store` tenant_ai, `internal/secretbox`, `internal/gemini`,
  `apps/tenant-web` AI settings
- design: [DES-0056](../02-design/56-tenant-owned-gemini-key-enforcement-spec.md)
- prior: [DES-0039](../02-design/39-tenant-ai-config-extensibility-spec.md)
