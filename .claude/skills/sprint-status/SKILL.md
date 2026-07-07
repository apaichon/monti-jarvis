---
name: sprint-status
description: Report progress on the current Monti Jarvis sprint — compare SPRINT-NNN.md commitment table against git state and task doc status. Use for standups, mid-sprint checks, or before release. (PM agent)
---

# sprint-status — current-sprint progress report

## Procedure
1. Run **`km-context`** for the active sprint.
2. Cross-check against reality:
   - `git log --oneline -10` and `git status`
   - Task frontmatter `status` in `docs/sdlc/04-tasks/`
   - `go test ./...` / `make build` if implementation claimed done
3. Emit the report:
   ```
   SPRINT-NNN — <goal>            <today>
   ✅ completed   : TASK-0006, TASK-0007
   🟡 in progress : TASK-0008
   ⚪ todo         : TASK-0009
   Points: X / Y committed
   Risks: <one-liners from sprint doc>
   DoD remaining: <list>
   ```
4. Run **`km-sync`** if you corrected any task statuses.

## Guardrails
- Read-only on code unless fixing doc drift.
- Flag work outside sprint scope as scope creep.

See `docs/sdlc/03-sprints/SPRINT-002.md`.