# Feature: Tenant Set Scope and KM   (FEAT-0015)
**Sprint:** SPRINT-015   **Owner:** DEV   **Status:** shipped (v1.6.0)

## Problem

Sprint 2 shipped **backend** KM ingest, ClickHouse embeddings, and scope-filtered RAG. Operators still manage knowledge via platform-oriented `/api/km/*` or seeds. **Active tenants** need a self-service portal to upload FAQs, assign **scopes** (`general` / `billing` / `technical`), inspect documents per agent, and delete or reset knowledge — without platform admin access.

## Scope

**In:**
- Tenant-admin APIs under `/api/tenant/km/*` (JWT `tenant_admin` + active tenant)
- List agents (workforce) with KM document counts for **this** tenant
- Upload document (multipart) with explicit `scope`
- List documents per agent; show status, chunk count, scope, version
- Update document scope (re-tag without full re-upload when feasible, or re-ingest)
- Delete one document (Postgres + MinIO object + ClickHouse embeddings)
- Reset all knowledge for one agent (tenant-scoped)
- Tenant UI `/tenant/km` — agent tabs, upload, scope picker, delete/reset
- **`km_gaps` table** — record questions with `missing_km` (FAQ backlog); list/update status for tenant
- Reuse S2 pipeline (`internal/km`), S13 KM quota/rate limits
- Design pack + manual UAT

**Out:**
- Auto FAQ extraction from websites / PDF OCR pipeline (future)
- Approval workflows / multi-reviewer publish
- Locale & call-time limits (**SPRINT-016**)
- Test & preview sandbox (**SPRINT-017**)
- Full conversation records / analytics dashboards (**SPRINT-022** — S15 only stores `km_gaps` backlog)
- Changing global hard-coded agent→scope matrix in retrieval (document-level `km_scope` remains the control; display agent defaults as help)
- Platform admin KM redesign (keep seed for ops)

## Acceptance criteria

1. Active `tenant_admin` can open `/tenant/km`, select an assigned agent, and upload a Markdown/text document with scope `general|billing|technical`.
2. Document appears in list with status ready/chunk count after ingest; chat RAG for that tenant+agent+topic uses the new chunks.
3. Document from another tenant is never listed or deletable.
4. Delete removes PG rows, MinIO object, and ClickHouse embeddings for that document.
5. Reset agent clears all KM for that agent under the tenant only.
6. KM quota (max documents / rate) returns clear 429/403 with existing quota codes.
7. Inactive / non-admin callers receive 401/403.
8. Manual UAT checklist exists under `06-manual-tests/SPRINT-015-manual.md`.
9. When RAG returns no chunks, a row is upserted into **`km_gaps`** (deduped by question hash); tenant can list/patch gaps via API (and UI panel).

## Test notes

- Unit: delete cascades; tenant isolation on list/delete; question hash normalize
- Integration: upload → list → chat citation path (optional); missing_km → km_gaps row
- Manual: tenant A upload; tenant B cannot see; delete; quota edge; open gaps after unknown question

## Dependencies

- SPRINT-002 KM + scope RAG
- SPRINT-006/007 tenant active
- SPRINT-005 avatar assignment (agent list for tenant)
- SPRINT-013 KM quota counters
- Design: [18-tenant-scope-km-spec.md](../02-design/18-tenant-scope-km-spec.md) (DES-0018)
- API: [04-api-spec.md](../02-design/04-api-spec.md) § Tenant KM / Scope
- UX: [05-ux-ui.md](../02-design/05-ux-ui.md) § T8
- Workflow: [02-workflow.md](../02-design/02-workflow.md) §40–44

## Links

- Sprint: [SPRINT-015](../03-sprints/SPRINT-015.md)
- Prior feature: [FEAT-0002](FEAT-0002-km-scope-rag.md)
- KM ops: [KM_SETUP.md](../../KM_SETUP.md)
