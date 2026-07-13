---
id: DES-0022
title: Customer Account Import and Domain Integration Specification
status: shipped
updated: 2026-07-13
sprint: SPRINT-019
owner: SA
---

# Customer Account Import and Domain Integration — Design Spec

**Sprint:** SPRINT-019 · **Release:** v2.0.0  
**Feature:** [FEAT-0021](../01-features/FEAT-0021-customer-account-import.md)  
**Depends on:** [06-auth-spec.md](06-auth-spec.md), [21-customer-tier-spec.md](21-customer-tier-spec.md)

## 1. Goals

1. Give each tenant an isolated customer directory before customer authentication exists.
2. Support safe manual creation and repeatable CSV imports with dry-run validation.
3. Bind customers to SPRINT-018 tiers and groups.
4. Store domain policy/defaults for SPRINT-020 registration and login enforcement.
5. Provide deterministic integration identity through `source` plus `external_id`.

## 2. Non-goals

- Passwords, OAuth identities, invitations, verification, JWT issuance, or customer sessions.
- Vendor-specific CRM connectors or outbound webhook delivery.
- Customer KYC, discounts, tickets, and customer-facing account screens.
- Production customer traffic or changes to the public no-auth conversation portal.

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `CUSTOMER_IMPORT_MAX_BYTES` | `2097152` | Maximum CSV upload size (2 MiB) |
| `CUSTOMER_IMPORT_MAX_ROWS` | `5000` | Maximum data rows per import |

No new Redis, ClickHouse, MinIO, or third-party service is required. CSV bytes are parsed in memory and are not retained after the request.

## 4. Data model — Postgres `monti_jarvis.callcenter`

### `customers`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `cust_{ulid}` |
| `tenant_id` | text FK | Required tenant isolation |
| `email` | text | Original trimmed value |
| `email_normalized` | text | Lower-case lookup key; nullable when no email |
| `phone` | text | Optional E.164-compatible text |
| `display_name` | text | Required human-readable name |
| `locale` | text | Empty, `en`, or `th` |
| `tier_id` | text FK | Optional SPRINT-018 tier |
| `source` | text | `manual`, `csv`, `api`, or future connector slug |
| `external_id` | text | Optional source-owned stable identifier |
| `status` | text | `active` or `inactive` |
| `metadata` | jsonb | Bounded integration attributes; no credentials |
| audit columns | timestamptz/text | `created_at`, `updated_at`, `created_by`, `updated_by` |

Indexes:

- unique `(tenant_id, email_normalized)` where normalized email is not null;
- unique `(tenant_id, source, external_id)` where external id is not null/empty;
- list index `(tenant_id, status, updated_at desc)`.

### `customer_group_members`

| Column | Type | Notes |
| --- | --- | --- |
| `customer_id` | text FK | Customer |
| `group_id` | text FK | SPRINT-018 group |
| `tenant_id` | text FK | Enables isolation checks/indexing |
| audit columns | timestamptz/text | Standard audit contract |

Primary key: `(customer_id, group_id)`. Service validation requires customer and group to share the JWT tenant.

### `customer_import_jobs`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `cimp_{ulid}` |
| `tenant_id` | text FK | Import owner |
| `filename` | text | Display only; no object storage |
| `mode` | text | `dry_run` or `commit` |
| `status` | text | `validating`, `validated`, `completed`, `failed` |
| `total_rows` | int | Parsed data rows |
| `created_rows` | int | Commit count |
| `updated_rows` | int | Commit count |
| `rejected_rows` | int | Validation failures |
| `errors` | jsonb | Max first 100 row errors |
| audit columns | timestamptz/text | Standard audit contract |

### `customer_domain_rules`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `cdr_{ulid}` |
| `tenant_id` | text FK | Rule owner |
| `domain` | text | Lower-case punycode/ASCII domain; unique per tenant |
| `policy` | text | `allow` or `deny`; enforcement deferred to SPRINT-020 |
| `default_tier_id` | text FK | Optional tenant tier |
| `default_group_id` | text FK | Optional tenant group |
| `active` | boolean | Default true |
| audit columns | timestamptz/text | Standard audit contract |

