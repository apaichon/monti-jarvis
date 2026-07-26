---
id: DES-0043
title: Product Web Growth and Conversion Specification
status: approved
updated: 2026-07-25
sprint: SPRINT-048
owner: SA
feature: FEAT-0040
---

# Product Web Growth and Conversion — Design Spec

**Sprint:** SPRINT-048 · **Release target:** v2.20.0  
**Feature:** [FEAT-0040](../01-features/FEAT-0040-product-web-growth.md)  
**Depends on:** [08-packages-spec.md](08-packages-spec.md),
[11-tenant-register-spec.md](11-tenant-register-spec.md),
[14-buy-package-spec.md](14-buy-package-spec.md),
[20-tenant-test-preview-spec.md](20-tenant-test-preview-spec.md),
[23-customer-auth-spec.md](23-customer-auth-spec.md),
[37-theme-color-customization-spec.md](37-theme-color-customization-spec.md)

## 1. Goals

- Public product website that markets Monti AI call-center value and converts
  visitors into leads, demos, registrations, and paid tenants.
- Reuse existing tenant registration, KYC, package catalog, checkout, and
  (when present) referral attribution — never fork commerce authorities.
- Consent-aware lead capture with sales lifecycle and funnel measurement.

## 2. Non-goals

- CRM replacement, marketing automation, or unsolicited email blasts
- Public payment processing or entitlement grants from marketing pages
- Hard-coded package prices or unapproved social-proof metrics
- Arbitrary external redirects or open redirectors
- Replacing customer portal at `/` or tenant portal at `/tenant/`

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `PRODUCT_WEB_DIR` | `apps/product-web/build` | Static product-web assets served at `/product/` |
| `PRODUCT_WEB_ENABLED` | `true` | When `false`, `/product/` returns a disabled page |
| `LEAD_CAPTURE_ENABLED` | `true` | When `false`, public lead POST returns `503` |
| `LEAD_RATE_LIMIT_PER_IP` | `10` | Max lead posts per IP per hour (Redis) |
| `FUNNEL_RATE_LIMIT_PER_IP` | `120` | Max funnel events per IP per hour |
| `LEAD_DEDUPE_WINDOW_HOURS` | `24` | Same email+kind dedupe window |

## 4. Surface topology

| Surface | Path | App |
| --- | --- | --- |
| Product web (marketing) | `/product/*` | `apps/product-web` |
| Live demo (existing) | `/` customer portal | `apps/customer-web` |
| Tenant register / billing | `/tenant/*` | `apps/tenant-web` |
| Sales lead console | `/admin/leads` | `apps/platform-admin-web` |

Production may map apex host to product-web via reverse proxy; local/dev uses
path `/product/` on `:8091`.

## 5. Data model (Postgres `callcenter`)

### 5.1 `marketing_leads`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `lead_{hex}` |
| `kind` | text | `contact` \| `book_demo` \| `newsletter` |
| `status` | text | lifecycle (see §6) |
| `email` | text | lowercased, required |
| `full_name` | text | optional |
| `company_name` | text | optional |
| `phone` | text | optional |
| `use_case` | text | free text, max 2k |
| `preferred_channel` | text | `email` \| `phone` \| `line` \| `other` |
| `language` | text | `en` \| `th` default `en` |
| `consent_marketing` | boolean | required true for newsletter |
| `consent_contact` | boolean | required true for contact/book_demo |
| `consent_at` | timestamptz | set on create |
| `utm_source` | text | nullable |
| `utm_medium` | text | nullable |
| `utm_campaign` | text | nullable |
| `utm_content` | text | nullable |
| `utm_term` | text | nullable |
| `referral_code` | text | optional S46 code when present |
| `landing_path` | text | first product-web path |
| `package_interest_id` | text | optional package id from pricing CTA |
| `dedupe_key` | text | unique `(kind, email_norm)` active window helper |
| `assigned_to` | text | optional platform user id |
| `converted_tenant_id` | text | nullable FK when linked post-register |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

**Indexes:** `(status, created_at desc)`, unique partial on active dedupe when
status not in (`lost`, `unsubscribed`, `converted`) within window — implement
as application dedupe + unique constraint on `(kind, email)` for simplicity in
v1, with status transitions recorded in notes/history.

### 5.2 `marketing_lead_notes`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `lnote_{hex}` |
| `lead_id` | text FK | → `marketing_leads.id` ON DELETE CASCADE |
| `body` | text | max 4k |
| audit cols | | actor in `created_by` |

### 5.3 `marketing_lead_events` (status history)

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | |
| `lead_id` | text FK | |
| `from_status` | text | nullable on create |
| `to_status` | text | required |
| `actor` | text | `system` or user id |
| `created_at` | timestamptz | |

### 5.4 `funnel_events`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | |
| `event_name` | text | allowlisted |
| `page_path` | text | |
| `cta_id` | text | optional |
| `utm_*` / `referral_code` | text | optional |
| `session_key` | text | opaque client key, not email |
| `client_ip_hash` | text | hashed IP for abuse analytics only |
| `created_at` | timestamptz | |

No conversation transcripts, passwords, or payment payloads.

### Migration

