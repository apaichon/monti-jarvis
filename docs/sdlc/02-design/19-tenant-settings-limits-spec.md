---
id: DES-0019
title: Tenant Settings, Locale, and Call Limits Specification
status: approved
updated: 2026-07-12
sprint: SPRINT-016
owner: SA
---

# Tenant Settings & Call Limits — Design Spec

**Sprint:** SPRINT-016 · **Release target:** v1.7.0  
**Feature:** [FEAT-0016](../01-features/FEAT-0016-tenant-settings-locale-limits.md)  
**Depends on:** [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md), tenant active (S6–7)

## 1. Goals

1. Tenant self-service **settings** (locale, timezone, display name).  
2. Tenant **usage vs package limits** (read-only, same data as platform S13 snapshot).  
3. **Operational voice caps**: per-call and per-day minutes under package monthly ceiling.  
4. Scaffold **tier/group labels** for ops (full product S18).

## 2. Non-goals

| Out | Sprint |
| --- | --- |
| Customer tiers product | S18 |
| Preview sandbox | S17 |
| Overage billing | Future |
| Full portal i18n | Future |
| Per-caller quotas | S19+ |

## 3. Data model

### `tenant_settings`

| Column | Type | Notes |
| --- | --- | --- |
| tenant_id | text PK | FK tenants |
| locale | text | `en` \| `th` |
| timezone | text | IANA e.g. `Asia/Bangkok` |
| display_name | text | optional workspace label |
| ai_reply_locale | text | empty = auto; else en/th |
| user_tier_label | text | scaffold |
| user_group_label | text | scaffold |
| audit | | created_at, updated_at, created_by, updated_by |

### `tenant_call_limits`

| Column | Type | Notes |
| --- | --- | --- |
| tenant_id | text PK | FK tenants |
| max_minutes_per_call | int | 0 = unset |
| max_call_minutes_per_day | int | 0 = unset |
| audit | | |

## 4. Redis

| Key | Value | TTL |
| --- | --- | --- |
| `{prefix}call_daily:{tenant}:{YYYYMMDD}` | minutes used (int) | ~48h |
| Day boundary | Tenant `timezone` calendar day | — |
| S13 monthly | unchanged `minutes:{YYYYMM}` | — |

## 5. API summary

| Method | Path | Auth |
| --- | --- | --- |
| GET/PUT | `/api/tenant/settings` | tenant_admin active |
| GET | `/api/tenant/usage` | tenant_admin active |
| GET/PUT | `/api/tenant/call-limits` | tenant_admin active |

### PUT call-limits body

```json
{ "max_minutes_per_call": 15, "max_call_minutes_per_day": 120 }
```

### GET usage 200

Reuse S13 snapshot shape for the JWT tenant (package, period, limits, usage).

## 6. Enforcement

```text
Voice open:
  S13 voice_enabled + concurrent + monthly remaining
  + S16 daily remaining (if max_call_minutes_per_day > 0)

Voice in progress:
  if max_minutes_per_call > 0 and elapsed >= cap → end call / signal client

Voice end:
  Add elapsed to S13 monthly + S16 daily
```

## 7. AI locale hint

When building chat/voice system prompt, if `ai_reply_locale` or `locale` is set:

```text
Prefer replying in {lang} when the caller's language is unclear.
```

## 8. UX

Tenant route **`/tenant/settings`** (T9). See [05-ux-ui.md](05-ux-ui.md).

## 9. Verification

```bash
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/settings | jq .
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/usage | jq .
```

## 10. Related

| Artifact | Path |
| --- | --- |
| Feature | [FEAT-0016](../01-features/FEAT-0016-tenant-settings-locale-limits.md) |
| Sprint | [SPRINT-016](../03-sprints/SPRINT-016.md) |
| S13 quota | [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md) |
