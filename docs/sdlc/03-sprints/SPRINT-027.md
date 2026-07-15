---
id: SPRINT-027
status: in_progress
start: 2026-07-15
end: 2026-07-16
updated: 2026-07-15
design_pack: approved
release_target: v2.8.0
goal: "Customer / Integrator: expose a stable mobile call API and typed SDK for inbound AI voice integration."
roadmap_sprint: 27
platform: Customer / Integrator
depends_on: [SPRINT-001, SPRINT-020]
---

# SPRINT-027 - Mobile Call API and SDK

## Goal

Give mobile integrators a versioned, tenant-safe contract for starting, operating, observing, rating, and ending inbound AI voice calls without coupling them to the web embed surface.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S24, S25, S26) | 16, 16, 16 -> **avg 16** |
| **Proposed commitment** | **16** |

## Proposed commitment

The roadmap-level work is committed at 16 points. Concrete TASK IDs will be decomposed and linked by the approved technical-spec pack before implementation begins.

| Work package | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| Versioned mobile call API contract | 4 | completed | dev | Authenticated session creation, tenant/avatar selection, call status, transcript events, end-call, and rating schemas |
| Mobile voice transport and lifecycle | 4 | completed | dev | Mobile-safe session handshake, reconnect behavior, audio permission/lifecycle contract, and bounded failure states |
| Typed SDK core and adapter decision | 4 | completed | dev | SDK lifecycle client with token refresh and callbacks, plus a documented native/React Native/Flutter/layered target decision |
| Policy enforcement, sample integration, and verification | 4 | completed | devops/tester | Tenant isolation, avatar assignment, auth/quota/rate limits, reference integration, compatibility docs, and contract tests |

## Design

The Sprint 27 technical design pack is approved for implementation. The build is decomposed into four task slices below.

| Artifact | Planned scope | Status |
| --- | --- | --- |
| Feature | [FEAT-0029 - Mobile Call API and SDK](../01-features/FEAT-0029-mobile-call-api-sdk.md) | `in_progress` |
| Deep spec | [30-mobile-call-api-sdk-spec.md](../02-design/30-mobile-call-api-sdk-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) Sprint 27 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 27 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Mobile Call API and SDK | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) Sprint 27 | `approved` |

## Implementation tasks

| Task | Scope | Points | Status |
| --- | --- | ---: | --- |
| [TASK-0124](../04-tasks/TASK-0124.md) | Versioned mobile API and bootstrap | 4 | `in_progress` |
| [TASK-0125](../04-tasks/TASK-0125.md) | Mobile voice WebSocket and lifecycle | 4 | `in_progress` |
| [TASK-0126](../04-tasks/TASK-0126.md) | Typed SDK core and adapters | 4 | `in_progress` |
| [TASK-0127](../04-tasks/TASK-0127.md) | Policy, contract tests, and reference integration | 4 | `in_progress` |

## Scope boundary

**In**

- Versioned mobile API schemas for authenticated call-session creation and lifecycle control.
- Tenant and avatar selection constrained by existing assignment, authentication, quota, rate-limit, and isolation rules.
- Transcript and call-status event delivery suitable for mobile clients.
- Explicit end-call and customer rating operations with stable client-facing errors.
- Mobile voice transport handshake, reconnect, audio permission/lifecycle guidance, and bounded failure behavior.
- Typed SDK surface for the selected mobile integration target and a small reference integration.
- Provider credential protection, raw infrastructure error redaction, compatibility documentation, and automated contract tests.

**Out**

- A public provider SDK or direct Gemini/LiveKit credential exposure.
- Replacing the existing web embed API or customer-web call flow.
- Native production applications for every mobile platform in one sprint.
- New billing, quota, identity, or avatar-assignment policy models.
- Cross-tenant administration, audit logs, platform monitoring, or analytics dashboards.

## Technical-spec gates

The Sprint 27 design pack must resolve these decisions before implementation:

1. Select native iOS/Android, React Native, Flutter, or a layered core-plus-adapters SDK shape.
2. Define authentication and token-refresh boundaries without exposing tenant or provider secrets.
3. Define the versioned API, WebSocket/session handshake, event envelopes, reconnect behavior, and idempotent end-call contract.
4. Map existing tenant/avatar assignment, customer auth, quota, rate-limit, and tenant-isolation checks into the mobile path.
5. Define reference integration, compatibility matrix, migration guidance, and contract/UAT coverage.

## Verification

```bash
make test
make build
# API contract tests cover authentication, tenant/avatar policy, lifecycle, rating, and redaction.
# Mobile session reconnect and explicit end-call behavior are deterministic and idempotent.
# Tenant A cannot create or observe a call for Tenant B.
# Provider credentials and raw infrastructure errors never cross the mobile API boundary.
# Reference integration follows the published SDK contract.
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Mobile audio and reconnect behavior differs by platform | Keep the transport contract explicit and test adapters against bounded lifecycle states |
| SDK couples callers to provider implementation details | Expose only versioned Monti session/event models and redact provider failures |
| Existing tenant policies are bypassed by a new entry point | Reuse the existing auth, avatar assignment, quota, rate-limit, and scope checks |
| Platform choice expands the sprint beyond 16 points | Decide the target shape in the design pack and keep other adapters out of this sprint |

## Links

- Depends: [SPRINT-001](SPRINT-001.md), [SPRINT-020](SPRINT-020.md)
- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Sprint 27
- Target: **v2.8.0**
- Next backlog: Sprint 28 cross-tenant audit log
