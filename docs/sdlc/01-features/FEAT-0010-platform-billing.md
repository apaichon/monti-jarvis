# Feature: Platform Billing Ledger (FEAT-0010)

**Sprint:** SPRINT-010 · **Status:** shipped · **Release:** v1.1.0

## Problem

Platform cannot see who paid for which package after tenant checkout.

## Scope

In: list/filter payment orders; order detail with documents.  
Out: metering counters, refunds.

## Acceptance criteria

1. `GET /api/platform/billing/orders` returns orders with package_name, tenant_name.
2. Filters `tenant_id`, `status` work.
3. `GET /api/platform/billing/orders/{id}` includes documents when paid.
4. Platform UI `/admin/billing` shows ledger.

## Links

- Sprint: [SPRINT-010](../03-sprints/SPRINT-010.md)
- Chain: [15-commerce-chain-plan.md](../02-design/15-commerce-chain-plan.md)
