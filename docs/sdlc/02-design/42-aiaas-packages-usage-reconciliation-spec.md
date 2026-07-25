---
id: DES-0042
title: AiaaS Packages and Usage Reconciliation Specification
status: approved
updated: 2026-07-25
sprint: SPRINT-045
owner: SA
feature: FEAT-0039
---

# AiaaS Packages and Usage Reconciliation — Design Spec

**Sprint:** SPRINT-045 · **Release target:** v2.18.0  
**Feature:** [FEAT-0039](../01-features/FEAT-0039-aiaas-packages-usage-reconciliation.md)  
**Depends on:** [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md),
[28-call-center-statistics-spec.md](28-call-center-statistics-spec.md),
[30-mobile-call-api-sdk-spec.md](30-mobile-call-api-sdk-spec.md),
[34-platform-billing-quota-ai-cost-spec.md](34-platform-billing-quota-ai-cost-spec.md),
[39-tenant-ai-config-extensibility-spec.md](39-tenant-ai-config-extensibility-spec.md)

## 1. Goals

- Define four package tiers as catalog data and entitlement snapshots.
- Add storage bytes and mobile call minutes as first-class dimensions.
- Make quota decisions and usage events idempotent and tenant-scoped.
- Reconcile Postgres enforcement, Redis counters, ClickHouse facts, MinIO
  storage, billing orders, and tenant/platform projections without treating a
  missing source as zero.
- Preserve historical usage context when a tenant changes package.

## 2. Non-goals

No overage billing, automatic upgrade, quota pooling, enterprise custom plan,
new LLM provider, customer generative workspace, CLI process execution, or
replacement of the Gemini inbound call path.

## 3. Initialization package catalog

| Slug | Monthly price | AI avatars | KM docs | Storage | Concurrent calls | Mobile minutes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `aiaas-500` | ฿500 | 1 | 100 | 5 GB | 1 | 100 |
| `aiaas-1000` | ฿1,000 | 3 | 300 | 20 GB | 2 | 300 |
| `aiaas-1500` | ฿1,500 | 5 | 750 | 50 GB | 5 | 750 |
| `aiaas-2000` | ฿2,000 | 10 | 1,500 | 100 GB | 10 | 1,500 |

These four rows are initialization defaults, not immutable commercial policy.
The seed/migration must be idempotent and insert missing packages without
overwriting an existing platform-admin change. The implementation reads the
active values from `packages` and `package_limits.rules`; no Svelte component
hard-codes price or quota values.

After initialization, a platform admin can create, edit, activate, deactivate,
or archive package catalog entries through the existing package-management
surface (`GET/POST/PUT /api/platform/packages*`). Package slugs/IDs used by
existing entitlements remain stable identifiers. A catalog edit affects future
assignments; it does not rewrite an existing tenant's immutable entitlement
snapshot. Applying a changed package to a tenant is an explicit entitlement
assignment/upgrade action and creates a new validity interval.

## 4. Dimension contract

| Dimension | Unit | Authority | Enforcement source |
| --- | --- | --- | --- |
| `ai_employees` | active assignments | Postgres | entitlement + assignment transaction |
| `monthly_call_minutes` | minutes/month | Redis + Postgres snapshot | voice lifecycle finalization |
| `km_documents` | active documents | Postgres | KM create/delete transaction |
| `storage_bytes` | bytes/monthly allowance | MinIO + Postgres projection | successful object write/delete |
| `concurrent_calls` | active calls | Redis lease | acquire/release with TTL |
| `mobile_call_minutes` | minutes/month | Redis + Postgres snapshot | mobile call finalization |

Every response reports `dimension`, `unit`, `period`, `limit`, `consumed`,
`remaining`, `source`, and `freshness`. Enforcement and reporting must not
silently merge mobile minutes with web minutes; the approved commercial rule
must be stored in the package rule schema.

## 5. Data model

Existing `packages`, `package_limits`, and `tenant_entitlements` remain the
catalog and current entitlement authorities. The implementation proposes these
additional Postgres tables in `callcenter`:

### `tenant_entitlement_snapshots`

Immutable package/rules context for each validity interval. It includes
`tenant_id`, `package_id`, `rules_schema_id`, `rules_snapshot`, `valid_from`,
`valid_until`, `source_order_id`, and audit columns.

### `usage_events`

Append-only logical usage events with unique `(tenant_id, idempotency_key)`;
fields include `dimension`, `unit`, `amount`, `period_start`, `period_end`,
`source_type`, `source_id`, `entitlement_snapshot_id`, `state`, and audit
columns. A duplicate event returns the original logical result.

### `usage_reconciliation_runs`

Operator-controlled run metadata: `run_id`, date range, source watermarks,
status, mismatch count, correction count, safe error, and audit columns.
Raw provider bodies and credentials are never stored.

Redis keys use DB 4 and the existing prefix:

| Key | Purpose |
| --- | --- |
| `monti_jarvis:quota:{tenant}:mobile_minutes:{YYYYMM}` | Mobile minute counter |
| `monti_jarvis:quota:{tenant}:storage_bytes:{YYYYMM}` | Storage projection counter |
| `monti_jarvis:quota:{tenant}:concurrent` | Active-call lease set |
| `monti_jarvis:usage:event:{tenant}:{idempotency}` | Short-lived duplicate guard |

## 6. API summary

Existing contracts are extended rather than replaced:

| Method | Path | Role | Change |
| --- | --- | --- | --- |
| `GET` | `/api/entitlements/me` | tenant admin/platform admin | Include catalog price and dimensioned allowance |
| `GET` | `/api/tenant/usage` | tenant admin | Return dimension rows and freshness/source metadata |
| `GET` | `/api/platform/tenants/{tenant_id}/usage` | platform admin | Return current vs historical usage by dimension |
| `GET` | `/api/platform/billing/usage` | platform admin | Include package, usage, and reconciliation freshness |
| `GET` | `/api/mobile/v1/bootstrap` | authenticated customer | Return mobile allowance and dimension-specific remaining values |
| `POST` | `/api/platform/usage/reconcile` | platform admin | Start bounded, idempotent reconciliation run |
| `GET` | `/api/platform/usage/reconcile/{run_id}` | platform admin | Read run status, mismatch counts, and safe errors |

Stable errors: `no_entitlement`, `quota_exceeded`, `quota_unavailable`,
`usage_duplicate`, `usage_source_stale`, `reconciliation_in_progress`, and
`reconciliation_forbidden`.

## 7. RBAC and isolation

| Action | platform admin | tenant admin | customer | anonymous |
| --- | --- | --- | --- | --- |
| Read package catalog | yes | yes | no | no |
| Read own entitlement/usage | yes | own tenant | own customer response only | no |
| Reconcile usage | yes | no | no | no |
| Read another tenant | controlled support filter | no | no | no |
| Change package | fulfillment policy only | no direct mutation | no | no |

Tenant IDs are derived from the authenticated context. Query/body tenant IDs
are accepted only for platform-admin support routes and must be authorized.

## 8. Verification

```bash
make test
make build
go test ./internal/quota ./internal/metering ./cmd/server -count=1
# seed all four tiers and verify dimensioned entitlement responses
# run two-tenant fixtures for web, embed, mobile, KM, storage, and concurrency
# replay the same usage event and verify one logical result
# force disconnect, timeout, failed start, delete, and retry paths
# verify stale/unavailable sources are labeled, never rendered as zero
git diff --check
```

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | | ☐ |
| Dev | | | ☐ |
| Tester | | | ☐ |
