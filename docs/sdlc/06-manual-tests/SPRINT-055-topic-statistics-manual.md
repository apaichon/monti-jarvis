---
id: UAT-055
title: Sprint 55 Manual UAT — Tenant Call-Center Topic Statistics
status: completed_auto
updated: 2026-07-29
sprint: SPRINT-055
release: v2.28.0
---

# Sprint 55 Manual UAT — Tenant Call-Center Topic Statistics

Sprint 55 shipped tenant call-center statistics grouped and filtered by topic.
Automated release verification passed locally; browser/staging UAT remains a
safe re-run checklist for the next tester pass.

## Automated release verification

- [x] `go test ./internal/clickhouse ./cmd/server -count=1`
- [x] `git diff --check`
- [x] `cd apps/tenant-web && npm run check`
- [x] `cd apps/tenant-web && npm run build`

Known non-blocking warnings:

- `apps/tenant-web/src/routes/avatars/+page.svelte` has two pre-existing
  Svelte a11y warnings on the create-avatar modal.
- `npm ci` reported existing audit findings: 3 low, 1 high.

## Scenario checklist

| Scenario | Expected | Status |
| --- | --- | --- |
| Seed billing, technical, general, and missing-topic completed conversations | Analytics returns separate `by_topic` buckets and missing values appear as `unknown` | Passed by unit/contract coverage; staging browser re-run pending |
| Filter `GET /api/tenant/call-center/statistics?topic=billing` | Totals, `by_channel`, `by_avatar`, and `by_topic` only include billing facts | Passed by unit/contract coverage |
| Invalid `topic=billing!` | API returns `400 validation_error` | Passed by helper coverage |
| Tenant dashboard topic selector | Selecting a topic reloads stats and keeps quota/channel/avatar sections visible | Build/typecheck passed; browser re-run pending |
| Tenant isolation | Tenant ID comes only from auth context; no tenant selector accepted | Preserved by endpoint implementation |
| Empty range / unavailable analytics | Empty and retry states remain compatible with S25 dashboard behavior | Regression preserved; browser re-run pending |

## Release sign-off

User authorized Sprint 55 close, merge to main, push, and tag on 2026-07-29.
