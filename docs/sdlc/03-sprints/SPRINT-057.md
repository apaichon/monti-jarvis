---
id: SPRINT-057
status: completed
start: 2026-07-30
end: 2026-08-08
updated: 2026-07-30
design_pack: approved
roadmap_sprint: 57
feature: FEAT-0049
platform: Customer / Product Web / Branding
depends_on: [SPRINT-039, SPRINT-048, SPRINT-054, SPRINT-056]
goal: "Use the approved Monti logo on root-domain/customer-facing surfaces and add Open Graph/Twitter metadata for rich link previews."
velocity_basis: "Last 3 recorded closed slices: S54=12, S55=12, S56=12 -> avg 12; commit 12"
release_target: v2.30.0
release: v2.30.0
---

# SPRINT-057 — Monti Logo + Social Preview Metadata

## Goal

Root-domain/customer-facing Monti surfaces use the approved Monti AI Call
Center logo, and pasted root-domain links render a correct rich preview image,
title, and description on Facebook and other social surfaces.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed slices (S54, S55, S56) | 12, 12, 12 -> **avg 12** |
| **Committed** | **12** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0205](../04-tasks/TASK-0205.md) Add approved Monti logo asset and root branding usage | 4 | completed | dev | New logo asset in public path and used by root/customer branding |
| [TASK-0206](../04-tasks/TASK-0206.md) Add Open Graph and Twitter social preview metadata | 4 | completed | dev | Root HTML exposes crawler-ready metadata |
| [TASK-0207](../04-tasks/TASK-0207.md) Social preview validation and release evidence | 4 | completed | tester/dev | Local checks and production rich-preview checklist |
| **Total** | **12** | **12/12** | | |

## Design Pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0049](../01-features/FEAT-0049-monti-logo-social-preview-metadata.md) | approved |
| Deep spec | **`approved`** | [DES-0053](../02-design/53-monti-logo-social-preview-metadata-spec.md) |
| Workflow | **`planned`** | See DES-0053 implementation sequence |
| ER | **`not_required`** | Static asset + HTML metadata only |
| API | **`not_required`** | No backend endpoint change expected |
| UX | **`planned`** | Root/customer visible logo and social-preview metadata |

## Scope Boundary

### In

- Approved Monti logo asset from `/Users/apaichon/Projects/libra/monti/images/logo.png`
- Root/customer-facing visible Monti logo usage
- Open Graph metadata for Facebook/link sharing
- Twitter/X summary-large-image metadata
- Absolute production image URL behavior
- Local checks and post-deploy rich-preview validation checklist

### Out

- Product-web page redesign
- Tenant-specific OG images
- Dynamic social image generation
- Tenant company logo changes
- Marketing tracking pixels

## Verification

```bash
cd apps/product-web && npm run check
# or affected root/customer app check if root domain maps elsewhere
# Inspect built/source HTML for og:* and twitter:* tags
# Verify preview image URL is public and image/* content type after deploy
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Root domain maps to a different frontend app in deployment | Identify root app before implementation; apply metadata at the app serving `/` |
| Social crawler requires absolute image URL | Use production origin config with same-origin public asset path |
| Facebook cache shows stale preview | Include debugger/cache-refresh checklist in TASK-0207 |
| Large image affects load/perf | Use static asset optimized for preview and visible usage; avoid layout-blocking oversized render |

## Worktree

```text
.worktrees/SPRINT-057
branch: feature/sprint-057-logo-social-preview
```

## Implementation Notes

- Approved source logo: `/Users/apaichon/Projects/libra/monti/images/logo.png`
- Candidate app surfaces: `apps/product-web` root domain, then customer-web if
  deployment maps `/` there.
- Metadata must include Open Graph and Twitter card tags for crawler previews.
- Production validation requires deployed root URL and Facebook Sharing Debugger
  or equivalent rich-link crawler.

## Built Summary

- Approved logo copied to customer-web and product-web static assets.
- Customer-web root metadata uses `https://monti.devclub.dev/`.
- Product-web metadata uses `https://monti.devclub.dev/product/`.
- Manual evidence: [SPRINT-057-logo-social-preview-manual.md](../06-manual-tests/SPRINT-057-logo-social-preview-manual.md)

## Shipped Summary (v2.30.0)

- New approved Monti logo installed for customer-web and product-web static assets.
- Open Graph and Twitter card metadata added for root and `/product/` sharing.
- Local customer-web and product-web checks/builds passed.
- Production Facebook Sharing Debugger verification remains a post-deploy cache check.

**Closed:** 2026-07-30
