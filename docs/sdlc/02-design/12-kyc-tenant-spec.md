---
id: DES-0012
sprint: SPRINT-007
status: review_pending
updated: 2026-07-09
owner: SA
---

# 12 — Platform KYC Tenant Review (Sprint 7)

> **Status:** `review_pending` — run **`sprint-tech-specs`** to complete before TASK-0031.

## Purpose

Platform admins review tenant-submitted KYC packages (Sprint 6) and approve or reject registrations, transitioning `pending_kyc` tenants to `active`.

## Depends on

- [11-tenant-register-spec.md](11-tenant-register-spec.md) — registration + tenant KYC submit APIs
- `tenant_kyc_profiles`, `tenant_registrations`, MinIO `kyc/` prefix

## Planned API surface (draft)

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `GET` | `/api/platform/tenants/{tenant_id}/kyc` | `platform_admin` | Full KYC package + asset URLs |
| `POST` | `/api/platform/tenants/{tenant_id}/kyc/approve` | `platform_admin` | Approve → tenant `active` |
| `POST` | `/api/platform/tenants/{tenant_id}/kyc/reject` | `platform_admin` | Reject with `{reason}` |

## Planned UI

- **P12** — `/admin/tenants/{id}/kyc` review screen (see [05-ux-ui.md](05-ux-ui.md) — to be added in sprint-tech-specs)
- Link from P11 tenants list row

## Out of scope

- HeyGen / LiveAvatar
- Auto package assign on approve (Sprint 9)

## Links

- Feature: [FEAT-0007](../01-features/FEAT-0007-kyc-tenant.md)
- Sprint: [SPRINT-007](../03-sprints/SPRINT-007.md)
- Workflow: [02-workflow.md](02-workflow.md) §22–24 *(pending)*
- API: [04-api-spec.md](04-api-spec.md) § Platform KYC *(pending)*