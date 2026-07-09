# Feature: Platform KYC Tenant Review   (FEAT-0007)
**Sprint:** SPRINT-007   **Owner:** DEV   **Status:** shipped   **Release:** v0.8.0

## Problem

Sprint 6 lets tenants self-register and submit KYC evidence, but tenants remain `pending_kyc` indefinitely. Platform operators need a **review queue** and **approve/reject** actions that transition tenants to `active` and unlock product features (KM writes, commerce prep).

## Scope

In:
- Extend `tenant_kyc_profiles` with review metadata (`reviewed_at`, `reviewed_by`, `rejection_reason`; status `approved` | `rejected`)
- Extend `tenant_registrations` with reviewer fields on approve/reject
- `GET /api/platform/tenants/{tenant_id}/kyc` — full package for `platform_admin`
- `POST /api/platform/tenants/{tenant_id}/kyc/approve` — `pending_kyc` → `active`
- `POST /api/platform/tenants/{tenant_id}/kyc/reject` — `{reason}` required
- List filter `?kyc_status=submitted` on `GET /api/platform/tenants`
- Platform admin UI `/admin/tenants/{id}/kyc` — preview photo/docs, approve/reject
- Resend email to tenant admin on decision
- Design pack via `sprint-tech-specs` (`12-kyc-tenant-spec.md`)

Out:
- Auto package entitlement on approve (Sprint 9)
- Payment / billing (Sprints 8–12)
- Formal audit log table (Sprint 27)
- HeyGen / video avatars
- Customer-tier KYC (Sprint 19)

## Acceptance criteria

1. `ensureSchema` migrates `tenant_kyc_profiles.status` to include `approved` | `rejected`; adds review columns idempotently.
2. `GET /api/platform/tenants/{tenant_id}/kyc` returns tenant metadata + KYC profile + asset URLs; `403` for non-platform-admin; `404` unknown tenant.
3. `POST .../kyc/approve` requires `tenants.status = pending_kyc` and `tenant_kyc_profiles.status = submitted`; atomically sets tenant `active`, registration `approved`, profile `approved`; returns `200`.
4. `POST .../kyc/reject` requires submitted KYC; sets registration `rejected` + reason; tenant remains `pending_kyc`; returns `200`.
5. After approve, `tenant_admin` `POST /api/km/*` writes succeed (no longer `403 tenant not active`).
6. `/admin/tenants` row links to `/admin/tenants/{id}/kyc`; review page shows contact, photo, documents, approve/reject with feedback dialog.
7. Resend sends approve/reject email to registration `admin_email` when `RESEND_API_KEY` configured.
8. `go test ./...`; manual UAT in `docs/sdlc/06-manual-tests/SPRINT-007-manual.md`.

## Test notes

- API: RBAC, 409 guards, transaction integrity, KM unblock after approve
- Browser UAT: tenant submits KYC → platform admin approves → tenant KM upload works
- E2E: platform login → tenants list → approve flow

## Links

- Sprint: [SPRINT-007](../03-sprints/SPRINT-007.md)
- Depends on: [FEAT-0006](FEAT-0006-tenant-register.md)
- Design: [12-kyc-tenant-spec.md](../02-design/12-kyc-tenant-spec.md) · [04-api-spec.md](../02-design/04-api-spec.md) § Platform KYC · [05-ux-ui.md](../02-design/05-ux-ui.md) § P12 · [02-workflow.md](../02-design/02-workflow.md) §22–24
- Roadmap: Sprint 7 · Phase C
- Next: Sprint 8 Payment Gateway