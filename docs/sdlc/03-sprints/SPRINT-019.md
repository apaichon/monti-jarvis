---
id: SPRINT-019
status: completed
start: 2026-07-12
end: 2026-07-13
closed: 2026-07-13
updated: 2026-07-13
design_pack: shipped
release_target: v2.0.0
release: v2.0.0
goal: "Tenant: import and manage customer accounts, domain rules, and integration-ready identities before customer authentication."
roadmap_sprint: 19
platform: Tenant
depends_on: [SPRINT-003, SPRINT-018]
---

# SPRINT-019 — Tenant: Customer Account Import and Integration

## Goal

Let active tenants import and manage customer identity records, assign tiers/groups, and define email-domain defaults before SPRINT-020 adds end-customer authentication.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S16–S18) | 16, 16, 16 → **avg 16** |
| **Commitment** | **16** |
| **Completed** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0087](../04-tasks/TASK-0087.md) | 3 | completed | devops | Customer, import-job, domain-rule, and group-membership schema |
| [TASK-0088](../04-tasks/TASK-0088.md) | 5 | completed | dev | Tenant customer, CSV import, and domain-rule APIs |
| [TASK-0089](../04-tasks/TASK-0089.md) | 4 | completed | dev | Tenant `/tenant/customers` management and import UI |
| [TASK-0090](../04-tasks/TASK-0090.md) | 3 | completed | dev | Tier/group binding and idempotent integration contracts |
| [TASK-0091](../04-tasks/TASK-0091.md) | 1 | completed | tester | Automated smoke coverage and signed two-tenant UAT |

**Committed:** 16 points · **Completed:** 16 points · **Completion:** 100%

## Shipped summary (v2.0.0)

| Area | Outcome |
| --- | --- |
| Data | Tenant-isolated customers, import jobs, group membership, and domain rules |
| APIs | Customer CRUD, atomic CSV dry-run/commit, import status, and domain-rule management |
| UI | `/tenant/customers` search, edit, deactivate, CSV import, and domain defaults |
| Integration | Idempotent `(tenant_id, source, external_id)` upsert with tier/group mapping |
| Verification | Two-tenant UAT, full Go tests/builds, infrastructure and runtime smoke green |

## Scope boundary

**In**

- Tenant-scoped customer records without credentials.
- CSV import with dry-run validation and row-level outcome counts.
- Customer tier and group assignment using SPRINT-018 catalogs.
- Email-domain rules that provide default tier/group and allow/deny intent for SPRINT-020.
- Idempotent integration identity through `(tenant_id, source, external_id)`.
- Tenant UI for search, create/edit/deactivate, import, and domain-rule management.

**Out**

- Customer login, passwords, OAuth, invitations, or JWT issuance (SPRINT-020).
- Dedicated Salesforce, HubSpot, LINE, or other vendor connector.
- Customer KYC, ticketing, discounts, or billing changes.
- Opening production customer traffic before the quota/rate-limit gate is signed off.

## Feature

- [FEAT-0021 — Customer Account Import and Domain Integration](../01-features/FEAT-0021-customer-account-import.md)

## Design pack

| Artifact | Path | Status |
| --- | --- | --- |
| Deep spec | [22-customer-account-import-spec.md](../02-design/22-customer-account-import-spec.md) | `shipped` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §55–58 | `shipped` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 19 | `shipped` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Customer Accounts & Imports | `shipped` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § T12 | `shipped` |

> **Gate:** Approved by user on 2026-07-12; implementation may proceed.

## Verification

```bash
make test && make build
# Tenant admin → /tenant/customers → dry-run CSV → import → assign tier/group
# Cross-tenant customer/domain IDs return 404
# Customer login remains unavailable until SPRINT-020
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-019-manual.md
```

## Risks

| Risk | Mitigation |
| --- | --- |
| PII leaks across tenants | Tenant id only from JWT; cross-tenant reads return 404 |
| Duplicate imports | Unique source/external id and normalized email; idempotent upsert policy |
| Bad CSV mutates partial data | Dry-run first; import transaction with row outcome summary |
| Domain rules imply auth too early | Persist policy only; enforcement is explicitly SPRINT-020 |
| Tier/group deletion breaks customers | Restrict delete while referenced or use inactive state |

## Production launch gate

SPRINT-019 prepares identity data but does not authorize customer production traffic. After SPRINT-020 authentication, verify rate limits, quotas, tier overrides, and tenant isolation under multi-user load.

## Links

- Depends: [SPRINT-003](SPRINT-003.md), [SPRINT-018](SPRINT-018.md)
- Next: SPRINT-020 Customer Auth
- Target: **v2.0.0**
