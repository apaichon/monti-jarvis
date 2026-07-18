---
id: DES-0037
title: Theme Color Customization Specification
status: review_pending
updated: 2026-07-18
sprint: SPRINT-039
owner: SA
release_target: v2.15.0
---

# Theme Color Customization — Design Spec

**Sprint:** SPRINT-039 · **Release target:** v2.15.0  
**Feature:** [FEAT-0035](../01-features/FEAT-0035-theme-color-customization.md)  
**Depends on:** [17-embed-to-web-spec.md](17-embed-to-web-spec.md), [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md)

## 1. Goals

1. Per-tenant **draft** and **published** color token documents.  
2. **Presets**: `dark` (default, pre-S39 parity), `light`, `branded` (custom).  
3. Apply **published** tokens as CSS custom properties on **customer** and **embed** surfaces.  
4. Tenant admin editor with **live preview**, **contrast warnings**, publish confirmation.  
5. Public resolve of published theme only (never draft).  
6. Platform admin **read-only** summary for support.

## 2. Non-goals

| Out | Notes |
| --- | --- |
| Logo / font system | Existing branding fields only |
| White-label CNAME | Separate backlog |
| Mobile theme API | Future consumer of same tokens |
| Arbitrary CSS upload | Token JSON only |
| Full platform theme marketplace | Out |

## 3. Token model

### 3.1 Required tokens

| Token key | CSS variable | Purpose |
| --- | --- | --- |
| `primary` | `--mj-primary` | Buttons, launcher gradient start, focus |
| `accent` | `--mj-accent` | Secondary highlight / gradient end |
| `surface` | `--mj-surface` | Panel / card background |
| `background` | `--mj-background` | Page background |
| `text` | `--mj-text` | Primary ink |
| `muted` | `--mj-muted` | Secondary text |
| `line` | `--mj-line` | Borders |
| `success` | `--mj-success` | Positive status |
| `warn` | `--mj-warn` | Warning status |
| `danger` | `--mj-danger` | Errors / destructive |

Format: `#RRGGBB` (preferred) or `#RGB` expanded server-side. Reject alpha / named CSS colors in v1.

### 3.2 Presets

| Preset | Notes |
| --- | --- |
| `dark` | Default; maps from current Monti dark palette (document exact hex in migration seed). |
| `light` | Light surface/background, dark text; primary/accent retain brand-safe blues unless overridden. |
| `branded` | Starts from dark or last draft; all tokens editable. |

`preset` field on the row indicates which base was last applied; `branded` means custom edits are the source of truth.

### 3.3 Contrast rules

Compute relative luminance per WCAG 2.x and contrast ratio:

| Pair | Minimum for “pass” |
| --- | --- |
| `text` on `surface` | 4.5 |
| `text` on `background` | 4.5 |
| `primary` on `surface` (for button label white/black auto-pick) | document chosen algorithm |

**Publish policy:** If any required pair fails, API returns `contrast_warnings[]`. Client must send `confirm_low_contrast: true` on publish to proceed (soft gate). Log audit on override.

## 4. Data model

### `tenant_themes`

| Column | Type | Notes |
| --- | --- | --- |
| `tenant_id` | text PK FK | → `tenants.id` ON DELETE CASCADE |
| `preset` | text | `dark` \| `light` \| `branded` |
| `draft_tokens` | jsonb | required keys object |
| `published_tokens` | jsonb | published snapshot; empty/`null` = use system dark default |
| `published_at` | timestamptz null | last successful publish |
| `draft_updated_at` | timestamptz | last draft save |
| audit | | `created_at`, `updated_at`, `created_by`, `updated_by` |

Redis (optional): `monti_jarvis:theme:pub:{tenant_id}` TTL 60s — fail-open to Postgres.

Migration: `scripts/migrations/` or `ensureThemeSchema` alongside settings pattern.

## 5. API summary

Auth: active `tenant_admin` unless noted.

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/tenant/theme` | tenant_admin | Draft + published + contrast report on draft |
| `PUT` | `/api/tenant/theme` | tenant_admin | Save draft (`preset`, `tokens`) |
| `POST` | `/api/tenant/theme/publish` | tenant_admin | Publish draft → published; body may include `confirm_low_contrast` |
| `POST` | `/api/tenant/theme/reset` | tenant_admin | Reset draft to preset defaults (`preset` in body) |
| `GET` | `/api/public/theme/{tenant_id}` | public | Published tokens only |
| `GET` | `/api/public/embed/{embed_key}` | public | Extend existing response with `theme: { preset, tokens }?` when published |
| `GET` | `/api/admin/tenants/{id}/theme` | platform_admin | Read-only summary |

### Errors

| HTTP | code |
| ---: | --- |
| 400 | `invalid_theme_tokens`, `invalid_preset` |
| 409 | `contrast_confirmation_required` (when warnings and no confirm flag) |
| 401/403 | auth / inactive tenant |
| 404 | tenant not found (admin) |

## 6. Application surfaces

```text
Published tokens
  → customer-web :root / [data-mj-theme] style attribute
  → embed route same map
  → launcher button in monti-embed.js may stay default OR read CSS vars if injected into host (host injection out of scope for vanilla loader; iframe interior uses tokens)
```

**Vanilla loader:** outer launcher chrome remains Monti default unless future host CSS hooks; **iframe interior** must use tenant published theme.

**Framework SDKs:** `theme` prop remains optional string hint; server published tokens win for iframe content.

## 7. RBAC

| Role | Capability |
| --- | --- |
| `tenant_admin` (active) | Full draft/publish/reset for own tenant |
| `platform_admin` | Read summary only |
| Public | Published tokens for known tenant_id or via embed key resolve |
| Customer end-user | No theme write |

## 8. Verification (curl sketch)

```bash
# Login tenant admin → TOKEN
curl -sS -H "Authorization: Bearer $TOKEN" "$MONTI/api/tenant/theme" | jq .

curl -sS -X PUT -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"preset":"branded","tokens":{"primary":"#ff5500", ...}}' \
  "$MONTI/api/tenant/theme" | jq .

curl -sS -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"confirm_low_contrast":true}' \
  "$MONTI/api/tenant/theme/publish" | jq .

curl -sS "$MONTI/api/public/theme/$TENANT_ID" | jq .
```

## 9. See also

- [02-workflow.md](02-workflow.md) §89–90  
- [03-er-diagram.md](03-er-diagram.md) Sprint 39  
- [04-api-spec.md](04-api-spec.md) Theme  
- [05-ux-ui.md](05-ux-ui.md) Sprint 39 T20  
- [SPRINT-039](../03-sprints/SPRINT-039.md)
