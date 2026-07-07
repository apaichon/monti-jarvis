---
name: sprint-tech-specs
description: Produce or update per-sprint technical design artifacts for Monti Jarvis — workflow sequences, ER diagram deltas, API contract, UX/UI ASCII wireframes with screen→API mapping, and optional feature deep-spec. Use when opening a sprint, before implementation, on design review, or when the user asks for technical specs, workflow, er-diagram, api-spec, or ux-ui for a sprint. (SA/PM agent)
---

# sprint-tech-specs — per-sprint technical design pack

Run **after `sprint-plan`** (or when grooming design before build). Delivers the five linked artifacts DEV/Tester implement against.

## Inputs

- Sprint id: `SPRINT-NNN` (from `docs/sdlc/03-sprints/SPRINT-NNN.md`)
- Feature: `docs/sdlc/01-features/FEAT-NNNN-*.md`
- Roadmap row: `docs/sdlc/00-roadmap/ROADMAP.md`
- Prior sprint specs in `docs/sdlc/02-design/` (incremental update — do not delete shipped sections)

## Procedure

1. Run **`km-context`** for the active sprint + feature scope.
2. Read the feature spec **In/Out** and task ACs — specs must not exceed sprint boundary.
3. Update **five artifacts** (templates in `references/`):
   | # | File | Action |
   |---|------|--------|
   | 1 | `docs/sdlc/02-design/02-workflow.md` | Append **§ Sprint N — \<title\>** mermaid sequence(s) for new flows |
   | 2 | `docs/sdlc/02-design/03-er-diagram.md` | Add/extend entities, relationships, audit columns, future-entity table |
   | 3 | `docs/sdlc/02-design/04-api-spec.md` | Add endpoint tables, request/response JSON, RBAC, error codes |
   | 4 | `docs/sdlc/02-design/05-ux-ui.md` | ASCII wireframe(s) + **Screen map → API** table + flow diagrams |
   | 5 | `docs/sdlc/02-design/NN-<slug>-spec.md` | Deep spec when sprint introduces a new domain (`NN` = next DES id) |
4. **Cross-link** every artifact:
   - Sprint doc → links to all five
   - Feature spec → links to deep spec + api-spec
   - Each design doc footer → sibling docs
5. Run **`doc-control`** — refresh frontmatter (`sprint`, `updated`, `status`).
6. Update `docs/sdlc/02-design/README.md` index row for each touched file.
7. Run **`km-sync`** — note files + `status: approved|review_pending` in sprint doc **Design** section.

## Artifact rules

### Workflow (`02-workflow.md`)

- One **numbered mermaid `sequenceDiagram`** per user-visible or ops flow introduced this sprint.
- Name participants: `Browser`, `Go :8091`, `Postgres`, `Redis`, package names (`internal/auth`, `internal/km`, …).
- Include **error/alt branches** when RBAC or optional deps matter (`alt forbidden`, `alt AUTH_DISABLED`).
- Append **state tables** when new entities have lifecycle (`active` → `revoked`, etc.).
- Never renumber prior sections — append at end before the “See also” line.

### ER diagram (`03-er-diagram.md`)

- Every Postgres table shows **audit columns**: `created_at`, `updated_at`, `created_by`, `updated_by`.
- Use mermaid `erDiagram` for new tables + FK relationships.
- Update **Future entities** table — move shipped tables out of “future”.
- ClickHouse / Redis / MinIO sections when sprint touches analytics, cache keys, or object paths.
- Reference migration script name when schema ships (`scripts/migrations/00N_*.sql`).

### API spec (`04-api-spec.md`)

- Group by domain (`## Auth`, `## Packages`, …).
- Per endpoint: method, path, auth role, request fields table, response JSON example, error codes.
- Note `AUTH_DISABLED` / dev-bypass behavior when customer paths stay public.
- Bump `sprint` in `/healthz` example when sprint ships customer-visible version string.
- WebSocket/SSE: document query params, event shapes, client handshake order.

### UX/UI ASCII (`05-ux-ui.md`)

Required for every sprint — even API-only sprints document the **operator surface** (curl/REST client ASCII).

**Must include:**

1. **Screen map → API** table:

   | UI zone | User action | API / WS |
   | --- | --- | --- |
   | … | … | `METHOD /path` |

2. **Full layout** ASCII box diagram (desktop); mobile collapse section if layout changes.
3. **Flow A/B/C** ASCII step diagrams for primary interactions.
4. **Sprint N — \<title\>** section stating what changed vs prior sprint (e.g. “no customer UI change”).
5. **Component → file** table when new Svelte routes/components ship.

Use box-drawing chars (`┌─┐│└┘├┤`). Label zones (A1, B2, …) consistently with the screen map.

### Deep spec (`NN-<slug>-spec.md`)

Filename **`NN-`** prefix matches `DES-NNNN` (oldest→newest index). Create when the sprint adds a **new bounded domain** (not a one-line API tweak):

```yaml
---
id: DES-NNNN
title: <Domain> Specification
status: approved|review_pending|shipped
updated: YYYY-MM-DD
sprint: SPRINT-NNN
owner: SA
---
```

Sections: Goals · Non-goals · Env vars · Data model · API summary · RBAC · Verification curl block.

Assign next free `DES-NNNN` id (check existing frontmatter in `02-design/`).

## Status gates

| Status | Meaning |
| --- | --- |
| `review_pending` | Draft complete; awaiting user/PM sign-off before TASK implementation |
| `approved` | Signed off; DEV may implement |
| `shipped` | Sprint closed; spec matches tagged release |

Sprint **Design** section should list each artifact + status. Implementation starts only when deep spec (if any) + api-spec are `approved`.

## Platform-specific notes (Monti Jarvis)

- **Customer portal** (`apps/customer-web/`) stays no-auth unless sprint explicitly adds customer identity (Sprints 19–20).
- **API-only sprints** (auth, packages): UX doc shows terminal/REST-client ASCII, not a fake admin SPA.
- **Audit columns** on every new Postgres table; **Redis** keys use prefix `monti_jarvis:`.
- **Thai + English** labels in UX when user-facing copy is specified.

## Checklist (copy before closing)

```
[ ] 02-workflow.md — Sprint N section with ≥1 sequence diagram
[ ] 03-er-diagram.md — new entities + audit cols + future table updated
[ ] 04-api-spec.md — endpoints match NN-*-spec / feature ACs
[ ] 05-ux-ui.md — screen map → API + ASCII layout + flows
[ ] NN-<slug>-spec.md — created if domain warrants deep spec
[ ] 02-design/README.md index updated
[ ] SPRINT-NNN.md Design links + statuses
[ ] km-sync run
```

## Guardrails

- Specs are **incremental** — preserve shipped Sprint 1–N content; append or extend.
- Do not invent endpoints/tasks not in the sprint commitment table.
- An endpoint in api-spec must appear in workflow or ux-ui mapping (or document as ops-only).
- ER entities must match Go `internal/store` DDL or migration scripts — no orphan tables.
- Prefer linking to blueprint `docs/monti_multi_tenant_ai_call_center_blueprint.md` for business context, not duplicating it.

## References

- Templates: `.claude/skills/sprint-tech-specs/references/`
- Examples: `06-auth-spec.md`, `08-packages-spec.md`, `02-workflow.md` §6–7, `05-ux-ui.md` § Sprint 3

Pairs with **`sprint-plan`** (before) and **`manual-test-doc`** (after implementation).