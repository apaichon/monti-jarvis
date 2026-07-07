---
name: doc-audit
description: Audit Monti Jarvis SDLC docs for consistency — sprint/task/feature links, roadmap alignment, stale status, broken paths. Use at sprint close or on request. Read-mostly; writes sprint/ROADMAP/AGENTS fixes.
---

# doc-audit — SDLC consistency check

Run at sprint close or when docs feel stale.

## Procedure

1. **Enumerate SDLC artifacts:**
   ```bash
   ls docs/sdlc/03-sprints/ docs/sdlc/04-tasks/ docs/sdlc/01-features/
   ```

2. **Sprint checks:**
   - Exactly one `status: in_progress` sprint (or zero between sprints).
   - Commitment table statuses match task frontmatter.
   - `ROADMAP.md` "Current sprint" points to the right file.
   - `_velocity.json` includes closed sprints.

3. **Task checks:**
   - Every task in sprint table has a `docs/sdlc/04-tasks/TASK-NNNN.md` file.
   - Frontmatter has `id`, `status`, `sprint`, `owner`, `points`.

4. **Feature checks:**
   - Sprint-linked features exist under `docs/sdlc/01-features/`.
   - ACs in feature spec are referenced in tasks where appropriate.

5. **Design docs:** `02-design/` has architecture, workflow, er-diagram, api-spec, ux-ui.

6. **Test & release docs:**
   - `05-test-scenarios/TEST-MATRIX.md` covers active sprint ACs
   - `06-manual-tests/SPRINT-NNN-manual.md` exists for closed/in-progress sprint
   - `07-deployment/LOCAL-DEV.md` matches Makefile targets
   - `08-readiness/RELEASE-READINESS.md` sprint section matches active sprint

7. **Link check:** markdown links to `docs/sdlc/*`, `docs/KM_SETUP.md`, blueprint resolve.

8. **AGENTS.md:** skills and current sprint line match reality.

9. **Report:**
   ```
   doc-audit  <date>
   sprint drift     : ...
   task mismatches  : ...
   broken links     : ...
   → proposed fixes : ...
   ```

10. **Reconcile (on approval):** apply fixes via `doc-control` / `km-sync`.

## Guardrails
- Never modify `.claude/agents|skills` bodies except when roster update requested.
- Report first; reconcile when asked.

Pairs with `doc-control`.