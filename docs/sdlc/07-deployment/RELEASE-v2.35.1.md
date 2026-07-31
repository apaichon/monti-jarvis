---
id: RELEASE-v2.35.1
status: ready
release: v2.35.1
date: 2026-07-31
sprints: [SPRINT-063]
git_tag: v2.35.1
---

# Release v2.35.1

Patch release for the Platform Admin Leads inbox rendering regression.

## Included

- Platform Admin Leads list now reads API rows from `items[]`.
- Product Web Book Demo leads render as inbox rows instead of producing
  `0 shown / 1 total`.
- List normalization preserves server count metadata and handles null item
  collections safely.

## Database

No schema or migration changes.

## Configuration

No environment changes.

## Verification

- `cd apps/platform-admin-web && npm test`: 3/3 passed.
- `git diff --check`: passed.
- Sprint implementation verification also passed `npm run check`,
  `npm run build`, and focused Go server/store regression tests before merge.

## Deferred manual check

Run credentialed target-environment browser UAT from
`docs/sdlc/06-manual-tests/SPRINT-063-leads-inbox-manual.md`.
