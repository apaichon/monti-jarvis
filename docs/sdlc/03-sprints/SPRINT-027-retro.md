---
id: SPRINT-027-RETRO
type: sprint
title: "Retro - SPRINT-027"
status: done
owner: sdlc-orchestrator
created: 2026-07-16
updated: 2026-07-16
related: [SPRINT-027, FEAT-0029]
release: v2.8.0
---

# Retrospective - SPRINT-027

## Metrics

| Metric | Value |
| --- | ---: |
| Committed points | 16 |
| Completed points | 16 |
| Velocity | 16 |
| Completion | 100% |
| Carry-over | 0 implementation points |
| Risk closed/opened | 4 / 0 |

## Per-role delivery

| Role | Points | Tasks |
| --- | ---: | --- |
| dev | 12 | TASK-0124, TASK-0125, TASK-0126 |
| tester | 4 | TASK-0127 |

## What went well

- The mobile facade reused existing tenant, customer, avatar, quota, and call-session policy instead of creating a parallel authorization path.
- Public brand discovery and mobile lifecycle contracts were delivered together with JSON-only API fallbacks and typed SDK support.
- Go tests, vet, server build, and standalone SDK compilation passed before release.

## What did not go as well

- APNs/FCM delivery remains a follow-up; the API accepts safe notification metadata but reports provider delivery as `not_configured`.
- End-to-end mobile device and clean-infrastructure smoke testing remains a deployment follow-up.

## Action items

- [ ] Add APNs/FCM provider adapters and device-level mobile smoke coverage in a future mobile sprint.
