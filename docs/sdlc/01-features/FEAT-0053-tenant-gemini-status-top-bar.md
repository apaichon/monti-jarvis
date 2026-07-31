---
id: FEAT-0053
title: "Tenant UX: remove system performance page; Gemini status in top bar"
status: completed
release: v2.35.0
roadmap_sprint: 61
priority: Q
depends_on: [SPRINT-026, SPRINT-029, SPRINT-043, SPRINT-060]
design: DES-0057
design_spec: ../02-design/57-tenant-gemini-status-top-bar-spec.md
updated: 2026-07-31
---

# FEAT-0053: Tenant Gemini Status in Top Bar

## Purpose

Simplify the tenant portal: remove the operational System Performance page from
normal tenant navigation and surface a compact Gemini readiness signal in the
top bar, linking to AI Settings when action is needed.

## Problem

S26 system performance is too operational for most tenant admins. Tenants care
whether Gemini is configured and reachable; deeper infra health belongs to
platform admin (S29).

## Scope

### In

- Remove System Performance from tenant nav/route discovery (keep API only if
  support needs it; document deprecation for tenant UI)
- Top-bar Gemini status: ready | key missing | validation failed | degraded
- Click-through to AI Settings when remediation needed
- Reuse S60/S43 key validation metadata (no second status model)
- Preserve platform-admin system performance dashboards

### Out

- Removing platform-admin performance monitoring
- Showing Redis/Postgres/MinIO/NATS/ClickHouse/LiveKit internals to tenants
- New key-management behavior owned by S60
- Customer call-page infrastructure status redesign

## Acceptance criteria

1. Tenant admin no longer sees System Performance in normal nav.
2. Top bar shows Gemini status without infra hostnames/raw errors.
3. Missing/invalid key status links to AI Settings.
4. Platform admin retains full system performance diagnostics.
5. S26 tenant performance API is either support-only or explicitly deprecated
   in design.

## Dependencies

- design: [DES-0057](../02-design/57-tenant-gemini-status-top-bar-spec.md)
- prior: [DES-0029](../02-design/29-tenant-system-performance-spec.md), DES-0056
