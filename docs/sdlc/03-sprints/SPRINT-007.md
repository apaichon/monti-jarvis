---
id: SPRINT-007
status: completed
start: 2026-07-09
end: 2026-07-22
closed: 2026-07-09
updated: 2026-07-09
release: v0.8.0
goal: "Platform Admin: KYC Tenant — review submitted packages and approve or reject pending_kyc tenants."
roadmap_sprint: 7
platform: Platform Admin
depends_on: [SPRINT-006]
release_target: v0.8.0
---

# SPRINT-007 — Platform Admin: KYC Tenant

## Goal

Let **platform admins** review tenant KYC packages submitted in Sprint 6, **approve** (`pending_kyc` → `active`) or **reject** with a reason, and surface the review queue in **`/admin/tenants/{id}/kyc`**. Approved tenants unlock KM writes and full tenant capabilities.

## Context from Sprint 6 (already shipped)

- `tenant_kyc_profiles` — contact, photo, documents; tenant submits via `POST /api/tenant/kyc/submit`
- `tenant_registrations.status` — `submitted` | `approved` | `rejected` (schema ready)
- `GET /api/platform/tenants` + `/admin/tenants` list with `pending_kyc` filter
- KYC assets in MinIO under `kyc/{tenant_id}/`; served at `/api/assets/kyc/...`
- `pending_kyc` tenants: login OK; `POST /api/km/*` writes blocked until approval

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0030 | 3 | completed | devops | KYC review schema — `tenant_kyc_profiles` review columns + registration reviewer fields |
| TASK-0031 | 5 | completed | dev | Platform KYC APIs — `GET /api/platform/tenants/{id}/kyc`, approve, reject |
| TASK-0032 | 3 | completed | dev | Lifecycle — atomic approve/reject; `tenants.status` transition; KM write unblock |
| TASK-0033 | 2 | completed | dev | Resend email to tenant admin on approve/reject |
| TASK-0034 | 3 | completed | dev | Platform admin UI — `/admin/tenants/{id}/kyc` review screen |

**Committed:** 16 points · **Completed:** 16 points · **Velocity:** 16

## Shipped (v0.8.0)

- Schema: `tenant_kyc_profiles` review columns (`reviewed_at`, `reviewed_by`, `rejection_reason`); status `approved` | `rejected`
- Platform APIs: `GET /api/platform/tenants/{id}/kyc`, `POST .../approve`, `POST .../reject` (`platform_admin` only)
- Lifecycle: single Postgres transaction on approve (`pending_kyc` → `active`, registration + profile `approved`); reject keeps tenant `pending_kyc`
- KM writes unblock for `tenant_admin` after approve (`RequireKMWrite` passes on `active` tenant)
- `GET /api/platform/tenants?kyc_status=submitted` filter; tenants list KYC column + link to review
- Platform admin UI: `/admin/tenants/{id}/kyc` — contact, photo, documents, approve/reject with feedback dialog
- Resend decision emails to tenant admin on approve/reject (no-op when `RESEND_API_KEY` unset)
- Tenant resubmit: rejected KYC resets to `submitted` on backoffice re-submit
- E2E: `e2e/tests/platform-kyc.spec.ts`

## Scope boundary

**In**
- Platform operator reviews KYC package (contact, photo, documents) for a `pending_kyc` tenant
- Approve: `tenants.status` → `active`, `tenant_registrations` → `approved`, `tenant_kyc_profiles` → `approved`, reviewer + timestamp
- Reject: `tenant_registrations` → `rejected`, optional `rejection_reason`; tenant stays `pending_kyc` (can resubmit in follow-up)
- `GET /api/platform/tenants` optional filter `?kyc_status=submitted` (join `tenant_kyc_profiles`)
- Platform admin review UI with asset preview and shadcn-style feedback dialog
- Resend notification email to tenant admin email on decision
- `sprint-tech-specs` design pack **before** TASK-0031 implementation

**Out** (→ backlog / later sprints)
- Auto-assign Starter package on approve (Sprint 9)
- Payment gateway, billing (Sprints 8–12)
- Tenant self-service resubmit after reject (stretch — document manual workaround)
- HeyGen / LiveAvatar avatars (not used)
- Cross-tenant audit log table (Sprint 27)
- CAPTCHA / fraud scoring

## Feature

- [FEAT-0007 — Platform KYC tenant review](../01-features/FEAT-0007-kyc-tenant.md)

## Design pack (`sprint-tech-specs`)

| Artifact | Path | Status |
| --- | --- | --- |
| KYC tenant deep spec | [12-kyc-tenant-spec.md](../02-design/12-kyc-tenant-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §22–24 | `approved` |
| ER diagram | [03-er-diagram.md](../02-design/03-er-diagram.md) | `approved` |
| API spec | [04-api-spec.md](../02-design/04-api-spec.md) § Platform KYC | `approved` |
| UX/UI ASCII | [05-ux-ui.md](../02-design/05-ux-ui.md) § P12, flows P-C/P-D | `approved` |

> Design pack approved — DEV may start **TASK-0030** (schema) then **TASK-0031**.

## Verification

```bash
make build && make test
make infra-init && make restart
# Tenant: register → verify email → submit KYC in /tenant/backoffice
open http://localhost:8091/admin/tenants?status=pending_kyc
# Platform admin: open tenant KYC → approve
curl -H "Authorization: Bearer $TENANT_TOKEN" -X POST http://localhost:8091/api/km/agents/ava/documents  # 403 before, 2xx after approve
```

- Manual: `docs/sdlc/06-manual-tests/SPRINT-007-manual.md` (Tester, at VERIFY)
- E2E: extend `e2e/tests/platform-tenants.spec.ts` or add `platform-kyc.spec.ts`

## Risks

| Risk | Mitigation |
| --- | --- |
| Approve without submitted KYC | API returns `409` unless `tenant_kyc_profiles.status = submitted` |
| Partial state on approve | Single Postgres transaction: tenant + registration + kyc profile |
| Asset access for reviewers | Platform admin serves KYC assets via existing `/api/assets/kyc/...` with auth gate |
| Reject without resubmit flow | Document operator message; tenant can edit backoffice and resubmit (reset to `submitted`) as stretch |

## Definition of done

- Design pack approved · code reviewed · ACs verified · `make build` · tag **v0.8.0** at sprint close ✅