`scripts/migrations/029_product_web_leads.sql` (or next free migration number).

## 6. Lead lifecycle

```text
new → contacted → demo_scheduled → demo_completed → qualified
  → registered → kyc_pending → package_selected → paid → converted
lost | unsubscribed (terminal-ish; may re-open only by admin)
```

Public create always inserts `new`. Only platform sales/admin transitions status.

## 7. API summary

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `POST` | `/api/public/leads` | public + rate limit | Create/dedupe lead |
| `GET` | `/api/public/packages` | public | Active sellable packages |
| `POST` | `/api/public/funnel/events` | public + rate limit | Funnel beacon |
| `GET` | `/api/platform/leads` | platform_admin | List/filter leads |
| `GET` | `/api/platform/leads/{id}` | platform_admin | Lead detail + notes |
| `PATCH` | `/api/platform/leads/{id}` | platform_admin | Status/assignment |
| `POST` | `/api/platform/leads/{id}/notes` | platform_admin | Append note |

### Public lead request

```json
{
  "kind": "book_demo",
  "email": "ops@example.com",
  "full_name": "Alex Example",
  "company_name": "Example Co",
  "phone": "+66...",
  "use_case": "Inbound support for e-commerce",
  "preferred_channel": "email",
  "language": "en",
  "consent_contact": true,
  "consent_marketing": false,
  "utm_source": "google",
  "utm_campaign": "aiaas-q3",
  "referral_code": "REFABC",
  "landing_path": "/product/pricing",
  "package_interest_id": "aiaas-1000",
  "website": ""
}
```

`website` is honeypot — must be empty.

### Errors

| Code | HTTP | When |
| --- | --- | --- |
| `LEAD_DISABLED` | 503 | capture disabled |
| `LEAD_VALIDATION` | 400 | missing/invalid fields |
| `LEAD_CONSENT_REQUIRED` | 400 | consent false |
| `LEAD_RATE_LIMITED` | 429 | IP over limit |
| `LEAD_SPAM` | 400 | honeypot filled |
| `FUNNEL_UNKNOWN_EVENT` | 400 | event not allowlisted |
| `PACKAGE_PUBLIC_UNAVAILABLE` | 503 | catalog disabled |

## 8. Attribution and redirect rules

Allowlisted outbound paths from product-web:

- `/` (demo/customer)
- `/product/*` (same site)
- `/tenant/register`
- `/tenant/login`
- `/tenant/billing` (only useful post-auth; still relative)

Allowlisted query keys: `utm_source`, `utm_medium`, `utm_campaign`,
`utm_content`, `utm_term`, `ref`, `package_id`, `lead_id`, `lang`.

Reject: absolute external URLs, `//evil`, `javascript:`, data URLs.

Referral code (`ref`) may be forwarded into register when S46 attribution API
accepts it; product-web does not invent qualification logic.

## 9. Public packages projection

Map active packages to public DTO:

```json
{
  "packages": [
    {
      "id": "aiaas-1000",
      "name": "AiaaS ฿1,000",
      "price_amount": 1000,
      "price_currency": "THB",
      "billing_period": "month",
      "highlights": ["3 AI avatars", "300 KM docs", "20 GB storage"],
      "rules_summary": { "ai_employees": 3, "km_documents": 300 }
    }
  ]
}
```

Omit internal cost rates, provider keys, and inactive/archived packages.

## 10. UX brand notes

- Dark navy/blue Monti brand; high-contrast CTAs (cyan/white)
- Reference artwork from product-web downloads is inspiration only
- Primary CTAs: Try live demo · Book a demo · Start free registration · Contact sales
- Pricing cards always show “Prices from catalog” source note when loaded

## 11. RBAC

| Role | Public product pages | Lead create | Lead admin | Public packages |
| --- | --- | --- | --- | --- |
| anonymous | yes | yes | no | yes |
| tenant_admin | yes | yes | no | yes |
| platform_admin | yes | yes | yes | yes |

## 12. Verification curl

```bash
# Public packages
curl -s http://localhost:8091/api/public/packages | jq '.packages | length'

# Create lead
curl -s -X POST http://localhost:8091/api/public/leads \
  -H 'content-type: application/json' \
  -d '{"kind":"contact","email":"a@example.com","consent_contact":true,"full_name":"A"}'

# Funnel
curl -s -X POST http://localhost:8091/api/public/funnel/events \
  -H 'content-type: application/json' \
  -d '{"event_name":"page_view","page_path":"/product/","session_key":"s1"}'

# Platform list (admin token)
curl -s http://localhost:8091/api/platform/leads -H "Authorization: Bearer $TOKEN"
```

## 13. Packages / files

| Area | Path |
| --- | --- |
| Product UI | `apps/product-web/` |
| Static serve | `internal/productweb/serve.go` |
| Lead domain | `internal/leads/` |
| Store | `internal/store/leads.go` |
| Handlers | `cmd/server/leads.go`, `cmd/server/public_packages.go` |
| Migration | `scripts/migrations/029_product_web_leads.sql` |
| Sales UI | `apps/platform-admin-web/src/routes/leads/` |

See workflow §105–108, ER Sprint 48, API Sprint 48, UX P48/A48.
