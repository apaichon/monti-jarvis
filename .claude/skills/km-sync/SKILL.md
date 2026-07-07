---
name: km-sync
description: Propagate SDLC status changes across linked docs — update sprint/task/feature frontmatter, commitment tables, velocity, and ROADMAP pointers. Use at the END of PM/DEV/Tester/DevOps tasks.
---

# km-sync — persist SDLC status changes

Run this **after** doing work, to keep `docs/sdlc/` consistent.

## Inputs
- Nodes changed: sprint, task(s), feature(s).
- New status values: `todo` | `in_progress` | `completed` | `cancelled`.
- Optional: points completed, release version, close date.

## Procedure
1. **Update task frontmatter** in `docs/sdlc/04-tasks/TASK-NNNN.md`:
   ```yaml
   status: completed
   updated: YYYY-MM-DD
   ```
2. **Update sprint commitment table** in `docs/sdlc/03-sprints/SPRINT-NNN.md` to match.
3. **Sprint close** (when all tasks done):
   - Set sprint `status: completed`, `closed`, `release: vX.Y.Z`
   - Append entry to `docs/sdlc/03-sprints/_velocity.json`
   - Update `docs/sdlc/00-roadmap/ROADMAP.md` — mark sprint shipped, point "Current" to next
   - Update `AGENTS.md` current sprint line
4. **Feature link**: ensure `FEAT-NNNN` is referenced from sprint and tasks.
5. Confirm what changed: list file paths + new statuses.

## Status propagation rules
- Task `completed` → sprint table row `completed`
- Sprint `completed` → ROADMAP sprint index gets ✅ version note
- Never delete task/sprint history — mark `cancelled` instead

## Guardrails
- Only one sprint `in_progress` at a time.
- Do not mark tasks done without verification note when Tester gate applies.

Pairs with `km-context`. See `docs/sdlc/03-sprints/SPRINT-001.md` for a closed example.