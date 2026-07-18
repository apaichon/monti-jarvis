---
id: FEAT-0033
title: "Platform Billing, Quota, and AI Infrastructure Cost Usage"
status: in_progress
owner: product
created: 2026-07-17
updated: 2026-07-17
sprint: SPRINT-031
---

# FEAT-0033: Platform Billing, Quota, and AI Infrastructure Cost Usage

## Purpose

Give platform administrators a read-only operational view of paid package value, current quota enforcement, historical call usage, and AI infrastructure cost coverage across tenants.

## Scope

- Reconcile existing payment orders and active entitlements without changing payment or fulfillment behavior.
- Show historical ClickHouse call usage separately from current Redis quota enforcement counters.
- Capture provider-neutral AI usage with explicit observed, estimated, and unavailable states.
- Show versioned rate metadata and observed/estimated cost by tenant and date range.
- Provide bounded platform-admin API and responsive billing usage dashboard with safe partial-failure states.

## Out of scope

- Charging, invoices, refunds, tax documents, auto-upgrades, or entitlement changes.
- Customer-facing cost display, customer-level records, raw prompts/responses, transcripts, audio, or provider payloads.
- Replacing Redis quota enforcement or introducing a second quota authority.

## Acceptance criteria

1. Paid order and active entitlement summaries come from their existing Postgres authorities and remain read-only.
2. Historical reporting minutes and current enforcement counters are displayed as separate, labeled metrics.
3. AI usage facts are idempotent and retain measurement state, rate version, and source timestamps.
4. Missing provider usage is marked estimated or unavailable, never silently counted as zero exact cost.
5. Platform-only authorization, bounded pagination, aggregate-only response fields, and safe source failure states are enforced.
6. Controlled fixtures cover payment/entitlement mismatch, quota divergence, duplicate metering, date boundaries, and observed/estimated/unavailable AI usage.

## Design links

- Sprint: [SPRINT-031](../03-sprints/SPRINT-031.md)
- Deep spec: [DES-0034 - Platform Billing, Quota, and AI Infrastructure Cost Usage](../02-design/34-platform-billing-quota-ai-cost-spec.md)
- [API specification](../02-design/04-api-spec.md) - Platform Billing Usage
