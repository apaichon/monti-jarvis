---
id: FEAT-0026
title: "Customer Satisfaction Review and Tenant Statistics"
status: completed
owner: product
created: 2026-07-14
updated: 2026-07-14
sprint: SPRINT-024
---

# FEAT-0026: Customer Satisfaction Review and Tenant Statistics

## Purpose

Let customers rate an AI conversation after it ends and give tenant operators a tenant-scoped view of satisfaction trends.

## Scope

- AI asks for a 1-5 rating after chat or voice conversation completion.
- Customer UI uses star icons and offers a follow-up prompt when no rating was submitted.
- Reviews are linked to the call/conversation and remain idempotent.
- Tenant users can filter review statistics by date, with today as the default.
- Tenant statistics include average rating, distribution, completion rate, avatar, and channel breakdowns.

## Out of scope

- Public ratings, cross-tenant platform analytics, sentiment inference, and free-form customer comments.

## Acceptance criteria

1. A completed conversation can receive one 1-5 satisfaction review.
2. Review submission works for chat and voice and does not block archive/session completion.
3. Tenant statistics are isolated, filterable, and consistent with stored reviews.
4. Unrated completed calls can be reviewed from a follow-up prompt without reopening the call.

## Design links

- [DES-0027 - Customer Satisfaction Review and Tenant Statistics](../02-design/27-customer-satisfaction-statistics-spec.md)
- [API specification](../02-design/04-api-spec.md) § Customer Satisfaction
