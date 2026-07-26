---
id: SPRINT-046-RETRO
type: sprint
title: "Retro - SPRINT-046"
status: done
owner: sdlc-orchestrator
created: 2026-07-25
updated: 2026-07-25
related: [SPRINT-046]
release: pending
---

# Retrospective - SPRINT-046

## Metrics

| Metric | Value |
| --- | ---: |
| Committed points | 17 |
| Completed points | 17 |
| Velocity | 17 |
| Completion | 100.0% |
| Carry-over | none |
| Risk closed/opened | 1 / 0 |

Sprint 46 was rebuilt to complete the carry-over and full referral reward
scope. Manual UAT is now refreshed and ready for a local fixture sign-off
before the pending v2.20.0 release cut.

## Per-role delivery

| Role | Points | Tasks |
| --- | ---: | --- |
| dev | 10 | TASK-0169, TASK-0175 |
| devops/dev/tester | 7 | TASK-0166, TASK-0167, TASK-0168 |

## What went well

- Referral code ownership, signup attribution, qualification gates, and the
  append-only event trail were delivered as a tenant-scoped foundation.
- Idempotency, self-referral, duplicate, circular, disabled-code, and payment
  qualification rules are covered by automated verification.
- Monti email templates now share a branded, inline-email-safe layout with the
  Monti logo and a prominent OTP hierarchy.

## What did not go well

- The original Sprint 46 close deferred the bonus-quota scope and manual UAT.
- Manual UAT was not executed during the rebuild and remains the release gate.
- Rebuilding the sprint made the bonus layer an explicit acceptance item while
  keeping purchased entitlements and bonus entitlements separate.

## Action items

- [ ] Execute the refreshed Sprint 46 referral manual UAT and attach evidence
      before the v2.20.0 release cut.
