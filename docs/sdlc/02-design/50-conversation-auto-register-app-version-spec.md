---
id: DES-0050
title: Conversation Auto-Register Email OTP and App Version Specification
status: shipped
release: v2.26.0
updated: 2026-07-29
sprint: SPRINT-053
owner: SA
feature: FEAT-0046
---

# DES-0050 — Conversation Auto-Register (Email OTP) + App Version

**Sprint:** SPRINT-053 · **Release target:** v2.25.0  
**Feature:** [FEAT-0046](../01-features/FEAT-0046-conversation-auto-register-app-version.md)  
**Tasks:** [TASK-0188](../04-tasks/TASK-0188.md), [TASK-0189](../04-tasks/TASK-0189.md),
[TASK-0198](../04-tasks/TASK-0198.md)  
**Depends on:** [23-customer-auth-spec.md](23-customer-auth-spec.md),
[19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md)

## 1. Goals

- Tenant-controlled **auto-register** when a customer enters email and verifies
  OTP **during conversation**.
- Reuse customer directory + OTP + session authorities (S19–S21).
- Show **app version** equal to git tag / `VERSION` on tenant and admin UIs
  (customer portal optional footer).

## 2. Non-goals

- SMS OTP, social login, password accounts
- Forcing email capture on every anonymous demo when setting is off
- Replacing require-auth-for-workforce policy
- S51 commercial catalog changes

## 3. Setting

Extend `tenant_customer_auth_settings`:

| Column | Type | Default | Notes |
| --- | --- | --- | --- |
| `auto_register_on_conversation_otp` | boolean NOT NULL | `false` | When true, conversation email OTP may create customer |

Exposed on existing tenant customer-auth settings GET/PUT as:

```json
"auto_register_on_conversation_otp": false
```

**Coexistence with `require_auth_for_workforce`:**

| require_auth_for_workforce | auto_register_on_conversation_otp | Behavior |
| --- | --- | --- |
| false | false | Anonymous conversation allowed; no auto-register UI |
| false | true | Anonymous start allowed; UI may prompt email+OTP mid-flow or before sensitive actions; verify auto-creates |
| true | true | Must auth before workforce; verify still auto-creates if missing (same OTP path) |
| true | false | Existing S21: OTP required; customer must already exist or follow existing verify path without auto-create if product currently requires pre-provisioned customers — **S53: when auto_register false, verify fails for unknown email with stable code `customer_not_found` unless existing product already creates; prefer auto-create only when flag true** |

**Rule:** Auto-create on successful OTP **only if** `auto_register_on_conversation_otp` is true. If false and email is unknown, return `404 customer_not_found` (or existing product error) without creating a row.

## 4. Conversation auto-register flow

```text
Customer in conversation (tenant context known)
  → (setting on) show email step
  → POST /api/customer/auth/request-otp { tenant_id, email, purpose: "conversation" }
  → email OTP (existing sender)
  → POST /api/customer/auth/verify-otp { challenge_id, otp }
  → if auto_register: ensure customer row (upsert by tenant+normalized email)
  → issue customer session
  → bind subsequent chat/voice to customer_id
```

Public discovery of setting for conversation UI (no secrets):

```http
GET /api/public/tenants/{tenant_id}/customer-auth-policy
→ { "enabled": true, "auto_register_on_conversation_otp": true, "require_auth_for_workforce": false }
```

Only safe booleans; never domains list if that would leak policy detail beyond needed — domains still enforced server-side on request-otp.

## 5. Data model

### Migration `036_conversation_auto_register.sql`

```sql
ALTER TABLE callcenter.tenant_customer_auth_settings
  ADD COLUMN IF NOT EXISTS auto_register_on_conversation_otp boolean NOT NULL DEFAULT false;
```

No new tables. Reuse:

- `customers`, `customer_auth_identities`, OTP challenges, sessions
- Redis rate keys `monti_jarvis:rate:customer_auth:{tenant}:{email_hash}`

## 6. App version

| Source | Canonical |
| --- | --- |
| Repo file | `VERSION` → `2.24.0` |
| Display | Prefer `v` + VERSION → `v2.24.0` |
| Git tag | `v2.24.0` must match |

**Server:** embed VERSION at build (`//go:embed` or ldflags).  
**Endpoints:**

- `GET /healthz` includes `"version": "v2.24.0"` (replace hard-coded sprint string)
- `GET /api/version` → `{ "version": "v2.24.0", "version_raw": "2.24.0" }` public

**UI:** tenant sidebar footer; platform admin footer; customer small footer.

## 7. API summary

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| GET/PUT | existing tenant customer-auth settings | tenant_admin | + auto_register flag |
| GET | `/api/public/tenants/{tenant_id}/customer-auth-policy` | public | safe booleans |
| POST | `/api/customer/auth/request-otp` | public | honor setting + domain + rate |
| POST | `/api/customer/auth/verify-otp` | public | auto-create iff setting on |
| GET | `/api/version` | public | app version |
| GET | `/healthz` | public | includes version |

## 8. RBAC

| Action | tenant_admin | customer | public |
| --- | --- | --- | --- |
| Change auto-register setting | yes | no | no |
| Request/verify conversation OTP | n/a | via OTP | yes (tenant scoped) |
| Read version | yes | yes | yes |

## 9. Verification

```bash
# setting off: unknown email verify → customer_not_found
# setting on: verify → customer created + session
# VERSION file 2.25.0 → /api/version shows v2.25.0
```

## 10. See also

- Workflow §120–122
- ER Sprint 53
- API Sprint 53
- UX T53 / C53 / A53
- [SPRINT-053](../03-sprints/SPRINT-053.md)
