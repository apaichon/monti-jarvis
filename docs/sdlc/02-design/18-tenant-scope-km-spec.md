---
id: DES-0018
title: Tenant Set Scope and KM Specification
status: shipped
updated: 2026-07-12
sprint: SPRINT-015
owner: SA
release: v1.6.0
---

# Tenant Scope & KM — Design Spec

**Sprint:** SPRINT-015 · **Release:** v1.6.0  
**Feature:** [FEAT-0015](../01-features/FEAT-0015-tenant-scope-km.md)  
**Depends on:** S2 KM/RAG, S5 avatars, S6–7 active tenant, S13 KM quota  
**Tasks:** TASK-0067 … TASK-0071

## 1. Goals

1. Active **tenant_admin** self-serves knowledge documents for their agents in the tenant portal.
2. Each document is tagged with a **scope** (`general` | `billing` | `technical`) used by S2 RAG filters.
3. Safe **delete** (one document) and **reset** (all docs for an agent) with multi-store cascade.
4. **Tenant isolation** — JWT `tenant_id` only; no body/query tenant override.
5. Reuse S2 `internal/km` pipeline and S13 quota hooks; dual-surface with legacy `/api/km/*`.

## 2. Non-goals

| Out | Sprint |
| --- | --- |
| PDF OCR / website crawl | Future |
| Multi-reviewer publish / `km_versions` UI | Future |
| Editable global `AgentScopes` hard-map | Keep read-only help |
| Customer portal KM admin | N/A |
| Locale & call-minute limits UI | S16 |
| Test & preview sandbox | S17 |
| Knowledge-gap analytics | S22 |

## 3. Env / infra

| Item | Value / notes |
| --- | --- |
| Postgres schema | `callcenter` — `knowledge_documents`, `knowledge_chunks` (S2) |
| MinIO | Bucket `monti-jarvis`; key `km/{tenant}/{agent}/{doc}/original/{file}` |
| ClickHouse | DB `monti_jarvis`, table `km_embeddings` |
| Redis | S13 keys `monti_jarvis:rl:…:km:…`, document count entitlements |
| New env vars | **None** |

## 4. Data model

### 4.1 Existing (S2)

`knowledge_documents`, `knowledge_chunks`, ClickHouse `km_embeddings`. See [03-er-diagram.md](03-er-diagram.md) § Sprint 15.

### 4.2 New: `callcenter.km_gaps`

Records **customer questions with no matching KM** (`missing_km`) so tenants can turn them into FAQs.

| Column | Notes |
| --- | --- |
| `question` + `question_hash` | Hash = sha256(lower(trim(q))); unique per `(tenant_id, agent_id, hash)` |
| `occurrence_count` / `last_seen_at` | Deduped repeats |
| `status` | `open` → `resolved` / `dismissed` / `converted` |
| `source` | `chat` \| `voice` \| `embed` |
| `resolved_document_id` | Optional link after upload |

**Writers:** chat handler when `rag.MissingKM` (ClickHouse `qa_events` retained for analytics).  
**Readers:** `GET /api/tenant/km/gaps` for tenant UI backlog.

### Status machine (`internal/km` constants)

```text
uploaded → indexing → indexed
                  ↘ failed
```

| Code status | UI label |
| --- | --- |
| `uploaded`, `indexing` | Processing |
| `indexed` | Ready |
| `failed` | Failed |

### Scope vs retrieval

- Document write: free choice of `general|billing|technical`.
- Retrieval still uses `scope.Resolve(agent_id, topic)` — agent only sees scopes in its hard-map.
- UI shows `default_scopes` per agent so tenants pick scopes that will be retrieved.

## 5. API summary

Full contract: [04-api-spec.md](04-api-spec.md) § **Tenant KM / Scope (Sprint 15)**.

| Method | Path |
| --- | --- |
| GET | `/api/tenant/km/scopes` |
| GET | `/api/tenant/km/agents` |
| GET | `/api/tenant/km/agents/{agent_id}/documents` |
| POST | `/api/tenant/km/agents/{agent_id}/documents` |
| PATCH | `/api/tenant/km/documents/{id}` |
| DELETE | `/api/tenant/km/documents/{id}` |
| POST | `/api/tenant/km/agents/{agent_id}/reset` |
| GET | `/api/tenant/km/gaps` |
| PATCH | `/api/tenant/km/gaps/{id}` |

