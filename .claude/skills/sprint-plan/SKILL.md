---
name: sprint-plan
description: Plan or groom a sprint for Monti Jarvis — scope a sprint goal, write docs/sdlc/sprints/SPRINT-NNN.md with tasks and acceptance criteria. Use when starting a new iteration or re-grooming the active one. (PM agent)
---

# sprint-plan — groom and open a sprint

## Procedure
1. Read completed tasks in `docs/sdlc/tasks/` and note carry-over items.
2. Pick a coherent **sprint goal** (one sentence) and the smallest set of features that achieve it (~5–12 points for greenfield).
3. Create `docs/sdlc/sprints/SPRINT-NNN.md` from the structure below.
4. Create linked `docs/sdlc/tasks/TASK-NNNN.md` files with testable ACs.

## sprint file structure
```markdown
# Sprint NNN — <title>   (<start> → <end>)
**Goal:** <one sentence>

## Commitment
| Task | Points | Outcome |
| --- | ---: | --- |

## Scope Boundary
- In scope: …
- Out of scope: …

## Verification
- `go test ./...`
- manual checks …

## Risks
- …
```

## Guardrails
- Only one sprint is "active" at a time.
- Every task must name an owner and testable ACs.
- Defer auth/KYC/ticketing/CRM unless explicitly in scope.

See `docs/sdlc/sprints/SPRINT-001.md`.