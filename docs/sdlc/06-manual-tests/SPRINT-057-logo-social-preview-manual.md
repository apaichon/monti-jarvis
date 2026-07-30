---
id: MANUAL-TEST-SPRINT-057
title: Sprint 57 logo and social preview metadata manual validation
status: completed
updated: 2026-07-30
sprint: SPRINT-057
feature: FEAT-0049
---

# Sprint 57 Logo and Social Preview Metadata Manual Validation

## Local Build Evidence

Commands run:

```bash
cd apps/customer-web && npm run check && npm run build
cd apps/product-web && npm run check && npm run build
```

Result:

- `customer-web` check: 0 errors, 0 warnings.
- `customer-web` build: passed.
- `product-web` check: 0 errors, 4 existing Svelte warnings unrelated to S57.
- `product-web` build: passed with the same existing warnings.

## Source / Build Checks

- `apps/customer-web/build/index.html` includes:
  - `og:title`
  - `og:description`
  - `og:image`
  - `og:image:alt`
  - `og:url`
  - `og:type`
  - `og:site_name`
  - `twitter:card`
  - `twitter:title`
  - `twitter:description`
  - `twitter:image`
  - `twitter:image:alt`
- `apps/product-web/build/index.html` includes the same Open Graph and Twitter
  tags.
- Customer root metadata points to:
  `https://monti.devclub.dev/images/monti-social-preview.png`
- Product metadata points to:
  `https://monti.devclub.dev/product/images/monti-social-preview.png`

## Asset Checks

- `apps/customer-web/build/images/monti-logo.png` is PNG image data,
  1254 x 1254.
- `apps/customer-web/build/images/monti-social-preview.png` is PNG image data,
  1254 x 1254.
- `apps/product-web/build/images/monti-logo.png` is PNG image data,
  1254 x 1254.
- `apps/product-web/build/images/monti-social-preview.png` is PNG image data,
  1254 x 1254.
- Static assets were copied from:
  `/Users/apaichon/Projects/libra/monti/images/logo.png`

## Post-Deploy Checklist

- [ ] Deploy customer-web and product-web builds to production.
- [ ] Open `https://monti.devclub.dev/images/monti-social-preview.png` and
  confirm it returns the Monti PNG without authentication.
- [ ] Open `https://monti.devclub.dev/product/images/monti-social-preview.png`
  and confirm it returns the Monti PNG without authentication.
- [ ] Run Facebook Sharing Debugger against `https://monti.devclub.dev/`.
- [ ] Confirm Facebook preview shows Monti AI Call Center title, description,
  and the approved logo image.
- [ ] If Facebook shows stale content, click scrape/refresh in the debugger and
  re-check the preview.
