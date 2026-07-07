---
id: TEST-MATRIX
status: active
updated: 2026-07-07
sprints: [SPRINT-001, SPRINT-002]
---

# Test Scenario Matrix — Monti Jarvis

Maps feature acceptance criteria to executable scenarios. **Auto** = `go test` or build gate; **Manual** = browser/curl UAT in [`06-manual-tests/`](../06-manual-tests/).

## Legend

| Type | Meaning |
| --- | --- |
| Auto | Covered by unit/integration test or CI build step |
| Manual | Requires running stack + human or scripted curl/browser check |
| Smoke | Short post-deploy check in readiness checklist |

---

## FEAT-0001 — Workforce + Inbound Q&A (SPRINT-001)

| ID | AC | Scenario | Type | Test / Command |
| --- | ---: | --- | --- | --- |
| T1-01 | 1 | Workforce API returns four agents with images and roles | Auto | `go test ./internal/workforce/...` · `TestAllReturnsWorkforce` |
| T1-02 | 1 | Portal lists Ava, Max, Luna, Neo with avatar photos | Manual | [S1](../06-manual-tests/SPRINT-001-manual.md#s1--agent-selection-feat-0001--ac-1) |
| T1-03 | 2 | Chat response shaped by agent system prompt | Auto | `TestSystemPromptIncludesRole` |
| T1-04 | 2 | Text chat on General/Billing/Technical tabs | Manual | [S2](../06-manual-tests/SPRINT-001-manual.md#s2--text-chat-per-topic-feat-0001--ac-2) |
| T1-05 | 3 | Voice WebSocket connects with selected agent voice | Manual | [S3](../06-manual-tests/SPRINT-001-manual.md#s3--voice-call-feat-0001--ac-3) |
| T1-06 | 3 | Multi-turn voice Q&A without disconnect | Manual | [S3](../06-manual-tests/SPRINT-001-manual.md#s3--voice-call-feat-0001--ac-3) |
| T1-07 | 4 | Transcript shows user and assistant turns (text) | Manual | [S2](../06-manual-tests/SPRINT-001-manual.md#s2--text-chat-per-topic-feat-0001--ac-2) |
| T1-08 | 4 | Transcript shows voice turns in Caller Desk | Manual | [S3](../06-manual-tests/SPRINT-001-manual.md#s3--voice-call-feat-0001--ac-3) |
| T1-09 | 5 | Call session persists when Postgres available | Auto | `go test ./cmd/server/...` · calls tests |
| T1-10 | 5 | Session metadata via call APIs | Manual | [S4](../06-manual-tests/SPRINT-001-manual.md#s4--call-session-persistence-feat-0001--ac-5) |
| T1-11 | — | Health includes sprint flag and customer web | Auto | `TestHealthIncludesSprint002` |
| T1-12 | — | LiveKit token route registered | Auto | `TestIssueCallTokenRoutePattern` |
| T1-13 | — | Legacy UI available at `/legacy/` | Manual | [S5](../06-manual-tests/SPRINT-001-manual.md#s5--legacy-ui-feat-0001--regression) |
| T1-14 | Safety | Agent refuses credential/OTP requests | Manual | [S6](../06-manual-tests/SPRINT-001-manual.md#s6--safety-refusal-feat-0001--safety) |

---

## FEAT-0002 — KM and Scope RAG (SPRINT-002)

| ID | AC | Scenario | Type | Test / Command |
| --- | ---: | --- | --- | --- |
| T2-01 | 1 | Document upload → chunks in ClickHouse after pipeline | Manual | [S1](../06-manual-tests/SPRINT-002-manual.md#s1--km-ingest-feat-0002--ac-1) |
| T2-02 | 1 | `POST /api/km/seed` ingests all four sample KBs | Manual | [S1](../06-manual-tests/SPRINT-002-manual.md#s1--km-ingest-feat-0002--ac-1) |
| T2-03 | 1 | Chunk splitter splits paragraphs | Auto | `go test ./internal/km/...` · `TestChunkTextSplitsParagraphs` |
| T2-04 | 2 | General tab searches `general` scope only | Auto | `TestResolveAvaOnTechnicalTab` |
| T2-05 | 2 | Ava on Technical tab still uses general scope | Manual | [S3](../06-manual-tests/SPRINT-002-manual.md#s3--scope-enforcement-feat-0002--ac-2) |
| T2-06 | 3 | Billing tab + Max returns grounded answer with sources | Manual | [S2](../06-manual-tests/SPRINT-002-manual.md#s2--scoped-rag-chat-feat-0002--ac-3) |
| T2-07 | 3 | Citation chips in portal transcript | Manual | [S2](../06-manual-tests/SPRINT-002-manual.md#s2--scoped-rag-chat-feat-0002--ac-3) |
| T2-08 | 4 | No matching chunks → safe fallback | Manual | [S4](../06-manual-tests/SPRINT-002-manual.md#s4--missing-km-fallback-feat-0002--ac-4) |
| T2-09 | 4 | `qa_events` row for missing KM | Manual | [S4](../06-manual-tests/SPRINT-002-manual.md#s4--missing-km-fallback-feat-0002--ac-4) |
| T2-10 | 5 | Voice path uses same scope resolver as chat | Auto | `TestResolveBillingAgent`, `TestResolveNeoTriage` |
| T2-11 | 5 | Voice turn grounded when KB seeded | Manual | [S5](../06-manual-tests/SPRINT-002-manual.md#s5--voice-rag-feat-0002--ac-5) |
| T2-12 | 6 | Turn metadata stores source chunk IDs | Manual | [S6](../06-manual-tests/SPRINT-002-manual.md#s6--turn-metadata-feat-0002--ac-6) |
| T2-13 | — | Per-agent KB status API | Manual | [S7](../06-manual-tests/SPRINT-002-manual.md#s7--km-admin-apis-feat-0002--ops) |
| T2-14 | — | Per-agent reset clears Postgres + MinIO + ClickHouse | Manual | [S7](../06-manual-tests/SPRINT-002-manual.md#s7--km-admin-apis-feat-0002--ops) |
| T2-15 | — | RAG degrades when ClickHouse offline | Manual | [S8](../06-manual-tests/SPRINT-002-manual.md#s8--degraded-mode-feat-0002--regression) |
| T2-16 | Safety | RAG does not bypass scope; credential refusal holds | Manual | [S3](../06-manual-tests/SPRINT-002-manual.md#s3--scope-enforcement-feat-0002--ac-2) |

---

## Build & regression gates (all sprints)

| ID | Scenario | Type | Command |
| --- | --- | --- | --- |
| G-01 | Go unit tests pass | Auto | `make test` or `go test ./...` |
| G-02 | Customer portal builds | Auto | `make customer-web` |
| G-03 | Server binary builds | Auto | `make build` |
| G-04 | Healthz reachable | Smoke | `curl -fsS http://localhost:8091/healthz` |
| G-05 | Infra dependencies ok | Smoke | `make infra-check` |
| G-06 | Infra API reports stores | Smoke | `curl -fsS http://localhost:8091/api/infra` |

---

## Coverage gaps (planned)

| Area | Gap | Target |
| --- | --- | --- |
| ClickHouse | No mocked integration test in CI | Add `internal/clickhouse` unit tests when client stabilizes |
| RAG | End-to-end ingest→search in test | Integration test with testcontainers (Sprint 3+) |
| Voice | Automated WebSocket smoke | `cmd/voicecheck` style harness (deferred) |