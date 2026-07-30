---
id: DES-0058
title: Referral Code Redemption Specification
status: approved
updated: 2026-07-30
sprint: SPRINT-062
owner: SA
feature: FEAT-0054
release_target: v2.35.0
---

# DES-0058 — Referral Code Redemption (Bonus Quota)

**Sprint:** SPRINT-062 · **Release target:** v2.35.0  
**Feature:** [FEAT-0054](../01-features/FEAT-0054-referral-code-redemption.md)  
**Tasks:** TASK-0224–0227  
**Prior:** SPRINT-046 referral attribution + bonus ledger

## 1. Goals

1. Tenant admin redeems a referral **code** for bounded bonus quota.
2. Grants are idempotent, audited, and isolated per tenant.
3. Usage UI shows base + referral bonus + expiry.
4. Platform can reverse a grant.

## 2. Non-goals

- Cash affiliate payouts
- Mutating purchased package entitlements
- Unlimited concurrency without capacity policy

## 3. Concepts

| Term | Meaning |
| --- | --- |
| Referral code | Existing S46 tenant-scoped public code (referrer) |
| Redeemer | Active tenant applying someone else's code |
| Grant | Bonus ledger entry(s) for redeemer, source=`referral_redeem` |

## 4. Rules

1. Redeemer cannot redeem **own** code (self-referral).
2. Code must belong to an eligible referrer (active tenant; optional KYC/package rules align S46).
3. One successful redemption per redeemer per code (unique constraint).
4. Optional campaign caps: max redemptions per code / per period (config).
5. Retry after success returns same grant (idempotent).
6. Bonus dimensions/amounts from platform config (default table) or campaign rules.
7. Expiry on bonus rows (e.g. 30/90 days) from S46 patterns.

## 5. Data model

### 5.1 `referral_redemptions` (new)

| Column | Type | Notes |
| --- | --- | --- |
| `id` | uuid PK | |
| `redeemer_tenant_id` | uuid FK | auth tenant |
| `referrer_tenant_id` | uuid FK | code owner |
| `referral_code` | text | denormalized |
| `status` | text | `applied` \| `reversed` |
| `idempotency_key` | text null | optional client key |
| `bonus_grant_ids` | jsonb/uuid[] | ledger refs |
| `applied_at` | timestamptz | |
| `reversed_at` | timestamptz null | |
| `created_at` `updated_at` `created_by` `updated_by` | audit | |

Unique: `(redeemer_tenant_id, referral_code)` where status=applied (partial unique).

### 5.2 Bonus ledger

Reuse S46 bonus quota ledger tables/APIs; source tag `referral_redeem:{redemption_id}`.

## 6. API

### Tenant

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/api/tenant/referrals/validate` | Dry-run eligibility |
| POST | `/api/tenant/referrals/redeem` | Apply code |
| GET | `/api/tenant/referrals/redemptions` | List own redemptions |

#### Redeem request

```json
{ "code": "REFABC", "idempotency_key": "optional-uuid" }
```

#### Redeem success

```json
{
  "redemption_id": "…",
  "status": "applied",
  "bonus": [
    { "dimension": "call_minutes", "amount": 100, "expires_at": "2026-10-30T00:00:00Z" }
  ]
}
```

#### Error codes

| HTTP | code | Meaning |
| --- | --- | --- |
| 400 | `validation_error` | empty code |
| 404 | `referral_not_found` | unknown code |
| 409 | `already_redeemed` | redeemer already applied this code |
| 409 | `self_referral` | own code |
| 422 | `referral_ineligible` | expired/disabled/cap |
| 429 | `rate_limited` | abuse |

### Platform

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/api/platform/referrals/redemptions` | Filter by tenant/code |
| POST | `/api/platform/referrals/redemptions/{id}/reverse` | Reverse grant |

## 7. RBAC

- Redeem: `tenant_admin` active tenant  
- Reverse: `platform_admin`  
- No cross-tenant redeemer spoofing

## 8. Verification

```bash
# apply
curl -X POST /api/tenant/referrals/redeem -d '{"code":"REFABC"}' -H "Authorization: Bearer $T"
# retry same → same redemption_id
# self code → 409 self_referral
```

## See also

Workflow §137–139 · ER Sprint 62 · API Sprint 62 · UX T62/A62
