# Feature: Tenant Self-Registration   (FEAT-0006)
**Sprint:** SPRINT-006   **Owner:** DEV   **Status:** shipped   **Release:** v0.7.0

## Problem

Tenants today exist only via dev seeds (`demo`) or manual SQL. Phase C onboarding requires a **public registration path** so a business can sign up, create its tenant record, and obtain a **tenant_admin** login — without platform operator intervention. Sprint 7 (KYC) will gate full activation; this sprint establishes the registration pipeline and `pending_kyc` state.

## Scope

In:
- Postgres `tenant_registrations` audit trail + extend `tenants.status` with `pending_kyc`
- Minimal `brands` stub (one default brand per new tenant)
- Public `POST /api/public/tenant/register` — company name, slug, admin email, password, display name
- Atomic create: tenant + brand stub + user + `tenant_admin` role + registration row
- Issue JWT access + refresh for new tenant_admin (same auth service as Sprint 3)
- `apps/tenant-web` at `/tenant` — registration form + success / login redirect
- Platform admin: `GET /api/platform/tenants` list with status filter (visibility for KYC prep)
- Design pack via `sprint-tech-specs` before implementation

Out:
- KYC review / approve / reject workflow (Sprint 7)
- Email verification, SMTP, magic links
- Auto package entitlement on signup (Sprint 9)
- Full tenant admin dashboard (Sprint 15+)
- `brands` full CRUD, channels, locales
- Customer self-registration (Sprint 19)
- CAPTCHA (rate-limit stub only)

## Acceptance criteria

1. `ensureSchema` creates `tenant_registrations` and `brands`; `tenants.status` accepts `pending_kyc` | `active` | `suspended`.
2. `POST /api/public/tenant/register` validates slug (lowercase alphanumeric + hyphen, unique), email (unique), password (min length); returns `201` with `tenant_id`, `registration_id`, and auth tokens.
3. Duplicate slug or email returns `409` with clear error; invalid payload returns `400`.
4. New tenant is `pending_kyc`; tenant_admin can log in and call `GET /api/auth/me` with correct `tenant_id`.
5. `GET /api/platform/tenants` (`platform_admin`) lists tenants including `pending_kyc` with pagination/filter.
6. `/tenant/register` Svelte form submits to API; success shows confirmation and stores session (or redirects to tenant shell).
7. `AUTH_DISABLED=true` dev mode unchanged for customer portal `/`; registration API works regardless.
8. `go test ./...`; manual UAT in `docs/sdlc/06-manual-tests/SPRINT-006-manual.md`.

## Test notes

- API: slug collision, email collision, password policy, RBAC on platform list
- Browser: register → login → `/api/auth/me` shows new tenant
- Regression: platform admin `/admin`, customer `/`, demo tenant seeds intact

## Links

- Sprint: [SPRINT-006](../03-sprints/SPRINT-006.md)
- Design: [11-tenant-register-spec.md](../02-design/11-tenant-register-spec.md) · [04-api-spec.md](../02-design/04-api-spec.md) · [05-ux-ui.md](../02-design/05-ux-ui.md) § T1–T3
- Depends on: [FEAT-0003](FEAT-0003-auth-rbac.md)
- Roadmap: Sprint 6 · Phase C
- Shipped v0.7.0 · Next: [FEAT-0007](FEAT-0007-kyc-tenant.md) Sprint 7 KYC Tenant