**Auth:** `RequireTenantAdminActive` on all of the above.

### Service methods to add/extend (`internal/km`)

| Method | Notes |
| --- | --- |
| `Ingest` | Existing |
| `ListAgentDocuments` / `AgentKnowledge` | Existing |
| `ResetAgent` | Existing (ensure tenant-scoped) |
| **`DeleteDocument(tenantID, docID)`** | **New** (TASK-0067) |
| **`UpdateDocumentScope(tenantID, docID, scope)`** | **New** — PG + CH `km_scope` |

### Store / ClickHouse

| Layer | Need |
| --- | --- |
| `DocumentStore` | `DeleteKnowledgeDocument` (or get + delete agent objects single-doc) |
| `clickhouse.Client` | Existing delete-by-document helper (verify/wrap) |

## 6. RBAC

| Action | platform_admin | tenant_admin | public |
| --- | --- | --- | --- |
| `/api/tenant/km/*` | no | yes (own) | no |
| `/api/km/*` legacy | seed / ops | optional bearer patterns | no product UI |
| Chat RAG consume | — | — | yes (tenant from call path) |

## 7. UX

Tenant route **`/tenant/km`** (T8). Wireframes + zone map: [05-ux-ui.md](05-ux-ui.md) § Sprint 15.

| File | Role |
| --- | --- |
| `apps/tenant-web/src/routes/km/+page.svelte` | Page |
| `apps/tenant-web/src/lib/api/km.ts` | Client |
| `+layout.svelte` | Nav **Knowledge** |

## 8. Workflows

| § | Flow |
| --- | --- |
| 40 | Upload |
| 41 | Delete |
| 42 | List agents/docs |
| 43 | Patch scope |
| 44 | Reset agent |

→ [02-workflow.md](02-workflow.md)

## 9. Security & multi-tenant

1. Every mutation: `doc.TenantID == claims.TenantID` else **404** (no existence leak across tenants).
2. Never return MinIO `object_key` to the browser.
3. Multipart size cap (~8MB) unchanged.
4. Quota failures use existing S13 error JSON.
5. Confirm UI on reset/delete (destructive).

## 10. Verification

```bash
# Login tenant_admin → TOKEN
curl -sS -H "Authorization: Bearer $TOKEN" \
  http://localhost:8091/api/tenant/km/scopes | jq .

curl -sS -H "Authorization: Bearer $TOKEN" \
  -F file=@docs/samples/km/ava.md -F scope=general \
  http://localhost:8091/api/tenant/km/agents/ava/documents | jq .

DOC=… # id from upload
curl -sS -X PATCH -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"km_scope":"billing"}' \
  http://localhost:8091/api/tenant/km/documents/$DOC | jq .

curl -sS -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8091/api/tenant/km/documents/$DOC | jq .

# Other tenant TOKEN2 → same DELETE → 404
go test ./internal/km/ ./internal/store/ ./cmd/server/ -count=1 -run 'KM|Knowledge|TenantKM'
```

## 11. Implementation order

1. **TASK-0067** — `DeleteDocument` cascade + register routes  
2. **TASK-0068** — full handlers + tests + quota  
3. **TASK-0069** — `/tenant/km` UI  
4. **TASK-0070** — matrix + scope polish  
5. **TASK-0071** — manual UAT doc  

## 12. Related

| Artifact | Path |
| --- | --- |
| Feature | [FEAT-0015](../01-features/FEAT-0015-tenant-scope-km.md) |
| Sprint | [SPRINT-015](../03-sprints/SPRINT-015.md) |
| Workflow | [02-workflow.md](02-workflow.md) §40–44 |
| ER | [03-er-diagram.md](03-er-diagram.md) § Sprint 15 |
| API | [04-api-spec.md](04-api-spec.md) § Tenant KM |
| UX | [05-ux-ui.md](05-ux-ui.md) § T8 |
| Prior KM | [FEAT-0002](../01-features/FEAT-0002-km-scope-rag.md), [KM_SETUP.md](../../KM_SETUP.md) |
| Quota | [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md) |
