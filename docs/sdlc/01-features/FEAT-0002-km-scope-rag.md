# Feature: KM and Scope RAG   (feature:km-scope-rag)
**Sprint:** SPRINT-002   **Owner:** DEV

## Problem

Sprint 1 agents answer from general LLM knowledge only. Callers need **grounded answers** from approved tenant knowledge, limited by **scope** (general, billing, technical) and agent role.

## Scope

In:
- KM document upload and storage in MinIO
- Chunking + embedding pipeline into ClickHouse `km_embeddings`
- Postgres scope model linking documents/chunks to `km_scope` tags
- Scoped vector retrieval before each AI turn (chat + voice)
- Citation of source chunks in the customer portal transcript
- Missing-KM detection logged to ClickHouse `qa_events`

Out:
- Tenant admin KM wizard, versioning UI, approval workflows (Sprint 15)
- Multi-tenant auth and RBAC (Sprint 3)
- pgvector or Postgres-native vectors (blueprint uses ClickHouse only)
- Automatic KM improvement suggestions

## Acceptance criteria

1. Operator can ingest a text/Markdown document for `demo` tenant; chunks appear in ClickHouse within one pipeline run.
2. Caller question on **General** tab retrieves only chunks tagged `general` (and shared scopes).
3. Caller question on **Billing** tab with billing-scoped KM returns a grounded answer citing the source document.
4. Question with no matching scoped chunks yields a safe fallback and writes a `qa_events` missing-KM record.
5. Voice and text paths share the same RAG retrieval service and scope filter.
6. Retrieved chunk IDs are logged on the call session turn metadata when Postgres is available.

## Test notes

- Unit: scope resolver, chunk splitter, ClickHouse query builder with tenant + scope filters.
- Integration: ingest sample FAQ → embed → search → orchestrator prompt contains chunk text.
- Manual: compare answer with/without KM ingest for the same question.
- Safety: RAG context must not bypass scope; agents still refuse credential requests.

## Dependencies

- Sprint 1: `/api/chat`, `/ws/voice`, call sessions, topic tabs, `DEMO_TENANT_ID`
- Infra: MinIO, ClickHouse (`poc-gml-clickhouse` or `CLICKHOUSE_URL`)
- Blueprint §12.3, §17.4
- Design: `docs/sdlc/02-design/02-workflow.md`, `04-api-spec.md`, `03-er-diagram.md`