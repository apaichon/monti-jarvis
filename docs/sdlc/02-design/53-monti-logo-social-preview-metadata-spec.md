---
id: DES-0053
title: Monti Logo and Social Preview Metadata Specification
status: completed
release: v2.30.0
updated: 2026-07-30
sprint: SPRINT-057
owner: SA
feature: FEAT-0049
release_target: v2.30.0
---

# DES-0053 — Monti Logo + Social Preview Metadata

**Sprint:** [SPRINT-057](../03-sprints/SPRINT-057.md) · **Release target:** v2.30.0  
**Feature:** [FEAT-0049](../01-features/FEAT-0049-monti-logo-social-preview-metadata.md)  
**Tasks:** [TASK-0205](../04-tasks/TASK-0205.md), [TASK-0206](../04-tasks/TASK-0206.md),
[TASK-0207](../04-tasks/TASK-0207.md)  
**Source logo:** `/Users/apaichon/Projects/libra/monti/images/logo.png`  
**Related:** [DES-0043](43-product-web-growth-spec.md) (product web),
[DES-0049](49-customer-portal-tenant-list-spec.md) (root tenant list),
[DES-0052](52-caller-desk-branding-audio-devices-spec.md) (customer branding)

## 1. Goals

1. Use the approved Monti AI Call Center logo as the canonical root-domain
   Monti mark.
2. Publish crawler-ready Open Graph and Twitter card metadata on the root page.
3. Ensure shared root-domain links show the correct image, title, description,
   and site name on Facebook and other rich-preview surfaces.
4. Avoid backend/data migrations; this is static asset and frontend metadata
   work.

## 2. Non-Goals

- Product-web landing page redesign
- Tenant-specific Open Graph previews
- Dynamic image generation per tenant or route
- Tenant uploaded logo changes
- Marketing tracking pixels, campaign UTM logic, or ad attribution

## 3. Surface Ownership

| Surface | Rule |
| --- | --- |
| Root domain `/` | Must expose Monti logo and social metadata |
| Product web | Primary target if deployment routes root domain to `apps/product-web` |
| Customer web | Secondary target if deployment routes root domain to `apps/customer-web` |
| Tenant/platform admin | No required visible changes unless shared root shell owns metadata |

Implementation must confirm which app owns production `/` before code changes.
If both product-web and customer-web can serve root links, both may need the
metadata head tags and the public logo asset.

## 4. Asset Policy

| Concern | Rule |
| --- | --- |
| Source | `/Users/apaichon/Projects/libra/monti/images/logo.png` |
| Destination | App public/static asset path, same origin as root page |
| Filename | Stable and descriptive, e.g. `monti-social-preview.png` |
| Access | Public, unauthenticated, crawler-readable |
| Content type | Served as `image/png` or equivalent image content type |
| Tenant isolation | Do not replace tenant/company uploaded logo assets |

The social preview image should be same-origin where practical. A same-origin
public asset avoids crawler failures from private storage, signed URLs, or CORS
mistakes.

## 5. Metadata Contract

Root page HTML must include:

| Tag | Required value policy |
| --- | --- |
| `title` | Monti AI Call Center |
| `meta[name="description"]` | Concise product description |
| `meta[property="og:title"]` | Monti AI Call Center |
| `meta[property="og:description"]` | Same positioning as description |
| `meta[property="og:image"]` | Absolute production URL to the logo/preview image |
| `meta[property="og:image:alt"]` | Monti AI Call Center logo |
| `meta[property="og:url"]` | Absolute production root URL |
| `meta[property="og:type"]` | `website` |
| `meta[property="og:site_name"]` | Monti AI Call Center |
| `meta[name="twitter:card"]` | `summary_large_image` |
| `meta[name="twitter:title"]` | Monti AI Call Center |
| `meta[name="twitter:description"]` | Same positioning as description |
| `meta[name="twitter:image"]` | Absolute production URL to the logo/preview image |
| `meta[name="twitter:image:alt"]` | Monti AI Call Center logo |

Suggested description:

```text
Monti AI Call Center helps customers start AI voice and text conversations with tenant-branded AI agents.
```

## 6. URL and Environment Rules

| Environment | Rule |
| --- | --- |
| Local dev | Metadata may resolve from local origin for inspection |
| Production | `og:url`, `og:image`, and `twitter:image` must be absolute HTTPS URLs |
| Missing origin config | Fallback safely; do not emit malformed `undefined/...` URLs |
| Crawlers | Image URL must not require auth, cookies, signed headers, or JS execution |

Preferred implementation is a small helper that joins a configured public site
origin with a same-origin public asset path.

## 7. Data Model

No Postgres, Redis, ClickHouse, or MinIO schema change.

| Store | Change |
| --- | --- |
| App public/static folder | Add Monti logo/preview image asset |
| Frontend page/head metadata | Add static metadata tags |
| Config/env | Optional public site origin for production absolute URLs |

## 8. API Summary

No new REST, WebSocket, NATS, or background worker endpoint is expected.

| Method | Path | Change |
| --- | --- | --- |
| GET | `/` | Returns HTML with visible logo and social-preview metadata |
| GET | `/assets-or-static/monti-social-preview.png` | Public image asset served by frontend/static host |

## 9. UX / HTML Mapping

```text
Root page
├─ <head>
│  ├─ title / description
│  ├─ og:title / og:description / og:image / og:url / og:type / og:site_name
│  └─ twitter:card / twitter:title / twitter:description / twitter:image
└─ <body>
   └─ Monti visible brand mark uses approved logo where root/product shell shows logo
```

## 10. Implementation Sequence

1. **TASK-0205** — Add approved logo asset and use it for root/customer-visible
   Monti branding.
2. **TASK-0206** — Add Open Graph and Twitter card metadata with absolute
   production URL handling.
3. **TASK-0207** — Validate local output and document production rich-preview
   debugger checks.

## 11. Verification

```bash
cd apps/product-web && npm run check
# If root domain maps to customer-web instead:
cd apps/customer-web && npm run check
```

Manual/source checks:

1. Inspect root HTML source and confirm every required `og:*` and `twitter:*`
   tag is present.
2. Open the preview image URL directly and confirm it returns an image without
   auth.
3. Confirm the visible root logo uses the approved image.
4. After deployment, run Facebook Sharing Debugger or equivalent rich-preview
   crawler and refresh cache if needed.

## 12. Risks

| Risk | Mitigation |
| --- | --- |
| Production root app differs from local assumption | Confirm deployment mapping before code edit |
| Social crawlers cache old metadata | Add production debugger/cache refresh to TASK-0207 |
| Relative image URL is ignored by crawlers | Emit absolute HTTPS production image URL |
| Logo file is too large for visible use | Use static optimized public asset while preserving preview quality |

## 13. Acceptance Mapping

| Acceptance | Covered by |
| --- | --- |
| New logo used on root/customer-facing surface | TASK-0205 |
| OG/Twitter metadata present | TASK-0206 |
| Absolute public preview image URL | TASK-0206, TASK-0207 |
| Local/source validation | TASK-0207 |
| Production rich-preview checklist | TASK-0207 |
