---
id: SPRINT-020
status: completed
start: 2026-07-13
end: 2026-07-14
closed: 2026-07-13
updated: 2026-07-13
design_pack: shipped
release_target: v2.1.0
release: v2.1.0
goal: "Customer: authenticate imported customer identities, enforce tenant domain policy, and prove quota/rate-limit isolation before production customer traffic."
roadmap_sprint: 20
platform: Customer
depends_on: [SPRINT-003, SPRINT-013, SPRINT-016, SPRINT-019]
---

# SPRINT-020 - Customer: Auth and Domain Enforcement

## Goal

Allow tenants to enable customer authentication on top of SPRINT-019 imported customer records, bind customer sessions to tenant/tier/group context, and verify quota/rate-limit isolation under authenticated multi-user load.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S17-S19) | 16, 16, 16 -> **avg 16** |
| **Commitment** | **16** |
| **Completed** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0092](../04-tasks/TASK-0092.md) | 3 | completed | devops | Customer OTP identity/session schema and auth settings |
| [TASK-0093](../04-tasks/TASK-0093.md) | 5 | completed | dev | Customer OTP request/verify, session, claim, and profile APIs |
| [TASK-0094](../04-tasks/TASK-0094.md) | 3 | completed | dev | Tenant customer-auth configuration UI |
| [TASK-0095](../04-tasks/TASK-0095.md) | 4 | completed | dev | Customer login/account UX and authenticated context wiring |
| [TASK-0096](../04-tasks/TASK-0096.md) | 1 | completed | tester | Authenticated tenant smoke, manual checklist, and production gate evidence |

**Committed:** 16 points · **Completed:** 16 points · **Completion:** 100%

## Shipped summary (v2.1.0)

| Area | Outcome |
| --- | --- |
| Data | Tenant-scoped customer auth settings, OTP challenges, auth identities, customer sessions, and auth events |
| APIs | Customer email OTP request/verify, refresh, logout, current profile, and tenant customer-auth settings |
| UI | Customer portal OTP sign-in/account state, tenant-context URL support, avatar picker popup, and tenant settings auth controls |
| Integration | Authenticated customer tokens carry tenant/customer role context into chat and call creation |
| Verification | Full Go tests/build, customer/tenant web checks/builds, browser OTP smoke on Libra Tech tenant, and S20 manual checklist |

## Scope boundary

**In**

- Customer OTP identities and sessions linked to tenant-scoped customer records.
- Customer email OTP request/verify, logout, session refresh, and current-profile APIs.
- Tenant auth settings and domain-rule enforcement using SPRINT-019 domain rules.
- Customer portal login/account state while preserving no-auth conversation access.
- Quota/rate-limit/tier/group isolation checks for authenticated customer traffic.

**Out**

- Customer OAuth/social login.
- Password login, password reset, magic links, invitation workflow, and customer email preference management.
- Customer tickets, conversation history, KYC, discounts, or billing changes.
- Production customer launch until readiness gate is signed off.

## Feature

- [FEAT-0022 - Customer Authentication and Domain Enforcement](../01-features/FEAT-0022-customer-auth.md)

## Design pack

| Artifact | Path | Status |
| --- | --- | --- |
| Deep spec | [23-customer-auth-spec.md](../02-design/23-customer-auth-spec.md) | `shipped` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §59–63 | `shipped` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) § Sprint 20 | `shipped` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) § Customer Authentication & Domain Enforcement | `shipped` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) § T13 | `shipped` |

> **Gate:** Approved by user instruction to close and release on 2026-07-13 after browser OTP/tenant smoke checks and automated validation.

## Verification

```bash
make test && make build
# Tenant admin enables customer auth and domain enforcement
# Customer login binds to a tenant customer, tier, groups, quota, and rate limits
# Cross-tenant customer/session/quota access returns 401/403/404 without leakage
# Manual UAT: docs/sdlc/06-manual-tests/SPRINT-020-manual.md
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Credential data leaks across tenants | Tenant id from session/JWT; cross-tenant customer ids return 404 |
| Domain policy blocks valid customers | Tenant setting controls enforcement; allow/deny behavior covered by UAT |
| Public no-auth portal regresses | Preserve public conversation path and add regression tests |
| Quota/rate-limit attribution is wrong | Multi-user UAT verifies tenant, tier, group, and package counters |
| Weak production gate | S20 cannot close until readiness records auth plus quota/rate-limit isolation |

## Production launch gate

SPRINT-020 produced customer-authenticated tenant smoke evidence and a manual UAT checklist. Before broad production customer traffic, run the documented multi-session/load variant to re-confirm package quota, API rate-limit, operational call caps, tier overrides, and tenant isolation in the target deployment.

## Links

- Depends: [SPRINT-003](SPRINT-003.md), [SPRINT-013](SPRINT-013.md), [SPRINT-016](SPRINT-016.md), [SPRINT-019](SPRINT-019.md)
- Release: **v2.1.0**
