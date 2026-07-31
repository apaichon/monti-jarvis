---
id: FEAT-0055
title: "Platform Admin Leads inbox renders API items"
status: completed
release: pending
roadmap_sprint: 63
priority: BUG/P1
depends_on: [SPRINT-048, SPRINT-051]
updated: 2026-07-31
---

# FEAT-0055: Platform Admin Leads Inbox Rendering

## Purpose

Show Product Web Book Demo submissions in the Platform Admin Leads inbox when
the platform API returns them.

## Problem

`GET /api/platform/leads` returns lead rows in `items[]`, but the Platform
Admin client expected a `leads` field. The total count remained correct while
the rendered collection became empty, producing `0 shown / 1 total`.

## Scope

### In

- Align the TypeScript list response with `items`, `total`, `limit`, `offset`.
- Normalize null or omitted collection/count values safely.
- Render supported lead kinds, including `book_demo`.
- Preserve filters, detail drawer, status, assignment, notes, and history.
- Add a regression test for non-empty, fallback-count, and empty responses.

### Out

- Backend API or database changes.
- Product Web form changes.
- Lead lifecycle redesign or external CRM integration.

## Acceptance criteria

1. A response containing one Book Demo lead in `items[]` renders one inbox row.
2. The inbox reports `1 shown / 1 total` for that response.
3. Missing `total` falls back to the number of rendered items.
4. Null `items` produces a true empty state without an exception.
5. Existing detail and mutation flows remain available for the rendered row.

## Design

- Existing domain: Sprint 48 marketing leads.
- API contract: [04-api-spec.md](../02-design/04-api-spec.md), Sprint 63.
- UX mapping: [05-ux-ui.md](../02-design/05-ux-ui.md), A63.
