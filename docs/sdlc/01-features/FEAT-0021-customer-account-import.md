---
id: FEAT-0021
title: Customer Account Import and Domain Integration
status: completed
sprint: SPRINT-019
owner: PM
updated: 2026-07-13
---

# Feature: Customer Account Import and Domain Integration

## Problem

Tenants can define tiers and groups, but they cannot attach real customer records to them. SPRINT-020 authentication needs a tenant-isolated identity directory, safe bulk import, and domain defaults before credentials are issued.

## Scope

**In**

- Tenant-scoped customer directory with normalized email, optional phone, locale, metadata, status, source, and external id.
- Optional tier plus many-to-many group membership.
- CSV dry-run and commit with deterministic validation and import summaries.
- Domain rules with `allow` or `deny` policy intent and optional default tier/group.
- Tenant CRUD/import UI and integration-safe upsert semantics.
- Design DES-0022 and UAT documentation.

**Out**

- Customer credential creation, login, invitation, email verification, OAuth, and session issuance.
- Vendor-specific CRM connectors or background synchronization.
- Domain-rule enforcement on public authentication before SPRINT-020.
- Discounts, KYC, tickets, and conversation-history UI.

## Acceptance criteria

1. Active tenant admin can create, search, update, and deactivate a customer record scoped to its JWT tenant.
2. CSV dry-run returns accepted/rejected rows without writes; commit returns created/updated/rejected counts.
3. Reimporting the same `(source, external_id)` updates rather than duplicates the customer.
4. Tier and group references must belong to the same tenant and be active/available.
5. Domain rules normalize case and reject duplicates within a tenant; `allow`/`deny` is stored for SPRINT-020 enforcement.
6. Tenant A cannot read or mutate Tenant B customers, imports, rules, tiers, or groups.
7. Customer portal remains public/no-auth and no customer token is issued.
8. Manual UAT is recorded under `docs/sdlc/06-manual-tests/SPRINT-019-manual.md` during implementation verification.

## Dependencies

- [FEAT-0003 — Auth and RBAC](FEAT-0003-auth-rbac.md)
- [FEAT-0020 — Customer Tier](FEAT-0020-customer-tier.md)
- Packages: `internal/store`, `cmd/server`, `apps/tenant-web`

## Design

- [DES-0022 — Customer Account Import Specification](../02-design/22-customer-account-import-spec.md)
- [API contract — Customer Accounts & Imports](../02-design/04-api-spec.md#customer-accounts--imports-sprint-19)
