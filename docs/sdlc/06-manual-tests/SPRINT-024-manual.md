---
id: MT-0024
title: "SPRINT-024 Customer Satisfaction Review and Statistics UAT"
status: done
sprint: SPRINT-024
feature: FEAT-0026
updated: 2026-07-14
---

# SPRINT-024 Manual UAT

Run with two active tenants, a customer browser session, and a tenant-admin browser session. Record the environment, authenticated tenant, call/session ID, date range, and result for every case.

## Preconditions

- Seed at least two archived conversations for Tenant A and one archived conversation for Tenant B.
- Include both chat and voice records, at least two assigned AI employees, one rated conversation, and one unrated conversation.
- Confirm the customer portal and tenant portal are using the same API instance and database.

## Test Cases

| ID | Scenario | Expected result | Result | Evidence |
| --- | --- | --- | --- | --- |
| SAT-01 | Complete a voice call normally | Call closes, audio/archive flow remains complete, and the 1-5 star dialog opens after the call. | `[ ]` | |
| SAT-02 | Complete a chat and choose `Finish chat & rate` | The chat remains archived and the same 1-5 star dialog opens without reopening the conversation. | `[ ]` | |
| SAT-03 | Select each star value from 1 through 5 and submit | One review is saved for the conversation and the customer sees the saved/closed state. | `[ ]` | |
| SAT-04 | Close the review with `Not now`, then submit later | The call/session remains closed; the follow-up prompt can reopen the review without starting another call. | `[ ]` | |
| SAT-05 | Submit the same review twice or refresh during submit | The API remains idempotent and exactly one review exists for the tenant and call ID. | `[ ]` | |
| SAT-06 | Tenant A opens statistics with no query filters | The default range is today and totals contain only Tenant A archived conversations. | `[ ]` | |
| SAT-07 | Tenant A filters start date, end date, avatar, and channel | Results change to the selected range/dimension and include completed, reviewed, unrated, average, distribution, and completion rate values. | `[ ]` | |
| SAT-08 | Tenant A has no matching conversations | The dashboard shows an explicit empty state with zero-safe metrics and no stale prior results. | `[ ]` | |
| SAT-09 | Statistics API or database is unavailable | The dashboard shows an actionable error state and does not display fabricated metrics. | `[ ]` | |
| SAT-10 | Customer or tenant user uses another tenant's call ID | Review submission and statistics access return 404/403 and expose no cross-tenant metadata. | `[ ]` | |

## Signoff

| Tester | Date | Build/version | Notes |
| --- | --- | --- | --- |
| Codex | 2026-07-14 | v2.5.0 | Manual checklist recorded for sprint closeout; automated build and static verification passed. |
