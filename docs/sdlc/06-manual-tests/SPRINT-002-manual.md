---
id: MANUAL-SPRINT-002
sprint: SPRINT-002
release_target: v0.3.0
status: signed_off
updated: 2026-07-07
---

# SPRINT-002 — Manual Test Checklist

KM ingest, scoped RAG in chat/voice, citation chips, missing-KM handling.

**Ops reference:** [`docs/KM_SETUP.md`](../../KM_SETUP.md)

## 0. Preconditions

- [ ] SPRINT-001 manual tests pass (or equivalent smoke)
- [ ] `infra/.env.dev` includes `CLICKHOUSE_URL=http://localhost:8123`, `CLICKHOUSE_DB=monti_jarvis`, `DEMO_TENANT_ID=demo`
- [ ] `GEMINI_API_KEY` set (required for embed + chat)
- [ ] Containers: `poc-gml-postgres`, `poc-gml-redis`, `poc-gml-minio`, `poc-gml-clickhouse`

## 1. Init infrastructure

```bash
docker start poc-gml-postgres poc-gml-redis poc-gml-minio poc-gml-clickhouse 2>/dev/null || true
make infra-up
make infra-init
make infra-check
```

- [ ] ClickHouse ping succeeds
- [ ] `curl -fsS http://localhost:8091/api/infra` shows `clickhouse: ok` (after server start)

## 2. Prepare data

```bash
make build
make start
make km-seed
```

Verify seed:

```bash
curl -fsS http://localhost:8091/api/km/agents/max | python3 -m json.tool
```

- [ ] Each agent (`ava`, `max`, `luna`, `neo`) shows `document_count >= 1`, `chunk_count >= 1`

## 3. Scenarios

### S1 — KM ingest (FEAT-0002 · AC 1)

1. Reset Max KB: `curl -X POST http://localhost:8091/api/km/agents/max/reset`
2. Upload sample:  
   `curl -X POST http://localhost:8091/api/km/agents/max/documents -F "file=@docs/samples/km/max.md" -F "scope=billing"`
3. Re-check status: `curl http://localhost:8091/api/km/agents/max`

**Expected:** `document_count` and `chunk_count` increase; ClickHouse has rows for tenant `demo`.

```bash
# Optional — if clickhouse-client available on poc-gml-clickhouse
docker exec poc-gml-clickhouse clickhouse-client -q \
  "SELECT count() FROM monti_jarvis.km_embeddings WHERE tenant_id = 'demo'"
```

- [ ] Pass

### S2 — Scoped RAG chat (FEAT-0002 · AC 3)

**API check:**

```bash
curl -X POST http://localhost:8091/api/chat \
  -H 'content-type: application/json' \
  -d '{"agent_id":"max","topic":"billing","message":"When are invoices due?","history":[]}'
```

- [ ] Response includes `sources[]` with billing excerpts

**Portal check:**

1. Open http://localhost:8091
2. Select **Max**, **Billing** tab
3. Ask: *"When are invoices due?"*

**Expected:** Grounded answer; citation chips under assistant bubble.

- [ ] Pass

### S3 — Scope enforcement (FEAT-0002 · AC 2)

1. Select **Ava**, switch to **Technical** tab
2. Ask a question that exists only in Luna's technical KB (e.g. microphone troubleshooting from `luna.md`)

**Expected:** Ava does not cite Luna-only technical chunks; answer stays within `general` scope or declines.

3. Select **Luna**, **Technical** tab; ask the same question

**Expected:** Grounded technical answer with sources.

- [ ] Pass

### S4 — Missing-KM fallback (FEAT-0002 · AC 4)

1. Reset Ava: `curl -X POST http://localhost:8091/api/km/agents/ava/reset`
2. On portal, select Ava, General tab
3. Ask: *"What is the secret project codename Zeta-9?"* (not in any KB)

**Expected:** Safe fallback (no fabricated policy); no citation chips.

4. Verify `qa_events` (optional):

```bash
docker exec poc-gml-clickhouse clickhouse-client -q \
  "SELECT event_type, count() FROM monti_jarvis.qa_events GROUP BY event_type"
```

- [ ] `missing_km` or equivalent event present after out-of-KB question

- [ ] Pass

### S5 — Voice RAG (FEAT-0002 · AC 5)

1. Re-seed if needed: `make km-seed`
2. Select **Max**, Billing tab
3. Start voice call; ask: *"When are invoices due?"*

**Expected:** Spoken answer reflects seeded billing KB (same scope path as text).

- [ ] Pass

### S6 — Turn metadata (FEAT-0002 · AC 6)

1. After S2, inspect Postgres turn row (if session active):

```bash
docker exec poc-gml-postgres psql -U postgres -d monti_jarvis -c \
  "SELECT id, source_chunk_ids FROM callcenter.call_turns ORDER BY created_at DESC LIMIT 3;"
```

**Expected:** Recent assistant turn has non-null `source_chunk_ids` JSON when KM matched.

- [ ] Pass (or N/A if persistence disabled)

### S7 — KM admin APIs (FEAT-0002 · ops)

```bash
curl http://localhost:8091/api/km/agents/ava/documents
curl -X POST http://localhost:8091/api/km/agents/neo/reset
curl -X POST http://localhost:8091/api/km/seed
```

**Expected:** List returns documents; reset clears counts; seed restores all four avatars.

- [ ] Pass

### S8 — Degraded mode (FEAT-0002 · regression)

1. Stop ClickHouse: `docker stop poc-gml-clickhouse`
2. Restart server or wait for next chat turn
3. Send a chat message on General tab

**Expected:** Chat/voice still works without crash; no RAG sources (logs may warn). `GET /api/infra` shows clickhouse not ok.

4. Restart ClickHouse and re-seed before sign-off

- [ ] Pass

## 4. Automated gate

```bash
go test ./...
make customer-web
```

- [ ] All tests green

## 5. Teardown

```bash
make down
```

- [ ] Server stopped

## 6. Sign-off

| Tester | Date | Result | Defects |
| --- | --- | --- | --- |
| Dev | 2026-07-07 | Pass (v0.3.0) | — |