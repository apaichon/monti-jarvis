# Feature: Tenant Settings, Locale, and Call Limits   (FEAT-0016)
**Sprint:** SPRINT-016   **Owner:** DEV   **Status:** shipped (v1.7.0)

## Production note

Before **customer production launch** (after tenant **customer-user auth** — S19–20), re-verify **rate limits + quota management** under multi-user load. S16 ships tenant operational caps; S13 package quotas remain the platform ceiling — both must be proven before open customer traffic.

## Problem

Package quotas (S13) are enforced platform-side, but **tenant admins** cannot:

1. Set workspace **locale / timezone** preferences  
2. See their **usage vs package limits** without platform admin  
3. Apply **tighter operational limits** (per-call minutes, daily call minutes) under the package ceiling  

S18 will add full customer tiers/groups; S16 ships the settings foundation and tenant-level call time controls.

## Scope

**In:**
- `tenant_settings` (locale, timezone, display preferences, default language for AI replies hint)
- `tenant_call_limits` (optional caps: `max_minutes_per_call`, `max_call_minutes_per_day` — must be ≤ package where applicable)
- Tenant APIs (active `tenant_admin`):
  - `GET/PUT /api/tenant/settings`
  - `GET /api/tenant/usage` (package limits + usage snapshot)
  - `GET/PUT /api/tenant/call-limits`
- Enforce per-call and daily minutes on **voice** open/end (Redis), fail-open consistent with S13
- Tenant UI `/tenant/settings` — locale, timezone, usage meters, call limits form
- Scaffold only: **user tier / group** labels as free-text tags on settings (no full RBAC matrix — S18)
- Design pack DES-0019 + manual UAT

**Out:**
- Full customer identity tiers (→ **SPRINT-018**)
- Monetary overage / auto-upgrade
- Changing platform package catalog
- Test & preview sandbox (→ **SPRINT-017**)
- Full tenant portal i18n of every string (locale preference stored; TH/EN labels for settings page primary)
- Per-end-customer quotas (needs S19+)

## Acceptance criteria

1. Active tenant admin opens `/tenant/settings` and saves locale (`th`|`en`) + timezone (IANA).
2. `GET /api/tenant/usage` returns package name, period, limits, and current usage (mirrors S13 snapshot for **own** tenant).
3. Tenant can set `max_minutes_per_call` and `max_call_minutes_per_day`; values `0` = use package default / unlimited within package.
4. Voice session denied or auto-ends when per-call cap exceeded; daily cap blocks new voice opens.
5. Limits cannot exceed package `max_monthly_call_minutes` semantics in a documented way (daily/call caps are soft operational ≤ remaining monthly).
6. Inactive tenant / non-admin → 401/403.
7. Manual UAT checklist under `06-manual-tests/SPRINT-016-manual.md`.

## Test notes

- Unit: limit clamp, Redis daily key, per-call duration check  
- Manual: set low per-call cap → start voice → exceed → end/deny  
- Embed chat text remains available when voice capped  

## Dependencies

- SPRINT-013 quota/rate limit  
- SPRINT-006/007 active tenant  
- SPRINT-015 tenant portal shell  

## Links

- Sprint: [SPRINT-016](../03-sprints/SPRINT-016.md)  
- Design: [19-tenant-settings-limits-spec.md](../02-design/19-tenant-settings-limits-spec.md)  
- Prior: [FEAT-0013](FEAT-0013-quota-rate-limit.md)  
