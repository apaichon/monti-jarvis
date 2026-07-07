---
name: obsidian-graph
description: Optional — build an Obsidian-friendly view of Monti Jarvis SDLC links (sprint → task → feature → packages). Use when the team wants a visual graph of docs/sdlc relationships. Owner doc-manager.
---

# obsidian-graph — SDLC relationship vault (optional)

Monti Jarvis primary source of truth is **markdown under `docs/sdlc/`**, not a Postgres
knowledge graph. This skill generates an optional Obsidian vault for visual navigation.

## Vault layout (`docs/sdlc/05-obsidian/` — if created)
```
docs/sdlc/05-obsidian/
  README.md
  00-MOC.md
  03-sprints/  # one note per SPRINT-NNN
  04-tasks/    # one note per TASK-NNNN
  01-features/ # one note per FEAT-NNNN
  02-design/   # 01-architecture … NN-<slug>-spec (ordered)
  packages/    # internal/* and apps/customer-web
```

## Node template
```markdown
# TASK-0008

implements:: [[FEAT-0002-km-scope-rag]]
belongs_to:: [[SPRINT-002]]
touches:: [[internal/km]], [[internal/rag]]
```

## Procedure
1. Scan `docs/sdlc/03-sprints/`, `04-tasks/`, `01-features/`, `02-design/`.
2. Emit wikilink notes for each artifact.
3. Add `implements`, `belongs_to`, `touches` edges from task content.
4. Regenerate only the changed slice on sprint updates.

## Guardrails
- Generated view only — `docs/sdlc/*.md` stays authoritative.
- Skip until explicitly requested; not required for day-to-day agent work.

See `km-context` / `km-sync` for the primary SDLC workflow.