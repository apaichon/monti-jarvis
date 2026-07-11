# Feature: Platform Receipt Ops (FEAT-0011)

**Sprint:** SPRINT-011 · **Status:** shipped · **Release:** v1.2.0

## Problem

S9 auto-issues receipts/tax invoices but platform cannot void, reissue, or brand seller block.

## Scope

In: list documents, void, reissue, seller branding, HTML print.  
Out: government e-Tax gateway.

## Acceptance criteria

1. Document status `issued|voided`; only one active (order_id, doc_type).
2. Void sets status + reason; reissue creates new issued row.
3. Seller branding used on issue/reissue.
4. UI `/admin/billing/receipts` + `/admin/billing/seller`.

## Links

- Sprint: [SPRINT-011](../03-sprints/SPRINT-011.md)