Schema is added through the repository's idempotent `internal/store` ensure-schema pattern. If a numbered SQL migration is introduced during TASK-0087, record its filename in [03-er-diagram.md](03-er-diagram.md).

## 5. Import and precedence rules

CSV required columns: `display_name` plus at least one of `email` or `external_id`. Optional columns: `phone`, `locale`, `tier_slug`, `group_slugs`, `source`, `external_id`.

```text
explicit CSV/API tier or groups
    > matching active domain defaults
    > no assignment

idempotency match
    1. tenant + source + external_id, when supplied
    2. tenant + normalized email for manual/csv rows
```

Dry-run performs every parse, normalization, reference, duplicate, and tenant check but creates no customer/group membership. Commit processes valid rows in one database transaction; invalid rows are excluded and summarized. A parser-level failure aborts the whole request.

## 6. Integration event contract

The service records/logs a stable event shape; external delivery is deferred.

```json
{
  "type": "customer.upserted",
  "tenant_id": "tenant_123",
  "customer_id": "cust_123",
  "source": "csv",
  "external_id": "crm-42",
  "occurred_at": "2026-07-12T12:00:00Z"
}
```

If NATS emission is added, use `monti.customer.upserted` and treat publish failure as non-fatal after the Postgres transaction. No Redis key is added.

## 7. API summary

See [04-api-spec.md](04-api-spec.md#customer-accounts--imports-sprint-19).

| Method | Path | Role |
| --- | --- | --- |
| `GET`, `POST` | `/api/tenant/customers` | active `tenant_admin` |
| `GET`, `PUT`, `DELETE` | `/api/tenant/customers/{id}` | active `tenant_admin` |
| `POST` | `/api/tenant/customer-imports` | active `tenant_admin` |
| `GET` | `/api/tenant/customer-imports/{id}` | active `tenant_admin` |
| `GET`, `POST` | `/api/tenant/customer-domain-rules` | active `tenant_admin` |
| `PUT`, `DELETE` | `/api/tenant/customer-domain-rules/{id}` | active `tenant_admin` |

`DELETE /customers/{id}` means deactivate; `DELETE /customer-domain-rules/{id}` hard-deletes an unneeded configuration rule.

## 8. RBAC

| Action | `platform_admin` | active `tenant_admin` | `customer` / public |
| --- | ---: | ---: | ---: |
| List/manage tenant customers | no | yes | no |
| Import CSV | no | yes | no |
| Manage domain rules | no | yes | no |
| Issue credentials/session | no | no | no |

Tenant id is always read from JWT. `AUTH_DISABLED=true` does not bypass these tenant-admin routes. Missing, inactive, or wrong-role actors receive 401/403; cross-tenant ids receive 404.

## 9. Verification

```bash
make test && make build

curl -sS -H "Authorization: Bearer $TOKEN" \
  http://localhost:8091/api/tenant/customers | jq .

curl -sS -X POST -H "Authorization: Bearer $TOKEN" \
  -F dry_run=true -F file=@customers.csv \
  http://localhost:8091/api/tenant/customer-imports | jq .
```

## 10. Related artifacts

| Artifact | Link |
| --- | --- |
| Sprint | [SPRINT-019](../03-sprints/SPRINT-019.md) |
| Feature | [FEAT-0021](../01-features/FEAT-0021-customer-account-import.md) |
| Workflow | [02-workflow.md](02-workflow.md) §55–58 |
| ER | [03-er-diagram.md](03-er-diagram.md) § Sprint 19 |
| API | [04-api-spec.md](04-api-spec.md) § Customer Accounts & Imports |
| UX | [05-ux-ui.md](05-ux-ui.md) § T12 |

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | | ☐ |
| Dev | | | ☐ |

## Release status

Shipped in v2.0.0 on 2026-07-13. Two-tenant UAT and automated regression gates passed; customer authentication and domain policy enforcement remain SPRINT-020 scope.
