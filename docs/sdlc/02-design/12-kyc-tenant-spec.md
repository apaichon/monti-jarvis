---
id: DES-0012
title: Platform KYC Tenant Review Specification
status: approved
updated: 2026-07-09
sprint: SPRINT-007
owner: SA
---

# Platform KYC Tenant Review — Design Spec

**Sprint:** SPRINT-007 · **Release target:** v0.8.0  
**Feature:** [FEAT-0007](../01-features/FEAT-0007-kyc-tenant.md)  
**Depends on:** [11-tenant-register-spec.md](11-tenant-register-spec.md) (v0.7.0 tenant submit flow)

## 1. Goals

- Platform operators **review** tenant-submitted KYC packages (contact, photo, documents).
- **Approve** transitions `pending_kyc` → `active` and unlocks KM writes.
- **Reject** records reason; notifies tenant admin via Resend when configured.
- Platform admin UI at `/admin/tenants/{id}/kyc` with standardized feedback dialogs.

## 2. Non-goals (Sprint 7)

- Auto-assign Starter package on approve (Sprint 9).
- Payment gateway / billing (Sprints 8–12).
- Formal cross-tenant audit log table (Sprint 27).
- HeyGen / LiveAvatar / video avatars.
- Authenticated asset proxy hardening beyond existing `/api/assets/kyc/...` (future security sprint).
- Automated tenant resubmit after reject (stretch — manual re-upload + operator guidance).

## 3. Environment

| Variable | Purpose |
| --- | --- |
| `RESEND_API_KEY` | Optional — approve/reject emails; no-op when unset |
| `RESEND_FROM_EMAIL` | Sender address for KYC decision emails |
| `APP_PUBLIC_URL` | Link to `/tenant/backoffice` in reject email |

## 4. Data model (Postgres `callcenter`)

### `tenant_kyc_profiles` (extend Sprint 6)

| Column | Type | Notes |
| --- | --- | --- |
| `tenant_id` | text PK FK → `tenants` | One profile per tenant |
| `contact_name` | text | |
| `contact_phone` | text | |
| `contact_address` | text | |
| `photo_object_key` | text | MinIO key under `kyc/{tenant_id}/` |
| `business_doc_keys` | jsonb | Array of MinIO object keys |
| `status` | text | `draft` \| `submitted` \| `approved` \| `rejected` |
| `submitted_at` | timestamptz | Set on tenant submit |
| `reviewed_at` | timestamptz | Set on approve/reject *(Sprint 7)* |
| `reviewed_by` | text | Platform admin user id *(Sprint 7)* |
| `rejection_reason` | text | Set on reject *(Sprint 7)* |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

**Migration (TASK-0030):** idempotent `ALTER` to widen `status` CHECK and add review columns.

### `tenant_registrations` (extend)

| Column | Type | Notes |
| --- | --- | --- |
| `rejection_reason` | text | Copied on reject *(Sprint 7)* |
| `reviewed_at` | timestamptz | *(Sprint 7)* |
| `reviewed_by` | text | *(Sprint 7)* |

`status` already allows `submitted` \| `approved` \| `rejected`.

### MinIO keys

```
monti-jarvis/kyc/{tenant_id}/photo.{ext}
monti-jarvis/kyc/{tenant_id}/docs/{filename}
```

Served at `GET /api/assets/kyc/{tenant_id}/{kind}/{file}`.

## 5. API summary

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `GET` | `/api/platform/tenants` | `platform_admin` | List; add `?kyc_status=submitted` |
| `GET` | `/api/platform/tenants/{tenant_id}/kyc` | `platform_admin` | Full review package |
| `POST` | `/api/platform/tenants/{tenant_id}/kyc/approve` | `platform_admin` | Activate tenant |
| `POST` | `/api/platform/tenants/{tenant_id}/kyc/reject` | `platform_admin` | Reject with `{reason}` |

Tenant-side APIs (shipped Sprint 6): `GET/PUT/POST /api/tenant/kyc*`.

Full contract: [04-api-spec.md](04-api-spec.md) § Platform KYC review.

## 6. RBAC & lifecycle

### Approve preconditions

| Check | Error |
| --- | --- |
| `tenants.status = pending_kyc` | `409` |
| `tenant_kyc_profiles.status = submitted` | `409` |
| Caller `platform_admin` | `403` |

### Approve effects (single transaction)

1. `tenants.status` → `active`
2. `tenant_registrations.status` → `approved`; `reviewed_at`, `reviewed_by`
3. `tenant_kyc_profiles.status` → `approved`; `reviewed_at`, `reviewed_by`

### Reject preconditions

Same as approve except tenant may already be `rejected` registration from prior attempt → `409` if KYC not `submitted`.

### Reject effects

1. `tenant_registrations.status` → `rejected`; `rejection_reason`, `reviewed_*`
2. `tenant_kyc_profiles.status` → `rejected`; `rejection_reason`, `reviewed_*`
3. `tenants.status` stays `pending_kyc`

### Post-approve

| Route | `tenant_admin` on `active` tenant |
| --- | --- |
| `POST /api/km/*` write | ✅ Allowed (`IsTenantActive` = true) |
| `GET /api/workforce` | ✅ Unchanged |

## 7. Email (Resend)

| Event | Recipient | Subject (draft) |
| --- | --- | --- |
| Approve | `tenant_registrations.admin_email` | Your Monti workspace is active |
| Reject | same | Action required — KYC review update |

Email send is **best-effort** after DB commit; failures logged, not rolled back.

## 8. NATS event (optional)

Subject: `monti.auth.tenant.kyc_decided`

```json
{
  "event": "tenant.kyc.approved",
  "tenant_id": "acme",
  "reviewed_by": "usr_platform_admin",
  "tenant_status": "active"
}
```

Publish when bus enabled; no-op otherwise (same pattern as registration).

## 9. Verification

```bash
make build && make test
make infra-init && make restart

# 1) Tenant submits KYC (Sprint 6 flow)
# Register + verify + login at /tenant/backoffice → Submit for KYC review

# 2) Platform review
PLATFORM=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)

curl -s -H "Authorization: Bearer $PLATFORM" \
  http://localhost:8091/api/platform/tenants/acme/kyc | jq .

curl -s -X POST -H "Authorization: Bearer $PLATFORM" \
  http://localhost:8091/api/platform/tenants/acme/kyc/approve | jq .

# 3) KM write unblocked
TENANT=$(curl -s -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@acme.test","password":"..."}' | jq -r .access_token)
curl -s -o /dev/null -w "%{http_code}" -X POST \
  -H "Authorization: Bearer $TENANT" \
  -F "file=@docs/samples/km/ava.md" \
  http://localhost:8091/api/km/agents/ava/documents
# expect 201 (not 403)
```

## 10. UI

- **P12** review screen: [05-ux-ui.md](05-ux-ui.md) § Screen P12
- **Flows P-C, P-D** — queue → approve/reject
- Feedback via `FeedbackDialog` (not inline status banners)

## Links

- Workflow: [02-workflow.md](02-workflow.md) §22–24
- ER: [03-er-diagram.md](03-er-diagram.md)
- Sprint: [SPRINT-007](../03-sprints/SPRINT-007.md)
- Tasks: TASK-0030–TASK-0034