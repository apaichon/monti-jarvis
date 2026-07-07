---
name: manual-test-doc
description: Generate a step-by-step manual test checklist for a Monti Jarvis sprint — init infra, seed KM, per-scenario steps with expected results and pass/fail boxes. Use at sprint VERIFY/sign-off. Owner Tester.
---

# manual-test-doc — sprint UAT checklist

Run **after** implementation and `go test ./...` pass, **before** sprint close.

## Inputs
- `docs/sdlc/03-sprints/SPRINT-NNN.md` — tasks + ACs
- Implementation surface: `Makefile`, `docs/KM_SETUP.md`, new API routes
- Portal URL: `http://localhost:8091`

## Procedure
1. Read sprint ACs and real commands (`make up`, `make km-seed`, curl examples).
2. Write `docs/sdlc/06-manual-tests/SPRINT-NNN-manual.md` from the template below.
3. Add scenario rows to `docs/sdlc/05-test-scenarios/TEST-MATRIX.md` if new ACs appear.
3. Run **`km-sync`** to link the manual test doc from the sprint verification section.

## Template
```markdown
# SPRINT-NNN — Manual Test Checklist

## 0. Preconditions
Tools, GEMINI_API_KEY, branch.

## 1. Init infrastructure
make up / infra-check steps with [ ] boxes.

## 2. Prepare data
make km-seed, per-agent upload examples.

## 3. Scenarios
### S1 — <objective> (TASK-NNNN · AC n)
Steps, expected, [ ] result.

## 4. Teardown
make down

## 5. Sign-off
| Tester | Date | Result | Defects |
```

## Guardrails
- Steps must be executable as written — real ports, routes, make targets.
- Every AC maps to at least one scenario.
- Failing step → defect task; sprint not signed off until green.

See `docs/KM_SETUP.md`, `docs/sdlc/07-deployment/LOCAL-DEV.md`, and `docs/sdlc/08-readiness/RELEASE-READINESS.md`.