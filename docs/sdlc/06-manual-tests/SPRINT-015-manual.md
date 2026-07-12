# SPRINT-015 Manual UAT — Tenant Scope and KM

**Sprint:** SPRINT-015 · **Release target:** v1.6.0 · **Feature:** FEAT-0015

## 0. Preconditions

```bash
make restart   # or infra up + server
# Active tenant_admin JWT (KYC approved)
```

- [ ] Server healthy on `:8091`
- [ ] Gemini embed configured if testing real index (or expect 502 on upload without key)
- [ ] `go test ./internal/store/ ./internal/km/ ./internal/scope/ ./cmd/server/ -count=1` green

## S1 — Scopes & agents API

```bash
TOKEN=…
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/km/scopes | jq .
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/km/agents | jq .
```

- [ ] Scopes include general, billing, technical
- [ ] Agents list with `doc_count` / `by_scope` / `default_scopes`

## S2 — Upload document

```bash
curl -sS -H "Authorization: Bearer $TOKEN" \
  -F file=@docs/samples/km/ava.md -F scope=general \
  http://localhost:8091/api/tenant/km/agents/ava/documents | jq .
```

- [ ] 201 with `status` indexed (or failed if no Gemini)
- [ ] No `object_key` in JSON
- [ ] List documents shows the file

## S3 — Scope patch + delete

```bash
DOC=… # id from upload
curl -sS -X PATCH -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"km_scope":"billing"}' http://localhost:8091/api/tenant/km/documents/$DOC | jq .
curl -sS -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8091/api/tenant/km/documents/$DOC | jq .
```

- [ ] PATCH updates scope
- [ ] DELETE returns deleted true; list no longer has doc

## S4 — Reset agent

```bash
curl -sS -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8091/api/tenant/km/agents/ava/reset | jq .
```

- [ ] status reset; documents empty

## S5 — Knowledge gaps (`km_gaps`)

1. Ensure agent has **no** matching KM for a weird question (or empty KB).
2. `POST /api/chat` with `X-Tenant-Id` / session as that tenant, message unlikely to match.
3. List gaps:

```bash
curl -sS -H "Authorization: Bearer $TOKEN" \
  'http://localhost:8091/api/tenant/km/gaps?status=open' | jq .
```

- [ ] Gap row appears with question text
- [ ] Repeat same question bumps `occurrence_count`
- [ ] PATCH dismiss works

## S6 — Tenant UI

Open `http://localhost:8091/tenant/km`

- [ ] Nav **Knowledge** visible
- [ ] Agent chips + overview matrix
- [ ] Upload, list, change scope, delete, reset confirm
- [ ] Gaps panel lists open gaps

## S7 — Isolation

- [ ] Second tenant cannot list/delete first tenant’s documents or gaps (404)

## S8 — Regression

- [ ] `/tenant/embed` still works
- [ ] Customer chat still answers when KM present

## Sign-off

| Role | Name | Date | Result |
| --- | --- | --- | --- |
| Tester | | | ☐ Pass ☐ Fail |
| Notes | | | |
