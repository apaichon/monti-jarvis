---
name: debug
description: Systematic debugging for Monti Jarvis — isolate browser → Go server → store → ClickHouse → Gemini layers, test one hypothesis at a time, fix at the right altitude, prevent with regression tests. Use for errors, bugs, crashes, or "not working" reports. Owner DEV.
---

# debug — systematic debugging for Monti Jarvis

## The five phases (in order)

1. **Reproduce** — exact steps; capture HTTP status + body, `.run/server.log`, browser Console + Network.
2. **Isolate** — binary-search to the failing layer.
3. **Hypothesize & test** — one change at a time; check Known traps first.
4. **Fix** — fix the shared mechanism; `make restart` after Go/UI changes.
5. **Prevent** — add regression test in the relevant `internal/` package.

## Isolate: which layer?

```
browser (apps/customer-web) ── REST/WS ──► Go :8091 ─┬─ /api/chat, /api/km/*
                                                     ├─ /ws/voice ──► Gemini Live
                                                     ├─ /api/calls ──► Postgres/Redis/NATS
                                                     └─ ClickHouse km_embeddings
```

```bash
make status
curl -s http://localhost:8091/healthz | jq .
curl -s http://localhost:8091/api/infra | jq .
tail -n 50 .run/server.log
go test ./... -run TestName -v
```

## Known traps

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| `GEMINI_API_KEY is not configured` | Missing key in `infra/.env.dev` | Set key, `make restart` |
| Voice connects but no audio | Mic permission; stale build | Browser settings; `make restart` |
| Chat 502 | Gemini API key or model | Check `.run/server.log` |
| No KB citations | ClickHouse down or not seeded | `docker start poc-gml-clickhouse`; `make km-seed` |
| `clickhouse: disabled` | Container stopped | `docker start poc-gml-clickhouse` |
| Postgres warnings on start | Shared containers offline | `docker start poc-gml-postgres` |
| Port in use | Wrong service on 8091 | `make stop`; Jarvis Chat uses 8090 |
| Stale UI | Portal not rebuilt | `make customer-web && make restart` |
| CORS from Vite dev | Missing header | Dev proxy or use built portal at :8091 |

## Output template
```
Root cause: <one sentence>
Evidence:   <curl / log / console>
Fix:        <file; why>
Verify:     <command that now passes>
Prevent:    <regression test>
```

See `docs/KM_SETUP.md` and `Makefile`.