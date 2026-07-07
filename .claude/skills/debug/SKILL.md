---
name: debug
description: Systematic debugging for Monti Jarvis — isolate browser → Go server → store → Gemini layers, test one hypothesis at a time, fix at the right altitude, add regression tests. Use for errors, bugs, crashes, or "not working" reports.
---

# debug — systematic debugging for Monti Jarvis

## The five phases (in order)

1. **Reproduce** — exact steps; capture HTTP status + body, `.run/server.log`, browser Console + Network.
2. **Isolate** — binary-search to the failing layer.
3. **Hypothesize & test** — one change at a time; check Known traps first.
4. **Fix** — fix the shared mechanism; rebuild + restart before re-testing.
5. **Prevent** — add a regression test in the relevant `internal/` package.

## Isolate: which layer?

```
browser (internal/web/public) ── REST/WS ──► Go :8091 ─┬─ /api/workforce, /api/chat
                                                      ├─ /ws/voice ──► Gemini Live
                                                      └─ store (Postgres/Redis/MinIO, optional)
```

## Known traps

| Symptom | Likely cause |
| --- | --- |
| `GEMINI_API_KEY is not configured` | Missing key in `infra/.env.dev` |
| Voice connects but no audio | Mic permission; check browser console |
| Chat 502 | Gemini API key or model name |
| Postgres warnings on start | Shared containers offline — app still runs |
| Port in use | Jarvis Chat on 8090; Monti uses 8091 |
| Stale UI | Rebuild with `make start` after Go/static changes |

## Quick checks

```bash
make status
curl -s http://localhost:8091/healthz
curl -s http://localhost:8091/api/workforce
make test
```