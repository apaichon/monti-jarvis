---
id: TEST-MATRIX
status: active
updated: 2026-07-13
sprints: [SPRINT-001, SPRINT-002, SPRINT-013, SPRINT-014, SPRINT-019, SPRINT-020]
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

## FEAT-0013 — Quota & Rate Limit (SPRINT-013)

| ID | AC | Scenario | Type | Test / Command |
| --- | ---: | --- | --- | --- |
| T13-01 | 1 | KM ingest over `max_km_documents` → 429 | Manual | [S2](../06-manual-tests/SPRINT-013-manual.md#s2--km-document-quota-exceed-task-0059--feat-ac-1) |
| T13-02 | 1 | Structured error body `code` + `dimension` | Auto | `TestWriteQuotaError_*` · `go test ./cmd/server/ -run Quota` |
| T13-03 | 2 | Concurrent voice over `max_concurrent_calls` rejected | Manual | [S3](../06-manual-tests/SPRINT-013-manual.md#s3--concurrent-voice-slots-task-0059--feat-ac-2) |
| T13-04 | 2 | Concurrent acquire/release | Auto | `TestAcquireConcurrent_Release` |
| T13-05 | 3 | `voice_enabled=false` blocks voice | Manual | [S5](../06-manual-tests/SPRINT-013-manual.md#s5--feature-flags-task-0059--feat-ac-3) |
| T13-06 | 3 | `rag_enabled=false` skips RAG, chat continues | Manual | [S5](../06-manual-tests/SPRINT-013-manual.md#s5--feature-flags-task-0059--feat-ac-3) |
| T13-07 | 4 | Redis key shapes + UTC month | Auto + docs | `TestAddCallMinutesAndMonthlyCheck` · LOCAL-DEV |
| T13-08 | 5 | Platform `GET .../usage` + admin UI | Manual | [S1](../06-manual-tests/SPRINT-013-manual.md#s1--platform-usage-snapshot-task-0060--feat-ac-5) |
| T13-09 | 6 | Chat rate limit → 429 + Retry-After | Manual / Auto | [S4](../06-manual-tests/SPRINT-013-manual.md#s4--rate-limit-burst-task-0059--feat-ac-6--optional-but-recommended) · `TestAllowRate` |
| T13-10 | 7 | Fail-open / disabled skips enforcement | Auto | `TestNoEntitlementFailOpen` · `TestDisabledSkipsChecks` |
| T13-11 | — | `/api/infra` quota + rate_limit | Manual / Smoke | [§1](../06-manual-tests/SPRINT-013-manual.md#1-init-infrastructure) |
| T13-12 | — | Avatar assign over `max_ai_employees` | Manual | [S6](../06-manual-tests/SPRINT-013-manual.md#s6--avatar-assign-cap-task-0059--feat-ac) |
| T13-13 | — | Regression login / packages / customer chat | Manual | [S8](../06-manual-tests/SPRINT-013-manual.md#s8--regression-must) |

## FEAT-0014 — Embed to Web (SPRINT-014)

| ID | AC | Scenario | Type | Test / Command |
| --- | ---: | --- | --- | --- |
| T14-01 | 1 | Tenant enables embed + copies snippet | Manual | [S1](../06-manual-tests/SPRINT-014-manual.md#s1--tenant-enable-embed--lazy-config-task-0065--feat-ac-1-6) |
| T14-02 | 2 | Unknown/disabled key → 404 | Manual + Auto | [S2](../06-manual-tests/SPRINT-014-manual.md#s2--public-resolve-api-task-0063--feat-ac-2) · `TestWriteEmbedError_*` |
| T14-03 | 3 | Origin allowlist deny | Manual + Auto | [S3](../06-manual-tests/SPRINT-014-manual.md#s3--origin-allowlist-deny-task-0063--feat-ac-3) · `TestOriginAllowed` |
| T14-04 | 1,4 | Loader + iframe chat | Manual | [S4](../06-manual-tests/SPRINT-014-manual.md#s4--loader--iframe-chat-task-0064--feat-ac-1-4-7) |
| T14-05 | 6 | Rotate key invalidates old | Manual | [S5](../06-manual-tests/SPRINT-014-manual.md#s5--rotate-key-task-0065--feat-ac-6) |
| T14-06 | 7 | Full portal `/` unchanged | Manual | [S8](../06-manual-tests/SPRINT-014-manual.md#s8--regression-must) |
| T14-07 | 5 | Quota applies to embed tenant | Manual optional | [S7](../06-manual-tests/SPRINT-014-manual.md#s7--quota-still-applies-on-embed-path-feat-ac-5--optional) |
| T14-08 | — | Loader asset 200 | Smoke | `curl /embed/monti-embed.js` |
| T14-09 | — | Origin helpers / key format | Auto | `go test ./internal/store/ -run Embed` |

## FEAT-0021 — Customer Account Import and Domain Integration (SPRINT-019)

| ID | AC | Scenario | Type | Test / Command | Result |
| --- | ---: | --- | --- | --- | --- |
| T19-01 | 1 | Manual customer create/edit/deactivate | Manual | [S1](../06-manual-tests/SPRINT-019-manual.md#s1--manual-customer-crud-and-deactivation-task-00880089--feat-ac-1) | Pass |
| T19-02 | 2 | CSV parser validates rows and limits | Auto | `go test ./internal/customerimport` | Pass |
| T19-03 | 2 | Dry-run reports errors and performs no customer writes | Manual | [S2](../06-manual-tests/SPRINT-019-manual.md#s2--csv-dry-run-performs-no-customer-writes-task-00880089--feat-ac-2) | Pass |
| T19-04 | 3 | Repeat source/external id preserves customer id | Manual | [S3](../06-manual-tests/SPRINT-019-manual.md#s3--csv-commit-and-idempotent-reimport-task-00880090--feat-ac-3) | Pass |
| T19-05 | 4–5 | Explicit assignment wins over domain default | Manual | [S4](../06-manual-tests/SPRINT-019-manual.md#s4--explicit-assignment-and-domain-default-precedence-task-0090--feat-ac-45) | Pass |
| T19-06 | 6 | Cross-tenant customer/import/rule ids return 404 | Manual | [S5](../06-manual-tests/SPRINT-019-manual.md#s5--tenant-isolation-task-008700880090--feat-ac-6) | Pass |
| T19-07 | 5 | Email/domain/source normalization | Auto | `go test ./internal/store -run NormalizeCustomer` | Pass |
| T19-08 | 7 | No customer credential/token surface | Manual | [S7](../06-manual-tests/SPRINT-019-manual.md#s7--customer-authentication-remains-unavailable-task-0091--feat-ac-7) | Pass |
| T19-09 | 8 | Tenant UI checks and full build pass | Auto | `npm --prefix apps/tenant-web run check && make build` | Pass |

---

## FEAT-0022 — Customer Authentication and Domain Enforcement (SPRINT-020)

| ID | AC | Scenario | Type | Test / Command | Result |
| --- | ---: | --- | --- | --- | --- |
| T20-01 | 1 | Customer auth settings tables and hashes-only storage | Manual + Auto | [S11](../06-manual-tests/SPRINT-020-manual.md#s11--storage-safety-hashes-only-no-plaintext-credentials-task-0092--ac-2-5) · `go test ./internal/store` | Pass |
| T20-02 | 2 | OTP request returns safe challenge metadata and sends email | Manual | [S2](../06-manual-tests/SPRINT-020-manual.md#s2--otp-request-returns-safe-challenge-metadata-task-0093--ac-2-5) | Pass |
| T20-03 | 3 | OTP verify issues customer JWT/session and `/me` profile | Manual | [S3](../06-manual-tests/SPRINT-020-manual.md#s3--otp-verify-issues-customer-jwtsession-and-me-profile-task-0093--ac-3-4) | Pass |
| T20-04 | 4 | Existing imported customer claim binds by tenant/email | Manual | [S3](../06-manual-tests/SPRINT-020-manual.md#s3--otp-verify-issues-customer-jwtsession-and-me-profile-task-0093--ac-3-4) | Pass |
| T20-05 | 5 | Domain allow/deny blocks invalid customer login | Manual | [S4](../06-manual-tests/SPRINT-020-manual.md#s4--denied-domain-blocks-otp-before-delivery-task-0093--ac-5) | Pass |
| T20-06 | 6 | Cross-tenant challenge/session isolation | Manual | [S5](../06-manual-tests/SPRINT-020-manual.md#s5--cross-tenant-challengesession-isolation-task-0093--ac-1-6-task-0096--ac-2) | Pass |
| T20-07 | TASK-0094 | Tenant settings UI saves customer-auth config | Manual | [S1](../06-manual-tests/SPRINT-020-manual.md#s1--customer-auth-settings-ui-and-api-survive-reload-task-0094--ac-1-4-5) | Pass |
| T20-08 | TASK-0095 | Customer portal OTP UX preserves no-auth path | Manual | [S8](../06-manual-tests/SPRINT-020-manual.md#s8--customer-portal-otp-ux-preserves-no-auth-path-task-0095--ac-1-2-5) | Pass |
| T20-09 | TASK-0096 | Authenticated chat rate-limit attribution | Manual | [S9](../06-manual-tests/SPRINT-020-manual.md#s9--authenticated-chat-consumes-correct-tenant-rate-limit-keys-task-0095--ac-3-task-0096--ac-4) | Pass |
| T20-10 | TASK-0096 | Authenticated call tenant/quota attribution | Manual | [S10](../06-manual-tests/SPRINT-020-manual.md#s10--authenticated-call-session-uses-customer-tenant-context-task-0095--ac-3-task-0096--ac-4) | Pass |
| T20-11 | — | Refresh/logout revokes customer sessions | Manual | [S6](../06-manual-tests/SPRINT-020-manual.md#s6--refresh-and-logout-revoke-customer-session-task-0093--ac-3-6) | Pass |
| T20-12 | — | Full build/regression gate | Auto | [S12](../06-manual-tests/SPRINT-020-manual.md#s12--automatedbuild-regression-gate) · `make test && make build` | Pass |

---

## FEAT-0023 — Authenticated Workforce Selection and Customer Quota Enforcement (SPRINT-021)

| ID | AC | Scenario | Type | Test / Command | Result |
| --- | ---: | --- | --- | --- | --- |
| T21-01 | 1–2 | Required-auth tenant blocks workforce selection until OTP | Manual | [S2](../06-manual-tests/SPRINT-021-manual.md#s2--required-tenant-blocks-workforce-until-otp-task-0097-task-0098) | Pending |
| T21-02 | 2 | Optional-auth tenant preserves no-auth chat/call | Manual | [S1](../06-manual-tests/SPRINT-021-manual.md#s1--optional-tenant-preserves-no-auth-flow-task-0097-task-0098) | Pending |
| T21-03 | 3–4 | Signed-in customer can select active assigned avatars only | Manual + Auto | [S2](../06-manual-tests/SPRINT-021-manual.md#s2--required-tenant-blocks-workforce-until-otp-task-0097-task-0098) · `go test ./cmd/server` | Pending |
| T21-04 | 5–7 | Customer quota status and exhausted limit behavior | Manual + Auto | [S3](../06-manual-tests/SPRINT-021-manual.md#s3--customer-quota-state-is-visible-and-enforced-task-0099-task-0100) · `go test ./internal/store` | Pending |
| T21-05 | 8 | Tenant settings persist workforce-auth/quota fields | Manual | [S4](../06-manual-tests/SPRINT-021-manual.md#s4--tenant-settings-persist-s21-fields-task-0100) | Pending |

---

## FEAT-0024 — Conversation Records and Knowledge Gap Review (SPRINT-022)

| ID | AC | Scenario | Type | Test / Command | Result |
| --- | ---: | --- | --- | --- | --- |
| T22-01 | 1–3 | Chat/call creates conversation record and archive object metadata | Manual + Auto | [S1](../06-manual-tests/SPRINT-022-manual.md#s1--chat-creates-conversation-record-and-archive-metadata-task-0102-task-0103) · `go test ./internal/store` | Pending |
| T22-02 | 4–5 | Missing-KM turn creates knowledge gap and tenant can update lifecycle | Manual | [S3](../06-manual-tests/SPRINT-022-manual.md#s3--knowledge-gap-candidate-lifecycle-task-0104-task-0105) | Pending |
| T22-03 | 6 | Cross-tenant record/gap access is denied | Manual | [S4](../06-manual-tests/SPRINT-022-manual.md#s4--cross-tenant-isolation-task-0106) | Pending |
| T22-04 | 7 | Tenant records/gaps UI loads and archive retry works | Manual | [S2](../06-manual-tests/SPRINT-022-manual.md#s2--tenant-records-ui-works-task-0105) | Pending |

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

## FEAT-0028 — Tenant System Performance Monitoring (SPRINT-026)

| ID | AC | Scenario | Type | Test / Command |
| --- | ---: | --- | --- | --- |
| T26-01 | 1–3 | Normalized dependency statuses, ordering, timeout handling, and redaction | Auto | `go test ./internal/observability ./cmd/server` |
| T26-02 | 1–4 | Tenant-admin auth, tenant scoping, and safe monitoring response | Auto + Manual | `go test ./cmd/server` · [UAT-026](../06-manual-tests/SPRINT-026-manual.md) |
| T26-03 | 1–6 | Tenant monitoring route, loading/error/retry states, and responsive layout | Auto + Manual | `npm run check && npm run build` · [UAT-026](../06-manual-tests/SPRINT-026-manual.md) |
| T26-04 | 4 | Existing call, archive, quota, statistics, and `/healthz` compatibility | Smoke / Manual | [UAT-026](../06-manual-tests/SPRINT-026-manual.md#uat-026-10) |

---

## Coverage gaps (planned)

| Area | Gap | Target |
| --- | --- | --- |
| ClickHouse | No mocked integration test in CI | Add `internal/clickhouse` unit tests when client stabilizes |
| RAG | End-to-end ingest→search in test | Integration test with testcontainers (Sprint 3+) |
| Voice | Automated WebSocket smoke | `cmd/voicecheck` style harness (deferred) |
