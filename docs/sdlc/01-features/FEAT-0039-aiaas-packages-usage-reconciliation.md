---
id: FEAT-0039
title: "AiaaS mass-market packages and usage reconciliation"
status: in_progress
roadmap_sprint: 45
priority: L
depends_on: [SPRINT-013, SPRINT-016, SPRINT-025, SPRINT-027, SPRINT-030, SPRINT-031, SPRINT-043]
updated: 2026-07-25
---

# FEAT-0039: AiaaS Mass-Market Packages and Usage Reconciliation

## Purpose

Offer four understandable monthly AiaaS packages while making quota,
metering, billing, tenant statistics, platform statistics, and mobile usage
use the same dimension definitions and tenant boundaries.

## Acceptance criteria

1. Initialization seeds four THB package defaults with price, billing period,
   included features, and quota dimensions without hard-coding commercial
   values in Svelte components; platform admins can change the catalog later.
2. Every quota response identifies dimension, unit, period, limit, consumed,
   remaining, and source. A rejected request does not consume quota.
3. Web, embed, and mobile paths enforce the same tenant entitlement snapshot;
   mobile minutes and concurrency have an explicitly documented relationship
   to the package allowance.
4. Usage events for calls, mobile calls, KM documents, avatars, storage bytes,
   and concurrent calls are idempotent and safe across retries, disconnects,
   timeouts, failed starts, deletes, and package changes.
5. Tenant and platform usage views distinguish current enforcement counters from
   historical activity and show unavailable or stale sources explicitly.
6. Two-tenant tests cover all four tiers, package changes, quota exhaustion,
   reconciliation replay, and concurrent multi-user load.

## Scope

In scope: idempotent package initialization plus platform-admin package CRUD,
entitlement rules, storage and mobile quota
dimensions, usage event ledger, reconciliation projections, tenant/platform
usage contracts, billing/package UI updates, mobile quota metadata, and
verification fixtures.

Out of scope: overage billing, automatic upgrades, enterprise custom pricing,
cross-tenant quota pooling, changing the Gemini voice pipeline, and the held
S44 customer generative/CLI workspace.

## Design links

- [DES-0042 — AiaaS Packages and Usage Reconciliation](../02-design/42-aiaas-packages-usage-reconciliation-spec.md)
- [API contract](../02-design/04-api-spec.md) — Sprint 45
- [Workflow](../02-design/02-workflow.md) — §100–104
- [UX/UI](../02-design/05-ux-ui.md) — T24/A24/M2
