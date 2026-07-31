# SPRINT-063 Manual UAT - Platform Admin Leads Inbox

**Status:** automated verification passed; target browser confirmation pending

## Automated evidence

- [x] `apps/platform-admin-web`: `npm test` (3/3)
- [x] `apps/platform-admin-web`: `npm run check` (0 errors)
- [x] `apps/platform-admin-web`: `npm run build`
- [x] `git diff --check`

## Book Demo flow

- [ ] Submit a new Book Demo form from Product Web.
- [ ] Confirm `GET /api/platform/leads?status=new` returns the lead in `items[]`.
- [ ] Open Platform Admin Leads with Status `new` and Kind `all`.
- [ ] Confirm the inbox reports `1 shown / 1 total` when the response has one item.
- [ ] Confirm email, name, company, kind, status, source, and created time render.
- [ ] Open the row and verify submitted phone, preferred channel, language, and use case.
- [ ] Change status and assignment; confirm the row remains available under the new filter.
- [ ] Add a note and confirm notes/history refresh.

## Filters and states

- [ ] Filter Kind to `book_demo`; confirm the lead remains visible.
- [ ] Search by email and company; confirm matching results.
- [ ] Use a non-matching search; confirm a true zero-result state and accurate total.
- [ ] Refresh repeatedly; confirm rows are not cleared or duplicated.
- [ ] Confirm API errors show feedback and do not leave stale counts.

## Regression boundary

- [ ] Confirm Contact and Newsletter lead kinds still render.
- [ ] Confirm tenant-admin and unauthenticated users cannot access platform leads.
