---
id: SPRINT-041
status: in_progress
start: 2026-08-22
end: 2026-08-28
updated: 2026-07-26
design_pack: review_pending
release_target: v2.22.0
roadmap_sprint: 41
feature: FEAT-0041
platform: Security / Platform
depends_on: [SPRINT-019, SPRINT-020, SPRINT-032, SPRINT-033]
risk: high
---

# SPRINT-041 — AI Call-Center Security Hardening

## Goal

Harden the browser, runtime configuration, database access, and tenant
boundaries so a compromised client, leaked read credential, malformed query,
or authorization regression cannot cross the intended security boundary.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed slices (S42, S45, S46) | 12, 13, 17 → **avg 14.0** |
| **Committed** | **14** |

The commitment matches the trailing average but keeps implementation bounded
to defense-in-depth controls and verification. Provider, network, and IAM
procurement remain outside the sprint.

## Commitment

| Task | Scope | Points | Owner | Status |
| --- | --- | ---: | --- | --- |
| [TASK-0176](../04-tasks/TASK-0176.md) | Browser session-storage migration and Web Crypto boundary | 3 | dev | in_progress |
| [TASK-0177](../04-tasks/TASK-0177.md) | Environment, secret validation, redaction, and rotation contract | 3 | devops | in_progress |
| [TASK-0178](../04-tasks/TASK-0178.md) | Dedicated PostgreSQL read-only role and read pool | 3 | devops | in_progress |
| [TASK-0179](../04-tasks/TASK-0179.md) | Tenant-scoped policy, query allowlists, and isolation enforcement | 4 | dev | in_progress |
| [TASK-0180](../04-tasks/TASK-0180.md) | Threat-model regression suite and two-tenant UAT evidence | 1 | tester | in_progress |
| **Total** | | **14** | | |

## Scope boundary

### In

- Remove plaintext persistent session credentials from tenant and customer web
  storage; preserve safe return-from-checkout behavior.
- Fail closed for unsafe production configuration and redact startup/health
  diagnostics.
- Route selected AI/reporting reads through a least-privilege read pool.
- Apply transaction-local tenant context, PostgreSQL policies where applicable,
  parameterized queries, and bounded query controls.
- Coordinate API capability users: AI conversation KM retrieval uses
  `monti_ai_km_ro`; confirmed ticket creation uses `monti_ticket_rw`; neither
  principal can cross tenant policy boundaries.
- Verify two-tenant denial, read-role denial, migration/logout cleanup, and
  secret absence from browser bundles and logs.

### Out

- New customer-facing product behavior or new AI provider.
- Full database migration to row-level security for every historical table.
- Network/WAF redesign, managed KMS rollout, or enterprise SSO replacement.
- Changing quota, billing, referral, or package authority.

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | `proposed` | [FEAT-0041](../01-features/FEAT-0041-ai-call-center-security-hardening.md) |
| Deep spec | `review_pending` | [DES-0044](../02-design/44-ai-call-center-security-hardening-spec.md) |
| Workflow | `review_pending` | [02-workflow.md](../02-design/02-workflow.md) §109–112 |
| ER / policy map | `review_pending` | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 41 |
| API | `review_pending` | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 41 |
| UX/UI | `review_pending` | [05-ux-ui.md](../02-design/05-ux-ui.md) S41 |

Implementation is gated on approval of the deep spec, API contract, and
database-role/RLS policy design.

## Verification target

```bash
go test ./... -count=1
go vet ./...
make test
make build
cd apps/tenant-web && npm run check && npm run build
cd ../customer-web && npm run check && npm run build
git diff --check
# production config rejects missing/weak secrets and unsafe database roles
# browser migration removes plaintext legacy session keys
# read-only role cannot write or DDL
# two tenants cannot read each other's scoped records
# pooled connections do not retain tenant context between requests
# security posture diagnostics contain no secret values
```

Manual checklist: [SPRINT-041-security-manual.md](../06-manual-tests/SPRINT-041-security-manual.md)

## Risks

| Risk | Mitigation |
| --- | --- |
| Cookie/session migration logs users out | Versioned migration, explicit fallback window, and logout cleanup UAT |
| Read pool misses a write path | Start with explicit read-only call sites and assert role privileges in CI |
| RLS context leaks across pooled requests | `SET LOCAL` only inside transactions, reset/rollback tests, and deny-by-default policy |
| Existing public embed breaks | Preserve current public behavior and test `AUTH_DISABLED`/public paths separately |
| Security checks leak secrets | Metadata-only diagnostics, redaction tests, and bundle/log scans |

## Release gate

Target **v2.22.0** after design approval, automated verification, two-tenant
manual UAT, and security-owner sign-off. No production rollout is implied by
this planning document.
