---
id: FEAT-0024
title: Conversation Records and Knowledge Gap Review
status: in_progress
sprint: SPRINT-022
owner: PM
updated: 2026-07-13
---

# Feature: Conversation Records and Knowledge Gap Review

## Problem

Tenants need durable conversation evidence for operations, compliance review, and support quality. They also need a structured way to detect when the AI could not answer from tenant knowledge so tenant admins can close knowledge gaps through KM updates.

## Scope

**In**

- Tenant-scoped conversation record metadata in Postgres.
- Transcript/audio/object archive writes to MinIO under the existing `calls/` prefix.
- Tenant setting for archive encryption mode where supported by the local MinIO/object contract.
- Knowledge-gap detection records linked to conversation turns and tenant KM scope.
- Tenant UI/API to search conversation records and review knowledge gaps.

**Out**

- Human ticket workflow and SLA routing.
- Satisfaction survey/statistics dashboards.
- Long-term cold archive/retention automation.
- Production observability dashboards beyond sprint verification evidence.

## Acceptance criteria

1. Chat and voice conversations produce tenant-scoped record metadata.
2. Transcript and call artifacts are written to MinIO using deterministic tenant/call paths.
3. Archive encryption can be configured per tenant or deployment without changing public APIs.
4. Failed/low-confidence/RAG-miss turns create knowledge-gap candidates.
5. Tenant admins can list, inspect, and resolve/snooze knowledge gaps.
6. Cross-tenant record and object access returns 404/403 without metadata leakage.
7. Manual UAT covers archive object creation, encrypted-mode smoke, and knowledge-gap lifecycle.

## Dependencies

- [FEAT-0001 - Workforce QA](FEAT-0001-workforce-qa.md)
- [FEAT-0002 - KM Scope RAG](FEAT-0002-km-scope-rag.md)
- [FEAT-0003 - Auth and RBAC](FEAT-0003-auth-rbac.md)
- [FEAT-0023 - Authenticated Workforce Selection and Customer Quota Enforcement](FEAT-0023-authenticated-workforce-selection.md)

## Design

- [DES-0025 — Conversation Records and Knowledge Gaps Specification](../02-design/25-conversation-records-knowledge-gaps-spec.md)
- [API contract — Conversation Records & Knowledge Gaps](../02-design/04-api-spec.md#conversation-records--knowledge-gaps-sprint-22)
- [Workflow §66–67](../02-design/02-workflow.md#66-conversation-archive-write-sprint-22)
- [UX § Sprint 22](../02-design/05-ux-ui.md#sprint-22--conversation-records-and-knowledge-gaps-t15)
- Sprint plan: [SPRINT-022](../03-sprints/SPRINT-022.md)
