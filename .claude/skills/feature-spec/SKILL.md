---
name: feature-spec
description: Scaffold a feature specification for Monti Jarvis — problem statement, scope, numbered testable acceptance criteria, test notes. Use when turning a backlog idea into a buildable, testable feature. (PM agent)
---

# feature-spec — write a buildable feature spec

## Procedure
1. Run **`km-context`** to check for related features in `docs/sdlc/01-features/`.
2. Draft the spec using the template below. ACs must be **independently testable**.
3. Save to `docs/sdlc/01-features/FEAT-NNNN-<slug>.md`.
4. Link from the active sprint commitment table.
5. Run **`km-sync`** to register the feature link.

## Template
```markdown
# Feature: <title>   (FEAT-NNNN)
**Sprint:** SPRINT-NNN   **Owner:** DEV

## Problem
<who needs this, what pain, why now>

## Scope
In:  <bullets>
Out: <bullets>

## Acceptance criteria
1. <falsifiable statement>
2. …

## Test notes
- Functional: <how to exercise each AC>
- KM/voice/chat: cite docs/KM_SETUP.md where relevant
- Languages: Thai + English where user-facing.

## Dependencies
- packages: internal/...
- blueprint: docs/monti_multi_tenant_ai_call_center_blueprint.md
```

## Guardrails
- No code, no effort estimates — that's DEV/sprint-plan.
- An AC that can't be tested is not an AC.

See `docs/sdlc/01-features/FEAT-0002-km-scope-rag.md`.