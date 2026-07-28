---
id: UAT-041
title: "Sprint 41 security hardening manual UAT"
status: ready
sprint: SPRINT-041
updated: 2026-07-26
---

# SPRINT-041 Security Hardening Manual UAT

## Purpose

Verify that browser session migration, production configuration checks,
read-only database access, secret redaction, and tenant isolation hold across
two isolated tenants without using production credentials or real customer
data.

## Preconditions

- [ ] Use a disposable local/staging database and two test tenants.
- [ ] Capture sanitized command output only; never record tokens, cookies,
      credentials, prompts, SQL, or customer data.
- [ ] Configure the writer and read-only database roles separately.
- [ ] Record build/version, timestamp, environment, and tester.

## Scenarios

| ID | Scenario | Expected result | Evidence |
| --- | --- | --- | --- |
| U41-01 | Login/refresh with secure cookie policy | Access works; refresh credential is not in localStorage/sessionStorage or response body | [ ] |
| U41-02 | Legacy plaintext session key migration | Keys are removed and user is safely re-authenticated; no infinite redirect | [ ] |
| U41-03 | Logout, expiry, tenant switch | Current and legacy keys clear; server session is revoked where supported | [ ] |
| U41-04 | Production missing/weak `JWT_SECRET` | Startup/readiness fails closed with a remediation code and no secret value | [ ] |
| U41-05 | Unsafe CORS/cookie/dev bypass | Production rejects the configuration; development reports degraded only | [ ] |
| U41-06 | Read-only role DML/DDL attempt | INSERT/UPDATE/DELETE/CREATE fail; approved SELECT succeeds | [ ] |
| U41-07 | Tenant A reads own data | Only Tenant A rows are returned | [ ] |
| U41-08 | Tenant A requests Tenant B data | Safe deny/empty response; no B existence or data leakage | [ ] |
| U41-09 | Pooled connection reuse | Tenant context does not survive commit/rollback into the next request | [ ] |
| U41-10 | Posture/readiness diagnostics | Only booleans/codes appear; no credentials, URLs, SQL, prompts, or tenant IDs | [ ] |
| U41-11 | Browser bundle/log scan | No server secret, token, refresh credential, or raw provider key is present | [ ] |
| U41-12 | AI chat/voice KM retrieval | `/api/chat` and `/ws/voice` use `monti_ai_km_ro`; KM lookup succeeds for the current tenant and any write attempt is denied | [ ] |
| U41-13 | Customer-confirmed ticket | `/api/customer/tickets` uses `monti_ticket_rw`; own-tenant ticket creation succeeds, wrong-tenant source/ticket access is denied | [ ] |

## Release gate

All scenarios must pass, or an explicit exception must identify the affected
boundary, owner, mitigation, and follow-up sprint. Attach sanitized evidence to
TASK-0180 before security-owner approval and the v2.22.0 release decision.
