---
id: FEAT-0043
title: "Tenant self-service avatar library with package active cap"
status: shipped
roadmap_sprint: 52
priority: D+
depends_on: [SPRINT-005, SPRINT-013, SPRINT-015, SPRINT-016, SPRINT-045, SPRINT-050]
design: DES-0047
design_spec: ../02-design/47-tenant-avatar-create-active-cap-spec.md
release: v2.24.0
updated: 2026-07-28
---

# FEAT-0043: Tenant Avatar Create & Active Cap

## Purpose

Let tenant admins **create and manage their own AI avatars** without a hard
create-count limit. Only **active** avatars count against the package
`max_ai_employees` limit (plus valid bonus quota).

## Problem

Platform admins own the global avatar catalog and assign agents to tenants.
Package `max_ai_employees` already caps **active** assignments, but tenants
cannot build a private library of drafts/personas and activate only what they
need. Sales and onboarding need self-service avatar creation without relaxing
the commercial **active** cap.

## Scope

### In

- Tenant-admin create / update / list / archive of **tenant-owned** avatars
- Library vs active: create defaults to **inactive** (not workforce-selectable)
- Activate / deactivate with hard cap from effective entitlement
  `max_ai_employees` (+ S46 bonus when present)
- Create is **not** limited by package avatar count
- Storage quota (KM/bytes) does not redefine avatar **count** policy
- Cap meter in tenant UI (active / limit / remaining)
- Workforce / customer / embed / mobile only expose **active** avatars
- Tenant isolation on all read/write paths

### Out

- HeyGen or third-party live avatar generation
- Cross-tenant avatar marketplace
- Removing platform catalog or platform assign paths
- Unlimited concurrent voice or changing S51 plan matrix
- Changing storage billing rules beyond clarifying independence from avatar count

## Acceptance criteria

1. Tenant admin can create more avatars than `max_ai_employees` without error
   (auth + field validation only).
2. Activating succeeds only while
   `active_count < effective max_ai_employees` (package + valid bonus).
3. At cap, activate returns stable `409` / `quota_exceeded` (or equivalent);
   no half-active state.
4. Deactivate frees a slot so another library avatar can activate.
5. Workforce and customer selection surfaces only **active** avatars for the
   tenant (plus existing platform-catalog assignment rules).
6. Tenant A cannot list, edit, or activate tenant B’s avatars.
7. Storage quota exhaustion is handled on portrait upload if applicable; it
   must not silently reuse the avatar **count** cap messaging or block create
   solely because “avatar count” is high.

## Test notes

- Functional: over-create, activate to cap, activate over cap, deactivate,
  reactivate
- Isolation: two tenants with private libraries
- Workforce: inactive avatars never appear in `GET /api/workforce` for that
  tenant context
- Bonus: when S46 AI-employee bonus exists, active cap uses base + remaining
  bonus as already resolved by quota service

## Dependencies

- packages: `internal/store` avatars, `internal/quota` CheckAIEmployees,
  `internal/entitlements`, `cmd/server`, `apps/tenant-web`
- design: [DES-0047](../02-design/47-tenant-avatar-create-active-cap-spec.md),
  [10-avatars-spec.md](../02-design/10-avatars-spec.md)
