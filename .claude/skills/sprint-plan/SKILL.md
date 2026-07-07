---
name: sprint-plan
description: Plan or groom a sprint for Monti Jarvis — read the roadmap, scope a sprint goal, write docs/sdlc/03-sprints/SPRINT-NNN.md with tasks and acceptance criteria, and create linked TASK-NNNN.md files. Use when starting a new iteration or re-grooming the active one. (PM agent)
---

# sprint-plan — groom and open a sprint

## Procedure
1. Run **`km-context`** (previous sprint) to see what carried over.
2. Read `docs/sdlc/00-roadmap/ROADMAP.md`. Pick the **next roadmap sprint** — do not skip ahead
   without explicit user approval.
3. Read completed tasks in `docs/sdlc/04-tasks/` and `docs/sdlc/03-sprints/_velocity.json`.
4. Create `docs/sdlc/03-sprints/SPRINT-NNN.md` from the structure below (NNN = roadmap number).
5. Create linked `docs/sdlc/04-tasks/TASK-NNNN.md` and `docs/sdlc/01-features/FEAT-NNNN-*.md`.
6. Run **`sprint-tech-specs`** — workflow, er-diagram, api-spec, ux-ui ASCII + mapping, and `<slug>-spec.md` if needed.
7. Update `docs/sdlc/00-roadmap/ROADMAP.md` "Current sprint" / "Next sprint" sections.
8. Run **`km-sync`** to set sprint `in_progress` and tasks `todo`.

## sprint file structure
```markdown
---
id: SPRINT-NNN
status: in_progress
start: YYYY-MM-DD
end: YYYY-MM-DD
goal: "..."
roadmap_sprint: N
platform: Customer|Backend|...
depends_on: [SPRINT-NNN]
---

# SPRINT-NNN — <title>

## Goal
<one sentence>

## Commitment
| Task | Points | Status | Owner | Outcome |

## Scope boundary
In / Out

## Verification
make test, manual checks

## Risks
```

## Guardrails
- Only one sprint is `in_progress` at a time.
- Every task must name an owner (dev/devops/tester) and testable ACs.
- Defer auth/KYC/ticketing unless explicitly in roadmap scope.
- Use **`sprint-tech-specs`** for all `02-design/` updates — do not hand-edit specs ad hoc during sprint-plan.

See `docs/sdlc/README.md`, `docs/sdlc/00-roadmap/ROADMAP.md`, `docs/sdlc/02-design/`, **`sprint-tech-specs`**.