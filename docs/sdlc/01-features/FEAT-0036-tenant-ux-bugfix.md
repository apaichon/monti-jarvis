# Feature: Tenant UX Bug Fix Pack   (FEAT-0036)
**Sprint:** SPRINT-042   **Owner:** DEV   **Status:** shipped · **Release:** v2.16.0

## Problem

Operators hit four recurring tenant-console defects that block trust and daily use:

1. **Session expired** — JWT/session expiry is inconsistent (silent fail, lost deep link, or incomplete logout).
2. **First login menu blank** — after first successful login the sidebar/menu does not paint until a full refresh.
3. **Tenant menu UX** — long nav is flat (no groups) and does not scroll unless the user clicks the last visible item.
4. **Document scope** — KM documents cannot reliably be assigned to / filtered by scope; RAG may ignore tenant scope allowlists.

## Scope

**In:**
- Tenant web session expiry: 401 → clear session → login with `reason=session_expired` + `next`
- Consistent `handleUnauthorized` / refresh-token path (no double-redirect loops)
- Post-login navigation: full shell + nav on first client paint without hard reload
- Tenant nav: **grouped sections** + **overflow scroll** (sidebar `overflow-y: auto`, sticky brand/footer as needed)
- KM document **scope** assign on upload/edit; list/filter by scope; store + API enforce scope on document metadata
- Manual UAT checklist for all four defects

**Out:**
- Customer portal redesign, platform admin nav redesign
- New product features (embed auth, tenant Gemini key — S43)
- Full IA/information architecture product rewrite beyond grouping existing links
- Changing JWT TTLs as a product policy (may document defaults only)

## Acceptance criteria

1. With expired access token (and failed/missing refresh), any authenticated API call redirects tenant UI to `/tenant/login?reason=session_expired&next=…`; after re-login, user lands on `next`.
2. Clear site data → login as active tenant admin → **all** primary nav groups/links visible **without** F5.
3. Sidebar scrolls when content exceeds viewport; user can reach Theme / Preview / last items without clicking a link first; nav items are grouped (e.g. Operations, Knowledge, Commerce, Settings).
4. Tenant KM: create/update document with a scope id; list shows scope; documents without allowed scope are not returned to chat RAG for agents outside that scope (or are excluded per existing scope rules).
5. Manual UAT `docs/sdlc/06-manual-tests/SPRINT-042-manual.md` covers S1–S4.

## Test notes

- Browser: two sessions, DevTools Application → clear storage for first-login case  
- Force 401: shorten access TTL in dev or delete cookie/token mid-session  
- CSS: short viewport height for scroll regression  
- API: document scope field on list + upload  

## Dependencies

- SPRINT-003 Auth, SPRINT-015 Tenant KM/scope, SPRINT-020 customer auth patterns (session redirect parity where shared)  
- packages: `apps/tenant-web`, `internal/store` / KM handlers, optional shared auth helpers  

## Notes

Roadmap **#42** · Quality phase **Q** · bug-fix-only sprint — no feature creep from S43.
