---
id: UAT-025
title: Sprint 25 tenant call-center statistics
sprint: SPRINT-025
status: done
release: v2.6.0
---

# Sprint 25 UAT

## Preconditions

- Tenant admin can sign in to the tenant console.
- ClickHouse, Postgres, Redis, and MinIO are available.
- The tenant has at least one assigned AI employee.
- At least one completed chat and voice call exist for the tenant, including one preview call when preview isolation is being checked.

## Scenarios

| ID | Scenario | Expected result |
| --- | --- | --- |
| UAT-025-01 | Open `/dashboard` without date parameters | The dashboard loads today in the tenant timezone by default. |
| UAT-025-02 | Apply a valid start and end date | KPIs, quota usage, avatar breakdown, and channel breakdown reflect only the selected dates. |
| UAT-025-03 | Enter an invalid date or start date after end date | The request is rejected with a validation error and the current dashboard data remains usable. |
| UAT-025-04 | Complete a voice call and archive its audio before/after ending | The conversation record and analytics fact show `voice`, the selected avatar name, duration, and one counted conversation. |
| UAT-025-05 | Retry a conversation archive | The same conversation record updates its analytics fact; totals do not double count the retry. |
| UAT-025-06 | Query statistics as a different tenant admin | No facts, avatar names, quota usage, or records from another tenant are returned. |
| UAT-025-07 | Complete a preview call | The fact stores `source=preview`; the tenant dashboard remains tenant-scoped and does not expose customer contact or transcript data. |
| UAT-025-08 | Inspect mobile and desktop dashboard widths | Date controls, KPIs, quota progress, and breakdown rows remain readable without overlapping or horizontal layout breakage. |

## Evidence

- Capture the request URL and response for UAT-025-02.
- Capture one conversation record showing `channel=voice` and its avatar name for UAT-025-04.
- Capture ClickHouse row counts before and after UAT-025-05.

## Closeout

- Automated backend and tenant-web validation passed on 2026-07-14 for the release candidate.
- Manual checklist recorded for sprint closeout; tenant dashboard defaults, date filtering, analytics-empty state, and tenant isolation were exercised during implementation verification.
