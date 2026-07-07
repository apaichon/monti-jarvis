---
name: km-context
description: Load the relevant SDLC slice for the current sprint and task — active sprint doc, linked tasks, features, and roadmap context. Use at the START of any PM/DEV/Tester/DevOps task instead of reading the whole repo history.
---

# km-context — load the relevant SDLC slice

Run this **before** doing work. Monti Jarvis uses markdown SDLC docs (not a Postgres
knowledge graph yet).

## Inputs
- Current sprint (default: find `status: in_progress` in `docs/sdlc/03-sprints/`).
- The task / feature id if known (e.g. `TASK-0008`, `FEAT-0002`).

## Procedure
1. Read `docs/sdlc/00-roadmap/ROADMAP.md` — current and next sprint pointers.
2. Read the active `docs/sdlc/03-sprints/SPRINT-NNN.md` (goal, commitment, scope, risks).
3. Load linked artifacts:
   - `docs/sdlc/04-tasks/TASK-*.md` referenced in the commitment table
   - `docs/sdlc/01-features/FEAT-*.md` linked from the sprint
   - `docs/sdlc/03-sprints/_velocity.json` for past velocity
   - For VERIFY/release tasks: `05-test-scenarios/`, `06-manual-tests/`, `07-deployment/`, `08-readiness/`
4. If a specific task was given, read only that task + its feature + touched packages
   (grep `internal/`, `cmd/server/`, `apps/customer-web/` paths in the task).
5. Produce a briefing under ~40 lines:
   - Sprint goal + status
   - Task table (id, status, owner)
   - In-scope / out-of-scope reminders
   - Key files and APIs for this sprint (`docs/sdlc/02-design/api-spec.md`)

## Output
A markdown briefing. Do **not** dump full task bodies unless the task is the focus.

## Guardrails
- Never read all of `docs/sdlc/04-tasks/` without filtering to the active sprint.
- If sprint doc is missing, read `ROADMAP.md` and ask PM to run `sprint-plan`.

Pairs with `km-sync` (run at END of task to write status back).