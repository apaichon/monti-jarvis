---
id: DES-0021
title: Tenant Customer Tier Specification
status: shipped
updated: 2026-07-12
sprint: SPRINT-018
owner: SA
---

# Customer Tier — Design Spec

**Sprint:** SPRINT-018 · **Release target:** v1.9.0  
**Feature:** [FEAT-0020](../01-features/FEAT-0020-customer-tier.md)  
**Depends on:** [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md)

## 1. Goals

1. Tenant-defined **tier catalog** (VIP / Standard / Guest …).  
2. Optional **groups** for ops labels.  
3. Per-tier defaults: agent, AI locale, call-cap overrides.  
4. Preview can simulate a tier before customer auth exists.

## 2. Non-goals

| Out | Sprint |
| --- | --- |
| Customer accounts | S19–20 |
| Assign customer → tier | S19+ |
| Discounts / billing | Future |

## 3. Data model

### `customer_tiers`

| Column | Type | Notes |
| --- | --- | --- |
| id | text PK | `tier_{ulid}` |
| tenant_id | text | FK tenants |
| name | text | display |
| slug | text | unique per tenant |
| priority | int | higher = more VIP for future routing |
| description | text | |
| default_agent_id | text | optional |
| ai_reply_locale | text | `` \| en \| th |
| max_minutes_per_call | int | 0 = inherit |
| max_call_minutes_per_day | int | 0 = inherit |
| active | bool | |
| audit | | |

### `customer_groups`

| Column | Type | Notes |
| --- | --- | --- |
| id | text PK | `grp_{ulid}` |
| tenant_id | text | |
| name, slug, description | text | |
| audit | | |

## 4. API summary

| Method | Path | Auth |
| --- | --- | --- |
| GET/POST | `/api/tenant/tiers` | tenant_admin active |
| GET/PUT/DELETE | `/api/tenant/tiers/{id}` | tenant_admin active |
| GET/POST | `/api/tenant/groups` | tenant_admin active |
| GET/PUT/DELETE | `/api/tenant/groups/{id}` | tenant_admin active |

### POST tier body

```json
{
  "name": "VIP",
  "slug": "vip",
  "priority": 100,
  "description": "Priority support",
  "default_agent_id": "ava",
  "ai_reply_locale": "th",
  "max_minutes_per_call": 30,
  "max_call_minutes_per_day": 0,
  "active": true
}
```

## 5. Apply rules (S18)

```text
If request has tier_id (preview):
  locale = tier.ai_reply_locale if set else tenant settings
  max_per_call = tier.max_minutes_per_call if >0 else tenant_call_limits
  max_per_day  = tier.max_call_minutes_per_day if >0 else tenant_call_limits
Still clamp under package monthly where applicable.
```

Production embed without customer identity ignores tier until S19+.

## 6. UX

Route **`/tenant/tiers`** (T11). See [05-ux-ui.md](05-ux-ui.md).

## 7. Verification

```bash
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/tiers | jq .
```

## 8. Related

| Artifact | Path |
| --- | --- |
| Feature | [FEAT-0020](../01-features/FEAT-0020-customer-tier.md) |
| Sprint | [SPRINT-018](../03-sprints/SPRINT-018.md) |
| Settings | [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md) |
