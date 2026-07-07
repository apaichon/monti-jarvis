---
name: doc-control
description: Stamp or refresh SDLC artifact metadata — sprint/task/feature frontmatter, ROADMAP pointers, AGENTS.md sprint line. Use when creating or editing docs under docs/sdlc/.
---

# doc-control — stamp one SDLC artifact

Run when you **create** or **edit** a sprint, task, or feature doc.

## Inputs
- Document path under `docs/sdlc/`
- Create vs update
- Status change if any

## Procedure

### Sprint (`docs/sdlc/03-sprints/SPRINT-NNN.md`)
```yaml
---
id: SPRINT-NNN
status: in_progress|completed
start: YYYY-MM-DD
end: YYYY-MM-DD
updated: YYYY-MM-DD
goal: "..."
roadmap_sprint: N
---
```
On close: add `closed`, `release: vX.Y.Z`.

### Task (`docs/sdlc/04-tasks/TASK-NNNN.md`)
```yaml
---
id: TASK-NNNN
title: "..."
status: todo|in_progress|completed
sprint: SPRINT-NNN
points: N
owner: dev|devops|tester
updated: YYYY-MM-DD
---
```

### After create/update
- Sync commitment table in sprint doc via `km-sync`.
- If sprint changed, update `docs/sdlc/00-roadmap/ROADMAP.md` and `AGENTS.md`.

## Date sourcing
- Prefer `git log -1 --format=%as -- <file>` for `updated` when reconciling.

## Guardrails
- Task ids are stable; don't renumber shipped tasks.
- Deprecation: mark `cancelled`, don't delete files.

Pairs with `doc-audit`.