---
id: SPRINT-002
status: completed
start: 2026-07-07
end: 2026-07-07
closed: 2026-07-07
updated: 2026-07-07
release: v0.3.0
goal: "Customer: Add KM and Scope — tenant-approved knowledge with ClickHouse RAG in conversation."
roadmap_sprint: 2
platform: Customer
depends_on: [SPRINT-001]
---

# SPRINT-002 — Customer: Add KM and Scope

## Goal

AI answers in the customer portal use **tenant-approved knowledge** with **scope enforcement**, retrieved via **ClickHouse vector search** before each chat or voice turn.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0006 | 3 | completed | devops | ClickHouse `km_embeddings` schema + Go search client |
| TASK-0007 | 5 | completed | dev | Postgres KM documents, chunks, and scope model |
| TASK-0008 | 5 | completed | dev | KM ingest API — MinIO → chunk → embed → ClickHouse |
| TASK-0009 | 3 | completed | dev | Scoped RAG in chat/voice orchestrator + UI citations |

**Committed:** 16 points · **Completed:** 16 points · **Velocity:** 16

## Shipped (v0.3.0)

- Per-avatar KM ingest API (`/api/km/agents/{id}/documents`, `/reset`, `/seed`)
- Gemini `gemini-embedding-001` pipeline → ClickHouse `km_embeddings`
- Scoped RAG in `/api/chat` with citation chips in Caller Desk
- Voice RAG preload with parallel dial, cache, and `ready`-gated mic streaming
- `monti-clickhouse` in compose with published `:8123` and dev auth
- Scope resolver (`internal/scope`) for topic tab + agent role
- Sample KB under `docs/samples/km/` · ops guide `docs/KM_SETUP.md` · REST client `docs/km-setup.http`
- SDLC test scenarios, manual tests, deployment, and readiness checklists

## Feature

- [FEAT-0002 — KM and Scope RAG](../01-features/FEAT-0002-km-scope-rag.md)
- Design: [api-spec](../02-design/04-api-spec.md) · [workflow](../02-design/02-workflow.md) · [ux-ui](../02-design/05-ux-ui.md)

## Stack delivered

```text
internal/km          ingest, chunking, per-agent reset
internal/rag         scoped retrieval, voice preload cache
internal/clickhouse  km_embeddings, qa_events, vector search
internal/scope       agent/topic → km_scope
internal/gemini      embed (gemini-embedding-001)
cmd/server/km.go     KM HTTP handlers
apps/customer-web    citation chips, voice ready handshake
infra/docker-compose monti-clickhouse (:8123)
```

## Verification

- `go test ./...` · `make build` · `make km-seed`
- Manual: [SPRINT-002 UAT](../06-manual-tests/SPRINT-002-manual.md)
- Readiness: [RELEASE-READINESS](../08-readiness/RELEASE-READINESS.md)

## Retrospective notes

- `text-embedding-004` deprecated — migrated to `gemini-embedding-001` with RETRIEVAL task types.
- Shared `poc-gml-clickhouse` has no host port; `monti-clickhouse` added to Monti compose.
- Voice latency improved by parallel RAG preload, smaller KB context, and client `ready` gate before mic send.

## Risks (closed)

- ClickHouse dev uses Go-side cosine ranking on ≤50 candidates; sufficient for demo corpus.
- Voice RAG is preload-only (not per-utterance embed); query-specific voice RAG deferred.