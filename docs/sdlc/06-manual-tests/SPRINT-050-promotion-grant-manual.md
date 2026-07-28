# SPRINT-050 — Promotion grant manual UAT

**Feature:** FEAT-0042 · **Design:** DES-0046  
**Updated:** 2026-07-28

## Preconditions

- Platform admin credentials
- Postgres schema applied (server bootstrap or `033_promotion_grants.sql`)
- At least one **active** catalog package
- Target tenant exists

## Scenarios

### 1. Grant promotional package (happy path)

- [X] Open `/admin/tenants/{tenant_id}/entitlement`
- [X] Under **Promotion grant**, select an active package
- [X] Enter a non-empty reason
- [X] Click **Grant promotion**
- [X] **Expected:** success toast; **Current plan** shows the package; tax invoice number shown
- [X] Open **Receipts & tax invoices** and find `tax_invoice` for the tenant

### 2. Supersede existing plan

- [X] Tenant already has active plan A
- [X] Grant promotion for package B
- [X] **Expected:** active plan is B; plan A no longer active; new tax invoice issued

### 3. Idempotent retry

- [X] Grant with the same `idempotency_key` and body twice
- [X] **Expected:** second call succeeds as replay; no second tax invoice number

### 4. Validation

- [X] Empty reason → error, no plan change
- [X] Non-admin token → 403

### 5. Isolation

- [X] Grant for tenant A
- [X] As platform admin, confirm tenant B entitlement/invoices do not show A’s grant

## Sign-off

| Role | Date | Result |
| --- | --- | --- |
| Tester | | |
| Dev | | |
