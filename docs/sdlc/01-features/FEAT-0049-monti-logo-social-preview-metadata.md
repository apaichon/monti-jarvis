---
id: FEAT-0049
title: "Monti logo and social preview metadata"
status: completed
release: v2.30.0
roadmap_sprint: 57
priority: O+
depends_on: [SPRINT-039, SPRINT-048, SPRINT-054, SPRINT-056]
design: DES-0053
design_spec: ../02-design/53-monti-logo-social-preview-metadata-spec.md
updated: 2026-07-30
---

# FEAT-0049: Monti Logo and Social Preview Metadata

## Purpose

Use the approved Monti AI Call Center logo on root-domain/customer-facing brand
surfaces and publish rich social metadata so pasted root-domain links show the
right image, title, and description on Facebook and other preview surfaces.

## Problem

The root-domain preview can render as generic or stale when shared into social
apps. The product also needs the new approved Monti logo asset to be the
canonical root-domain brand image.

## Scope

### In

- Copy the approved source logo from
  `/Users/apaichon/Projects/libra/monti/images/logo.png` into the app public
  asset path with a stable, cache-safe filename.
- Update root/customer-facing Monti visible branding to consume the new asset.
- Add Open Graph metadata: `og:title`, `og:description`, `og:image`,
  `og:image:alt`, `og:url`, `og:type`, and `og:site_name`.
- Add Twitter/X card metadata: `twitter:card`, `twitter:title`,
  `twitter:description`, `twitter:image`, and `twitter:image:alt`.
- Ensure production metadata emits absolute crawler-readable URLs.
- Validate local build/source output and prepare a Facebook Sharing Debugger
  checklist for production verification.

### Out

- Full product-web redesign.
- Tenant-specific social preview generation.
- Dynamic route-by-route OG image generation.
- Changing tenant company logos or uploaded brand assets.
- Marketing pixels, ad attribution, or campaign tracking.

## Acceptance Criteria

1. Root-domain/customer-facing page uses the approved Monti logo asset.
2. HTML source includes Open Graph and Twitter metadata with title, description,
   image, image alt, site name, URL, and content type.
3. `og:image` and `twitter:image` resolve to an absolute production URL for
   deployed builds.
4. Preview image URL is public, unauthenticated, and served with an image
   content type.
5. Validation evidence covers local source/build checks and the production
   rich-preview debugger checklist.

## Test Notes

- Local: inspect built HTML/source for meta tags.
- Local: verify image asset is loadable via the dev server.
- Production: verify Facebook Sharing Debugger or equivalent crawler sees the
  Monti title, description, and image.

## Dependencies

- S39 theme/branding conventions
- S48 product web root-domain surface
- S54 customer root tenant-list routing
- S56 caller desk Monti/customer branding
- Design: [DES-0053](../02-design/53-monti-logo-social-preview-metadata-spec.md)
