---
id: FEAT-0041
title: "AI call-center security hardening"
status: proposed
roadmap_sprint: 41
priority: H
depends_on: [SPRINT-019, SPRINT-020, SPRINT-032, SPRINT-033]
updated: 2026-07-26
release: v2.22.0
---

# FEAT-0041: AI Call-Center Security Hardening

## Purpose

Reduce the blast radius of browser compromise, configuration mistakes,
database credential misuse, unsafe queries, and cross-tenant authorization
defects across the AI call-center platform.

## Acceptance criteria

1. Sensitive browser session material is not kept as plaintext in persistent
   storage; existing sessions migrate safely and logout/revocation clears all
   legacy keys.
2. Web Crypto protects only approved low-sensitivity client state; refresh
   credentials use an HttpOnly, Secure, SameSite cookie or an explicitly
   documented equivalent, and the browser never receives server secrets.
3. Production startup fails closed when required secrets, secure cookie policy,
   or database role configuration is missing or unsafe; diagnostics redact
   values.
4. AI, reporting, and other read paths can use a dedicated PostgreSQL
   read-only pool whose role cannot INSERT, UPDATE, DELETE, or DDL.
5. Tenant-scoped reads set an explicit transaction-local tenant context and
   database policies deny cross-tenant rows; pooled connections never retain
   tenant context between requests.
6. SQL inputs use parameters and allowlists; bounded limits, sort fields, and
   identifiers reject injection-shaped or unbounded values.
7. Automated tests prove secret redaction, read-role denial, transaction
   context cleanup, cross-tenant denial, and fail-closed production startup.
8. The manual UAT uses at least two tenants and records evidence for browser
   storage migration, configuration failure, read-role denial, and isolation.
9. AI agent conversation KM retrieval uses a read-only database principal and
   cannot create tickets, alter KM, or mutate tenant data.
10. Ticket creation and ticket-event writes use a separate ticket-write
    principal with only the required ticket permissions.
11. Every database principal is tenant-bound by authenticated context and
    policy; a principal for Tenant A cannot read or write Tenant B data.

## Scope

### In

- Tenant/platform browser session-storage migration and Web Crypto wrapper.
- Production environment validation, secret redaction, and rotation guidance.
- Dedicated read-only PostgreSQL role/pool for AI and reporting reads.
- Explicit API-to-database-user routing for AI/KM reads versus ticket writes.
- Parameterized/allowlisted query review for touched read paths.
- PostgreSQL tenant policies where applicable plus application authorization.
- Security posture diagnostics, regression tests, and manual UAT evidence.

### Out

- Full identity-provider replacement or customer-auth redesign.
- WAF, network segmentation, managed KMS procurement, or endpoint-MDM work.
- Encrypting arbitrary browser data with a key that is also persisted beside it.
- Rewriting every historical query in one sprint; high-risk tenant reads are
  prioritized and remaining inventory is tracked as follow-up work.

## Design links

- [DES-0044 — Security Hardening](../02-design/44-ai-call-center-security-hardening-spec.md)
- [Workflow](../02-design/02-workflow.md) — §109–112
- [ER / policy map](../02-design/03-er-diagram.md) — Sprint 41
- [API contract](../02-design/04-api-spec.md) — Sprint 41
- [UX/UI operator surface](../02-design/05-ux-ui.md) — S41
