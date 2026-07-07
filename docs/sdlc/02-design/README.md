# 02 — Design

System design artifacts for Monti Jarvis (v0.2.x / Sprint 1–2 scope).

| Doc | Description |
| --- | --- |
| [architecture.md](architecture.md) | Layers, packages, infra topology |
| [workflow.md](workflow.md) | Chat, voice, call, KM ingest sequences |
| [er-diagram.md](er-diagram.md) | Postgres + ClickHouse + MinIO entities |
| [api-spec.md](api-spec.md) | REST, WebSocket, SSE contract |
| [ux-ui.md](ux-ui.md) | ASCII wireframes with API mapping |

**Stack:** Go `net/http` · SvelteKit customer portal · Postgres · Redis 8 · MinIO · ClickHouse · NATS · LiveKit · Gemini

**Verify:** [test matrix](../05-test-scenarios/TEST-MATRIX.md) · [manual tests](../06-manual-tests/) · [deploy](../07-deployment/LOCAL-DEV.md)