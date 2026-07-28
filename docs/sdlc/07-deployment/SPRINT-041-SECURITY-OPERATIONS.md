---
id: OPS-0041
title: "Sprint 41 security operations and secret rotation"
status: ready
sprint: SPRINT-041
updated: 2026-07-26
---

# Sprint 41 Security Operations

## Required production configuration

Set `COOKIE_SECURE=true`, `COOKIE_SAMESITE=lax` or `strict`, an explicit
comma-separated `ALLOWED_ORIGINS` list, and `POSTGRES_RLS_ENFORCED=true`.
Production also requires distinct writer, KM-read, and ticket-write database
principals. Startup fails closed when any of these checks is unsafe.

## JWT and database credential rotation

1. Provision replacement secrets and capability-role credentials without
   removing the currently active values.
2. Update the deployment secret references, restart one instance, and verify
   `/readyz` plus the metadata-only posture endpoint.
3. Drain and restart the remaining instances. Existing refresh sessions are
   revoked through logout/session rotation; users re-authenticate if the
   database session store or JWT signing secret is rotated.
4. Remove the previous secret only after the overlap window and authentication
   metrics are healthy.

Never print secret values, connection URLs, cookies, tokens, prompts, or
request bodies during rotation. The application redacts common credential
patterns from startup and infrastructure diagnostics.

## Rollback

If readiness fails, restore the previous secret reference and restart the
affected instances. Do not weaken `COOKIE_SECURE`, disable RLS, or reuse the
general writer as a production read pool to recover service. Record the
remediation code and sanitized timestamps in the Sprint 41 UAT evidence.
