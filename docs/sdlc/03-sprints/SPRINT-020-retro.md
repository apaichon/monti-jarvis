---
id: SPRINT-020-RETRO
type: sprint
title: Retro - SPRINT-020
status: done
owner: sdlc-orchestrator
created: 2026-07-13
updated: 2026-07-13
related: [SPRINT-020, FEAT-0022]
release: v2.1.0
---

# Retrospective - SPRINT-020

## Metrics

| Metric | Value |
| --- | ---: |
| Committed points | 16 |
| Completed points | 16 |
| Velocity | 16 |
| Completion | 100% |
| Carry-over | 0 |
| Risk closed/opened | 4 / 0 |

## Per-role delivery

| Role | Points | Tasks |
| --- | ---: | --- |
| devops | 3 | TASK-0092 |
| dev | 12 | TASK-0093, TASK-0094, TASK-0095 |
| tester | 1 | TASK-0096 |

## What went well

- Customer auth shipped as a complete vertical slice: schema, OTP APIs, tenant settings, customer portal UX, and authenticated tenant context for chat/calls.
- Browser smoke on the Libra Tech tenant caught the missing tenant context in the public customer portal before release.
- Usability feedback at 100% browser scale caught the hidden avatar-list problem; the popup selector fixed the release UX.

## What did not go as well

- Tenant context was implicit in the first implementation; customer portal testing against non-demo tenants exposed the gap late.
- Manual UAT documentation was created after implementation instead of before the first browser smoke.
- The sprint accumulated a broad dirty worktree, increasing release packaging and review overhead.

## Risk burn-down

| Risk | Result |
| --- | --- |
| Credential data leaks across tenants | Mitigated; customer auth records and sessions are tenant scoped |
| Domain policy blocks valid customers | Mitigated; tenant auth settings and domain rules are explicit and testable |
| Public no-auth portal regresses | Mitigated; no-auth path remains available and OTP sign-in is optional |
| Quota/rate-limit attribution is wrong | Mitigated for tenant routing; manual checklist retains deeper multi-session pre-production re-run |
| Weak production gate | Reduced; release evidence is recorded, and broader production traffic still requires target-environment UAT |

## Action items

- Add a first-class tenant selector or brand route for customer portal testing instead of relying only on `?tenant_id=...`.
- Automate customer OTP API tests with a fake mailer to remove manual email dependency.
- Keep UI smoke checks at 100% browser scale and narrow viewport before sprint close.
