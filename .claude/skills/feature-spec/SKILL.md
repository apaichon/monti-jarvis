---
name: feature-spec
description: Scaffold a feature specification for Monti Jarvis — problem statement, scope, numbered testable acceptance criteria, test notes, and feature nodes. Use when turning a backlog idea into a buildable, testable feature. (PM agent)
---

# feature-spec — write a buildable feature spec

## Procedure
1. Check `docs/sdlc/features/` for related/overlapping features — extend instead of duplicating.
2. Draft the spec using the template below. ACs must be **independently testable** and reference concrete behaviour.
3. Save to `docs/sdlc/features/FEAT-NNNN-<slug>.md` and link from the active sprint.

## Template
```markdown
# Feature: <title>   (feature:<key>)
**Sprint:** <current or backlog>   **Owner:** DEV

## Problem
<who needs this, what pain, why now>

## Scope
In:  <bullets>
Out: <bullets — link to backlog if deferred>

## Acceptance criteria
1. <falsifiable statement>
2. …

## Test notes
- Functional: <how to exercise each AC>
- Languages: Thai + English where user-facing.

## Dependencies
- components: <packages>   decisions: <ADRs if any>
```

## Guardrails
- No code, no effort estimates here — that's DEV/sprint-plan.
- An AC that can't be tested is not an AC. Rewrite or split it.

See `docs/sdlc/features/FEAT-0001-workforce-qa.md` for the first example.